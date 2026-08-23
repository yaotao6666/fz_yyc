package merchant

import (
	"net/http"
	"strings"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// PaymentConfigRequest 支付配置请求
type PaymentConfigRequest struct {
	SubMchID string `json:"sub_mch_id"`
}

// UpdatePaymentConfig 更新商家支付配置
func UpdatePaymentConfig(c *gin.Context) {
	var req PaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	subMchID := strings.TrimSpace(req.SubMchID)
	if subMchID == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "子商户号不能为空")
		return
	}

	updates := map[string]interface{}{
		"sub_mch_id":            subMchID,
		"payment_config_status": 1,
	}
	if err := database.DB.Model(&models.Merchant{}).Where("id = ?", utils.DefaultMerchantID).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新支付配置失败")
		return
	}

	var merchant models.Merchant
	database.DB.First(&merchant, utils.DefaultMerchantID)
	response.SuccessWithMessage(c, "支付配置已更新", merchant)
}
