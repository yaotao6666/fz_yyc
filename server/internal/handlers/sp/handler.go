package sp

import (
	"encoding/json"
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

func parseModelJSON(value models.JSON) any {
	if len(value) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil
	}
	return out
}

func parseStringSlice(value any) []string {
	rawList, ok := value.([]any)
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(rawList))
	for _, item := range rawList {
		if s, ok := item.(string); ok && s != "" {
			result = append(result, s)
		}
	}
	return result
}

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

	token, _ := utils.GenerateToken(admin.ID, "sp", admin.Username)
	database.DB.Preload("ServiceProvider").First(&admin, admin.ID)
	serviceProviderName := ""
	if admin.ServiceProvider != nil {
		serviceProviderName = admin.ServiceProvider.Name
	}
	if serviceProviderName == "" {
		serviceProviderName = admin.Username
	}
	adminName := admin.Name
	if adminName == "" {
		adminName = admin.Username
	}
	response.Success(c, gin.H{
		"token": token,
		"service_provider": gin.H{
			"id":         admin.ServiceProviderID,
			"name":       serviceProviderName,
			"admin_name": adminName,
		},
	})
}

func Logout(c *gin.Context) {
	response.Success(c, gin.H{"message": "退出成功"})
}

func GetDashboard(c *gin.Context) {
	var totalMerchants int64
	var todayOrders int64
	var todayRevenue float64
	var pendingMerchants int64

	database.DB.Model(&models.Merchant{}).Count(&totalMerchants)
	database.DB.Model(&models.Merchant{}).Where("audit_status = ?", 0).Count(&pendingMerchants)
	database.DB.Model(&models.Order{}).Where("DATE(created_at) = CURDATE()").Count(&todayOrders)
	database.DB.Model(&models.Order{}).Where("DATE(created_at) = CURDATE() AND status >= 2").Select("COALESCE(SUM(pay_amount), 0)").Scan(&todayRevenue)

	var distribution []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&models.Merchant{}).Select("business_category as category, count(*) as count").Group("business_category").Scan(&distribution)

	var trend []struct {
		Date   string `json:"date"`
		Orders int64  `json:"orders"`
	}
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var orders int64
		database.DB.Model(&models.Order{}).Where("DATE(created_at) = ?", date).Count(&orders)
		trend = append(trend, struct {
			Date   string `json:"date"`
			Orders int64  `json:"orders"`
		}{Date: date, Orders: orders})
	}

	response.Success(c, gin.H{
		"total_merchants":   totalMerchants,
		"pending_merchants": pendingMerchants,
		"today_orders":      todayOrders,
		"today_revenue":     todayRevenue,
		"distribution":      distribution,
		"trend":             trend,
	})
}

func GetPendingMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Merchant{})
	if status != "" {
		statusInt, err := strconv.Atoi(status)
		if err == nil {
			query = query.Where("audit_status = ?", statusInt)
		}
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			"name LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?",
			like, like, like,
		)
	}

	var total int64
	query.Count(&total)

	var merchants []models.Merchant
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&merchants).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取待审核商家失败")
		return
	}

	type pendingMerchantItem struct {
		ID               uint64    `json:"id"`
		Name             string    `json:"name"`
		ContactName      string    `json:"contact_name"`
		ContactPhone     string    `json:"contact_phone"`
		BusinessCategory string    `json:"business_category"`
		AppliedAt        time.Time `json:"applied_at"`
		Status           uint8     `json:"status"`
		RejectReason     string    `json:"reject_reason,omitempty"`
	}
	list := make([]pendingMerchantItem, 0, len(merchants))
	for _, merchant := range merchants {
		item := pendingMerchantItem{
			ID:               merchant.ID,
			Name:             merchant.Name,
			ContactName:      merchant.ContactName,
			ContactPhone:     merchant.ContactPhone,
			BusinessCategory: merchant.BusinessCategory,
			AppliedAt:        merchant.CreatedAt,
			Status:           merchant.AuditStatus,
			RejectReason:     merchant.AuditRemark,
		}
		list = append(list, item)
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetMerchantDetail(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var application models.MerchantApplication
	_ = database.DB.Where("merchant_id = ?", id).Order("created_at DESC").First(&application).Error

	var totalOrders int64
	database.DB.Model(&models.Order{}).Where("merchant_id = ?", id).Count(&totalOrders)

	var totalAmount float64
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2", id).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&totalAmount)

	businessLicenseInfo := parseModelJSON(application.BusinessLicenseInfo)
	legalPersonInfo := parseModelJSON(application.LegalPersonInfo)
	bankAccountInfo := parseModelJSON(application.BankAccountInfo)
	storeInfo := parseModelJSON(application.StoreInfo)

	storeName := ""
	storeImages := []string{}
	if storeMap, ok := storeInfo.(map[string]any); ok {
		if v, ok := storeMap["store_name"].(string); ok {
			storeName = v
		}
		storeImages = parseStringSlice(storeMap["store_images"])
	}

	response.Success(c, gin.H{
		"id":                merchant.ID,
		"name":              merchant.Name,
		"logo":              merchant.Logo,
		"contact_name":      merchant.ContactName,
		"contact_phone":     merchant.ContactPhone,
		"address":           merchant.Address,
		"business_category": merchant.BusinessCategory,
		"status":            merchant.Status,
		"audit_status":      merchant.AuditStatus,
		"audit_remark":      merchant.AuditRemark,
		"qrcode_url":        merchant.QRCodeURL,
		"created_at":        merchant.CreatedAt,
		"store_name":        storeName,
		"application": gin.H{
			"business_license_info": businessLicenseInfo,
			"legal_person_info":     legalPersonInfo,
			"bank_account_info":     bankAccountInfo,
			"store_info":            storeInfo,
		},
		"license": businessLicenseInfo,
		"settings": gin.H{
			"bank_account": bankAccountInfo,
			"store_images": storeImages,
		},
		"total_orders": totalOrders,
		"total_amount": totalAmount,
		"total_users":  0,
	})
}

type AuditRequest struct {
	AuditStatus uint8  `json:"audit_status" binding:"required,oneof=1 2"`
	AuditRemark string `json:"audit_remark"`
}

func AuditMerchant(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var req AuditRequest
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
		"status":       req.AuditStatus,
	}

	if err := database.DB.Model(&merchant).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "审核失败")
		return
	}

	database.DB.First(&merchant, id)
	response.Success(c, merchant)
}

func GetMerchantDistribution(c *gin.Context) {
	var businessDistribution []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&models.Merchant{}).Select("business_category as category, count(*) as count").Group("business_category").Scan(&businessDistribution)

	var statusDistribution []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	database.DB.Model(&models.Merchant{}).Select("CASE WHEN status = 1 THEN 'active' ELSE 'inactive' END as status, count(*) as count").Group("status").Scan(&statusDistribution)

	response.Success(c, gin.H{
		"business": businessDistribution,
		"status":   statusDistribution,
	})
}

func GetMerchantList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Merchant{})

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}

	var total int64
	query.Count(&total)

	var merchants []models.Merchant
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&merchants).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取商家列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": merchants,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetOrderAnalytics(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	daysInt, _ := strconv.Atoi(days)

	var trends []struct {
		Date   string  `json:"date"`
		Orders int64   `json:"orders"`
		Amount float64 `json:"amount"`
	}

	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var orders int64
		var amount float64

		database.DB.Model(&models.Order{}).Where("DATE(created_at) = ?", date).Count(&orders)
		database.DB.Model(&models.Order{}).Where("DATE(created_at) = ? AND status >= 2", date).Select("COALESCE(SUM(pay_amount), 0)").Scan(&amount)

		trends = append(trends, struct {
			Date   string  `json:"date"`
			Orders int64   `json:"orders"`
			Amount float64 `json:"amount"`
		}{Date: date, Orders: orders, Amount: amount})
	}

	var totalOrders int64
	var totalAmount float64
	database.DB.Model(&models.Order{}).Count(&totalOrders)
	database.DB.Model(&models.Order{}).Where("status >= 2").Select("COALESCE(SUM(pay_amount), 0)").Scan(&totalAmount)

	response.Success(c, gin.H{
		"trends":       trends,
		"total_orders": totalOrders,
		"total_amount": totalAmount,
	})
}

func GetAmountAnalytics(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	daysInt, _ := strconv.Atoi(days)

	var trends []struct {
		Date   string  `json:"date"`
		Amount float64 `json:"amount"`
	}

	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var amount float64
		database.DB.Model(&models.Order{}).Where("DATE(created_at) = ? AND status >= 2", date).Select("COALESCE(SUM(pay_amount), 0)").Scan(&amount)
		trends = append(trends, struct {
			Date   string  `json:"date"`
			Amount float64 `json:"amount"`
		}{Date: date, Amount: amount})
	}

	response.Success(c, gin.H{
		"trends": trends,
	})
}

