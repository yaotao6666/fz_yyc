package merchant

import (
	"net/http"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type subscriptionItem struct {
	ID         uint64 `json:"id"`
	NotifyType string `json:"notify_type"`
	Enabled    bool   `json:"enabled"`
	PushOpenID string `json:"push_openid"`
}

func GetSubscriptions(c *gin.Context) {
	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工身份失败")
		return
	}

	items := []subscriptionItem{
		{ID: 1, NotifyType: "order_new", Enabled: staff.NotifyEnabled, PushOpenID: staff.OpenID},
		{ID: 2, NotifyType: "order_paid", Enabled: staff.NotifyEnabled, PushOpenID: staff.OpenID},
		{ID: 3, NotifyType: "order_refund", Enabled: staff.NotifyEnabled, PushOpenID: staff.OpenID},
	}
	response.Success(c, items)
}

type updateSubscriptionsRequest struct {
	Subscriptions []struct {
		NotifyType string `json:"notify_type" binding:"required"`
		Enabled    *bool  `json:"enabled" binding:"required"`
		PushOpenID string `json:"push_openid"`
	} `json:"subscriptions" binding:"required"`
}

func UpdateSubscriptions(c *gin.Context) {
	var req updateSubscriptionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工身份失败")
		return
	}

	var enabled *bool
	for _, item := range req.Subscriptions {
		if item.Enabled == nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
			return
		}
		if enabled == nil {
			enabled = item.Enabled
			continue
		}
		if *enabled != *item.Enabled {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "暂不支持不同通知类型分别开关")
			return
		}
	}
	if enabled == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订阅配置不能为空")
		return
	}

	updates := map[string]interface{}{
		"notify_enabled": *enabled,
	}
	if staff.OpenID == "" {
		for _, item := range req.Subscriptions {
			if item.PushOpenID != "" {
				updates["openid"] = item.PushOpenID
				break
			}
		}
	}

	if err := database.DB.Model(&models.MerchantStaff{}).Where("id = ?", staff.ID).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新订阅配置失败")
		return
	}

	GetSubscriptions(c)
}
