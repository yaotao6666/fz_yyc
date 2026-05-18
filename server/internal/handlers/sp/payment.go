package sp

import (
	"context"
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/config"
	wsHandler "fz_yyc_api/internal/handlers/ws"
	"math"
	"net/http"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	profitSharingPending uint8 = 0
	profitSharingSuccess uint8 = 1
	profitSharingFailed  uint8 = 2
	profitSharingSkipped uint8 = 3
)

func PaymentNotify(c *gin.Context) {
	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "读取回调内容失败"})
		return
	}

	eventType, plaintext, err := client.ParseAndVerifyNotifyEvent(c.Request.Header, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	if strings.Contains(eventType, "REFUND") {
		if err := processRefundNotify(context.Background(), plaintext); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	var resource struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        struct {
			PayerTotal int64 `json:"payer_total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plaintext, &resource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "解析支付回调资源失败"})
		return
	}

	if resource.TradeState != "SUCCESS" {
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "忽略非成功支付通知"})
		return
	}

	successTime, _ := time.Parse(time.RFC3339, resource.SuccessTime)
	notifyResult := &wechatpay.NotifyResult{
		EventType:     eventType,
		OrderNo:       resource.OutTradeNo,
		TransactionID: resource.TransactionID,
		TradeState:    resource.TradeState,
		SuccessTime:   successTime,
		PayAmount:     resource.Amount.PayerTotal,
		RawPayload:    plaintext,
	}

	if err := processPaymentSuccess(context.Background(), client, notifyResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

func processPaymentSuccess(ctx context.Context, client *wechatpay.ServiceProviderClient, notifyResult *wechatpay.NotifyResult) error {
	var shouldBroadcastOrderNotify bool
	var notifyMerchantID uint64
	var notifyOrderNo string

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var order models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ?", notifyResult.OrderNo).
			First(&order).Error; err != nil {
			return fmt.Errorf("订单不存在: %w", err)
		}

		var merchant models.Merchant
		if err := tx.First(&merchant, order.MerchantID).Error; err != nil {
			return fmt.Errorf("商家不存在: %w", err)
		}

		rawPayload, _ := json.Marshal(notifyResult)
		paySuccessAt := notifyResult.SuccessTime
		if paySuccessAt.IsZero() {
			paySuccessAt = time.Now()
		}

		payUpdates := map[string]any{
			"transaction_id":     notifyResult.TransactionID,
			"pay_notify_payload": models.JSON(rawPayload),
		}
		if order.Status < 2 {
			payUpdates["status"] = 2
			payUpdates["paid_at"] = paySuccessAt
			shouldBroadcastOrderNotify = true
			notifyMerchantID = order.MerchantID
			notifyOrderNo = order.OrderNo
		} else if order.PaidAt == nil {
			payUpdates["paid_at"] = paySuccessAt
		}
		if err := tx.Model(&order).Updates(payUpdates).Error; err != nil {
			return fmt.Errorf("更新订单支付状态失败: %w", err)
		}
		order.TransactionID = notifyResult.TransactionID

		if order.ProfitSharingStatus == profitSharingSuccess || order.ProfitSharingStatus == profitSharingSkipped {
			return nil
		}

		// 以订单维度锁定并复用已有记录，避免重复回调再次发起分账。
		var existingRecord models.MerchantProfitSharingRecord
		recordErr := tx.Where("order_id = ?", order.ID).First(&existingRecord).Error
		if recordErr == nil {
			return syncOrderProfitSharingFromRecord(tx, &order, &existingRecord)
		}
		if recordErr != nil && recordErr != gorm.ErrRecordNotFound {
			return fmt.Errorf("查询分账记录失败: %w", recordErr)
		}

		ratio := merchant.ProfitSharingRatio
		profitSharingOrderNo := fmt.Sprintf("ps_%s_%d", order.OrderNo, time.Now().Unix())
		profitSharingAmount := roundAmount(order.PayAmount * ratio / 100)
		merchantReceivedAmount := roundAmount(order.PayAmount - profitSharingAmount)

		if !merchant.ProfitSharingEnabled || ratio <= 0 {
			return createSkippedProfitSharingRecord(tx, &order, &merchant, notifyResult.TransactionID, profitSharingOrderNo, "商家未开启分账")
		}
		if profitSharingAmount <= 0 {
			return createSkippedProfitSharingRecord(tx, &order, &merchant, notifyResult.TransactionID, profitSharingOrderNo, "分账金额为0，已跳过")
		}
		if merchant.SubMchID == "" {
			return createSkippedProfitSharingRecord(tx, &order, &merchant, notifyResult.TransactionID, profitSharingOrderNo, "商家未配置子商户号")
		}

		record := models.MerchantProfitSharingRecord{
			ServiceProviderID:      merchant.ServiceProviderID,
			MerchantID:             merchant.ID,
			OrderID:                order.ID,
			OrderNo:                order.OrderNo,
			TransactionID:          notifyResult.TransactionID,
			ProfitSharingOrderNo:   profitSharingOrderNo,
			ProfitSharingDate:      time.Now(),
			PayAmount:              order.PayAmount,
			ProfitSharingRatio:     ratio,
			ProfitSharingAmount:    profitSharingAmount,
			MerchantReceivedAmount: merchantReceivedAmount,
			Status:                 profitSharingPending,
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("创建分账记录失败: %w", err)
		}
		if err := tx.Model(&order).Updates(map[string]any{
			"profit_sharing_status":   profitSharingPending,
			"profit_sharing_amount":   profitSharingAmount,
			"profit_sharing_order_no": profitSharingOrderNo,
			"profit_sharing_at":       nil,
			"profit_sharing_error":    "",
		}).Error; err != nil {
			return fmt.Errorf("写入订单待分账状态失败: %w", err)
		}

		result, err := client.CreateProfitSharingOrder(ctx, wechatpay.ProfitSharingRequest{
			AppID:         config.Config.Wechat.AppID,
			SubMchID:      merchant.SubMchID,
			TransactionID: notifyResult.TransactionID,
			OrderNo:       profitSharingOrderNo,
			Receivers: []wechatpay.ProfitSharingReceiver{
				{
					Type:        "MERCHANT_ID",
					Account:     client.GetSPMchID(),
					Amount:      amountToCents(profitSharingAmount),
					Description: "服务商抽佣",
				},
			},
		})
		if err != nil {
			updateErr := tx.Model(&order).Updates(map[string]any{
				"profit_sharing_status":   profitSharingFailed,
				"profit_sharing_amount":   profitSharingAmount,
				"profit_sharing_order_no": profitSharingOrderNo,
				"profit_sharing_error":    err.Error(),
			}).Error
			if updateErr != nil {
				return fmt.Errorf("分账失败且写入订单状态失败: %v, %w", updateErr, err)
			}
			_ = tx.Model(&record).Updates(map[string]any{
				"status":        profitSharingFailed,
				"error_message": err.Error(),
			}).Error
			return nil
		}

		now := time.Now()
		if err := tx.Model(&order).Updates(map[string]any{
			"profit_sharing_status":   profitSharingSuccess,
			"profit_sharing_amount":   profitSharingAmount,
			"profit_sharing_order_no": result.OrderID,
			"profit_sharing_at":       now,
			"profit_sharing_error":    "",
		}).Error; err != nil {
			return fmt.Errorf("更新订单分账状态失败: %w", err)
		}

		if err := tx.Model(&record).Updates(map[string]any{
			"status":                  profitSharingSuccess,
			"profit_sharing_order_no": result.OrderID,
			"profit_sharing_date":     now,
			"error_message":           "",
		}).Error; err != nil {
			return fmt.Errorf("更新分账记录失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if shouldBroadcastOrderNotify {
		wsHandler.BroadcastOrderNotify(notifyMerchantID, notifyOrderNo)
	}
	return nil
}

func processRefundNotify(ctx context.Context, plaintext []byte) error {
	var resource struct {
		OutRefundNo  string `json:"out_refund_no"`
		RefundID     string `json:"refund_id"`
		RefundStatus string `json:"refund_status"`
		SuccessTime  string `json:"success_time"`
	}
	if err := json.Unmarshal(plaintext, &resource); err != nil {
		return fmt.Errorf("解析退款回调资源失败: %w", err)
	}
	if resource.OutRefundNo == "" {
		return fmt.Errorf("退款回调缺少退款单号")
	}

	now := time.Now()
	if resource.SuccessTime != "" {
		if parsed, err := time.Parse(time.RFC3339, resource.SuccessTime); err == nil {
			now = parsed
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var refund models.Refund
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("refund_no = ?", resource.OutRefundNo).
			First(&refund).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}

		refundUpdates := map[string]any{
			"refund_id": resource.RefundID,
		}

		switch strings.ToUpper(resource.RefundStatus) {
		case "SUCCESS":
			refundUpdates["status"] = 2
			refundUpdates["refunded_at"] = now
			if err := tx.Model(&refund).Updates(refundUpdates).Error; err != nil {
				return err
			}
			return tx.Model(&models.Order{}).Where("id = ?", refund.OrderID).Updates(map[string]any{
				"status":      6,
				"refunded_at": now,
			}).Error
		case "CLOSED", "ABNORMAL":
			refundUpdates["status"] = 3
			if err := tx.Model(&refund).Updates(refundUpdates).Error; err != nil {
				return err
			}
			return nil
		default:
			refundUpdates["status"] = 1
			return tx.Model(&refund).Updates(refundUpdates).Error
		}
	})
}

func roundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func amountToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func createSkippedProfitSharingRecord(
	tx *gorm.DB,
	order *models.Order,
	merchant *models.Merchant,
	transactionID string,
	profitSharingOrderNo string,
	reason string,
) error {
	now := time.Now()
	record := models.MerchantProfitSharingRecord{
		ServiceProviderID:      merchant.ServiceProviderID,
		MerchantID:             merchant.ID,
		OrderID:                order.ID,
		OrderNo:                order.OrderNo,
		TransactionID:          transactionID,
		ProfitSharingOrderNo:   profitSharingOrderNo,
		ProfitSharingDate:      now,
		PayAmount:              order.PayAmount,
		ProfitSharingRatio:     merchant.ProfitSharingRatio,
		ProfitSharingAmount:    0,
		MerchantReceivedAmount: order.PayAmount,
		Status:                 profitSharingSkipped,
		ErrorMessage:           reason,
	}
	if err := tx.Create(&record).Error; err != nil {
		return fmt.Errorf("创建跳过分账记录失败: %w", err)
	}
	if err := tx.Model(order).Updates(map[string]any{
		"profit_sharing_status":   profitSharingSkipped,
		"profit_sharing_amount":   0,
		"profit_sharing_order_no": profitSharingOrderNo,
		"profit_sharing_at":       now,
		"profit_sharing_error":    reason,
	}).Error; err != nil {
		return fmt.Errorf("更新订单跳过分账状态失败: %w", err)
	}
	return nil
}

func syncOrderProfitSharingFromRecord(tx *gorm.DB, order *models.Order, record *models.MerchantProfitSharingRecord) error {
	updates := map[string]any{
		"profit_sharing_status":   record.Status,
		"profit_sharing_amount":   record.ProfitSharingAmount,
		"profit_sharing_order_no": record.ProfitSharingOrderNo,
		"profit_sharing_error":    record.ErrorMessage,
	}
	if record.Status == profitSharingSuccess || record.Status == profitSharingSkipped {
		updates["profit_sharing_at"] = record.ProfitSharingDate
	} else {
		updates["profit_sharing_at"] = nil
	}
	if err := tx.Model(order).Updates(updates).Error; err != nil {
		return fmt.Errorf("同步订单分账状态失败: %w", err)
	}
	return nil
}
