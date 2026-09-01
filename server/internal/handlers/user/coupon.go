package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/coupon"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================================
// C端优惠券接口（store 组）
// PRD V2.0 阶段二：优惠券营销闭环
// ============================================================

// buildStoreCouponTemplateResponse C端券模板展示字段（含剩余量）
func buildStoreCouponTemplateResponse(t models.CouponTemplate, receivedCount int64) gin.H {
	remain := -1 // -1 表示不限量
	if t.TotalCount > 0 {
		remain = t.TotalCount - int(receivedCount)
		if remain < 0 {
			remain = 0
		}
	}
	return gin.H{
		"id":               t.ID,
		"name":             t.Name,
		"type":             t.Type,
		"threshold_amount": t.ThresholdAmount,
		"discount_amount":  t.DiscountAmount,
		"discount_rate":    t.DiscountRate,
		"total_count":      t.TotalCount,
		"received_count":   t.ReceivedCount,
		"remain_count":     remain,
		"per_user_limit":   t.PerUserLimit,
		"valid_type":       t.ValidType,
		"valid_start_at":   t.ValidStartAt,
		"valid_end_at":     t.ValidEndAt,
		"valid_days":       t.ValidDays,
	}
}

// GetAvailableCoupons 可领券列表（首页领券入口；未登录也可浏览）
func GetAvailableCoupons(c *gin.Context) {
	now := time.Now()
	query := database.DB.Model(&models.CouponTemplate{}).
		Where("status = ?", utils.CouponTemplateStatusEnabled)

	// 固定期限券只展示领取窗口内的；领取后N天有效的全部展示
	query = query.Where(
		"(valid_type = ?) OR (valid_type = ? AND valid_start_at IS NOT NULL AND valid_start_at <= ? AND valid_end_at IS NOT NULL AND valid_end_at >= ?)",
		utils.CouponValidTypeAfter, utils.CouponValidTypeFixed, now, now,
	)

	var templates []models.CouponTemplate
	if err := query.Order("id DESC").Limit(50).Find(&templates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询可领券失败")
		return
	}

	// 已登录用户附带已领数量，便于前端控制限领
	var userID uint64
	if utils.GetUserID(c) != 0 {
		userID = utils.GetUserID(c)
	}

	list := make([]gin.H, 0, len(templates))
	for _, t := range templates {
		var received int64
		if userID > 0 {
			database.DB.Model(&models.UserCoupon{}).
				Where("user_id = ? AND template_id = ?", userID, t.ID).
				Count(&received)
		}
		item := buildStoreCouponTemplateResponse(t, int64(t.ReceivedCount))
		item["my_received_count"] = received
		item["can_receive"] = t.PerUserLimit > 0 && received < int64(t.PerUserLimit) &&
			(t.TotalCount == 0 || t.ReceivedCount < t.TotalCount)
		list = append(list, item)
	}
	response.Success(c, gin.H{"list": list})
}

// ReceiveCoupon 领取优惠券（需登录）
func ReceiveCoupon(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("template_id"), 10, 64)
	if templateID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	uc, err := coupon.Receive(userID, templateID, utils.UserCouponSourceSelf)
	if err != nil {
		switch {
		case errors.Is(err, coupon.ErrTemplateNotFound):
			response.Fail(c, http.StatusNotFound, response.CodeCouponNotFound, "券模板不存在")
		case errors.Is(err, coupon.ErrTemplateDisabled):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券已停用")
		case errors.Is(err, coupon.ErrTemplateNotStarted):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券未到领取时间")
		case errors.Is(err, coupon.ErrTemplateEnded):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券已过领取时间")
		case errors.Is(err, coupon.ErrTemplateSoldOut):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponUnavailable, "券已抢光")
		case errors.Is(err, coupon.ErrLimitExceeded):
			response.Fail(c, http.StatusBadRequest, response.CodeCouponLimitExceeded, "已超出限领数量")
		default:
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "领取失败")
		}
		return
	}
	response.Success(c, gin.H{
		"user_coupon_id": uc.ID,
		"template_id":    uc.TemplateID,
		"expired_at":    uc.ExpiredAt,
	})
}

// GetMyCoupons 我的券列表（status: 1=未使用 2=已使用 3=已过期，默认未使用）
func GetMyCoupons(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	status := uint8(1)
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > 4 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "状态参数无效")
			return
		}
		status = uint8(v)
	}

	// 惰性刷新：查未使用券时先把已过期的置为过期，保证口径准确
	if status == utils.UserCouponStatusUnused {
		database.DB.Model(&models.UserCoupon{}).
			Where("user_id = ? AND status = ? AND expired_at < ?", userID, utils.UserCouponStatusUnused, time.Now()).
			Update("status", utils.UserCouponStatusExpired)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query := database.DB.Model(&models.UserCoupon{}).
		Where("user_id = ? AND status = ?", userID, status)
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询优惠券失败")
		return
	}

	var ucs []models.UserCoupon
	if err := query.
		Preload("Template").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ucs).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询优惠券失败")
		return
	}

	list := make([]gin.H, 0, len(ucs))
	for _, uc := range ucs {
		item := gin.H{
			"id":          uc.ID,
			"template_id": uc.TemplateID,
			"status":      uc.Status,
			"source":      uc.Source,
			"received_at": uc.ReceivedAt,
			"expired_at":  uc.ExpiredAt,
			"used_at":     uc.UsedAt,
			"order_no":    uc.OrderNo,
		}
		if uc.Template != nil {
			item["name"] = uc.Template.Name
			item["type"] = uc.Template.Type
			item["threshold_amount"] = uc.Template.ThresholdAmount
			item["discount_amount"] = uc.Template.DiscountAmount
			item["discount_rate"] = uc.Template.DiscountRate
			item["apply_scope"] = uc.Template.ApplyScope
		}
		list = append(list, item)
	}
	response.Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// GetOrderUsableCoupons 下单可用券预览（结算页选券，最优在前）
func GetOrderUsableCoupons(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	totalAmount, _ := strconv.ParseFloat(c.Query("total_amount"), 64)
	if totalAmount <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单金额参数无效")
		return
	}

	// 商品ID集合（逗号分隔）；分类未显式传入时由商品反查，保证指定分类券也能匹配
	productIDs := parseUint64CSV(c.Query("product_ids"))
	categoryIDs := parseUint64CSV(c.Query("category_ids"))
	if len(categoryIDs) == 0 && len(productIDs) > 0 {
		var products []models.Product
		if err := database.DB.Select("id, category_id").Where("id IN ?", productIDs).Find(&products).Error; err == nil {
			for _, p := range products {
				if p.CategoryID != nil && *p.CategoryID > 0 {
					categoryIDs = append(categoryIDs, *p.CategoryID)
				}
			}
		}
	}

	list, err := coupon.UsableForOrder(userID, totalAmount, productIDs, categoryIDs)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询可用券失败")
		return
	}
	response.Success(c, gin.H{"list": list})
}

// parseUint64CSV 解析逗号分隔的 uint64 参数
func parseUint64CSV(raw string) []uint64 {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]uint64, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}
