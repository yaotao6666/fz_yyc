package admin

import (
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
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.Where("username = ? AND status = ?", req.Username, 1).First(&admin).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	now := time.Now()
	database.DB.Model(&admin).Update("last_login_at", now)

	token, _ := utils.GenerateToken(admin.ID, "admin", admin.Username)
	response.Success(c, gin.H{
		"token":  token,
		"admin": admin,
	})
}

func GetServiceProvider(c *gin.Context) {
	var sp models.ServiceProvider
	if err := database.DB.First(&sp).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}
	response.Success(c, sp)
}

type UpdateServiceProviderRequest struct {
	Name         string `json:"name"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	CallbackURL  string `json:"callback_url"`
}

func UpdateServiceProvider(c *gin.Context) {
	var req UpdateServiceProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var sp models.ServiceProvider
	if err := database.DB.First(&sp).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ContactName != "" {
		updates["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}
	if req.CallbackURL != "" {
		updates["callback_url"] = req.CallbackURL
	}

	if err := database.DB.Model(&sp).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新服务商信息失败")
		return
	}

	database.DB.First(&sp)
	response.Success(c, sp)
}

func GetMerchantApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.MerchantApplication{}).Preload("Merchant")

	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}

	var total int64
	query.Count(&total)

	var applications []models.MerchantApplication
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&applications).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取申请列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": applications,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

func GetMerchantApplicationDetail(c *gin.Context) {
	id := c.Param("id")

	var application models.MerchantApplication
	if err := database.DB.Preload("Merchant").First(&application, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "申请不存在")
		return
	}

	response.Success(c, application)
}

func SubmitMerchantApplication(c *gin.Context) {
	id := c.Param("id")

	var application models.MerchantApplication
	if err := database.DB.First(&application, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "申请不存在")
		return
	}

	now := time.Now()
	if err := database.DB.Model(&application).Updates(map[string]interface{}{
		"status":       1,
		"submit_time":  now,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "提交申请失败")
		return
	}

	response.Success(c, gin.H{"message": "提交成功"})
}

func GetMerchantApplicationStatus(c *gin.Context) {
	id := c.Param("id")

	var application models.MerchantApplication
	if err := database.DB.Select("id, status, audit_detail, audit_time").First(&application, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "申请不存在")
		return
	}

	response.Success(c, application)
}

func GetPendingMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	database.DB.Model(&models.Merchant{}).Where("audit_status = ?", 0).Count(&total)

	var merchants []models.Merchant
	offset := (page - 1) * pageSize
	if err := database.DB.Where("audit_status = ?", 0).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&merchants).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取待审核商家失败")
		return
	}

	response.Success(c, gin.H{
		"list": merchants,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

type AuditMerchantRequest struct {
	AuditStatus uint8  `json:"audit_status" binding:"required,oneof=1 2"`
	AuditRemark string `json:"audit_remark"`
}

func AuditMerchant(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var req AuditMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
		return
	}

	updates := map[string]interface{}{
		"audit_status": req.AuditStatus,
		"audit_remark": req.AuditRemark,
		"status":        req.AuditStatus,
	}

	if err := database.DB.Model(&merchant).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "审核失败")
		return
	}

	database.DB.First(&merchant)
	response.Success(c, merchant)
}

func GetDashboard(c *gin.Context) {
	var totalMerchants int64
	var pendingMerchants int64
	var totalOrders int64
	var todayOrders int64
	var totalRevenue float64
	var todayRevenue float64

	database.DB.Model(&models.Merchant{}).Count(&totalMerchants)
	database.DB.Model(&models.Merchant{}).Where("audit_status = ?", 0).Count(&pendingMerchants)
	database.DB.Model(&models.Order{}).Count(&totalOrders)
	database.DB.Model(&models.Order{}).Where("DATE(created_at) = CURDATE()").Count(&todayOrders)

	database.DB.Model(&models.Order{}).Where("status >= 2").Select("COALESCE(SUM(pay_amount), 0)").Scan(&totalRevenue)
	database.DB.Model(&models.Order{}).Where("DATE(created_at) = CURDATE() AND status >= 2").Select("COALESCE(SUM(pay_amount), 0)").Scan(&todayRevenue)

	response.Success(c, gin.H{
		"total_merchants":   totalMerchants,
		"pending_merchants": pendingMerchants,
		"total_orders":      totalOrders,
		"today_orders":      todayOrders,
		"total_revenue":     totalRevenue,
		"today_revenue":     todayRevenue,
	})
}

func PaymentNotify(c *gin.Context) {
	// 微信支付回调处理
	// 具体实现待完成
	response.Success(c, gin.H{"message": "回调处理成功"})
}
