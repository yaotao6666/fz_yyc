package merchant

import (
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CompleteOrderRequest struct {
	VerifyCode string `json:"verify_code" binding:"required"`
}

func GetOrders(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	deliveryType := c.Query("delivery_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Order{}).Where("merchant_id = ?", merchantID)

	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}
	if deliveryType != "" {
		deliveryTypeInt, _ := strconv.Atoi(deliveryType)
		query = query.Where("delivery_type = ?", deliveryTypeInt)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		endDateTime, _ := time.Parse("2006-01-02", endDate)
		query = query.Where("created_at <= ?", endDateTime.Add(24*time.Hour))
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取订单列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": orders,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetOrderDetail(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var order models.Order
	if err := database.DB.Preload("User").Preload("Items").Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	response.Success(c, order)
}

func CompleteOrder(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var req CompleteOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	verifyCode := strings.TrimSpace(req.VerifyCode)
	if len(verifyCode) != 6 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "核销码应为6位数字")
		return
	}
	for _, r := range verifyCode {
		if r < '0' || r > '9' {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "核销码应为6位数字")
			return
		}
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	if order.VerifyCode != verifyCode {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "核销码错误")
		return
	}

	if order.Status != 2 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单状态不正确")
		return
	}

	now := time.Now()
	if err := database.DB.Model(&order).Updates(map[string]interface{}{
		"status":       3,
		"completed_at": now,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "完成订单失败")
		return
	}

	database.DB.Preload("User").Preload("Items").First(&order, id)
	response.Success(c, order)
}

type RefundRequest struct {
	RefundAmount float64 `json:"refund_amount" binding:"required"`
	Reason       string  `json:"reason"`
}

func RefundOrder(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	if order.Status != 2 && order.Status != 3 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单状态不正确")
		return
	}

	refundNo := strconv.FormatInt(time.Now().UnixNano(), 10)
	refund := models.Refund{
		OrderID:      id,
		RefundNo:     refundNo,
		RefundAmount: req.RefundAmount,
		RefundReason: req.Reason,
		Status:       0,
	}

	now := time.Now()
	tx := database.DB.Begin()
	if err := tx.Create(&refund).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建退款记录失败")
		return
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      5,
		"refunded_at": now,
	}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新订单状态失败")
		return
	}

	tx.Commit()

	response.Success(c, refund)
}

func GetOrderStatistics(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var totalOrders int64
	var totalAmount float64
	var todayOrders int64
	var todayAmount float64
	var pendingOrders int64
	var completedOrders int64
	var refundedAmount float64

	database.DB.Model(&models.Order{}).Where("merchant_id = ?", merchantID).Count(&totalOrders)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND status >= 2", merchantID).Select("COALESCE(SUM(pay_amount), 0)").Scan(&totalAmount)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE()", merchantID).Count(&todayOrders)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID).Select("COALESCE(SUM(pay_amount), 0)").Scan(&todayAmount)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND status = 1", merchantID).Count(&pendingOrders)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND status = 3", merchantID).Count(&completedOrders)
	database.DB.Model(&models.Refund{}).Joins("JOIN orders ON orders.id = refunds.order_id").Where("orders.merchant_id = ? AND refunds.status = 2", merchantID).Select("COALESCE(SUM(refund_amount), 0)").Scan(&refundedAmount)

	response.Success(c, gin.H{
		"total_orders":     totalOrders,
		"total_amount":     totalAmount,
		"today_orders":     todayOrders,
		"today_amount":     todayAmount,
		"pending_orders":   pendingOrders,
		"completed_orders": completedOrders,
		"refunded_amount":  refundedAmount,
	})
}

