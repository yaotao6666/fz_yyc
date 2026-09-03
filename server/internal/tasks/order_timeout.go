// Package tasks 订单/业务超时预警定时任务（PRD V2.0 阶段五：预警中心扩展）。
//   - 每 5 分钟：按 alert_settings 配置阈值扫描 6 类订单/业务超时预警（幂等写入 service_alert_events）
//   - 预警总开关 enabled=false 时跳过全部扫描
package tasks

import (
	"fmt"
	"log"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/alertcfg"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
)

// loadAlertSettings 读取预警配置（从通用 system_configs 聚合并回退默认值）
func loadAlertSettings() models.AlertSettings {
	return alertcfg.Load()
}

// ensureAlert 幂等写入订单级预警：同一 order_id + alert_type 仅保留一条未闭环(status!=3)事件
func ensureAlert(orderID uint64, staffID *uint64, alertType uint8, summary string) (bool, error) {
	var count int64
	if err := database.DB.Model(&models.ServiceAlertEvent{}).
		Where("order_id = ? AND alert_type = ? AND status != ?", orderID, alertType, utils.AlertStatusHandled).
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	var order models.Order
	if err := database.DB.Select("delivery_address").
		Where("id = ?", orderID).First(&order).Error; err != nil {
		return false, nil // 订单已不存在则跳过
	}

	event := models.ServiceAlertEvent{
		OrderID:   &orderID,
		StaffID:   staffID,
		AlertType: alertType,
		Address:   order.DeliveryAddress,
		Summary:   summary,
		Status:    utils.AlertStatusPending,
	}
	if err := database.DB.Create(&event).Error; err != nil {
		return false, err
	}
	return true, nil
}

