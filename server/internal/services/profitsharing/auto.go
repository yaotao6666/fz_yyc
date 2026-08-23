package profitsharing

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"

	"gorm.io/gorm"
)

// 分账记录状态
const (
	RecordStatusPending    uint8 = 0 // 待分账
	RecordStatusProcessing uint8 = 1 // 分账中
	RecordStatusSuccess    uint8 = 2 // 分账成功
	RecordStatusFailed     uint8 = 3 // 分账失败
	RecordStatusSkipped    uint8 = 4 // 已跳过
)

// 订单分账状态（与 orders.profit_sharing_status 对应）
const (
	OrderNoShare    uint8 = 0 // 未分账
	OrderProcessing uint8 = 1 // 分账中
	OrderSuccess    uint8 = 2 // 分账成功
	OrderFailed     uint8 = 3 // 分账失败
	OrderSkipped    uint8 = 4 // 已跳过
)

// 微信分账明细结果取值（缓存于明细表 result_status）
const (
	ResultStatusPending = "PENDING"
	ResultStatusSuccess = "SUCCESS"
	ResultStatusClosed  = "CLOSED"
)

// ReceiverWechatType 将内部接收方类型映射为微信枚举
func ReceiverWechatType(receiverType uint8) string {
	if receiverType == 2 {
		return "PERSONAL_OPENID"
	}
	return "MERCHANT_ID"
}

// AutoProfitShareForPaidOrder 支付成功后自动分账入口。
// 因为在支付回调中调用，内部在独立 goroutine 执行，避免阻塞回调响应。
func AutoProfitShareForPaidOrder(orderNo, transactionID string) {
	ctx := context.Background()
	if err := autoProfitShare(ctx, orderNo, transactionID); err != nil {
		log.Printf("[PROFIT-SHARING] 订单 %s 自动分账失败: %v", orderNo, err)
	}
}