func GetTopMerchants(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	limitInt, _ := strconv.Atoi(limit)

	var topMerchants []struct {
		MerchantID   uint64  `json:"merchant_id"`
		MerchantName string  `json:"merchant_name"`
		TotalAmount  float64 `json:"total_amount"`
		OrderCount   int64   `json:"order_count"`
	}

	database.DB.Table("orders").
		Select("orders.merchant_id, merchants.name as merchant_name, SUM(orders.pay_amount) as total_amount, COUNT(*) as order_count").
		Joins("JOIN merchants ON merchants.id = orders.merchant_id").
		Where("orders.status >= 2").
		Group("orders.merchant_id").
		Order("total_amount DESC").
		Limit(limitInt).
		Scan(&topMerchants)

	response.Success(c, topMerchants)
}

func GetAuditRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	database.DB.Model(&models.MerchantAuditRecord{}).Count(&total)

	var records []models.MerchantAuditRecord
	offset := (page - 1) * pageSize
	if err := database.DB.Preload("Merchant").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取审核记录失败")
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

func GetMerchantFee(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var fees []models.MerchantFee
	database.DB.Where("merchant_id = ?", id).Order("year DESC").Find(&fees)

	response.Success(c, gin.H{
		"merchant_id": id,
		"fees":        fees,
	})
}

func GetMerchantRate(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var rates []models.MerchantRate
	database.DB.Where("merchant_id = ?", id).Order("effective_time DESC").Find(&rates)

	response.Success(c, gin.H{
		"merchant_id": id,
		"rates":       rates,
	})
}

type SetRateRequest struct {
	Rate          float64 `json:"rate" binding:"required"`
	EffectiveTime string  `json:"effective_time"`
	ExpireTime    string  `json:"expire_time"`
	Remark        string  `json:"remark"`
}

func SetMerchantRate(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var req SetRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	effectiveTime, _ := time.Parse("2006-01-02 15:04:05", req.EffectiveTime)
	var expireTime *time.Time
	if req.ExpireTime != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.ExpireTime)
		expireTime = &t
	}

	rate := models.MerchantRate{
		MerchantID:    id,
		RateType:      "promotion",
		Rate:          req.Rate,
		EffectiveTime: effectiveTime,
		ExpireTime:    expireTime,
		Remark:        req.Remark,
	}

	if err := database.DB.Create(&rate).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置费率失败")
		return
	}

	response.Success(c, gin.H{"message": "设置成功"})
}

func GetMerchantQRCode(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.Select("id", "name").First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	qrCodeURL := "/qrcode/" + strconv.FormatUint(id, 10)

	response.Success(c, gin.H{
		"merchant_id":   id,
		"merchant_name": merchant.Name,
		"qrcode_url":    qrCodeURL,
		"page_path":     "pages/shop/index?merchant_id=" + strconv.FormatUint(id, 10),
	})
}

func GetRefunds(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	merchantID := c.Query("merchant_id")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Refund{}).Preload("Order")

	if merchantID != "" {
		id, _ := strconv.ParseUint(merchantID, 10, 64)
		query = query.Joins("JOIN orders ON orders.id = refunds.order_id").Where("orders.merchant_id = ?", id)
	}
	if status != "" {
		query = query.Where("refunds.status = ?", status)
	}

	var total int64
	query.Count(&total)

	var refunds []models.Refund
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&refunds).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取退款列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": refunds,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetSettings(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}
	adminID, ok := userIDValue.(uint64)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.Preload("ServiceProvider").First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}
	if admin.ServiceProvider == nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}

	adminName := admin.Name
	if adminName == "" {
		adminName = admin.Username
	}

	response.Success(c, gin.H{
		"service_provider_id": admin.ServiceProvider.ID,
		"name":                admin.ServiceProvider.Name,
		"admin_name":          adminName,
		"contact_phone":       admin.ServiceProvider.ContactPhone,
		"contact_email":       "",
		"created_at":          admin.ServiceProvider.CreatedAt,
	})
}

