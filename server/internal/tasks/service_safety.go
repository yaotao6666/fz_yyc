// Package tasks 定时任务（PRD V2.0 阶段三：服务过程安全）。
//  - 每 15 分钟：服务超时未结束预警（服务中超 4 小时 → 写预警事件 alert_type=2）
//  - 每日：服务录音 30 天保留期清理（删七牛文件 + 置删除标记，失败次日重试）
package tasks

import (
	"log"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
)

// ============================================
// 服务超时未结束预警
// ============================================

// RunServiceTimeoutAlert 扫描服务中超时的订单并写预警事件（幂等：同订单同类型仅一条未闭环事件）
func RunServiceTimeoutAlert() (int, error) {
	threshold := time.Now().Add(-time.Duration(utils.ServiceTimeoutAlertHours) * time.Hour)

	// 服务中（biz_status=3）且开始时间早于阈值、已有指派人员的订单
	var orders []models.Order
	if err := database.DB.
		Where("biz_status = ? AND assigned_staff_id IS NOT NULL AND actual_started_at IS NOT NULL AND actual_started_at < ?", 3, threshold).
		Find(&orders).Error; err != nil {
		return 0, err
	}

	created := 0
	for _, order := range orders {
		if order.AssignedStaffID == nil {
			continue
		}

		// 已存在未闭环的超时预警则跳过
		var count int64
		database.DB.Model(&models.ServiceAlertEvent{}).
			Where("order_id = ? AND alert_type = ? AND status != ?",
				order.ID, utils.AlertTypeTimeout, utils.AlertStatusHandled).
			Count(&count)
		if count > 0 {
			continue
		}

		event := models.ServiceAlertEvent{
			OrderID:   &order.ID,
			StaffID:   order.AssignedStaffID,
			AlertType: utils.AlertTypeTimeout,
			Address:   order.DeliveryAddress,
			Status:    utils.AlertStatusPending,
		}
		if err := database.DB.Create(&event).Error; err != nil {
			log.Printf("[SAFETY-TASK] 超时预警写入失败 order=%d: %v", order.ID, err)
			continue
		}
		// 同步标记服务记录异常
		database.DB.Model(&models.ServiceRecord{}).
			Where("order_id = ?", order.ID).
			Update("status", utils.ServiceRecordStatusAbnormal)
		created++
	}
	return created, nil
}

// ============================================
// 服务录音 30 天清理
// ============================================

// RunServiceAudioCleanup 清理超过保留期的服务录音（失败次日重试：仅标记成功删除的）
// 保留天数读取 alert_settings.service_audio_retain_days（可视化配置），未配置时用默认值
func RunServiceAudioCleanup() (int, error) {
	retainDays := loadAlertSettings().ServiceAudioRetainDays
	if retainDays < 1 {
		retainDays = utils.AlertDefaultServiceAudioRetainDays
	}
	deadline := time.Now().AddDate(0, 0, -retainDays)

	// 已到保留期且未标记删除的录音
	var records []models.ServiceRecord
	if err := database.DB.
		Where("audio_url != '' AND audio_deleted_at IS NULL AND audio_uploaded_at IS NOT NULL AND audio_uploaded_at < ?", deadline).
		Find(&records).Error; err != nil {
		return 0, err
	}

	cleaned := 0
	for _, record := range records {
		// 先删七牛文件，删除成功后才置标记（失败留待次日重试）
		if err := qiniu.GetService().DeleteFile(record.AudioURL); err != nil {
			log.Printf("[SAFETY-TASK] 录音删除失败(次日重试) record=%d: %v", record.ID, err)
			continue
		}
		now := time.Now()
		if err := database.DB.Model(&models.ServiceRecord{}).
			Where("id = ?", record.ID).
			Updates(map[string]interface{}{
				"audio_deleted_at": &now,
			}).Error; err != nil {
			log.Printf("[SAFETY-TASK] 录音删除标记失败 record=%d: %v", record.ID, err)
			continue
		}
		cleaned++
	}
	return cleaned, nil
}

// StartServiceSafetyTasks 启动服务安全定时任务（阻塞式，需以 goroutine 调用）
func StartServiceSafetyTasks() {
	// 启动时先执行一次超时预警，保证重启后口径准确
	if n, err := RunServiceTimeoutAlert(); err != nil {
		log.Printf("[SAFETY-TASK] 启动超时预警失败: %v", err)
	} else if n > 0 {
		log.Printf("[SAFETY-TASK] 启动超时预警完成: %d 单", n)
	}

	// 超时预警：每 15 分钟
	timeoutTicker := time.NewTicker(15 * time.Minute)
	defer timeoutTicker.Stop()

	// 录音清理：每 24 小时
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-timeoutTicker.C:
			if n, err := RunServiceTimeoutAlert(); err != nil {
				log.Printf("[SAFETY-TASK] 超时预警失败: %v", err)
			} else if n > 0 {
				log.Printf("[SAFETY-TASK] 超时预警完成: %d 单", n)
			}
		case <-cleanupTicker.C:
			if n, err := RunServiceAudioCleanup(); err != nil {
				log.Printf("[SAFETY-TASK] 录音清理失败: %v", err)
			} else if n > 0 {
				log.Printf("[SAFETY-TASK] 录音清理完成: %d 条", n)
			}
		}
	}
}
