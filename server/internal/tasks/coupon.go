// Package tasks 定时任务（PRD V2.0 阶段二：优惠券营销闭环）。
// 轻量实现：进程内 goroutine + time.Ticker，无需外部依赖。
//  - 每小时：用户券过期状态刷新（status 1→3）
//  - 每日：30 天未下单用户自动发券（source=2，按用户+模板去重）
package tasks

import (
	"log"
	"time"

	"fz_yyc_api/internal/services/coupon"
)

// StartCouponTasks 启动优惠券定时任务（阻塞式，需以 goroutine 调用）
func StartCouponTasks() {
	// 启动时先执行一次过期刷新，保证重启后口径准确
	if n, err := coupon.RefreshExpired(); err != nil {
		log.Printf("[COUPON-TASK] 启动过期刷新失败: %v", err)
	} else if n > 0 {
		log.Printf("[COUPON-TASK] 启动过期刷新完成: %d 张券置为已过期", n)
	}

	// 过期刷新：每小时
	refreshTicker := time.NewTicker(time.Hour)
	defer refreshTicker.Stop()

	// 唤回发券：每 24 小时
	grantTicker := time.NewTicker(24 * time.Hour)
	defer grantTicker.Stop()

	for {
		select {
		case <-refreshTicker.C:
			if n, err := coupon.RefreshExpired(); err != nil {
				log.Printf("[COUPON-TASK] 过期刷新失败: %v", err)
			} else if n > 0 {
				log.Printf("[COUPON-TASK] 过期刷新完成: %d 张券置为已过期", n)
			}
		case <-grantTicker.C:
			if n, err := coupon.GrantToInactiveUsers(); err != nil {
				log.Printf("[COUPON-TASK] 30天唤回发券失败: %v", err)
			} else if n > 0 {
				log.Printf("[COUPON-TASK] 30天唤回发券完成: %d 位用户", n)
			}
		}
	}
}
