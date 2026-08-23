package merchant

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/profitsharing"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================
// 分账接收方管理
// ============================================

// ProfitSharingReceiverRequest 接收方新增/编辑请求
type ProfitSharingReceiverRequest struct {
	ReceiverType uint8   `json:"receiver_type" binding:"required,gte=1,lte=2"` // 1=商户号 2=个人微信
	Name         string  `json:"name" binding:"required"`
	Account      string  `json:"account" binding:"required"`
	PersonalName string  `json:"personal_name"`
	RelationType string  `json:"relation_type"`
	DefaultRatio float64 `json:"default_ratio"`
	Status       uint8   `json:"status"`
	Sort         uint    `json:"sort"`
	Remark       string  `json:"remark"`
}

// GetProfitSharingReceivers 分账接收方列表
func GetProfitSharingReceivers(c *gin.Context) {
	var list []models.ProfitSharingReceiver
	if err := database.DB.Where("merchant_id = ?", utils.DefaultMerchantID).
		Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		response.ServerError(c, "查询分账接收方失败")
		return
	}
	response.Success(c, list)
}

// CreateProfitSharingReceiver 新增分账接收方（并建立微信关系）
func CreateProfitSharingReceiver(c *gin.Context) {
	var req ProfitSharingReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	req.Account = strings.TrimSpace(req.Account)
	req.Name = strings.TrimSpace(req.Name)
	if req.Account == "" || req.Name == "" {
		response.ParamError(c, "接收方名称与账号不能为空")
		return
	}
	if req.RelationType == "" {
		req.RelationType = "SERVICE_PROVIDER"
	}
	if req.Status == 0 {
		req.Status = 1
	}

	receiver := models.ProfitSharingReceiver{
		MerchantID:   utils.DefaultMerchantID,
		ReceiverType: req.ReceiverType,
		Name:         req.Name,
		Account:      req.Account,
		PersonalName: strings.TrimSpace(req.PersonalName),
		RelationType: req.RelationType,
		DefaultRatio: req.DefaultRatio,
		Status:       req.Status,
		Sort:         req.Sort,
		Remark:       req.Remark,
	}

	// 建立微信接收方关系（失败不阻断创建，记录错误状态供后台查看）
	if err := bindWechatReceiver(&receiver); err != nil {
		log.Printf("[PROFIT-SHARING] 接收方 %s 建立微信关系失败: %v", receiver.Name, err)
	}

	if err := database.DB.Create(&receiver).Error; err != nil {
		response.ServerError(c, "保存分账接收方失败")
		return
	}
	response.Success(c, receiver)
}

// UpdateProfitSharingReceiver 编辑分账接收方（名称/比例/启停/备注等）
func UpdateProfitSharingReceiver(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.ParamError(c, "参数错误")
		return
	}
	var receiver models.ProfitSharingReceiver
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).First(&receiver).Error; err != nil {
		response.NotFound(c, "分账接收方不存在")
		return
	}

	var req ProfitSharingReceiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	updates := map[string]interface{}{
		"name":          strings.TrimSpace(req.Name),
		"personal_name": strings.TrimSpace(req.PersonalName),
		"default_ratio": req.DefaultRatio,
		"sort":          req.Sort,
		"remark":        req.Remark,
	}
	if req.ReceiverType >= 1 && req.ReceiverType <= 2 {
		updates["receiver_type"] = req.ReceiverType
	}
	if strings.TrimSpace(req.Account) != "" {
		updates["account"] = strings.TrimSpace(req.Account)
	}
	if req.Status == 0 || req.Status == 1 {
		updates["status"] = req.Status
	}
	// 账号或类型发生变化时，需要重新绑定微信关系
	if strings.TrimSpace(req.Account) != "" && req.Account != receiver.Account {
		updates["wechat_bound"] = 0
		updates["wechat_error"] = ""
	}

	if err := database.DB.Model(&receiver).Updates(updates).Error; err != nil {
		response.ServerError(c, "更新分账接收方失败")
		return
	}
	database.DB.First(&receiver, id)
	response.Success(c, receiver)
}

