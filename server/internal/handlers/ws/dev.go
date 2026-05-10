package ws

import (
	"encoding/json"
	"net/http"

	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type DevOrderNotifyRequest struct {
	MerchantID uint64 `json:"merchant_id" binding:"required"`
	OrderNo    string `json:"order_no"`
}

type DevStoreVisitNotifyRequest struct {
	MerchantID    uint64 `json:"merchant_id" binding:"required"`
	VisitorOpenID string `json:"visitor_openid"`
	Source        string `json:"source"`
}

func DevOrderNotify(c *gin.Context) {
	if gin.Mode() == gin.ReleaseMode {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "接口不可用")
		return
	}

	var req DevOrderNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	payload := gin.H{
		"type": "order_notify",
		"payload": gin.H{
			"merchant_id": req.MerchantID,
			"order_no":    req.OrderNo,
		},
	}

	msg, _ := json.Marshal(payload)
	delivered := BroadcastToMerchant(req.MerchantID, msg)

	response.Success(c, gin.H{
		"delivered": delivered,
	})
}

func DevStoreVisitNotify(c *gin.Context) {
	if gin.Mode() == gin.ReleaseMode {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "接口不可用")
		return
	}

	var req DevStoreVisitNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.VisitorOpenID == "" {
		req.VisitorOpenID = "mock_visitor"
	}
	if req.Source == "" {
		req.Source = "dev"
	}

	delivered := BroadcastStoreVisitNotify(req.MerchantID, req.VisitorOpenID, req.Source)
	response.Success(c, gin.H{
		"delivered": delivered,
	})
}
