package sp

import (
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"sort"
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
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var totalMerchants int64
	var todayOrders int64
	var todayRevenue float64
	merchantIDsQuery := database.DB.Model(&models.Merchant{}).
		Where("service_provider_id = ?", serviceProviderID).
		Select("id")

	database.DB.Model(&models.Merchant{}).
		Where("service_provider_id = ?", serviceProviderID).
		Count(&totalMerchants)
	database.DB.Model(&models.Order{}).
		Where("merchant_id IN (?) AND DATE(created_at) = CURDATE()", merchantIDsQuery).
		Count(&todayOrders)
	database.DB.Model(&models.Order{}).
		Where("merchant_id IN (?) AND DATE(created_at) = CURDATE() AND status >= 2", merchantIDsQuery).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&todayRevenue)

	var distribution []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&models.Merchant{}).
		Where("service_provider_id = ?", serviceProviderID).
		Select("business_category as category, count(*) as count").
		Group("business_category").
		Scan(&distribution)

	var trend []struct {
		Date   string `json:"date"`
		Orders int64  `json:"orders"`
	}
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var orders int64
		database.DB.Model(&models.Order{}).
			Where("merchant_id IN (?) AND DATE(created_at) = ?", merchantIDsQuery, date).
			Count(&orders)
		trend = append(trend, struct {
			Date   string `json:"date"`
			Orders int64  `json:"orders"`
		}{Date: date, Orders: orders})
	}

	response.Success(c, gin.H{
		"total_merchants":   totalMerchants,
		"pending_merchants": 0,
		"today_orders":      todayOrders,
		"today_revenue":     todayRevenue,
		"distribution":      distribution,
		"trend":             trend,
	})
}

func GetMerchantDetail(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ? AND service_provider_id = ?", id, serviceProviderID).First(&merchant).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	var totalOrders int64
	database.DB.Model(&models.Order{}).Where("merchant_id = ?", id).Count(&totalOrders)

	var totalAmount float64
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2", id).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&totalAmount)

	response.Success(c, gin.H{
		"id":                     merchant.ID,
		"name":                   merchant.Name,
		"logo":                   merchant.Logo,
		"cover_image":            merchant.CoverImage,
		"contact_name":           merchant.ContactName,
		"contact_phone":          merchant.ContactPhone,
		"contact_email":          merchant.ContactEmail,
		"address":                merchant.Address,
		"business_category":      merchant.BusinessCategory,
		"business_hours":         merchant.BusinessHours,
		"announcement":           merchant.Announcement,
		"status":                 merchant.Status,
		"qrcode_url":             merchant.QRCodeURL,
		"created_at":             merchant.CreatedAt,
		"sub_mch_id":             merchant.SubMchID,
		"profit_sharing_enabled": merchant.ProfitSharingEnabled,
		"profit_sharing_ratio":   merchant.ProfitSharingRatio,
		"payment_config_status":  merchant.PaymentConfigStatus,
		"total_orders":           totalOrders,
		"total_amount":           totalAmount,
		"total_users":            0,
	})
}

type UpdateMerchantAssetsRequest struct {
	Logo       string `json:"logo"`
	CoverImage string `json:"cover_image"`
}

func UpdateMerchantAssets(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	var req UpdateMerchantAssetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if len(updates) == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "未提供可更新内容")
		return
	}

	if err := database.DB.Model(&models.Merchant{}).
		Where("id = ? AND service_provider_id = ?", id, serviceProviderID).
		Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商家图片失败")
		return
	}

	GetMerchantDetail(c)
}

func GetMerchantDistribution(c *gin.Context) {
	metrics, totals := buildSpMerchantMetrics()
	response.Success(c, gin.H{
		"merchants": metrics,
		"totals":    totals,
	})
}