// DeleteProfitSharingReceiver 删除分账接收方（并尝试解绑微信关系）
func DeleteProfitSharingReceiver(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.ParamError(c, "参数错误")
		return
	}
	var receiver models.ProfitSharingReceiver
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).First(&receiver).Error; err != nil {
		response.NotFound(c, "分账接收方不存在")
		return
	}

	if err := unbindWechatReceiver(&receiver); err != nil {
		log.Printf("[PROFIT-SHARING] 接收方 %s 解绑微信关系失败: %v", receiver.Name, err)
	}

	if err := database.DB.Delete(&receiver).Error; err != nil {
		response.ServerError(c, "删除分账接收方失败")
		return
	}
	response.SuccessWithMessage(c, "分账接收方已删除", nil)
}

// SyncProfitSharingReceiver 重试建立分账接收方微信关系
func SyncProfitSharingReceiver(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.ParamError(c, "参数错误")
		return
	}
	var receiver models.ProfitSharingReceiver
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).First(&receiver).Error; err != nil {
		response.NotFound(c, "分账接收方不存在")
		return
	}
	if err := bindWechatReceiver(&receiver); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeProviderConfigError, "建立微信关系失败: "+err.Error())
		return
	}
	response.Success(c, receiver)
}

// bindWechatReceiver 建立/重试建立接收方微信关系，并回写绑定状态
func bindWechatReceiver(receiver *models.ProfitSharingReceiver) error {
	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		return err
	}
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return err
	}
	var merchant models.Merchant
	if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
		return err
	}
	if strings.TrimSpace(merchant.SubMchID) == "" {
		return errSubMchIDEmpty()
	}
	if err := client.AddProfitSharingReceiver(context.Background(), wechatpay.AddProfitSharingReceiverRequest{
		SubMchID:     merchant.SubMchID,
		AppID:        appIdentity.AppID,
		Type:         profitsharing.ReceiverWechatType(receiver.ReceiverType),
		Account:      receiver.Account,
		Name:         receiver.PersonalName,
		RelationType: receiver.RelationType,
	}); err != nil {
		database.DB.Model(receiver).Updates(map[string]interface{}{
			"wechat_bound": 0,
			"wechat_error": err.Error(),
		})
		return err
	}
	database.DB.Model(receiver).Updates(map[string]interface{}{
		"wechat_bound": 1,
		"wechat_error": "",
	})
	return nil
}

// unbindWechatReceiver 解绑接收方微信关系（尽力而为）
func unbindWechatReceiver(receiver *models.ProfitSharingReceiver) error {
	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		return err
	}
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return err
	}
	var merchant models.Merchant
	if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
		return err
	}
	if strings.TrimSpace(merchant.SubMchID) == "" {
		return errSubMchIDEmpty()
	}
	return client.DeleteProfitSharingReceiver(context.Background(), wechatpay.AddProfitSharingReceiverRequest{
		SubMchID: merchant.SubMchID,
		AppID:    appIdentity.AppID,
		Type:     profitsharing.ReceiverWechatType(receiver.ReceiverType),
		Account:  receiver.Account,
	})
}

func errSubMchIDEmpty() error {
	return errors.New("商家未配置微信子商户号")
}

// ============================================
// 分账配置
// ============================================

// ProfitSharingConfig 分账配置
type ProfitSharingConfig struct {
	ProfitSharingEnabled bool   `json:"profit_sharing_enabled"`
	SubMchID             string `json:"sub_mch_id"`
	MaxRatioPct          string `json:"max_ratio_pct"` // 微信允许最大分账比例(%)
	MaxRatioRaw          string `json:"max_ratio_raw"`
}

// UpdateProfitSharingConfigRequest 更新自动分账开关
type UpdateProfitSharingConfigRequest struct {
	ProfitSharingEnabled bool `json:"profit_sharing_enabled"`
}