func autoProfitShare(ctx context.Context, orderNo, transactionID string) error {
	db := database.DB

	var order models.Order
	if err := db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}
	// 幂等：已有分账状态则直接返回
	if order.ProfitSharingStatus != OrderNoShare {
		return nil
	}
	// 先占位为「分账中」，防止并发重复触发
	db.Model(&models.Order{}).Where("id = ? AND profit_sharing_status = 0", order.ID).
		Update("profit_sharing_status", OrderProcessing)

	var merchant models.Merchant
	if err := db.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
		return fmt.Errorf("查询商家失败: %w", err)
	}
	subMchID := strings.TrimSpace(merchant.SubMchID)
	if !merchant.ProfitSharingEnabled {
		return markOrderSkipped(order.ID, "商家未开启自动分账")
	}
	if subMchID == "" {
		return markOrderSkipped(order.ID, "商家未配置微信子商户号")
	}

	// 启用中的接收方
	var receivers []models.ProfitSharingReceiver
	if err := db.Where("status = 1").Order("sort ASC, id ASC").Find(&receivers).Error; err != nil {
		return fmt.Errorf("查询分账接收方失败: %w", err)
	}
	if len(receivers) == 0 {
		return markOrderSkipped(order.ID, "未配置可用的分账接收方")
	}

	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		return err
	}
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return err
	}

	// 微信允许的分账最大比例(单位为万分比)，查询失败时兜底 30%
	maxRatioBp := int64(3000)
	if q, qErr := client.QueryProfitSharingMerchantRatio(ctx, subMchID); qErr == nil && q > 0 {
		maxRatioBp = q
	}

	// 计算各方分账金额(单位:分)
	payCents := amountToCents(order.PayAmount)
	totalRatioPct := 0.0
	for _, r := range receivers {
		totalRatioPct += r.DefaultRatio
	}
	if int64(totalRatioPct*100) > maxRatioBp {
		return markOrderSkipped(order.ID,
			fmt.Sprintf("分账总比例 %.2f%% 超过微信允许最大比例 %.2f%%", totalRatioPct, float64(maxRatioBp)/100))
	}
	if totalRatioPct <= 0 {
		return markOrderSkipped(order.ID, "分账总比例为 0，已跳过")
	}

	items := make([]wechatpay.ProfitSharingReceiverItem, 0, len(receivers))
	details := make([]models.ProfitSharingRecordReceiver, 0, len(receivers))
	var totalShareCents int64
	for _, r := range receivers {
		if r.DefaultRatio <= 0 {
			continue
		}
		amountCents := int64(math.Floor(float64(payCents) * r.DefaultRatio / 100))
		if amountCents <= 0 {
			continue
		}
		items = append(items, wechatpay.ProfitSharingReceiverItem{
			Type:        ReceiverWechatType(r.ReceiverType),
			Account:     r.Account,
			Name:        r.PersonalName,
			Amount:      amountCents,
			Description: "服务商分账",
		})
		details = append(details, models.ProfitSharingRecordReceiver{
			ReceiverID:   &r.ID,
			ReceiverType: r.ReceiverType,
			ReceiverName: r.Name,
			Account:      r.Account,
			Amount:       centsToYuan(amountCents),
			ResultStatus: ResultStatusPending,
		})
		totalShareCents += amountCents
	}
	if len(items) == 0 {
		return markOrderSkipped(order.ID, "各接收方分账金额均为 0，已跳过")
	}

	outOrderNo := fmt.Sprintf("ps_%s_%d", order.OrderNo, time.Now().Unix())
	record := models.ProfitSharingRecord{
		MerchantID:       utils.DefaultMerchantID,
		OrderID:          order.ID,
		OrderNo:          order.OrderNo,
		SPMchID:          client.SPMchID(),
		SubMchID:         subMchID,
		AppID:            appIdentity.AppID,
		TransactionID:    transactionID,
		OutOrderNo:       outOrderNo,
		TotalAmount:      order.PayAmount,
		TotalShareAmount: centsToYuan(totalShareCents),
		Status:           RecordStatusProcessing,
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		for i := range details {
			details[i].RecordID = record.ID
		}
		if len(details) > 0 {
			if err := tx.Create(&details).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Order{}).Where("id = ? AND profit_sharing_status = 1", order.ID).
			Updates(map[string]interface{}{
				"profit_sharing_amount":   centsToYuan(totalShareCents),
				"profit_sharing_order_no": outOrderNo,
				"profit_sharing_error":    "",
			}).Error
	}); err != nil {
		return fmt.Errorf("写入分账记录失败: %w", err)
	}

	// 对尚未建立微信关系的接收方，先建立关系
	if err := ensureWechatRelations(ctx, client, appIdentity.AppID, subMchID, receivers); err != nil {
		return markRecordFailed(record.ID, order.ID, err)
	}

	// 发起分账
	if _, err := client.CreateProfitSharingOrder(ctx, wechatpay.ProfitSharingRequest{
		SubMchID:      subMchID,
		AppID:         appIdentity.AppID,
		TransactionID: transactionID,
		OutOrderNo:    outOrderNo,
		Receivers:     items,
	}); err != nil {
		return markRecordFailed(record.ID, order.ID, err)
	}

	// 查询并回写分账结果
	result, err := client.QueryProfitSharingOrder(ctx, subMchID, transactionID, outOrderNo)
	if err != nil {
		return markRecordFailed(record.ID, order.ID, fmt.Errorf("分账结果查询失败: %w", err))
	}
	return applyQueryResult(record.ID, order.ID, result)
}

