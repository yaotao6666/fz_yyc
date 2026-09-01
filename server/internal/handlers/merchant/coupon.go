package merchant

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/coupon"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 优惠券模板管理（商家端 CRUD + 手动发放 + 领取记录）
// PRD V2.0 阶段二：优惠券营销闭环
// ============================================================

type CouponTemplateRequest struct {
	Name            string      `json:"name" binding:"required,max=64"`
	Type            uint8       `json:"type" binding:"required,oneof=1 2"`
	ThresholdAmount float64     `json:"threshold_amount" binding:"min=0"`
	DiscountAmount  float64     `json:"discount_amount" binding:"min=0"`
	DiscountRate    float64     `json:"discount_rate" binding:"min=0,max=1"`
	TotalCount      int         `json:"total_count" binding:"min=0"`
	PerUserLimit    int         `json:"per_user_limit" binding:"min=1"`
	ValidType       uint8       `json:"valid_type" binding:"required,oneof=1 2"`
	ValidStartAt    *time.Time  `json:"valid_start_at"`
	ValidEndAt      *time.Time  `json:"valid_end_at"`
	ValidDays       int         `json:"valid_days" binding:"min=0"`
	ApplyScope      uint8       `json:"apply_scope" binding:"required,oneof=1 2 3"`
	ScopeIds        models.JSON `json:"scope_ids"`
	Remark          string      `json:"remark" binding:"max=512"`
}

func buildCouponTemplateResponse(t models.CouponTemplate) gin.H {
	return gin.H{
		"id":               t.ID,
		"name":             t.Name,
		"type":             t.Type,
		"threshold_amount": t.ThresholdAmount,
		"discount_amount":  t.DiscountAmount,
		"discount_rate":    t.DiscountRate,
		"total_count":      t.TotalCount,
		"received_count":   t.ReceivedCount,
		"per_user_limit":   t.PerUserLimit,
		"valid_type":       t.ValidType,
		"valid_start_at":   t.ValidStartAt,
		"valid_end_at":     t.ValidEndAt,
		"valid_days":       t.ValidDays,
		"apply_scope":      t.ApplyScope,
		"scope_ids":        t.ScopeIds,
		"status":           t.Status,
		"remark":           t.Remark,
		"created_at":       t.CreatedAt,
		"updated_at":       t.UpdatedAt,
	}
}

// validateCouponTemplate 校验模板业务规则
func validateCouponTemplate(req CouponTemplateRequest) string {
	if req.Type == utils.CouponTypeThreshold && req.DiscountAmount <= 0 {
		return "满减券必须指定满减面值"
	}
	if req.Type == utils.CouponTypeDiscount && (req.DiscountRate <= 0 || req.DiscountRate >= 1) {
		return "折扣率必须为 0~1 之间的小数（如 0.90）"
	}
	if req.ValidType == utils.CouponValidTypeFixed {
		if req.ValidStartAt == nil || req.ValidEndAt == nil {
			return "固定期限券必须指定起止时间"
		}
		if !req.ValidEndAt.After(*req.ValidStartAt) {
			return "固定期限结束时间必须晚于开始时间"
		}
	} else if req.ValidDays <= 0 {
		return "领取后有效天数必须大于 0"
	}
	if req.ApplyScope != utils.CouponScopeAll && len(req.ScopeIds) == 0 {
		return "指定适用范围时必须选择分类或商品"
	}
	return ""
}

// GetCouponTemplates 券模板列表（分页 + 名称/状态/类型筛选）
func GetCouponTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.CouponTemplate{})
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if typ := c.Query("type"); typ != "" {
		query = query.Where("type = ?", typ)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询券模板失败")
		return
	}

	var templates []models.CouponTemplate
	if err := query.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&templates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询券模板失败")
		return
	}

	list := make([]gin.H, 0, len(templates))
	for _, t := range templates {
		list = append(list, buildCouponTemplateResponse(t))
	}
	response.Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// GetCouponTemplate 券模板详情
func GetCouponTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var t models.CouponTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "券模板不存在")
		return
	}
	response.Success(c, buildCouponTemplateResponse(t))
}

// CreateCouponTemplate 新增券模板
func CreateCouponTemplate(c *gin.Context) {
	var req CouponTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if msg := validateCouponTemplate(req); msg != "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, msg)
		return
	}

	t := models.CouponTemplate{
		Name:            req.Name,
		Type:            req.Type,
		ThresholdAmount: req.ThresholdAmount,
		DiscountAmount:  req.DiscountAmount,
		DiscountRate:    req.DiscountRate,
		TotalCount:      req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		ValidType:       req.ValidType,
		ValidStartAt:    req.ValidStartAt,
		ValidEndAt:      req.ValidEndAt,
		ValidDays:       req.ValidDays,
		ApplyScope:      req.ApplyScope,
		ScopeIds:        req.ScopeIds,
		Status:          utils.CouponTemplateStatusEnabled,
		Remark:          req.Remark,
	}
	if err := database.DB.Create(&t).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建券模板失败")
		return
	}
	response.Success(c, buildCouponTemplateResponse(t))
}

