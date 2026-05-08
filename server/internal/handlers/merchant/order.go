package merchant

import (
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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
	if err := query.Preload("User").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "获取订单列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": orders,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
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
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	response.Success(c, order)
}

func CompleteOrder(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var order models.Order
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	if order.Status != 2 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单状态不正确")
		return
	}

	now := time.Now()
	if err := database.DB.Model(&order).Updates(map[string]interface{}{
		"status":       3,
		"completed_at": now,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "完成订单失败")
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
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	if order.Status != 2 && order.Status != 3 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单状态不正确")
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
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "创建退款记录失败")
		return
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      5,
		"refunded_at": now,
	}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新订单状态失败")
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
		"total_amount":      totalAmount,
		"today_orders":     todayOrders,
		"today_amount":     todayAmount,
		"pending_orders":   pendingOrders,
		"completed_orders": completedOrders,
		"refunded_amount": refundedAmount,
	})
}

func GetAnalyticsOverview(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var todayOrders int64
	var todayAmount float64
	var yesterdayOrders int64
	var yesterdayAmount float64
	var todayNewUsers int64
	var totalProducts int64
	var outOfStock int64

	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE()", merchantID).Count(&todayOrders)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID).Select("COALESCE(SUM(pay_amount), 0)").Scan(&todayAmount)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY)", merchantID).Count(&yesterdayOrders)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND status >= 2", merchantID).Select("COALESCE(SUM(pay_amount), 0)").Scan(&yesterdayAmount)
	database.DB.Model(&models.Order{}).Where("merchant_id = ? AND DATE(created_at) = CURDATE()", merchantID).Select("COUNT(DISTINCT user_id)").Scan(&todayNewUsers)
	database.DB.Model(&models.Product{}).Where("merchant_id = ?", merchantID).Count(&totalProducts)
	database.DB.Model(&models.Product{}).Where("merchant_id = ? AND stock = 0", merchantID).Count(&outOfStock)

	var orderChange float64
	if yesterdayOrders > 0 {
		orderChange = float64(todayOrders-yesterdayOrders) / float64(yesterdayOrders) * 100
	}

	var amountChange float64
	if yesterdayAmount > 0 {
		amountChange = (todayAmount - yesterdayAmount) / yesterdayAmount * 100
	}

	response.Success(c, gin.H{
		"today_orders":     todayOrders,
		"today_amount":     todayAmount,
		"order_change":     orderChange,
		"amount_change":    amountChange,
		"today_new_users":  todayNewUsers,
		"total_products":   totalProducts,
		"out_of_stock":     outOfStock,
	})
}

func GetSalesTrend(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
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

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ? AND status >= 2", merchantID, date).
			Count(&orders)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND DATE(created_at) = ? AND status >= 2", merchantID, date).
			Select("COALESCE(SUM(pay_amount), 0)").Scan(&amount)

		trends = append(trends, struct {
			Date   string  `json:"date"`
			Orders int64   `json:"orders"`
			Amount float64 `json:"amount"`
		}{
			Date:   date,
			Orders: orders,
			Amount: amount,
		})
	}

	response.Success(c, trends)
}

func GetProductRanking(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	limit := c.DefaultQuery("limit", "10")
	limitInt, _ := strconv.Atoi(limit)

	var rankings []struct {
		ProductID   uint64  `json:"product_id"`
		ProductName string  `json:"product_name"`
		Image       string  `json:"image"`
		TotalSales  uint    `json:"total_sales"`
		TotalAmount float64 `json:"total_amount"`
	}

	database.DB.Table("order_items").
		Select("order_items.product_id, order_items.product_name, order_items.image, SUM(order_items.quantity) as total_sales, SUM(order_items.subtotal) as total_amount").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.merchant_id = ? AND orders.status >= 2", merchantID).
		Group("order_items.product_id").
		Order("total_sales DESC").
		Limit(limitInt).
		Scan(&rankings)

	response.Success(c, rankings)
}

func GetHourlyAnalysis(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var hourlyData []struct {
		Hour   int   `json:"hour"`
		Orders int64 `json:"orders"`
		Amount float64 `json:"amount"`
	}

	for h := 0; h < 24; h++ {
		var orders int64
		var amount float64

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND HOUR(created_at) = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID, h).
			Count(&orders)
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND HOUR(created_at) = ? AND DATE(created_at) = CURDATE() AND status >= 2", merchantID, h).
			Select("COALESCE(SUM(pay_amount), 0)").Scan(&amount)

		hourlyData = append(hourlyData, struct {
			Hour   int    `json:"hour"`
			Orders int64  `json:"orders"`
			Amount float64 `json:"amount"`
		}{
			Hour:   h,
			Orders: orders,
			Amount: amount,
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
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "获取库存警告失败")
		return
	}

	response.Success(c, products)
}

func GetCustomerAnalysis(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var totalCustomers int64
	var newCustomers int64
	var repeatRate float64

	database.DB.Model(&models.Order{}).
		Where("merchant_id = ?", merchantID).
		Select("COUNT(DISTINCT user_id)").Scan(&totalCustomers)

	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND DATE(created_at) = CURDATE()", merchantID).
		Select("COUNT(DISTINCT user_id)").Scan(&newCustomers)

	var totalOrders int64
	var repeatOrders int64
	database.DB.Model(&models.Order{}).Where("merchant_id = ?", merchantID).Count(&totalOrders)
	database.DB.Model(&models.Order{}).
		Select("COUNT(*) FROM (SELECT user_id FROM orders WHERE merchant_id = ? GROUP BY user_id HAVING COUNT(*) > 1) as t", merchantID).
		Scan(&repeatOrders)

	if totalOrders > 0 {
		repeatRate = float64(repeatOrders) / float64(totalCustomers) * 100
	}

	response.Success(c, gin.H{
		"total_customers": totalCustomers,
		"new_customers":   newCustomers,
		"repeat_rate":      repeatRate,
	})
}

func GetCustomerTrend(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	days := c.DefaultQuery("days", "7")
	daysInt, _ := strconv.Atoi(days)

	var trends []struct {
		Date         string `json:"date"`
		TotalUsers   int64  `json:"total_users"`
		NewUsers     int64  `json:"new_users"`
		OrderCount   int64  `json:"order_count"`
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
			Date         string `json:"date"`
			TotalUsers   int64  `json:"total_users"`
			NewUsers     int64  `json:"new_users"`
			OrderCount   int64  `json:"order_count"`
		}{
			Date:       date,
			TotalUsers: totalUsers,
			NewUsers:   newUsers,
			OrderCount: orderCount,
		})
	}

	response.Success(c, trends)
}
