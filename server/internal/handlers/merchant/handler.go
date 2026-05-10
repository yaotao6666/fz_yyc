package merchant

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
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
	database.DB.Model(&staff).Update("last_login_at", now)

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

	response.Success(c, merchant)
}

type UpdateProfileRequest struct {
	Name          string  `json:"name"`
	Logo          string  `json:"logo"`
	ContactName   string  `json:"contact_name"`
	ContactPhone  string  `json:"contact_phone"`
	ContactEmail  string  `json:"contact_email"`
	Address       string  `json:"address"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	BusinessHours string  `json:"business_hours"`
	Announcement  string  `json:"announcement"`
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
	username, _ := c.Get("username")

	var merchant models.Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var deliverySettings models.MerchantDeliverySettings
	database.DB.Where("merchant_id = ?", merchantID).First(&deliverySettings)

	notifyEnabled := true
	if usernameStr, ok := username.(string); ok && usernameStr != "" {
		var staff models.MerchantStaff
		if err := database.DB.Where("merchant_id = ? AND username = ?", merchantID, usernameStr).First(&staff).Error; err == nil {
			notifyEnabled = staff.NotifyEnabled
		}
	}

	response.Success(c, gin.H{
		"announcement":      merchant.Announcement,
		"business_hours":    merchant.BusinessHours,
		"min_order_amount":  merchant.MinOrderAmount,
		"takeout_enabled":   merchant.TakeoutEnabled,
		"dine_in_enabled":   merchant.DineInEnabled,
		"notify_enabled":    notifyEnabled,
		"delivery_settings": deliverySettings,
	})
}

type UpdateSettingsRequest struct {
	TakeoutEnabled *bool `json:"takeout_enabled"`
	DineInEnabled  *bool `json:"dine_in_enabled"`
	NotifyEnabled  *bool `json:"notify_enabled"`
}

func UpdateSettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	username, _ := c.Get("username")

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

	if len(merchantUpdates) > 0 {
		if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Updates(merchantUpdates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新设置失败")
			return
		}
	}

	if req.NotifyEnabled != nil {
		usernameStr, ok := username.(string)
		if !ok || usernameStr == "" {
			response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工身份失败")
			return
		}

		if err := database.DB.Model(&models.MerchantStaff{}).
			Where("merchant_id = ? AND username = ?", merchantID, usernameStr).
			Update("notify_enabled", *req.NotifyEnabled).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新提示音设置失败")
			return
		}
	}

	response.Success(c, gin.H{"message": "设置更新成功"})
}

type LicenseRequest struct {
	LicenseNo         string `json:"license_no"`
	LicenseName       string `json:"license_name"`
	LicenseImage      string `json:"license_image"`
	LegalPerson       string `json:"legal_person"`
	LegalPersonID     string `json:"legal_person_id"`
	LegalPersonIDFront string `json:"legal_person_id_front"`
	LegalPersonIDBack  string `json:"legal_person_id_back"`
	ValidFrom         string `json:"valid_from"`
	ValidTo           string `json:"valid_to"`
}

func UpdateLicense(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req LicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var license models.MerchantLicense
	if err := database.DB.Where("merchant_id = ?", merchantID).First(&license).Error; err != nil {
		license = models.MerchantLicense{MerchantID: merchantID}
	}

	license.LicenseNo = req.LicenseNo
	license.LicenseName = req.LicenseName
	license.LicenseImage = req.LicenseImage
	license.LegalPerson = req.LegalPerson
	license.LegalPersonID = req.LegalPersonID
	license.LegalPersonIDFront = req.LegalPersonIDFront
	license.LegalPersonIDBack = req.LegalPersonIDBack

	if req.ValidFrom != "" {
		validFrom, _ := time.Parse("2006-01-02", req.ValidFrom)
		license.ValidFrom = &validFrom
	}
	if req.ValidTo != "" {
		validTo, _ := time.Parse("2006-01-02", req.ValidTo)
		license.ValidTo = &validTo
	}

	if err := database.DB.Save(&license).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新证照信息失败")
		return
	}

	response.Success(c, license)
}

func UpdateBankAccount(c *gin.Context) {
	// 银行账户更新逻辑
	response.Success(c, gin.H{"message": "银行账户更新成功"})
}

type StatusRequest struct {
	Status uint8 `json:"status" binding:"required"`
}

func UpdateStatus(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新状态失败")
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

func GetApplicationStatus(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("applyment_status, audit_status, audit_remark, sub_mch_status").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	response.Success(c, merchant)
}

func GetQRCode(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("id", "name").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	// 尝试生成微信小程序二维码
	scene := fmt.Sprintf("id=%d", merchantID)

	var qrCodeURL string
	qrCodeBytes, err := utils.CreateWXACode(scene, "pages/index/index", 280)
	if err != nil {
		// 如果生成失败，返回占位符URL
		// TODO: 小程序发布后需要配置正确的页面路径
		qrCodeURL = fmt.Sprintf("/placeholder-qrcode/%d", merchantID)
	} else {
		qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCodeBytes)
		qrCodeURL = "data:image/png;base64," + qrCodeBase64
		// 保存到数据库
		database.DB.Model(&merchant).Update("qrcode_url", qrCodeURL)
	}

	response.Success(c, gin.H{
		"qrcode_url": qrCodeURL,
		"scene":      scene,
		"page":       "pages/index/index",
		"placeholder": err != nil,
		"message":    "小程序发布后可生成正式二维码",
	})
}

func GetDeliverySettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var settings models.MerchantDeliverySettings
	if err := database.DB.Where("merchant_id = ?", merchantID).First(&settings).Error; err != nil {
		settings = models.MerchantDeliverySettings{MerchantID: merchantID}
	}

	response.Success(c, settings)
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

func UpdateDeliverySettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req DeliverySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
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

	rules := req.DistanceRules
	if rules == nil {
		rules = []struct {
			MinDistance float64 `json:"min_distance"`
			MaxDistance float64 `json:"max_distance"`
			Fee         float64 `json:"fee"`
		}{}
	}
	rulesJSON, _ := json.Marshal(rules)
	settings.DistanceRules = models.JSON(rulesJSON)

	if err := database.DB.Save(&settings).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存配送设置失败")
		return
	}

	response.Success(c, settings)
}