// UpdateCouponTemplate 编辑券模板（已领取数量不可回退）
func UpdateCouponTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var req CouponTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if msg := validateCouponTemplate(req); msg != "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, msg)
		return
	}

	var t models.CouponTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "券模板不存在")
		return
	}
	// 发行总量不可小于已领取数量
	if req.TotalCount > 0 && req.TotalCount < t.ReceivedCount {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "发行总量不可小于已领取数量")
		return
	}

	updates := map[string]interface{}{
		"name":             req.Name,
		"type":             req.Type,
		"threshold_amount": req.ThresholdAmount,
		"discount_amount":  req.DiscountAmount,
		"discount_rate":    req.DiscountRate,
		"total_count":      req.TotalCount,
		"per_user_limit":   req.PerUserLimit,
		"valid_type":       req.ValidType,
		"valid_start_at":   req.ValidStartAt,
		"valid_end_at":     req.ValidEndAt,
		"valid_days":       req.ValidDays,
		"apply_scope":      req.ApplyScope,
		"scope_ids":        req.ScopeIds,
		"remark":           req.Remark,
	}
	if err := database.DB.Model(&t).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新券模板失败")
		return
	}
	database.DB.First(&t, id)
	response.Success(c, buildCouponTemplateResponse(t))
}

// UpdateCouponTemplateStatus 启停券模板（停用后不可再领取，已领取的仍可使用）
func UpdateCouponTemplateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var req struct {
		Status uint8 `json:"status" binding:"required,oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var t models.CouponTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "券模板不存在")
		return
	}

	if err := database.DB.Model(&t).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新状态失败")
		return
	}
	database.DB.First(&t, id)
	response.Success(c, buildCouponTemplateResponse(t))
}

// DeleteCouponTemplate 删除券模板（已有用户领取则拒绝，需先停用）
func DeleteCouponTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var t models.CouponTemplate
	if err := database.DB.First(&t, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "券模板不存在")
		return
	}

	var received int64
	if err := database.DB.Model(&models.UserCoupon{}).
		Where("template_id = ?", id).Count(&received).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询领取记录失败")
		return
	}
	if received > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该券已有用户领取，不可删除，可停用")
		return
	}

	if err := database.DB.Delete(&t).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除券模板失败")
		return
	}
	response.Success(c, gin.H{"id": id})
}

// GrantCoupon 手动发放券给指定用户（source=3）
func GrantCoupon(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var req struct {
		UserID uint64 `json:"user_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeUserNotFound, "用户不存在")
		return
	}

	uc, err := coupon.Receive(req.UserID, id, utils.UserCouponSourceManual)
	if err != nil {
		switch {
		case errors.Is(err, coupon.ErrTemplateNotFound):
			response.Fail(c, http.StatusNotFound, response.CodeCouponNotFound, "券模板不存在")
		case errors.Is(err, coupon.ErrLimitExceeded):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponLimitExceeded, "该用户已超出限领数量")
		case errors.Is(err, coupon.ErrTemplateSoldOut):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券已抢光")
		case errors.Is(err, coupon.ErrTemplateDisabled):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券已停用")
		default:
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "发放失败")
		}
		return
	}
	response.Success(c, gin.H{
		"user_coupon_id": uc.ID,
		"user_id":        uc.UserID,
		"template_id":    uc.TemplateID,
		"expired_at":     uc.ExpiredAt,
	})
}

// buildUserCouponResponse 领取记录响应（带模板快照与用户昵称）
func buildUserCouponResponse(uc models.UserCoupon) gin.H {
	item := gin.H{
		"id":          uc.ID,
		"user_id":     uc.UserID,
		"template_id": uc.TemplateID,
		"status":      uc.Status,
		"source":      uc.Source,
		"received_at": uc.ReceivedAt,
		"expired_at":  uc.ExpiredAt,
		"used_at":     uc.UsedAt,
		"order_id":    uc.OrderID,
		"order_no":    uc.OrderNo,
	}
	if uc.Template != nil {
		item["template_name"] = uc.Template.Name
		item["template_type"] = uc.Template.Type
		item["threshold_amount"] = uc.Template.ThresholdAmount
		item["discount_amount"] = uc.Template.DiscountAmount
		item["discount_rate"] = uc.Template.DiscountRate
	}
	if uc.User != nil {
		item["nickname"] = uc.User.Nickname
		item["openid"] = uc.User.OpenID
	}
	return item
}

// GetUserCoupons 领取/使用记录（分页 + 用户/模板/状态/来源筛选）
func GetUserCoupons(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.UserCoupon{})
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if templateID := c.Query("template_id"); templateID != "" {
		query = query.Where("template_id = ?", templateID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询领取记录失败")
		return
	}

	var ucs []models.UserCoupon
	if err := query.
		Preload("Template").
		Preload("User").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ucs).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询领取记录失败")
		return
	}

	list := make([]gin.H, 0, len(ucs))
	for _, uc := range ucs {
		list = append(list, buildUserCouponResponse(uc))
	}
	response.Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}