func GetMerchantList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	serviceProviderID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Merchant{}).Where("service_provider_id = ?", serviceProviderID)

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
	location := time.Now().Location()
	now := time.Now().In(location)

	response.Success(c, gin.H{
		"day":   buildSpOrderBuckets(now, 7, func(base time.Time, offset int) (time.Time, time.Time, string) {
			start := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -(6 - offset))
			end := start.Add(24 * time.Hour)
			return start, end, start.Format("01-02")
		}),
		"week":  buildSpOrderBuckets(now, 8, func(base time.Time, offset int) (time.Time, time.Time, string) {
			weekdayOffset := (int(base.Weekday()) + 6) % 7
			weekStart := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -weekdayOffset)
			start := weekStart.AddDate(0, 0, -7*(7-offset))
			end := start.AddDate(0, 0, 7)
			year, week := start.ISOWeek()
			return start, end, fmt.Sprintf("%d-W%02d", year, week)
		}),
		"month": buildSpOrderBuckets(now, 12, func(base time.Time, offset int) (time.Time, time.Time, string) {
			start := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, location).AddDate(0, -(11 - offset), 0)
			end := start.AddDate(0, 1, 0)
			return start, end, start.Format("2006-01")
		}),
		"year":  buildSpOrderBuckets(now, 5, func(base time.Time, offset int) (time.Time, time.Time, string) {
			start := time.Date(base.Year()-(4-offset), 1, 1, 0, 0, 0, 0, location)
			end := start.AddDate(1, 0, 0)
			return start, end, start.Format("2006")
		}),
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
	if limitInt <= 0 {
		limitInt = 10
	}
	metric := c.DefaultQuery("metric", "order_amount")

	metrics, _ := buildSpMerchantMetrics()
	sort.Slice(metrics, func(i, j int) bool {
		left := getSpMerchantMetricValue(metrics[i], metric)
		right := getSpMerchantMetricValue(metrics[j], metric)
		if left == right {
			return metrics[i]["merchant_name"].(string) < metrics[j]["merchant_name"].(string)
		}
		return left > right
	})

	if len(metrics) > limitInt {
		metrics = metrics[:limitInt]
	}

	list := make([]gin.H, 0, len(metrics))
	for index, item := range metrics {
		list = append(list, gin.H{
			"rank":             index + 1,
			"metric":           metric,
			"merchant_id":      item["merchant_id"],
			"merchant_name":    item["merchant_name"],
			"merchant_logo":    item["merchant_logo"],
			"visit_rate":       item["visit_rate"],
			"order_rate":       item["order_rate"],
			"order_amount":     item["order_amount"],
			"avg_order_amount": item["avg_order_amount"],
			"visit_users":      item["visit_users"],
			"order_users":      item["order_users"],
			"paid_orders":      item["paid_orders"],
		})
	}

	response.Success(c, list)
}

func buildSpMerchantMetrics() ([]gin.H, gin.H) {
	var merchants []models.Merchant
	database.DB.Order("created_at DESC").Find(&merchants)

	totalVisitUsers := int64(0)
	totalOrderUsers := int64(0)
	totalPaidOrders := int64(0)
	totalOrderAmount := 0.0

	metrics := make([]gin.H, 0, len(merchants))
	for _, merchant := range merchants {
		var visitUsers int64
		var orderUsers int64
		var paidOrders int64
		var orderAmount float64

		database.DB.Model(&models.UserVisit{}).
			Where("merchant_id = ?", merchant.ID).
			Distinct("user_id").
			Count(&visitUsers)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status >= 2", merchant.ID).
			Distinct("user_id").
			Count(&orderUsers)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status >= 2", merchant.ID).
			Count(&paidOrders)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status >= 2", merchant.ID).
			Select("COALESCE(SUM(pay_amount), 0)").
			Scan(&orderAmount)

		totalVisitUsers += visitUsers
		totalOrderUsers += orderUsers
		totalPaidOrders += paidOrders
		totalOrderAmount += orderAmount

		avgOrderAmount := 0.0
		if paidOrders > 0 {
			avgOrderAmount = orderAmount / float64(paidOrders)
		}

		metrics = append(metrics, gin.H{
			"merchant_id":      merchant.ID,
			"merchant_name":    merchant.Name,
			"merchant_logo":    merchant.Logo,
			"visit_users":      visitUsers,
			"order_users":      orderUsers,
			"paid_orders":      paidOrders,
			"order_amount":     orderAmount,
			"avg_order_amount": avgOrderAmount,
			"visit_rate":       0.0,
			"order_rate":       0.0,
		})
	}

	for _, item := range metrics {
		visitUsers := item["visit_users"].(int64)
		orderUsers := item["order_users"].(int64)
		if totalVisitUsers > 0 {
			item["visit_rate"] = float64(visitUsers) / float64(totalVisitUsers) * 100
		}
		if visitUsers > 0 {
			item["order_rate"] = float64(orderUsers) / float64(visitUsers) * 100
		}
	}

	return metrics, gin.H{
		"merchant_count": totalCountInt64(len(metrics)),
		"visit_users":    totalVisitUsers,
		"order_users":    totalOrderUsers,
		"paid_orders":    totalPaidOrders,
		"order_amount":   totalOrderAmount,
	}
}

func totalCountInt64(value int) int64 {
	return int64(value)
}

func getSpMerchantMetricValue(item gin.H, metric string) float64 {
	switch metric {
	case "visit_rate":
		return item["visit_rate"].(float64)
	case "order_rate":
		return item["order_rate"].(float64)
	case "avg_order_amount":
		return item["avg_order_amount"].(float64)
	default:
		return item["order_amount"].(float64)
	}
}

func buildSpOrderBuckets(
	base time.Time,
	count int,
	rangeBuilder func(base time.Time, offset int) (time.Time, time.Time, string),
) []gin.H {
	result := make([]gin.H, 0, count)
	for index := 0; index < count; index++ {
		start, end, label := rangeBuilder(base, index)
		var orderCount int64
		database.DB.Model(&models.Order{}).
			Where("status >= 2 AND created_at >= ? AND created_at < ?", start, end).
			Count(&orderCount)
		result = append(result, gin.H{
			"label":       label,
			"order_count": orderCount,
		})
	}
	return result
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
