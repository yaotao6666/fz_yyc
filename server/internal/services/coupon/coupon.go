// Package coupon 优惠券共享业务逻辑（PRD V2.0 阶段二：优惠券营销闭环）。
// 领取/发放/核销/抵扣计算统一收口，merchant 后台与 C 端 store 复用。
package coupon

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"

	"gorm.io/gorm"
)

// Common errors，handler 层据此映射响应码
var (
	ErrTemplateNotFound   = errors.New("券模板不存在")
	ErrTemplateDisabled   = errors.New("券已停用")
	ErrTemplateSoldOut    = errors.New("券已抢光")
	ErrTemplateNotStarted = errors.New("券未到领取时间")
	ErrTemplateEnded      = errors.New("券已过领取时间")
	ErrLimitExceeded      = errors.New("已超出限领数量")
	ErrCouponNotUsable    = errors.New("优惠券不可用")
	ErrCouponNotFound     = errors.New("优惠券不存在")
)

// ComputeExpiredAt 根据模板有效期规则计算用户券过期时间
func ComputeExpiredAt(t models.CouponTemplate, receivedAt time.Time) time.Time {
	if t.ValidType == utils.CouponValidTypeFixed {
		if t.ValidEndAt != nil {
			return *t.ValidEndAt
		}
		return receivedAt.AddDate(1, 0, 0) // 兜底
	}
	days := t.ValidDays
	if days <= 0 {
		days = utils.CouponValidDaysDefault
	}
	return receivedAt.AddDate(0, 0, days)
}

// CanReceiveNow 校验模板当前是否可领（启用状态 + 固定期限领取窗口）
func CanReceiveNow(t models.CouponTemplate, now time.Time) error {
	if t.Status != utils.CouponTemplateStatusEnabled {
		return ErrTemplateDisabled
	}
	if t.ValidType == utils.CouponValidTypeFixed {
		if t.ValidStartAt != nil && now.Before(*t.ValidStartAt) {
			return ErrTemplateNotStarted
		}
		if t.ValidEndAt != nil && now.After(*t.ValidEndAt) {
			return ErrTemplateEnded
		}
	}
	if t.TotalCount > 0 && t.ReceivedCount >= t.TotalCount {
		return ErrTemplateSoldOut
	}
	return nil
}

// Receive 领取/发放一张券（自主领取 source=1 / 系统发放 2 / 手动发放 3）。
// 事务内：限领校验 → 乐观锁扣减模板余量 → 落库用户券。
func Receive(userID uint64, templateID uint64, source uint8) (*models.UserCoupon, error) {
	var result *models.UserCoupon
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var t models.CouponTemplate
		if err := tx.Where("id = ?", templateID).First(&t).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTemplateNotFound
			}
			return err
		}

		if err := CanReceiveNow(t, time.Now()); err != nil {
			return err
		}

		// 限领校验（不区分状态：已使用/过期的领取记录同样计数）
		var count int64
		if err := tx.Model(&models.UserCoupon{}).
			Where("user_id = ? AND template_id = ?", userID, templateID).
			Count(&count).Error; err != nil {
			return err
		}
		if t.PerUserLimit > 0 && count >= int64(t.PerUserLimit) {
			return ErrLimitExceeded
		}

		// 乐观锁扣减余量（total_count=0 不限总量，仅自增计数）
		now := time.Now()
		expiredAt := ComputeExpiredAt(t, now)
		if t.TotalCount > 0 {
			res := tx.Model(&models.CouponTemplate{}).
				Where("id = ? AND received_count < total_count", templateID).
				Update("received_count", gorm.Expr("received_count + 1"))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrTemplateSoldOut
			}
		} else {
			if err := tx.Model(&models.CouponTemplate{}).
				Where("id = ?", templateID).
				Update("received_count", gorm.Expr("received_count + 1")).Error; err != nil {
				return err
			}
		}

		uc := models.UserCoupon{
			UserID:     userID,
			TemplateID: templateID,
			Status:     utils.UserCouponStatusUnused,
			Source:     source,
			ExpiredAt:  expiredAt,
		}
		if err := tx.Create(&uc).Error; err != nil {
			return err
		}
		result = &uc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ScopeMatch 判断券模板适用范围是否命中订单商品（分类/商品ID集合）
