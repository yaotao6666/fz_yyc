package sp

import (
	"net/http"
	"strconv"
	"strings"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateMerchantRequest struct {
	Name                 string  `json:"name" binding:"required"`
	ContactName          string  `json:"contact_name"`
	ContactPhone         string  `json:"contact_phone"`
	ContactEmail         string  `json:"contact_email"`
	Address              string  `json:"address"`
	BusinessCategory     string  `json:"business_category"`
	BusinessHours        string  `json:"business_hours"`
	Announcement         string  `json:"announcement"`
	Username             string  `json:"username" binding:"required"`
	Password             string  `json:"password" binding:"required,min=6"`
	StaffName            string  `json:"staff_name"`
	StaffPhone           string  `json:"staff_phone"`
	SubMchID             string  `json:"sub_mch_id"`
	ProfitSharingEnabled bool    `json:"profit_sharing_enabled"`
	ProfitSharingRatio   float64 `json:"profit_sharing_ratio"`
}

type UpdateMerchantRequest struct {
	Name             *string  `json:"name"`
	ContactName      *string  `json:"contact_name"`
	ContactPhone     *string  `json:"contact_phone"`
	ContactEmail     *string  `json:"contact_email"`
	Address          *string  `json:"address"`
	BusinessCategory *string  `json:"business_category"`
	BusinessHours    *string  `json:"business_hours"`
	Announcement     *string  `json:"announcement"`
	Status           *uint8   `json:"status"`
}

type UpdateMerchantPaymentConfigRequest struct {
	SubMchID             string  `json:"sub_mch_id"`
	ProfitSharingEnabled bool    `json:"profit_sharing_enabled"`
	ProfitSharingRatio   float64 `json:"profit_sharing_ratio"`
}

func buildPaymentConfigStatus(subMchID string, enabled bool, ratio float64) uint8 {
	if strings.TrimSpace(subMchID) == "" {
		return 0
	}
	if enabled && ratio <= 0 {
		return 0
	}
	return 1
}

func CreateMerchant(c *gin.Context) {
	var req CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var existing models.MerchantStaff
	if err := database.DB.Where("username = ?", strings.TrimSpace(req.Username)).First(&existing).Error; err == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "商家登录账号已存在")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "生成密码失败")
		return
	}

	merchant := models.Merchant{
		ServiceProviderID:    serviceProviderID,
		Name:                 strings.TrimSpace(req.Name),
		ContactName:          strings.TrimSpace(req.ContactName),
		ContactPhone:         strings.TrimSpace(req.ContactPhone),
		ContactEmail:         strings.TrimSpace(req.ContactEmail),
		Address:              strings.TrimSpace(req.Address),
		BusinessCategory:     strings.TrimSpace(req.BusinessCategory),
		BusinessHours:        strings.TrimSpace(req.BusinessHours),
		Announcement:         strings.TrimSpace(req.Announcement),
		SubMchID:             strings.TrimSpace(req.SubMchID),
		ProfitSharingEnabled: req.ProfitSharingEnabled,
		ProfitSharingRatio:   req.ProfitSharingRatio,
		PaymentConfigStatus:  buildPaymentConfigStatus(req.SubMchID, req.ProfitSharingEnabled, req.ProfitSharingRatio),
		Status:               1,
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&merchant).Error; err != nil {
			return err
		}

		staff := models.MerchantStaff{
			MerchantID: merchant.ID,
			Username:   strings.TrimSpace(req.Username),
			Password:   string(passwordHash),
			Name:       strings.TrimSpace(req.StaffName),
			Phone:      strings.TrimSpace(req.StaffPhone),
			Role:       "owner",
			Status:     1,
		}
		if staff.Name == "" {
			staff.Name = merchant.ContactName
		}
		if staff.Phone == "" {
			staff.Phone = merchant.ContactPhone
		}
		return tx.Create(&staff).Error
	}); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建商家失败")
		return
	}

	c.Params = append(c.Params, gin.Param{Key: "merchant_id", Value: strconv.FormatUint(merchant.ID, 10)})
	GetMerchantDetail(c)
}

func UpdateMerchant(c *gin.Context) {
	merchantID, _ := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ? AND service_provider_id = ?", merchantID, serviceProviderID).First(&merchant).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var req UpdateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.ContactName != nil {
		updates["contact_name"] = strings.TrimSpace(*req.ContactName)
	}
	if req.ContactPhone != nil {
		updates["contact_phone"] = strings.TrimSpace(*req.ContactPhone)
	}
	if req.ContactEmail != nil {
		updates["contact_email"] = strings.TrimSpace(*req.ContactEmail)
	}
	if req.Address != nil {
		updates["address"] = strings.TrimSpace(*req.Address)
	}
	if req.BusinessCategory != nil {
		updates["business_category"] = strings.TrimSpace(*req.BusinessCategory)
	}
	if req.BusinessHours != nil {
		updates["business_hours"] = strings.TrimSpace(*req.BusinessHours)
	}
	if req.Announcement != nil {
		updates["announcement"] = strings.TrimSpace(*req.Announcement)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "未提供可更新内容")
		return
	}

	if err := database.DB.Model(&merchant).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商家失败")
		return
	}

	GetMerchantDetail(c)
}

func UpdateMerchantPaymentConfig(c *gin.Context) {
	merchantID, _ := strconv.ParseUint(c.Param("merchant_id"), 10, 64)
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ? AND service_provider_id = ?", merchantID, serviceProviderID).First(&merchant).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var req UpdateMerchantPaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if req.ProfitSharingEnabled && req.ProfitSharingRatio <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "开启分账时分账比例必须大于0")
		return
	}

	updates := map[string]any{
		"sub_mch_id":              strings.TrimSpace(req.SubMchID),
		"profit_sharing_enabled":  req.ProfitSharingEnabled,
		"profit_sharing_ratio":    req.ProfitSharingRatio,
		"payment_config_status":   buildPaymentConfigStatus(req.SubMchID, req.ProfitSharingEnabled, req.ProfitSharingRatio),
	}

	if err := database.DB.Model(&merchant).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新支付配置失败")
		return
	}

	GetMerchantDetail(c)
}

func GetProfitSharingRecords(c *gin.Context) {
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.MerchantProfitSharingRecord{}).Where("service_provider_id = ?", serviceProviderID)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		query = query.Where("merchant_id = ?", merchantID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(profit_sharing_date) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(profit_sharing_date) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var records []models.MerchantProfitSharingRecord
	if err := query.Order("profit_sharing_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取分账记录失败")
		return
	}

	response.Success(c, gin.H{
		"list": records,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetMerchantProfitSharingRecords(c *gin.Context) {
	merchantIDValue, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}
	merchantID, ok := merchantIDValue.(uint64)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "商家身份无效")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.MerchantProfitSharingRecord{}).Where("merchant_id = ?", merchantID)
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(profit_sharing_date) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(profit_sharing_date) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var records []models.MerchantProfitSharingRecord
	if err := query.Order("profit_sharing_date DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取分账记录失败")
		return
	}

	response.Success(c, gin.H{
		"list": records,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