// ensureWechatRelations 为未绑定微信关系的接收方建立关系
func ensureWechatRelations(
	ctx context.Context,
	client *wechatpay.ServiceProviderClient,
	appID, subMchID string,
	receivers []models.ProfitSharingReceiver,
) error {
	for _, r := range receivers {
		if r.DefaultRatio <= 0 {
			continue
		}
		if r.WechatBound == 1 {
			continue
		}
		if err := client.AddProfitSharingReceiver(ctx, wechatpay.AddProfitSharingReceiverRequest{
			SubMchID:     subMchID,
			AppID:        appID,
			Type:         ReceiverWechatType(r.ReceiverType),
			Account:      r.Account,
			Name:         r.PersonalName,
			RelationType: r.RelationType,
		}); err != nil {
			return fmt.Errorf("为接收方 %s 建立微信关系失败: %w", r.Name, err)
		}
		database.DB.Model(&models.ProfitSharingReceiver{}).Where("id = ?", r.ID).
			Updates(map[string]interface{}{"wechat_bound": 1, "wechat_error": ""})
	}
	return nil
}

// applyQueryResult 根据微信查询结果更新分账记录与订单状态
func applyQueryResult(recordID uint64, orderID uint64, result *wechatpay.ProfitSharingResult) error {
	db := database.DB
	var status = RecordStatusProcessing
	var shareTime *time.Time

	// 统计各方结果：出现失败则整单失败，全部成功则完成
	hasFailed := false
	allSuccess := len(result.Receivers) > 0
	for _, rc := range result.Receivers {
		if rc.Result == ResultStatusClosed {
			hasFailed = true
		}
		if rc.Result != ResultStatusSuccess {
			allSuccess = false
		}
		if rc.FinishTime != nil && shareTime == nil {
			t := *rc.FinishTime
			shareTime = &t
		}
	}
	switch {
	case hasFailed:
		status = RecordStatusFailed
	case allSuccess:
		status = RecordStatusSuccess
	}

	// 回写各方明细（按 account 匹配）
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, rc := range result.Receivers {
			var finishTime *time.Time
			if rc.FinishTime != nil {
				t := *rc.FinishTime
				finishTime = &t
			}
			updates := map[string]interface{}{
				"result_status": rc.Result,
				"detail_id":     rc.DetailID,
				"fail_reason":   rc.FailReason,
				"finish_time":   finishTime,
			}
			if err := tx.Model(&models.ProfitSharingRecordReceiver{}).
				Where("record_id = ? AND account = ?", recordID, rc.Account).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("回写分账明细失败: %w", err)
	}

	recUpdates := map[string]interface{}{
		"status":    status,
		"error_message": "",
	}
	orderUpdates := map[string]interface{}{
		"profit_sharing_status": status,
		"profit_sharing_error":  "",
	}
	if shareTime != nil {
		recUpdates["share_time"] = shareTime
		orderUpdates["profit_sharing_at"] = shareTime
	}
	if status == RecordStatusSuccess {
		recUpdates["error_message"] = ""
	}
	db.Model(&models.ProfitSharingRecord{}).Where("id = ?", recordID).Updates(recUpdates)
	db.Model(&models.Order{}).Where("id = ?", orderID).Updates(orderUpdates)
	return nil
}

// markRecordFailed 将分账记录与订单标记为失败
func markRecordFailed(recordID, orderID uint64, cause error) error {
	msg := cause.Error()
	if len(msg) > 512 {
		msg = msg[:512]
	}
	database.DB.Model(&models.ProfitSharingRecord{}).Where("id = ?", recordID).
		Updates(map[string]interface{}{"status": RecordStatusFailed, "error_message": msg})
	database.DB.Model(&models.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{"profit_sharing_status": OrderFailed, "profit_sharing_error": msg})
	return cause
}

// markOrderSkipped 将订单标记为「已跳过」，不产生分账记录
func markOrderSkipped(orderID uint64, msg string) error {
	if len(msg) > 512 {
		msg = msg[:512]
	}
	return database.DB.Model(&models.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{"profit_sharing_status": OrderSkipped, "profit_sharing_error": msg}).Error
}

// amountToCents 元转分
func amountToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// centsToYuan 分转元
func centsToYuan(cents int64) float64 {
	return float64(cents) / 100
}