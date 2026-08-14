package merchant

import (
	"context"
	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/orderquery"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ReturnRentalOrderRequest struct {
	DeductAmount float64 `json:"deduct_amount"`
	Remark       string  `json:"remark"`
}

// ReturnRentalOrder 归还租赁商品并退还押金
func ReturnRentalOrder(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var req ReturnRentalOrderRequest
	_ = c.ShouldBindJSON(&req)

	var order models.Order
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	// 校验订单为已支付状态且包含租赁商品
	if order.Status != 2 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单状态不正确，仅已支付订单可归还")
		return
	}
	if order.TotalDeposit <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该订单非租赁订单，无押金需退还")
		return
	}
	if order.DepositStatus == 2 || order.DepositStatus == 3 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "押金已退还，不可重复操作")
		return
	}

	// 校验扣除金额
	deductAmount := req.DeductAmount
	if deductAmount < 0 {
		deductAmount = 0
	}
	if deductAmount > order.TotalDeposit {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "扣除金额不能超过押金总额")
		return
	}

	refundAmount := order.TotalDeposit - deductAmount
	now := time.Now()

	// 更新订单状态为已完成（已归还）
	updates := map[string]interface{}{
		"status":                3,
		"rental_returned_at":    now,
		"rental_return_remark":  strings.TrimSpace(req.Remark),
		"completed_at":          now,
		"deposit_deduct_amount": deductAmount,
		"deposit_refund_amount": refundAmount,
		"deposit_refunded_at":   now,
	}
	if deductAmount > 0 {
		updates["deposit_status"] = 3 // 部分扣除
	} else {
		updates["deposit_status"] = 2 // 已退
	}

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新订单状态失败")
		return
	}

	// 若有退款金额，发起微信退款
	if refundAmount > 0 && strings.TrimSpace(order.TransactionID) != "" {
		var merchant models.Merchant
		if err := database.DB.First(&merchant, merchantID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
			return
		}

		notifyURL := config.Config.WechatPay.CallbackURL
		if notifyURL == "" {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "服务商支付回调地址未配置")
			return
		}

		client, err := wechatpay.NewServiceProviderClient()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, err.Error())
			return
		}

		refundNo := strconv.FormatInt(time.Now().UnixNano(), 10)
		depositRefundRecord := models.Refund{
			OrderID:      order.ID,
			RefundNo:     refundNo,
			RefundAmount: refundAmount,
			RefundReason: "租赁押金退还",
			Status:       0,
		}
		if err := database.DB.Create(&depositRefundRecord).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建押金退款记录失败")
			return
		}

		refundResp, callErr := client.CreatePartnerRefund(context.Background(), wechatpay.RefundRequest{
			SubMchID:     merchant.SubMchID,
			OrderNo:      order.OrderNo,
			RefundNo:     refundNo,
			Reason:       "租赁押金退还",
			NotifyURL:    notifyURL,
			RefundAmount: int64(refundAmount * 100),
			TotalAmount:  int64(order.PayAmount * 100),
		})
		if callErr != nil {
			_ = database.DB.Model(&depositRefundRecord).Updates(map[string]any{
				"status": 2,
			}).Error
			response.Fail(c, http.StatusBadRequest, response.CodeRefundFailed, callErr.Error())
			return
		}
		refundID := strings.TrimSpace(refundResp.RefundID)
		_ = orderquery.SyncRefundAndOrderStatus(database.DB, &order, &depositRefundRecord, refundResp.Status, refundID, refundResp.SuccessTime)
	}

	// 重新加载订单
	database.DB.Preload("Items").First(&order, id)
	accessibleOrder := orderquery.BuildAccessibleOrder(order)

	response.Success(c, accessibleOrder)
}
