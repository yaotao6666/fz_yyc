package merchant

import (
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var staff models.MerchantStaff
	if err := database.DB.Preload("Merchant").Where("username = ? AND status = ?", req.Username, 1).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	now := time.Now()
	database.DB.Model(&models.MerchantStaff{}).Where("id = ?", staff.ID).Update("last_login_at", now)

	// 使用商家ID而不是员工ID生成token
	token, _ := utils.GenerateToken(staff.MerchantID, "merchant", staff.Username)
	response.Success(c, gin.H{
		"token":       token,
		"merchant_id": staff.MerchantID,
		"staff":       staff,
	})
}

func GetProfile(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	qiniuService := qiniu.GetService()
	if qiniuService != nil {
		merchant.Logo = qiniuService.BuildPrivateURL(merchant.Logo)
		merchant.CoverImage = qiniuService.BuildPrivateURL(merchant.CoverImage)
	}

	staff, _ := getCurrentMerchantStaff(c)

	response.Success(c, gin.H{
		"staff":    staff,
		"merchant": merchant,
	})
}

type UpdateProfileRequest struct {
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	CoverImage     string  `json:"cover_image"`
	ContactName    string  `json:"contact_name"`
	ContactPhone   string  `json:"contact_phone"`
	ContactEmail   string  `json:"contact_email"`
	Address        string  `json:"address"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	BusinessHours  string  `json:"business_hours"`
	Announcement   string  `json:"announcement"`
	MinOrderAmount float64 `json:"min_order_amount"`
}

func UpdateProfile(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if req.ContactName != "" {
		updates["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}
	if req.ContactEmail != "" {
		updates["contact_email"] = req.ContactEmail
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Lat != 0 {
		updates["lat"] = req.Lat
	}
	if req.Lng != 0 {
		updates["lng"] = req.Lng
	}
	if req.BusinessHours != "" {
		updates["business_hours"] = req.BusinessHours
	}
	if req.Announcement != "" {
		updates["announcement"] = req.Announcement
	}
	if req.MinOrderAmount > 0 {
		updates["min_order_amount"] = req.MinOrderAmount
	}

	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商家信息失败")
		return
	}

	var merchant models.Merchant
	database.DB.First(&merchant, merchantID)
	response.Success(c, merchant)
}

func GetSettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var deliverySettings models.MerchantDeliverySettings
	database.DB.Where("merchant_id = ?", merchantID).First(&deliverySettings)

	notifyEnabled := true
	browseNotifyEnabled := true
	wechatBound := false
	unionID := ""
	var wechatBoundAt *time.Time
	if staff, err := getCurrentMerchantStaff(c); err == nil {
		notifyEnabled = staff.NotifyEnabled
		browseNotifyEnabled = staff.BrowseNotifyEnabled
		wechatBound = strings.TrimSpace(staff.OpenID) != ""
		unionID = strings.TrimSpace(staff.UnionID)
		wechatBoundAt = staff.WechatBoundAt
	}

	response.Success(c, gin.H{
		"announcement":          merchant.Announcement,
		"business_hours":        merchant.BusinessHours,
		"min_order_amount":      merchant.MinOrderAmount,
		"takeout_enabled":       merchant.TakeoutEnabled,
		"dine_in_enabled":       merchant.DineInEnabled,
		"pickup_enabled":        merchant.PickupEnabled,
		"notify_enabled":        notifyEnabled,
		"browse_notify_enabled": browseNotifyEnabled,
		"wechat_bound":          wechatBound,
		"unionid":               unionID,
		"wechat_bound_at":       wechatBoundAt,
		"delivery_settings":     deliverySettings,
	})
}

type UpdateSettingsRequest struct {
	TakeoutEnabled      *bool `json:"takeout_enabled"`
	DineInEnabled       *bool `json:"dine_in_enabled"`
	PickupEnabled       *bool `json:"pickup_enabled"`
	NotifyEnabled       *bool `json:"notify_enabled"`
	BrowseNotifyEnabled *bool `json:"browse_notify_enabled"`
}

func UpdateSettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	merchantUpdates := map[string]interface{}{}
	if req.TakeoutEnabled != nil {
		merchantUpdates["takeout_enabled"] = *req.TakeoutEnabled
	}
	if req.DineInEnabled != nil {
		merchantUpdates["dine_in_enabled"] = *req.DineInEnabled
	}
	if req.PickupEnabled != nil {
		merchantUpdates["pickup_enabled"] = *req.PickupEnabled
	}

	if len(merchantUpdates) > 0 {
		if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Updates(merchantUpdates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新设置失败")
			return
		}
	}

	if req.NotifyEnabled != nil || req.BrowseNotifyEnabled != nil {
		staff, err := getCurrentMerchantStaff(c)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工身份失败")
			return
		}

		staffUpdates := map[string]interface{}{}
		if req.NotifyEnabled != nil {
			staffUpdates["notify_enabled"] = *req.NotifyEnabled
		}
		if req.BrowseNotifyEnabled != nil {
			staffUpdates["browse_notify_enabled"] = *req.BrowseNotifyEnabled
		}

		if err := database.DB.Model(&models.MerchantStaff{}).
			Where("id = ?", staff.ID).
			Updates(staffUpdates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新提示音设置失败")
			return
		}
	}

	response.Success(c, gin.H{"message": "设置更新成功"})
}

type StatusRequest struct {
	Status *uint8 `json:"status"`
}

func UpdateStatus(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.Status == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "状态不能为空")
		return
	}
	if *req.Status != 0 && *req.Status != 1 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "状态值不合法")
		return
	}

	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Update("status", *req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新状态失败")
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

func GetQRCode(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("id", "name").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	qrCode, err := utils.GenerateMerchantStoreQRCode(merchantID, 280)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "生成微信小程序码失败: "+err.Error())
		return
	}
	// 保存到数据库，便于详情页直接复用最新二维码。
	database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Update("qrcode_url", qrCode.QRCodeURL)

	response.Success(c, gin.H{
		"qrcode_url":  qrCode.QRCodeURL,
		"scene":       qrCode.Scene,
		"page":        qrCode.Page,
		"placeholder": false,
		"message":     "微信小程序码生成成功",
	})
}

func GetDeliverySettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("id", "takeout_enabled", "dine_in_enabled", "pickup_enabled").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var settings models.MerchantDeliverySettings
	if err := database.DB.Where("merchant_id = ?", merchantID).First(&settings).Error; err != nil {
		settings = models.MerchantDeliverySettings{MerchantID: merchantID}
	}

	response.Success(c, gin.H{
		"enabled":              settings.Enabled,
		"base_fee":             settings.BaseFee,
		"free_delivery_amount": settings.FreeDeliveryAmount,
		"max_distance":         settings.MaxDistance,
		"distance_rules":       settings.DistanceRules,
		"takeout_enabled":      merchant.TakeoutEnabled,
		"dine_in_enabled":      merchant.DineInEnabled,
		"pickup_enabled":       merchant.PickupEnabled,
	})
}

type DeliverySettingsRequest struct {
	Enabled            bool    `json:"enabled"`
	BaseFee            float64 `json:"base_fee"`
	FreeDeliveryAmount float64 `json:"free_delivery_amount"`
	MaxDistance        uint    `json:"max_distance"`
	DistanceRules      []struct {
		MinDistance float64 `json:"min_distance"`
		MaxDistance float64 `json:"max_distance"`
		Fee         float64 `json:"fee"`
	} `json:"distance_rules"`
}

type normalizedDistanceRule struct {
	MinDistance float64 `json:"min_distance"`
	MaxDistance float64 `json:"max_distance"`
	Fee         float64 `json:"fee"`
}

func normalizeDeliverySettingsRules(req DeliverySettingsRequest) ([]normalizedDistanceRule, error) {
	if req.BaseFee < 0 {
		return nil, fmt.Errorf("基础配送费不能小于0")
	}
	if req.FreeDeliveryAmount < 0 {
		return nil, fmt.Errorf("满额免配送费门槛不能小于0")
	}
	if req.MaxDistance == 0 {
		return nil, fmt.Errorf("最大配送距离必须大于0")
	}

	rules := make([]normalizedDistanceRule, 0, len(req.DistanceRules))
	for index, rule := range req.DistanceRules {
		if rule.MinDistance < 0 {
			return nil, fmt.Errorf("第%d条规则起始距离不能小于0", index+1)
		}
		if rule.MaxDistance <= rule.MinDistance {
			return nil, fmt.Errorf("第%d条规则结束距离必须大于起始距离", index+1)
		}
		if rule.Fee < 0 {
			return nil, fmt.Errorf("第%d条规则配送费不能小于0", index+1)
		}
		if rule.MaxDistance > float64(req.MaxDistance) {
			return nil, fmt.Errorf("第%d条规则超出最大配送距离", index+1)
		}

		rules = append(rules, normalizedDistanceRule{
			MinDistance: rule.MinDistance,
			MaxDistance: rule.MaxDistance,
			Fee:         rule.Fee,
		})
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].MinDistance == rules[j].MinDistance {
			return rules[i].MaxDistance < rules[j].MaxDistance
		}
		return rules[i].MinDistance < rules[j].MinDistance
	})

	for index := 1; index < len(rules); index++ {
		if rules[index].MinDistance < rules[index-1].MaxDistance {
			return nil, fmt.Errorf("第%d条规则与前一条规则区间重叠", index+1)
		}
	}

	return rules, nil
}

func UpdateDeliverySettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("id", "takeout_enabled", "dine_in_enabled", "pickup_enabled").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var req DeliverySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	normalizedRules, err := normalizeDeliverySettingsRules(req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	var settings models.MerchantDeliverySettings
	if err := database.DB.Where("merchant_id = ?", merchantID).First(&settings).Error; err != nil {
		settings = models.MerchantDeliverySettings{MerchantID: merchantID}
	}

	settings.Enabled = req.Enabled
	settings.BaseFee = req.BaseFee
	settings.FreeDeliveryAmount = req.FreeDeliveryAmount
	settings.MaxDistance = req.MaxDistance

	rulesJSON, _ := json.Marshal(normalizedRules)
	settings.DistanceRules = models.JSON(rulesJSON)

	if err := database.DB.Save(&settings).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存配送设置失败")
		return
	}

	response.Success(c, gin.H{
		"enabled":              settings.Enabled,
		"base_fee":             settings.BaseFee,
		"free_delivery_amount": settings.FreeDeliveryAmount,
		"max_distance":         settings.MaxDistance,
		"distance_rules":       settings.DistanceRules,
		"takeout_enabled":      merchant.TakeoutEnabled,
		"dine_in_enabled":      merchant.DineInEnabled,
		"pickup_enabled":       merchant.PickupEnabled,
	})
}
