package fullreduction

import (
	"sort"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
)

func GetActiveRulesByMerchantID(merchantID uint64) ([]models.MerchantFullReductionRule, error) {
	var rules []models.MerchantFullReductionRule
	if err := database.DB.
		Where("merchant_id = ? AND status = 1", merchantID).
		Order("threshold_amount ASC, sort ASC, id ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return normalizeRules(rules), nil
}

func normalizeRules(rules []models.MerchantFullReductionRule) []models.MerchantFullReductionRule {
	normalizedRules := append([]models.MerchantFullReductionRule(nil), rules...)
	sort.SliceStable(normalizedRules, func(i, j int) bool {
		if normalizedRules[i].ThresholdAmount == normalizedRules[j].ThresholdAmount {
			if normalizedRules[i].Sort == normalizedRules[j].Sort {
				return normalizedRules[i].ID < normalizedRules[j].ID
			}
			return normalizedRules[i].Sort < normalizedRules[j].Sort
		}
		return normalizedRules[i].ThresholdAmount < normalizedRules[j].ThresholdAmount
	})
	return normalizedRules
}

func CalculateDiscount(totalAmount float64, rules []models.MerchantFullReductionRule) (float64, *models.MerchantFullReductionRule) {
	if totalAmount <= 0 || len(rules) == 0 {
		return 0, nil
	}

	normalizedRules := normalizeRules(rules)
	var bestRule *models.MerchantFullReductionRule
	bestDiscount := 0.0

	for index := range normalizedRules {
		rule := normalizedRules[index]
		if rule.Status != 1 {
			continue
		}
		if totalAmount < rule.ThresholdAmount {
			continue
		}
		if rule.DiscountAmount > bestDiscount {
			bestDiscount = rule.DiscountAmount
			bestRule = &normalizedRules[index]
		}
	}

	if bestDiscount < 0 {
		return 0, nil
	}
	if bestDiscount > totalAmount {
		bestDiscount = totalAmount
	}

	return bestDiscount, bestRule
}
