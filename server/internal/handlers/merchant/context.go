package merchant

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

func resolveTargetMerchantID(c *gin.Context) (uint64, bool) {
	userType := middleware.GetUserType(c)
	if userType == "merchant" {
		return middleware.GetMerchantID(c), true
	}

	if userType != "sp" {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "需要商家或服务商权限")
		return 0, false
	}

	merchantID, ok := extractMerchantIDForSp(c)
	if !ok {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "缺少 merchant_id")
		return 0, false
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return 0, false
	}
	spUserID, ok := userIDValue.(uint64)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return 0, false
	}

	var spUser models.ServiceProviderSp
	if err := database.DB.Select("service_provider_id").First(&spUser, spUserID).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "服务商身份无效")
		return 0, false
	}

	var merchant models.Merchant
	if err := database.DB.Select("id").Where(
		"id = ? AND service_provider_id = ?",
		merchantID,
		spUser.ServiceProviderID,
	).First(&merchant).Error; err != nil {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权操作该商家")
		return 0, false
	}

	return merchantID, true
}

func extractMerchantIDForSp(c *gin.Context) (uint64, bool) {
	if merchantID, ok := parseMerchantIDValue(c.Query("merchant_id")); ok {
		return merchantID, true
	}

	if c.Request == nil || c.Request.Body == nil {
		return 0, false
	}

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0, false
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
	if len(rawData) == 0 {
		return 0, false
	}

	var payload map[string]any
	if err := json.Unmarshal(rawData, &payload); err != nil {
		return 0, false
	}

	value, exists := payload["merchant_id"]
	if !exists {
		return 0, false
	}

	switch typedValue := value.(type) {
	case float64:
		if typedValue <= 0 {
			return 0, false
		}
		return uint64(typedValue), true
	case string:
		return parseMerchantIDValue(typedValue)
	default:
		return 0, false
	}
}

func parseMerchantIDValue(value string) (uint64, bool) {
	if value == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
