package service_staff

import (
	"net/http"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetMyQualityScore 我的质量分与近期评价
func GetMyQualityScore(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	// 正常展示评价总数与各维度平均分
	var stats struct {
		Cnt             int64
		AvgScore        float64
		AvgAttitude     float64
		AvgProfessional float64
		AvgPunctual     float64
	}
	database.DB.Model(&models.ServiceReview{}).
		Select("COUNT(*) AS cnt, COALESCE(AVG(score),0) AS avg_score, COALESCE(AVG(attitude_score),0) AS avg_attitude, COALESCE(AVG(professional_score),0) AS avg_professional, COALESCE(AVG(punctual_score),0) AS avg_punctual").
		Where("staff_id = ? AND status = ?", staff.ID, utils.ReviewStatusVisible).
		Scan(&stats)

	// 近期评价（近 10 条正常展示）
	var recent []models.ServiceReview
	database.DB.Select("id, score, content, created_at").
		Where("staff_id = ? AND status = ?", staff.ID, utils.ReviewStatusVisible).
		Order("created_at DESC").
		Limit(10).
		Find(&recent)

	response.Success(c, gin.H{
		"quality_score":   staff.QualityScore,
		"review_count":    stats.Cnt,
		"avg_score":       round1(stats.AvgScore),
		"avg_attitude":    round1(stats.AvgAttitude),
		"avg_professional": round1(stats.AvgProfessional),
		"avg_punctual":    round1(stats.AvgPunctual),
		"recent_reviews":  recent,
	})
}

// round1 保留 1 位小数
func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}