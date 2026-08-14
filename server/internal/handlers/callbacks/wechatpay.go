package callbacks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WechatPayPayCallback 微信支付成功回调
// 验签→解密→更新订单状态为已支付→写入transaction_id
func WechatPayPayCallback(c *gin.Context) {
	ctx := c.Request.Context()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WECHAT-PAY-CALLBACK] 读取请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid request"})
		return
	}

	client, cliErr := wechatpay.NewServiceProviderClient()
	if cliErr != nil {
		log.Printf("[WECHAT-PAY-CALLBACK] 支付客户端初始化失败: %v", cliErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "internal error"})
		return
	}

	var tx wechatpay.PayTransactionResource
	notifyReq, parseErr := client.ParseNotifyRequestRaw(ctx, c.Request.Header, body, &tx)
	if parseErr != nil {
		log.Printf("[WECHAT-PAY-CALLBACK] 通知解析/验签失败: %v, body=%s", parseErr, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "verify fail"})
		return
	}
	_ = notifyReq

	// 只处理支付成功通知
	if tx.TradeState != "SUCCESS" {
		log.Printf("[WECHAT-PAY-CALLBACK] 忽略非成功支付通知, out_trade_no=%s, trade_state=%s",
			tx.OutTradeNo, tx.TradeState)
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "ok"})
		return
	}

	payload, _ := json.Marshal(tx)
	if err := markOrderPaid(tx.OutTradeNo, tx.TransactionID, payload); err != nil {
		log.Printf("[WECHAT-PAY-CALLBACK] 更新订单状态失败: out_trade_no=%s, err=%v", tx.OutTradeNo, err)
		// 幂等：若已经是已支付状态则返回成功避免重复回调
		if isOrderAlreadyPaid(tx.OutTradeNo) {
			c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "ok"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "process error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "ok"})
}

// WechatPayRefundCallback 微信退款成功/失败回调
func WechatPayRefundCallback(c *gin.Context) {
	ctx := c.Request.Context()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WECHAT-REFUND-CALLBACK] 读取请求体失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid request"})
		return
	}

	client, cliErr := wechatpay.NewServiceProviderClient()
	if cliErr != nil {
		log.Printf("[WECHAT-REFUND-CALLBACK] 支付客户端初始化失败: %v", cliErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "internal error"})
		return
	}

	var refund wechatpay.RefundNotifyResource
	_, parseErr := client.ParseNotifyRequestRaw(ctx, c.Request.Header, body, &refund)
	if parseErr != nil {
		log.Printf("[WECHAT-REFUND-CALLBACK] 通知解析/验签失败: %v, body=%s", parseErr, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "verify fail"})
		return
	}

	if err := applyRefundResult(refund); err != nil {
		log.Printf("[WECHAT-REFUND-CALLBACK] 处理退款结果失败: out_refund_no=%s, err=%v",
			refund.OutRefundNo, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "process error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "ok"})
}

// markOrderPaid 更新订单为已支付（幂等，通过 DB 条件更新保证）
func markOrderPaid(orderNo, transactionID string, payload []byte) error {
	if orderNo == "" {
		return fmt.Errorf("out_trade_no 为空")
	}
	now := time.Now()
	db := database.DB
	var order models.Order
	if err := db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}
	if order.Status == 2 || order.Status == 3 || order.Status == 6 {
		// 已支付/已完成/已退款都视作重复回调
		return nil
	}

	updates := map[string]interface{}{
		"status":             2, // 已支付
		"transaction_id":     transactionID,
		"paid_at":            now,
		"pay_notify_payload": payload,
	}

	result := db.Model(&models.Order{}).
		Where("order_no = ? AND status IN (?)", orderNo, []uint8{1, 5}). // 待支付 / 退款中（仅极端情况）
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	// biz_status 初始流转：普通/租赁订单默认「待接单」=1
	if result.RowsAffected > 0 && (order.OrderType == 1 || order.OrderType == 2) {
		db.Model(&models.Order{}).Where("id = ?", order.ID).Update("biz_status", 1)
	}
	return nil
}

func isOrderAlreadyPaid(orderNo string) bool {
	if orderNo == "" {
		return false
	}
	var o models.Order
	if err := database.DB.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return false
	}
	return o.Status == 2 || o.Status == 3 || o.Status == 6
}

// applyRefundResult 根据退款通知写 refunds 表和关联订单状态
func applyRefundResult(r wechatpay.RefundNotifyResource) error {
	if r.OutRefundNo == "" {
		return fmt.Errorf("out_refund_no 为空")
	}
	success := r.RefundStatus == "SUCCESS"
	fail := r.RefundStatus == "CLOSED" || r.RefundStatus == "ABNORMAL"
	var status uint8
	switch {
	case success:
		status = 1
	case fail:
		status = 2
	default:
		// PROCESSING 不处理
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var refund models.Refund
		if err := tx.Where("refund_no = ?", r.OutRefundNo).First(&refund).Error; err != nil {
			return fmt.Errorf("退款记录不存在: %w", err)
		}
		if refund.Status == 1 {
			return nil
		}
		updates := map[string]interface{}{
			"status": status,
		}
		if r.RefundID != "" {
			updates["refund_id"] = r.RefundID
		}
		if success {
			if r.SuccessTime != "" {
				if t, err := time.Parse(time.RFC3339, r.SuccessTime); err == nil {
					updates["refunded_at"] = t
				} else {
					updates["refunded_at"] = time.Now()
				}
			} else {
				updates["refunded_at"] = time.Now()
			}
		}
		if err := tx.Model(&refund).Updates(updates).Error; err != nil {
			return err
		}

		// 同步订单状态：若最新的该订单退款都成功，标记订单为「已退款」
		if success {
			var pending int64
			tx.Model(&models.Refund{}).Where("order_id = ? AND status = 0", refund.OrderID).Count(&pending)
			if pending == 0 {
				tx.Model(&models.Order{}).Where("id = ?", refund.OrderID).
					Updates(map[string]interface{}{"status": 6, "refunded_at": updates["refunded_at"]})
			}
		}
		return nil
	})
}
