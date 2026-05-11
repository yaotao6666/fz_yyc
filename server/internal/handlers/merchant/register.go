package merchant

import (
	"encoding/json"
	"net/http"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Name             string `json:"name" binding:"required"`
	ContactName      string `json:"contact_name" binding:"required"`
	ContactPhone     string `json:"contact_phone" binding:"required"`
	ContactEmail     string `json:"contact_email"`
	Address          string `json:"address" binding:"required"`
	BusinessCategory string `json:"business_category" binding:"required"`
	InviteCode       string `json:"invite_code"`
	License          any    `json:"license"`
	BankAccount      any    `json:"bank_account"`
	StoreInfo        any    `json:"store_info"`
}

func Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var exists int64
	if err := database.DB.Model(&models.Merchant{}).Where("contact_phone = ?", req.ContactPhone).Count(&exists).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询入驻状态失败")
		return
	}
	if exists > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该手机号已提交入驻申请")
		return
	}

	tx := database.DB.Begin()

	merchant := models.Merchant{
		ServiceProviderID: 1,
		Name:              req.Name,
		ContactName:       req.ContactName,
		ContactPhone:      req.ContactPhone,
		ContactEmail:      req.ContactEmail,
		Address:           req.Address,
		BusinessCategory:  req.BusinessCategory,
		ApplymentStatus:   0,
		AuditStatus:       0,
		Status:            1,
	}
	if err := tx.Create(&merchant).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建商家失败")
		return
	}

	contactInfo, _ := json.Marshal(gin.H{
		"contact_name":  req.ContactName,
		"contact_phone": req.ContactPhone,
		"contact_email": req.ContactEmail,
		"invite_code":   req.InviteCode,
	})
	businessLicenseInfo, _ := json.Marshal(req.License)
	bankAccountInfo, _ := json.Marshal(req.BankAccount)
	storeInfo, _ := json.Marshal(req.StoreInfo)

	application := models.MerchantApplication{
		MerchantID:          merchant.ID,
		MerchantName:        req.Name,
		BusinessLicenseInfo: models.JSON(businessLicenseInfo),
		BankAccountInfo:     models.JSON(bankAccountInfo),
		StoreInfo:           models.JSON(storeInfo),
		ContactInfo:         models.JSON(contactInfo),
		Status:              0,
	}
	if err := tx.Create(&application).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建入驻申请失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "提交入驻申请失败")
		return
	}

	response.Success(c, gin.H{
		"merchant_id":    merchant.ID,
		"application_id": application.ID,
		"status":         "draft",
	})
}