func ScopeMatch(t models.CouponTemplate, productIDs []uint64, categoryIDs []uint64) bool {
	switch t.ApplyScope {
	case utils.CouponScopeAll:
		return true
	case utils.CouponScopeCategory:
		scope := parseUint64List(t.ScopeIds)
		if len(scope) == 0 {
			return true
		}
		for _, cid := range categoryIDs {
			for _, s := range scope {
				if cid == s {
					return true
				}
			}
		}
		return false
	case utils.CouponScopeProduct:
		scope := parseUint64List(t.ScopeIds)
		if len(scope) == 0 {
			return true
		}
		for _, pid := range productIDs {
			for _, s := range scope {
				if pid == s {
					return true
				}
			}
		}
		return false
	}
	return false
}

// ComputeDiscount 计算抵扣金额：满减券=面值封顶；折扣券=总额×(1-折扣率)。
// 押金与配送费不参与抵扣；抵扣不超过商品总金额。
func ComputeDiscount(t models.CouponTemplate, totalAmount float64) float64 {
	if totalAmount <= 0 {
		return 0
	}
	var discount float64
	switch t.Type {
	case utils.CouponTypeThreshold:
		discount = t.DiscountAmount
	case utils.CouponTypeDiscount:
		discount = totalAmount * (1 - t.DiscountRate)
	}
	if discount < 0 {
		discount = 0
	}
	if discount > totalAmount {
		discount = totalAmount
	}
	return discount
}

// UsableCheck 校验一张用户券对当前订单是否可用，返回抵扣金额。
// totalAmount 为商品总金额（不含配送费/押金），productIDs/categoryIDs 用于范围匹配。
func UsableCheck(uc models.UserCoupon, t models.CouponTemplate, totalAmount float64, productIDs, categoryIDs []uint64, now time.Time) (float64, error) {
	if uc.UserID == 0 || uc.TemplateID != t.ID {
		return 0, ErrCouponNotUsable
	}
	if uc.Status != utils.UserCouponStatusUnused {
		return 0, ErrCouponNotUsable
	}
	if now.After(uc.ExpiredAt) {
		return 0, ErrCouponNotUsable
	}
	if totalAmount < t.ThresholdAmount {
		return 0, fmt.Errorf("%w：未达到使用门槛", ErrCouponNotUsable)
	}
	if !ScopeMatch(t, productIDs, categoryIDs) {
		return 0, fmt.Errorf("%w：商品不在适用范围", ErrCouponNotUsable)
	}
	return ComputeDiscount(t, totalAmount), nil
}

// OrderPreview 下单可用券预览条目
type OrderPreview struct {
	UserCouponID uint64  `json:"user_coupon_id"`
	TemplateID    uint64  `json:"template_id"`
	Name          string  `json:"name"`
	Type          uint8   `json:"type"`
	ThresholdAmt  float64 `json:"threshold_amount"`
	DiscountAmt   float64 `json:"discount_amount"`
	DiscountRate  float64 `json:"discount_rate"`
	Discount      float64 `json:"discount"` // 本单预计抵扣金额
	ExpiredAt     time.Time `json:"expired_at"`
}

// UsableForOrder 返回用户在当前订单下可用券列表（按抵扣金额降序，最优在前）
func UsableForOrder(userID uint64, totalAmount float64, productIDs, categoryIDs []uint64) ([]OrderPreview, error) {
	now := time.Now()
	var ucs []models.UserCoupon
	if err := database.DB.
		Where("user_id = ? AND status = ?", userID, utils.UserCouponStatusUnused).
		Preload("Template").
		Find(&ucs).Error; err != nil {
		return nil, err
	}

	list := make([]OrderPreview, 0, len(ucs))
	for _, uc := range ucs {
		if uc.Template == nil {
			continue
		}
		if now.After(uc.ExpiredAt) {
			continue
		}
		if totalAmount < uc.Template.ThresholdAmount {
			continue
		}
		if !ScopeMatch(*uc.Template, productIDs, categoryIDs) {
			continue
		}
		list = append(list, OrderPreview{
			UserCouponID: uc.ID,
			TemplateID:    uc.TemplateID,
			Name:          uc.Template.Name,
			Type:          uc.Template.Type,
			ThresholdAmt:  uc.Template.ThresholdAmount,
			DiscountAmt:   uc.Template.DiscountAmount,
			DiscountRate:  uc.Template.DiscountRate,
			Discount:      ComputeDiscount(*uc.Template, totalAmount),
			ExpiredAt:     uc.ExpiredAt,
		})
	}
	// 最优在前：先满减面值/折扣后金额大者
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].Discount > list[j-1].Discount; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	return list, nil
}