func UpdateSettings(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}
	adminID, ok := userIDValue.(uint64)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.Preload("ServiceProvider").First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}
	if admin.ServiceProvider == nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}
	sp := admin.ServiceProvider

	var req struct {
		ContactName  string `json:"contact_name"`
		ContactPhone string `json:"contact_phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.ContactName != "" {
		updates["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		updates["contact_phone"] = req.ContactPhone
	}

	if len(updates) > 0 {
		database.DB.Model(sp).Updates(updates)
	}

	response.Success(c, gin.H{"message": "设置成功"})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangePassword(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}
	adminID, ok := userIDValue.(uint64)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "旧密码错误")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	if err := database.DB.Model(&admin).Update("password", string(hashed)).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改失败")
		return
	}

	response.Success(c, gin.H{"message": "修改成功"})
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
			"total":     total,
			"page":      page,
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
		"status":      1,
		"submit_time": now,
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

func GetActivities(c *gin.Context) {
	var banners []models.Activity
	var announcements []models.Activity

	database.DB.Where("type = ? AND status = ?", "banner", 1).Order("sort ASC, created_at DESC").Find(&banners)
	database.DB.Where("type = ? AND status = ?", "announcement", 1).Order("sort ASC, created_at DESC").Find(&announcements)

	response.Success(c, gin.H{
		"banners":       banners,
		"announcements": announcements,
	})
}

type CreateActivityRequest struct {
	Type      string `json:"type" binding:"required,oneof=banner announcement"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Image     string `json:"image"`
	LinkType  string `json:"link_type" binding:"omitempty,oneof=merchant webview none"`
	LinkValue string `json:"link_value"`
	Sort      uint   `json:"sort"`
	Status    uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

func CreateActivity(c *gin.Context) {
	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	activity := models.Activity{
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Image:     req.Image,
		LinkType:  req.LinkType,
		LinkValue: req.LinkValue,
		Sort:      req.Sort,
		Status:    1,
	}
	if req.Status > 0 {
		activity.Status = req.Status
	}

	if err := database.DB.Create(&activity).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建活动失败")
		return
	}

	response.Success(c, gin.H{"id": activity.ID, "message": "创建成功"})
}

func UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	activityID, _ := strconv.ParseUint(id, 10, 64)

	var req CreateActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var activity models.Activity
	if err := database.DB.First(&activity, activityID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "活动不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Image != "" {
		updates["image"] = req.Image
	}
	if req.LinkType != "" {
		updates["link_type"] = req.LinkType
	}
	if req.LinkValue != "" {
		updates["link_value"] = req.LinkValue
	}
	if req.Sort > 0 {
		updates["sort"] = req.Sort
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	if err := database.DB.Model(&activity).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新活动失败")
		return
	}

	database.DB.First(&activity, activityID)
	response.Success(c, activity)
}

func DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	activityID, _ := strconv.ParseUint(id, 10, 64)

	var activity models.Activity
	if err := database.DB.First(&activity, activityID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "活动不存在")
		return
	}

	if err := database.DB.Model(&activity).Update("status", 0).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除活动失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

type WechatConfig struct {
	AppID          string            `json:"app_id"`
	AppSecret      string            `json:"app_secret"`
	Token          string            `json:"token"`
	EncodingAESKey string            `json:"encoding_aes_key"`
	TemplateIDs    map[string]string `json:"template_ids"`
	Enabled        bool              `json:"enabled"`
}

func GetWechatConfig(c *gin.Context) {
	var sp models.ServiceProvider
	if err := database.DB.First(&sp).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}

	config := WechatConfig{
		AppID:   sp.MchID,
		Enabled: true,
		TemplateIDs: map[string]string{
			"order_new":    "",
			"order_paid":   "",
			"order_refund": "",
		},
	}
	response.Success(c, config)
}

type UpdateWechatConfigRequest struct {
	AppID          string            `json:"app_id"`
	AppSecret      string            `json:"app_secret"`
	Token          string            `json:"token"`
	EncodingAESKey string            `json:"encoding_aes_key"`
	TemplateIDs    map[string]string `json:"template_ids"`
}

func UpdateWechatConfig(c *gin.Context) {
	var req UpdateWechatConfigRequest
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
	if req.AppID != "" {
		updates["mch_id"] = req.AppID
	}
	if req.AppSecret != "" {
		updates["api_key"] = req.AppSecret
	}

	if err := database.DB.Model(&sp).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新配置失败")
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}