func GetAnalyticsOverview(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	period := c.DefaultQuery("period", "today")
	now := time.Now()
	location := now.Location()

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	end := start.Add(24 * time.Hour)
	prevStart := start.Add(-24 * time.Hour)
	prevEnd := start

	switch period {
	case "week":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -6)
		end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).Add(24 * time.Hour)
		prevStart = start.AddDate(0, 0, -7)
		prevEnd = start
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
		end = start.AddDate(0, 1, 0)
		prevStart = start.AddDate(0, -1, 0)
		prevEnd = start
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, location)
		end = start.AddDate(1, 0, 0)
		prevStart = start.AddDate(-1, 0, 0)
		prevEnd = start
	}

	var totalSales float64
	var totalOrders int64
	var totalCustomers int64
	var prevSales float64
	var prevOrders int64
	var prevCustomers int64
	var visitCount int64
	var visitUsers int64
	var paySuccessUsers int64

	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, start, end).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&totalSales)
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, start, end).
		Count(&totalOrders)
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, start, end).
		Distinct("user_id").
		Count(&totalCustomers)

	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, prevStart, prevEnd).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&prevSales)
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, prevStart, prevEnd).
		Count(&prevOrders)
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2 AND created_at >= ? AND created_at < ?", merchantID, prevStart, prevEnd).
		Distinct("user_id").
		Count(&prevCustomers)

	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ? AND created_at >= ? AND created_at < ?", merchantID, "store_visit", start, end).
		Count(&visitCount)
	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ? AND created_at >= ? AND created_at < ?", merchantID, "store_visit", start, end).
		Distinct("user_id").
		Count(&visitUsers)
	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ? AND created_at >= ? AND created_at < ?", merchantID, "pay_success", start, end).
		Distinct("user_id").
		Count(&paySuccessUsers)

	avgOrderAmount := 0.0
	if totalOrders > 0 {
		avgOrderAmount = totalSales / float64(totalOrders)
	}

	salesGrowth := 0.0
	if prevSales > 0 {
		salesGrowth = (totalSales - prevSales) / prevSales * 100
	}

	ordersGrowth := 0.0
	if prevOrders > 0 {
		ordersGrowth = float64(totalOrders-prevOrders) / float64(prevOrders) * 100
	}

	customersGrowth := 0.0
	if prevCustomers > 0 {
		customersGrowth = float64(totalCustomers-prevCustomers) / float64(prevCustomers) * 100
	}

	response.Success(c, gin.H{
		"total_sales":       totalSales,
		"total_orders":      totalOrders,
		"total_customers":   totalCustomers,
		"avg_order_amount":  avgOrderAmount,
		"sales_growth":      salesGrowth,
		"orders_growth":     ordersGrowth,
		"customers_growth":  customersGrowth,
		"visit_count":       visitCount,
		"visit_users":       visitUsers,
		"pay_success_users": paySuccessUsers,
	})
}

func GetSalesTrend(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	days := c.DefaultQuery("days", "7")
	daysInt, _ := strconv.Atoi(days)
	location := time.Now().Location()

	var start time.Time
	var end time.Time
	if startDate != "" {
		parsedStart, err := time.ParseInLocation("2006-01-02", startDate, location)
		if err == nil {
			start = parsedStart
		}
	}
	if endDate != "" {
		parsedEnd, err := time.ParseInLocation("2006-01-02", endDate, location)
		if err == nil {
			end = parsedEnd.Add(24 * time.Hour)
		}
	}
	if start.IsZero() {
		start = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, location).AddDate(0, 0, -(daysInt - 1))
	}
	if end.IsZero() {
		end = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, location).Add(24 * time.Hour)
	}

	var trends []struct {
		Date             string  `json:"date"`
		Orders           int64   `json:"orders"`
		Sales            float64 `json:"sales"`
		Customers        int64   `json:"customers"`
		VisitUsers       int64   `json:"visit_users"`
		SubmitOrderUsers int64   `json:"submit_order_users"`
	}

	for cursor := start; cursor.Before(end); cursor = cursor.Add(24 * time.Hour) {
		date := cursor.Format("2006-01-02")
		var orders int64
		var sales float64
		var customers int64
		var visitUsers int64
		var submitOrderUsers int64

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ? AND status >= 2", merchantID, date).
			Count(&orders)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ? AND status >= 2", merchantID, date).
			Select("COALESCE(SUM(pay_amount), 0)").Scan(&sales)
		database.DB.Model(&models.UserBehaviorEvent{}).
			Where("merchant_id = ? AND event_type = ? AND DATE(created_at) = ?", merchantID, "store_visit", date).
			Distinct("user_id").
			Count(&customers)
		database.DB.Model(&models.UserBehaviorEvent{}).
			Where("merchant_id = ? AND event_type = ? AND DATE(created_at) = ?", merchantID, "store_visit", date).
			Distinct("user_id").
			Count(&visitUsers)
		database.DB.Model(&models.UserBehaviorEvent{}).
			Where("merchant_id = ? AND event_type = ? AND DATE(created_at) = ?", merchantID, "submit_order", date).
			Distinct("user_id").
			Count(&submitOrderUsers)

		trends = append(trends, struct {
			Date             string  `json:"date"`
			Orders           int64   `json:"orders"`
			Sales            float64 `json:"sales"`
			Customers        int64   `json:"customers"`
			VisitUsers       int64   `json:"visit_users"`
			SubmitOrderUsers int64   `json:"submit_order_users"`
		}{
			Date:             date,
			Orders:           orders,
			Sales:            sales,
			Customers:        customers,
			VisitUsers:       visitUsers,
			SubmitOrderUsers: submitOrderUsers,
		})
	}

	response.Success(c, trends)
}

