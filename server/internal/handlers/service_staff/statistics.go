package service_staff

import (
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetStatistics 接单统计
func GetStatistics(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, 401, response.CodeUnauthorized, "获取信息失败")
		return
	}

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	var todayCount, totalCount int64
	var todayAmount, totalAmount float64

	// 今日接单数
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND actual_started_at >= ? AND actual_started_at < ?",
			staff.ID, today, tomorrow).Count(&todayCount)

	// 总接单数
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ?", staff.ID).Count(&totalCount)

	// 今日金额
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND actual_started_at >= ? AND actual_started_at < ?",
			staff.ID, today, tomorrow).
		Select("COALESCE(SUM(pay_amount), 0)").Scan(&todayAmount)

	// 总金额
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ?", staff.ID).
		Select("COALESCE(SUM(pay_amount), 0)").Scan(&totalAmount)

	response.Success(c, gin.H{
		"today_accepted": todayCount,
		"total_accepted": totalCount,
		"today_amount":   todayAmount,
		"total_amount":   totalAmount,
	})
}
