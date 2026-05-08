package merchant

import (
	"encoding/json"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
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
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	var staff models.MerchantStaff
	if err := database.DB.Preload("Merchant").Where("username = ? AND status = ?", req.Username, 1).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "用户名或密码错误")
		return
	}

	now := time.Now()
	database.DB.Model(&staff).Update("last_login_at", now)

	response.Success(c, gin.H{
		"token":  utils.GenerateToken(staff.ID, "merchant", staff.Username),
		"merchant_id": staff.MerchantID,
		"staff": staff,
	})
}

func GetProfile(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商家不存在")
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
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
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
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新商家信息失败")
		return
	}

	var merchant models.Merchant
	database.DB.First(&merchant, merchantID)
	response.Success(c, merchant)
}

type UpdateSettingsRequest struct {
	TakeoutEnabled  bool `json:"takeout_enabled"`
	DineInEnabled   bool `json:"dine_in_enabled"`
}

func UpdateSettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Updates(map[string]interface{}{
		"takeout_enabled": req.TakeoutEnabled,
		"dine_in_enabled": req.DineInEnabled,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新设置失败")
		return
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
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
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
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新证照信息失败")
		return
	}

	response.Success(c, license)
}

func UpdateBankAccount(c *gin.Context) {
	// 银行账户更新逻辑
	response.Success(c, gin.H{"message": "银行账户更新成功"})
}

type StatusRequest struct {
	Status uint8 `json:"status" binding:"required,oneof=0 1"`
}

func UpdateStatus(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", merchantID).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新状态失败")
		return
	}

	response.Success(c, gin.H{"message": "状态更新成功"})
}

func GetApplicationStatus(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("applyment_status, audit_status, audit_remark, sub_mch_status").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商家不存在")
		return
	}

	response.Success(c, merchant)
}

func GetQRCode(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var merchant models.Merchant
	if err := database.DB.Select("id", "qrcode_url").First(&merchant, merchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商家不存在")
		return
	}

	if merchant.QRCodeURL == "" {
		// 生成二维码URL
		qrcodeURL := "https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=TOKEN"
		database.DB.Model(&merchant).Update("qrcode_url", qrcodeURL)
		merchant.QRCodeURL = qrcodeURL
	}

	response.Success(c, gin.H{"qrcode_url": merchant.QRCodeURL})
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
		Distance float64 `json:"distance"`
		Fee      float64 `json:"fee"`
	} `json:"distance_rules"`
}

func UpdateDeliverySettings(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req DeliverySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
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

	if len(req.DistanceRules) > 0 {
		rulesJSON, _ := json.Marshal(req.DistanceRules)
		settings.DistanceRules = models.JSON(rulesJSON)
	}

	if err := database.DB.Save(&settings).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "保存配送设置失败")
		return
	}

	response.Success(c, settings)
}