func GetProductRanking(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	limit := c.DefaultQuery("limit", "10")
	limitInt, _ := strconv.Atoi(limit)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var rankings []struct {
		ProductID   uint64  `json:"product_id"`
		ProductName string  `json:"product_name"`
		Image       string  `json:"image"`
		SalesCount  uint    `json:"sales_count"`
		SalesAmount float64 `json:"sales_amount"`
	}

	query := database.DB.Table("order_items").
		Select("order_items.product_id, order_items.product_name, order_items.image, SUM(order_items.quantity) as sales_count, SUM(order_items.subtotal) as sales_amount").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.merchant_id = ? AND orders.status >= 2", merchantID)

	if startDate != "" {
		query = query.Where("orders.created_at >= ?", startDate)
	}
	if endDate != "" {
		endDateTime, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			query = query.Where("orders.created_at < ?", endDateTime.Add(24*time.Hour))
		}
	}

	query.
		Group("order_items.product_id").
		Order("sales_count DESC").
		Limit(limitInt).
		Scan(&rankings)

	response.Success(c, rankings)
}

func GetHourlyAnalysis(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var hourlyData []struct {
		Hour   int     `json:"hour"`
		Orders int64   `json:"orders"`
		Sales  float64 `json:"sales"`
	}

	for h := 0; h < 24; h++ {
		var orders int64
		var sales float64

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND HOUR(created_at) = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID, h).
			Count(&orders)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND HOUR(created_at) = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID, h).
			Select("COALESCE(SUM(pay_amount), 0)").Scan(&sales)

		hourlyData = append(hourlyData, struct {
			Hour   int     `json:"hour"`
			Orders int64   `json:"orders"`
			Sales  float64 `json:"sales"`
		}{
			Hour:   h,
			Orders: orders,
			Sales:  sales,
		})
	}

	response.Success(c, hourlyData)
}

func GetStockAlert(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	threshold := c.DefaultQuery("threshold", "10")
	thresholdInt, _ := strconv.Atoi(threshold)

	var products []models.Product
	if err := database.DB.Where("merchant_id = ? AND stock <= ? AND status = 1", merchantID, thresholdInt).Order("stock ASC").Find(&products).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取库存警告失败")
		return
	}

	response.Success(c, products)
}

func GetCustomerAnalysis(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var totalCustomers int64
	var newCustomers int64
	var repeatRate float64
	var visitUsers int64
	var visitCount int64
	var submitOrderUsers int64
	var paySuccessUsers int64

	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status >= 2", merchantID).
		Select("COUNT(DISTINCT user_id)").Scan(&totalCustomers)

	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ? AND DATE(created_at) = CURDATE()", merchantID, "store_visit").
		Distinct("user_id").
		Count(&newCustomers)

	var totalOrders int64
	var repeatOrders int64
	database.DB.Model(&models.Order{}).Where("merchant_id = ?", merchantID).Count(&totalOrders)
	database.DB.Model(&models.Order{}).
		Select("COUNT(*) FROM (SELECT user_id FROM orders WHERE merchant_id = ? GROUP BY user_id HAVING COUNT(*) > 1) as t", merchantID).
		Scan(&repeatOrders)

	if totalOrders > 0 {
		repeatRate = float64(repeatOrders) / float64(totalCustomers) * 100
	}

	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ?", merchantID, "store_visit").
		Distinct("user_id").
		Count(&visitUsers)
	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ?", merchantID, "store_visit").
		Count(&visitCount)
	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ?", merchantID, "submit_order").
		Distinct("user_id").
		Count(&submitOrderUsers)
	database.DB.Model(&models.UserBehaviorEvent{}).
		Where("merchant_id = ? AND event_type = ?", merchantID, "pay_success").
		Distinct("user_id").
		Count(&paySuccessUsers)

	response.Success(c, gin.H{
		"total_customers":    totalCustomers,
		"new_customers":      newCustomers,
		"repeat_rate":        repeatRate,
		"visit_users":        visitUsers,
		"visit_count":        visitCount,
		"submit_order_users": submitOrderUsers,
		"pay_success_users":  paySuccessUsers,
	})
}

func GetCustomerTrend(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	days := c.DefaultQuery("days", "7")
	daysInt, _ := strconv.Atoi(days)

	var trends []struct {
		Date       string `json:"date"`
		TotalUsers int64  `json:"total_users"`
		NewUsers   int64  `json:"new_users"`
		OrderCount int64  `json:"order_count"`
	}

	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		var totalUsers int64
		var newUsers int64
		var orderCount int64

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) <= ?", merchantID, date).
			Select("COUNT(DISTINCT user_id)").Scan(&totalUsers)

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ?", merchantID, date).
			Select("COUNT(DISTINCT user_id)").Scan(&newUsers)

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ?", merchantID, date).
			Count(&orderCount)

		trends = append(trends, struct {
			Date       string `json:"date"`
			TotalUsers int64  `json:"total_users"`
			NewUsers   int64  `json:"new_users"`
			OrderCount int64  `json:"order_count"`
		}{
			Date:       date,
			TotalUsers: totalUsers,
			NewUsers:   newUsers,
			OrderCount: orderCount,
		})
	}

	response.Success(c, trends)
}