// GetProfitSharingConfig 查询分账配置（开关 + 微信允许最大比例）
func GetProfitSharingConfig(c *gin.Context) {
	var merchant models.Merchant
	if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
		response.NotFound(c, "商家不存在")
		return
	}
	cfg := ProfitSharingConfig{
		ProfitSharingEnabled: merchant.ProfitSharingEnabled,
		SubMchID:             merchant.SubMchID,
		MaxRatioPct:          "30%",
	}
	if client, err := wechatpay.NewServiceProviderClient(); err == nil {
		if merchant.SubMchID != "" {
			if maxRatioBp, qErr := client.QueryProfitSharingMerchantRatio(context.Background(), merchant.SubMchID); qErr == nil {
				cfg.MaxRatioRaw = strconv.FormatInt(maxRatioBp, 10)
				cfg.MaxRatioPct = floatToPercentString(float64(maxRatioBp) / 100)
			}
		}
	} else {
		log.Printf("[PROFIT-SHARING] 初始化支付客户端失败: %v", err)
	}
	response.Success(c, cfg)
}

// UpdateProfitSharingConfig 更新自动分账开关
func UpdateProfitSharingConfig(c *gin.Context) {
	var req UpdateProfitSharingConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", utils.DefaultMerchantID).
		Update("profit_sharing_enabled", req.ProfitSharingEnabled).Error; err != nil {
		response.ServerError(c, "更新分账配置失败")
		return
	}
	var merchant models.Merchant
	database.DB.First(&merchant, utils.DefaultMerchantID)
	response.SuccessWithMessage(c, "分账配置已更新", merchant.ProfitSharingEnabled)
}

// ============================================
// 分账记录
// ============================================

// GetProfitSharingRecords 分账记录列表
func GetProfitSharingRecords(c *gin.Context) {
	pageInfo := utils.GetPagination(c)
	query := database.DB.Model(&models.ProfitSharingRecord{}).
		Where("merchant_id = ?", utils.DefaultMerchantID)

	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}
	if orderNo := strings.TrimSpace(c.Query("order_no")); orderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+orderNo+"%")
	}
	if start := c.Query("start_date"); start != "" {
		query = query.Where("created_at >= ?", start)
	}
	if end := c.Query("end_date"); end != "" {
		query = query.Where("created_at <= ?", end+" 23:59:59")
	}

	var total int64
	query.Count(&total)

	var list []models.ProfitSharingRecord
	if err := query.Order("id DESC").Offset(pageInfo.GetOffset()).Limit(pageInfo.PageSize).
		Preload("Order").Find(&list).Error; err != nil {
		response.ServerError(c, "查询分账记录失败")
		return
	}
	pageInfo.Total = total
	response.Success(c, gin.H{
		"list":       list,
		"pagination": pageInfo,
	})
}

// GetProfitSharingRecordDetail 分账记录明细（含各方）
func GetProfitSharingRecordDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.ParamError(c, "参数错误")
		return
	}
	var record models.ProfitSharingRecord
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Preload("Order").Preload("Receivers").First(&record).Error; err != nil {
		response.NotFound(c, "分账记录不存在")
		return
	}
	response.Success(c, record)
}

