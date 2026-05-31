package merchant

import (
	"net/http"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/fullreduction"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type fullReductionRuleRequest struct {
	ThresholdAmount float64 `json:"threshold_amount"`
	DiscountAmount  float64 `json:"discount_amount"`
	Status          *uint8  `json:"status"`
}

type updateFullReductionRulesRequest struct {
	Rules []fullReductionRuleRequest `json:"rules"`
}

func GetFullReductionRules(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	rules, err := fullreduction.GetActiveRulesByMerchantID(merchantID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取满减规则失败")
		return
	}

	var allRules []models.MerchantFullReductionRule
	if err := database.DB.
		Where("merchant_id = ?", merchantID).
		Order("threshold_amount ASC, sort ASC, id ASC").
		Find(&allRules).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取满减规则失败")
		return
	}

	response.Success(c, gin.H{
		"rules":        allRules,
		"active_rules": rules,
	})
}

func UpdateFullReductionRules(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req updateFullReductionRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if len(req.Rules) > 5 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "最多支持配置 5 档满减规则")
		return
	}

	thresholdSet := map[float64]struct{}{}
	for _, rule := range req.Rules {
		if rule.ThresholdAmount <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "满减门槛必须大于 0")
			return
		}
		if rule.DiscountAmount <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "减免金额必须大于 0")
			return
		}
		if rule.DiscountAmount >= rule.ThresholdAmount {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "减免金额必须小于满减门槛")
			return
		}
		if _, exists := thresholdSet[rule.ThresholdAmount]; exists {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "满减门槛不能重复")
			return
		}
		thresholdSet[rule.ThresholdAmount] = struct{}{}
	}

	tx := database.DB.Begin()
	if err := tx.Where("merchant_id = ?", merchantID).Delete(&models.MerchantFullReductionRule{}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存满减规则失败")
		return
	}

	createdRules := make([]models.MerchantFullReductionRule, 0, len(req.Rules))
	for index, rule := range req.Rules {
		createdRule := models.MerchantFullReductionRule{
			MerchantID:      merchantID,
			ThresholdAmount: rule.ThresholdAmount,
			DiscountAmount:  rule.DiscountAmount,
			Sort:            index + 1,
			Status:          1,
		}
		if rule.Status != nil {
			createdRule.Status = *rule.Status
		}
		if err := tx.Create(&createdRule).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存满减规则失败")
			return
		}
		createdRules = append(createdRules, createdRule)
	}

	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存满减规则失败")
		return
	}

	response.Success(c, gin.H{
		"rules": createdRules,
	})
}
