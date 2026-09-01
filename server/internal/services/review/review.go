// Package review 提供服务评价相关的业务逻辑（PRD V2.0 阶段四）
package review

import (
	"math"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
)

// 质量分权重：总体 60% + 态度 10% + 专业 10% + 准时 20%
const (
	weightScore        = 0.6
	weightAttitude     = 0.1
	weightProfessional = 0.1
	weightPunctual     = 0.2
)

// RecalculateQualityScore 重算服务人员质量分（近 100 条正常展示评价加权平均）
// 评价提交或隐藏后调用（可异步）。无有效评价时回落到默认 5.0。
func RecalculateQualityScore(staffID uint64) error {
	var stats struct {
		Cnt             int64
		AvgScore        float64
		AvgAttitude     float64
		AvgProfessional float64
		AvgPunctual     float64
	}

	// 近 N 条正常展示评价
	recent := database.DB.Model(&models.ServiceReview{}).
		Select("score, attitude_score, professional_score, punctual_score").
		Where("staff_id = ? AND status = ?", staffID, utils.ReviewStatusVisible).
		Order("created_at DESC").
		Limit(utils.ReviewNearLimit)

	if err := database.DB.Table("(?) AS recent", recent).
		Select("COUNT(*) AS cnt, COALESCE(AVG(recent.score),0) AS avg_score, COALESCE(AVG(recent.attitude_score),0) AS avg_attitude, COALESCE(AVG(recent.professional_score),0) AS avg_professional, COALESCE(AVG(recent.punctual_score),0) AS avg_punctual").
		Scan(&stats).Error; err != nil {
		return err
	}

	quality := 5.0
	if stats.Cnt > 0 {
		quality = stats.AvgScore*weightScore +
			stats.AvgAttitude*weightAttitude +
			stats.AvgProfessional*weightProfessional +
			stats.AvgPunctual*weightPunctual
		quality = math.Round(quality*10) / 10
	}

	return database.DB.Model(&models.ServiceStaff{}).
		Where("id = ?", staffID).
		Update("quality_score", quality).Error
}