// scanGoodsUnverified 实物零售订单支付后超时未核销（type=3）
func scanGoodsUnverified(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.GoodsUnverifiedHours) * time.Hour)
	var orders []models.Order
	if err := database.DB.Select("id, delivery_address, paid_at").
		Where("order_type = ? AND status = ? AND completed_at IS NULL AND paid_at IS NOT NULL AND paid_at < ?",
			utils.OrderTypeGoodsMin, 2, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, o := range orders {
		summary := fmt.Sprintf("实物订单支付后超过 %d 小时未核销（支付时间 %s）",
			cfg.GoodsUnverifiedHours, o.PaidAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(o.ID, nil, utils.AlertTypeGoodsUnverified, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 实物未核销预警写入失败 order=%d: %v", o.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// scanServiceUnassigned 服务订单支付后超时未指派人员（type=4）
func scanServiceUnassigned(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.ServiceUnassignedHours) * time.Hour)
	var orders []models.Order
	if err := database.DB.Select("id, delivery_address, paid_at").
		Where("order_type BETWEEN ? AND ? AND status = ? AND biz_status = ? AND assigned_staff_id IS NULL AND paid_at IS NOT NULL AND paid_at < ?",
			utils.OrderTypeServiceMin, utils.OrderTypeServiceMax, 2, 1, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, o := range orders {
		summary := fmt.Sprintf("服务订单支付后超过 %d 小时未指派人员（支付时间 %s）",
			cfg.ServiceUnassignedHours, o.PaidAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(o.ID, nil, utils.AlertTypeServiceUnassigned, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 服务未指派预警写入失败 order=%d: %v", o.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// scanEscortUnfinished 陪诊订单(预约服务)已签到但超时未签退（type=5）
func scanEscortUnfinished(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.EscortUnfinishedMinutes) * time.Minute)
	var orders []models.Order
	if err := database.DB.Select("id, delivery_address, assigned_staff_id, actual_started_at").
		Where("order_type = ? AND biz_status = ? AND assigned_staff_id IS NOT NULL AND actual_started_at IS NOT NULL AND actual_ended_at IS NULL AND actual_started_at < ?",
			4, 3, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, o := range orders {
		summary := fmt.Sprintf("陪诊服务开始超过 %d 分钟未完成（开始时间 %s）",
			cfg.EscortUnfinishedMinutes, o.ActualStartedAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(o.ID, o.AssignedStaffID, utils.AlertTypeEscortUnfinished, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 陪诊未完成预警写入失败 order=%d: %v", o.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// scanServiceUnstarted 服务订单已指派但超时未签到（type=6）
func scanServiceUnstarted(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.ServiceUnstartedMinutes) * time.Minute)
	var orders []models.Order
	if err := database.DB.Select("id, delivery_address, assigned_staff_id, assigned_at").
		Where("order_type BETWEEN ? AND ? AND assigned_staff_id IS NOT NULL AND actual_started_at IS NULL AND assigned_at IS NOT NULL AND assigned_at < ?",
			utils.OrderTypeServiceMin, utils.OrderTypeServiceMax, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, o := range orders {
		summary := fmt.Sprintf("已指派超过 %d 分钟未签到（指派时间 %s）",
			cfg.ServiceUnstartedMinutes, o.AssignedAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(o.ID, o.AssignedStaffID, utils.AlertTypeServiceUnstarted, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 指派未签到预警写入失败 order=%d: %v", o.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// scanRentalOverdue 租赁订单逾期未归还（type=7）
func scanRentalOverdue(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.RentalOverdueHours) * time.Hour)
	var orders []models.Order
	if err := database.DB.Select("id, delivery_address, rental_end_at").
		Where("order_type = ? AND status IN (?) AND deposit_status IN (?) AND rental_end_at IS NOT NULL AND rental_end_at < ?",
			2, []uint8{2, 3}, []uint8{0, 1}, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, o := range orders {
		summary := fmt.Sprintf("租赁订单逾期超过 %d 小时未归还（到期时间 %s）",
			cfg.RentalOverdueHours, o.RentalEndAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(o.ID, nil, utils.AlertTypeRentalOverdue, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 租赁逾期预警写入失败 order=%d: %v", o.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// scanRefundStuck 退款单处理中超时未回调（type=8）
func scanRefundStuck(cfg models.AlertSettings) (int, error) {
	threshold := time.Now().Add(-time.Duration(cfg.RefundStuckHours) * time.Hour)
	var refunds []models.Refund
	if err := database.DB.Select("id, order_id, created_at").
		Where("status = ? AND created_at < ?", 0, threshold).
		Find(&refunds).Error; err != nil {
		return 0, err
	}
	created := 0
	for _, rf := range refunds {
		summary := fmt.Sprintf("退款单处理中超过 %d 小时未回调（发起时间 %s）",
			cfg.RefundStuckHours, rf.CreatedAt.Format("2006-01-02 15:04"))
		ok, err := ensureAlert(rf.OrderID, nil, utils.AlertTypeRefundStuck, summary)
		if err != nil {
			log.Printf("[ALERT-TASK] 退款卡单预警写入失败 refund=%d: %v", rf.ID, err)
			continue
		}
		if ok {
			created++
		}
	}
	return created, nil
}

// StartOrderTimeoutTasks 启动订单/业务超时预警定时任务（阻塞式，需以 goroutine 调用）
func StartOrderTimeoutTasks() {
	runAll := func() {
		cfg := loadAlertSettings()
		if !cfg.Enabled {
			log.Printf("[ALERT-TASK] 预警总开关关闭，跳过本轮订单超时扫描")
			return
		}
		if n, err := scanGoodsUnverified(cfg); err != nil {
			log.Printf("[ALERT-TASK] 实物未核销扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 实物未核销预警新增: %d 单", n)
		}
		if n, err := scanServiceUnassigned(cfg); err != nil {
			log.Printf("[ALERT-TASK] 服务未指派扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 服务未指派预警新增: %d 单", n)
		}
		if n, err := scanEscortUnfinished(cfg); err != nil {
			log.Printf("[ALERT-TASK] 陪诊未完成扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 陪诊未完成预警新增: %d 单", n)
		}
		if n, err := scanServiceUnstarted(cfg); err != nil {
			log.Printf("[ALERT-TASK] 指派未签到扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 指派未签到预警新增: %d 单", n)
		}
		if n, err := scanRentalOverdue(cfg); err != nil {
			log.Printf("[ALERT-TASK] 租赁逾期扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 租赁逾期预警新增: %d 单", n)
		}
		if n, err := scanRefundStuck(cfg); err != nil {
			log.Printf("[ALERT-TASK] 退款卡单扫描失败: %v", err)
		} else if n > 0 {
			log.Printf("[ALERT-TASK] 退款卡单预警新增: %d 单", n)
		}
	}

	// 启动即扫一次，保证重启后口径准确
	runAll()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		runAll()
	}
}
