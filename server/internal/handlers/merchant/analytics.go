package merchant

import (
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserStats struct {
	TotalUsers      int64   `json:"total_users"`
	NewUsersToday   int64   `json:"new_users_today"`
	NewUsersWeek    int64   `json:"new_users_week"`
	NewUsersMonth   int64   `json:"new_users_month"`
	PaidUsers       int64   `json:"paid_users"`
	UnpaidUsers     int64   `json:"unpaid_users"`
	TotalVisitCount int64   `json:"total_visit_count"`
	AvgVisitPerUser float64 `json:"avg_visit_per_user"`
	ActiveUsers     int64   `json:"active_users"`
}

func GetUserStats(c *gin.Context) {
	merchantID := c.Query("merchant_id")
	mid, _ := strconv.ParseUint(merchantID, 10, 64)

	var stats UserStats

	today := time.Now().Format("2006-01-02")
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	dayAgo := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 总用户数（按openid去重，有订单的用户）
	var distinctUsers []uint64
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ?", mid).
		Distinct("user_id").
		Pluck("user_id", &distinctUsers)
	stats.TotalUsers = int64(len(distinctUsers))

	// 今日新增用户（有订单且首次访问在今天）
	var todayNewUsers int64
	database.DB.Model(&models.User{}).
		Where("id IN ? AND DATE(first_visit_at) = ?", distinctUsers, today).
		Count(&todayNewUsers)
	stats.NewUsersToday = todayNewUsers

	// 本周新增用户
	var weekNewUsers int64
	database.DB.Model(&models.User{}).
		Where("id IN ? AND first_visit_at >= ?", distinctUsers, weekAgo).
		Count(&weekNewUsers)
	stats.NewUsersWeek = weekNewUsers

	// 本月新增用户
	var monthNewUsers int64
	database.DB.Model(&models.User{}).
		Where("id IN ? AND first_visit_at >= ?", distinctUsers, monthAgo).
		Count(&monthNewUsers)
	stats.NewUsersMonth = monthNewUsers

	// 已支付用户数
	var paidUsers int64
	database.DB.Model(&models.User{}).
		Where("id IN ? AND has_paid = ?", distinctUsers, true).
		Count(&paidUsers)
	stats.PaidUsers = paidUsers

	// 未支付用户数
	stats.UnpaidUsers = stats.TotalUsers - stats.PaidUsers

	// 总访问次数
	database.DB.Model(&models.UserVisit{}).
		Where("merchant_id = ?", mid).
		Count(&stats.TotalVisitCount)

	// 人均访问次数
	if stats.TotalUsers > 0 {
		stats.AvgVisitPerUser = float64(stats.TotalVisitCount) / float64(stats.TotalUsers)
	}

	// 活跃用户（7天内有访问）
	var activeUsers int64
	database.DB.Model(&models.UserVisit{}).
		Where("merchant_id = ? AND visit_time >= ?", mid, dayAgo).
		Distinct("user_id").
		Count(&activeUsers)
	stats.ActiveUsers = activeUsers

	response.Success(c, stats)
}

type MerchantOverview struct {
	MerchantID   uint64  `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	TotalUsers   int64   `json:"total_users"`
	TotalOrders  int64   `json:"total_orders"`
	TotalSales   float64 `json:"total_sales"`
	TodayOrders  int64   `json:"today_orders"`
	TodaySales   float64 `json:"today_sales"`
}

func GetMerchantsOverview(c *gin.Context) {
	var overviews []MerchantOverview

	today := time.Now().Format("2006-01-02")

	var merchants []models.Merchant
	database.DB.Find(&merchants)

	for _, merchant := range merchants {
		var distinctUsers []uint64
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ?", merchant.ID).
			Distinct("user_id").
			Pluck("user_id", &distinctUsers)

		var totalOrders, todayOrders int64
		var totalSales, todaySales float64

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status IN (2, 3)", merchant.ID).
			Count(&totalOrders)

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status IN (2, 3)", merchant.ID).
			Select("COALESCE(SUM(pay_amount), 0)").
			Scan(&totalSales)

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status IN (2, 3) AND DATE(created_at) = ?", merchant.ID, today).
			Count(&todayOrders)

		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND status IN (2, 3) AND DATE(created_at) = ?", merchant.ID, today).
			Select("COALESCE(SUM(pay_amount), 0)").
			Scan(&todaySales)

		overviews = append(overviews, MerchantOverview{
			MerchantID:   merchant.ID,
			MerchantName: merchant.Name,
			TotalUsers:   int64(len(distinctUsers)),
			TotalOrders:  totalOrders,
			TotalSales:   totalSales,
			TodayOrders:  todayOrders,
			TodaySales:   todaySales,
		})
	}

	response.Success(c, overviews)
}
