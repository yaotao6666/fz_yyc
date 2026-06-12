package user

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetPickupPoints(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var merchant models.Merchant
	if err := database.DB.Select("id", "pickup_enabled").First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
		return
	}
	if !merchant.PickupEnabled {
		response.Success(c, []models.MerchantPickupPoint{})
		return
	}

	var points []models.MerchantPickupPoint
	if err := database.DB.
		Where("merchant_id = ? AND status = 1", id).
		Order("is_default DESC, sort ASC, id DESC").
		Find(&points).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取自提点失败")
		return
	}

	response.Success(c, points)
}

