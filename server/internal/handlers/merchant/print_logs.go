package merchant

import (
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetPrintLogs(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.PrintLog{}).Where("merchant_id = ?", merchantID)
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		endDateTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "结束日期格式错误")
			return
		}
		query = query.Where("created_at <= ?", endDateTime.Add(24*time.Hour))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取打印记录失败")
		return
	}

	var logs []models.PrintLog
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC, id DESC").
		Find(&logs).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取打印记录失败")
		return
	}

	response.Success(c, gin.H{
		"list": logs,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