// RetryProfitSharingRecord 重试失败的分账记录
func RetryProfitSharingRecord(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.ParamError(c, "参数错误")
		return
	}
	var record models.ProfitSharingRecord
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Preload("Receivers").First(&record).Error; err != nil {
		response.NotFound(c, "分账记录不存在")
		return
	}
	if record.Status == profitsharing.RecordStatusSuccess || record.Status == profitsharing.RecordStatusProcessing {
		response.ParamError(c, "该分账记录无需重试")
		return
	}
	if len(record.Receivers) == 0 {
		response.ParamError(c, "该分账记录无接收方，无法重试")
		return
	}
	if record.TransactionID == "" || record.AppID == "" || record.SubMchID == "" {
		response.ParamError(c, "该分账记录缺少支付信息，无法重试")
		return
	}

	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeProviderConfigError, "支付客户端初始化失败")
		return
	}

	items := make([]wechatpay.ProfitSharingReceiverItem, 0, len(record.Receivers))
	for _, rc := range record.Receivers {
		amountCents := int64(0)
		if a := centsOfYuan(rc.Amount); a > 0 {
			amountCents = a
		}
		if amountCents <= 0 {
			continue
		}
		items = append(items, wechatpay.ProfitSharingReceiverItem{
			Type:        profitsharing.ReceiverWechatType(rc.ReceiverType),
			Account:     rc.Account,
			Amount:      amountCents,
			Description: "服务商分账",
		})
	}
	if len(items) == 0 {
		response.ParamError(c, "分账金额均为 0，无法重试")
		return
	}

	database.DB.Model(&models.ProfitSharingRecord{}).Where("id = ?", record.ID).
		Updates(map[string]interface{}{"status": profitsharing.RecordStatusProcessing, "error_message": ""})
	database.DB.Model(&models.Order{}).Where("id = ?", record.OrderID).
		Update("profit_sharing_status", profitsharing.OrderProcessing)

	if _, err := client.CreateProfitSharingOrder(context.Background(), wechatpay.ProfitSharingRequest{
		SubMchID:      record.SubMchID,
		AppID:         record.AppID,
		TransactionID: record.TransactionID,
		OutOrderNo:    record.OutOrderNo,
		Receivers:     items,
	}); err != nil {
		msg := err.Error()
		if len(msg) > 512 {
			msg = msg[:512]
		}
		database.DB.Model(&models.ProfitSharingRecord{}).Where("id = ?", record.ID).
			Updates(map[string]interface{}{"status": profitsharing.RecordStatusFailed, "error_message": msg})
		database.DB.Model(&models.Order{}).Where("id = ?", record.OrderID).
			Updates(map[string]interface{}{"profit_sharing_status": profitsharing.OrderFailed, "profit_sharing_error": msg})
		response.Fail(c, http.StatusBadRequest, response.CodeProviderConfigError, "分账重试失败: "+err.Error())
		return
	}

	// 查询回写结果
	if result, qErr := client.QueryProfitSharingOrder(context.Background(), record.SubMchID, record.TransactionID, record.OutOrderNo); qErr == nil {
		_ = applyQueryResultForRetry(record.ID, record.OrderID, result)
	}

	var updated models.ProfitSharingRecord
	database.DB.Preload("Receivers").First(&updated, record.ID)
	response.SuccessWithMessage(c, "分账已重试", updated)
}

func applyQueryResultForRetry(recordID, orderID uint64, result *wechatpay.ProfitSharingResult) error {
	db := database.DB
	var status = profitsharing.RecordStatusProcessing
	var shareTime *time.Time
	hasFailed := false
	allSuccess := len(result.Receivers) > 0
	for _, rc := range result.Receivers {
		if rc.Result == profitsharing.ResultStatusClosed {
			hasFailed = true
		}
		if rc.Result != profitsharing.ResultStatusSuccess {
			allSuccess = false
		}
		if rc.FinishTime != nil && shareTime == nil {
			t := *rc.FinishTime
			shareTime = &t
		}
	}
	switch {
	case hasFailed:
		status = profitsharing.RecordStatusFailed
	case allSuccess:
		status = profitsharing.RecordStatusSuccess
	}

	for _, rc := range result.Receivers {
		var finishTime *time.Time
		if rc.FinishTime != nil {
			t := *rc.FinishTime
			finishTime = &t
		}
		db.Model(&models.ProfitSharingRecordReceiver{}).
			Where("record_id = ? AND account = ?", recordID, rc.Account).
			Updates(map[string]interface{}{
				"result_status": rc.Result,
				"detail_id":     rc.DetailID,
				"fail_reason":   rc.FailReason,
				"finish_time":   finishTime,
			})
	}
	recUpdates := map[string]interface{}{"status": status, "error_message": ""}
	orderUpdates := map[string]interface{}{"profit_sharing_status": status, "profit_sharing_error": ""}
	if shareTime != nil {
		recUpdates["share_time"] = shareTime
		orderUpdates["profit_sharing_at"] = shareTime
	}
	db.Model(&models.ProfitSharingRecord{}).Where("id = ?", recordID).Updates(recUpdates)
	db.Model(&models.Order{}).Where("id = ?", orderID).Updates(orderUpdates)
	return nil
}

// ============================================
// 工具
// ============================================

func floatToPercentString(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64) + "%"
}

func centsOfYuan(yuan float64) int64 {
	return int64((yuan + 0.000001) * 100)
}