// Redeem 核销用户券（在事务 tx 内执行）：状态幂等校验 + 写入核销信息。
// 返回抵扣金额；不可用时返回 ErrCouponNotUsable。
func Redeem(tx *gorm.DB, userID uint64, userCouponID uint64, totalAmount float64, productIDs, categoryIDs []uint64, orderID uint64, orderNo string) (float64, error) {
	var uc models.UserCoupon
	if err := tx.Where("id = ? AND user_id = ?", userCouponID, userID).First(&uc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrCouponNotFound
		}
		return 0, err
	}

	var t models.CouponTemplate
	if err := tx.Where("id = ?", uc.TemplateID).First(&t).Error; err != nil {
		return 0, ErrCouponNotFound
	}

	discount, err := UsableCheck(uc, t, totalAmount, productIDs, categoryIDs, time.Now())
	if err != nil {
		return 0, err
	}

	now := time.Now()
	res := tx.Model(&models.UserCoupon{}).
		Where("id = ? AND status = ?", uc.ID, utils.UserCouponStatusUnused).
		Updates(map[string]interface{}{
			"status":    utils.UserCouponStatusUsed,
			"used_at":  now,
			"order_id": orderID,
			"order_no": orderNo,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected == 0 {
		return 0, ErrCouponNotUsable
	}
	return discount, nil
}

// RefreshExpired 将过期未使用的用户券置为已过期状态（定时任务）
func RefreshExpired() (int64, error) {
	res := database.DB.Model(&models.UserCoupon{}).
		Where("status = ? AND expired_at < ?", utils.UserCouponStatusUnused, time.Now()).
		Update("status", utils.UserCouponStatusExpired)
	return res.RowsAffected, res.Error
}

// GrantToInactiveUsers 30 天未下单用户自动发券（定时任务，source=2）。
// 模板取启用中的满减券里最新一张（简单实现）；同一活动周期内每用户仅发一次。
func GrantToInactiveUsers() (int, error) {
	cutoff := time.Now().AddDate(0, 0, -30)

	// 目标用户：注册满30天、30天内无订单
	var users []models.User
	if err := database.DB.
		Where("created_at < ? AND id NOT IN (SELECT DISTINCT user_id FROM orders WHERE created_at >= ?)", cutoff, cutoff).
		Find(&users).Error; err != nil {
		return 0, err
	}
	if len(users) == 0 {
		return 0, nil
	}

	// 取启用中的唤回券模板（优先名称含"唤回"，否则最新启用模板）
	var template models.CouponTemplate
	if err := database.DB.
		Where("status = ? AND type = ?", utils.CouponTemplateStatusEnabled, utils.CouponTypeThreshold).
		Order("id DESC").
		First(&template).Error; err != nil {
		return 0, nil // 无可用模板则跳过
	}

	granted := 0
	for _, u := range users {
		// 活动周期去重：已发放过系统券的用户不再发
		var count int64
		if err := database.DB.Model(&models.UserCoupon{}).
			Where("user_id = ? AND template_id = ? AND source = ?", u.ID, template.ID, utils.UserCouponSourceSystem).
			Count(&count).Error; err != nil {
			continue
		}
		if count > 0 {
			continue
		}
		// 发放走独立事务，单用户失败不影响整体
		if _, err := Receive(u.ID, template.ID, utils.UserCouponSourceSystem); err != nil {
			continue
		}
		granted++
	}
	return granted, nil
}

// parseUint64List 解析 JSON 数组字段为 uint64 列表
func parseUint64List(raw models.JSON) []uint64 {
	if len(raw) == 0 {
		return nil
	}
	var ids []uint64
	if err := json.Unmarshal(raw, &ids); err == nil {
		return ids
	}
	// 兼容字符串数组
	var strs []string
	if err := json.Unmarshal(raw, &strs); err == nil {
		for _, s := range strs {
			var v uint64
			if _, err := fmt.Sscanf(s, "%d", &v); err == nil && v > 0 {
				ids = append(ids, v)
			}
		}
	}
	return ids
}
