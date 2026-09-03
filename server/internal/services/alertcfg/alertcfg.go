// Package alertcfg 预警配置聚合层。
//
// 预警相关阈值在通用表 system_configs 中存为多条 key-value 配置项，
// 本包负责把散落的配置键聚合为 typed 的 models.AlertSettings 视图模型，
// 并定义全部配置键常量，避免散落在各业务文件。
package alertcfg

import (
	"encoding/json"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/internal/services/systemconfig"
)

// 配置键常量。可按业务域前缀命名（alert.* / service.*），易于归类与扩展。
const (
	KeyEnabled                 = "alert.enabled"
	KeyGoodsUnverifiedHours    = "alert.goods_unverified_hours"
	KeyServiceUnassignedHours  = "alert.service_unassigned_hours"
	KeyEscortUnfinishedMinutes = "alert.escort_unfinished_minutes"
	KeyServiceUnstartedMinutes = "alert.service_unstarted_minutes"
	KeyRentalOverdueHours      = "alert.rental_overdue_hours"
	KeyRefundStuckHours        = "alert.refund_stuck_hours"
	KeyServiceAudioRetainDays  = "service.audio_retain_days"

	// defaultRemark 写入配置时的备注说明
	remark = "预警/服务安全配置"
)

// Load 从通用系统配置聚合出预警配置视图模型，缺省时回退内置默认值。
func Load() models.AlertSettings {
	all := systemconfig.All()
	s := models.AlertSettings{
		ID:                      1,
		Enabled:                 true,
		GoodsUnverifiedHours:    utils.AlertDefaultGoodsUnverifiedHours,
		ServiceUnassignedHours:  utils.AlertDefaultServiceUnassignedHours,
		EscortUnfinishedMinutes: utils.AlertDefaultEscortUnfinishedMinutes,
		ServiceUnstartedMinutes: utils.AlertDefaultServiceUnstartedMinutes,
		RentalOverdueHours:      utils.AlertDefaultRentalOverdueHours,
		RefundStuckHours:        utils.AlertDefaultRefundStuckHours,
		ServiceAudioRetainDays:  utils.AlertDefaultServiceAudioRetainDays,
	}
	s.Enabled = boolOf(all, KeyEnabled, s.Enabled)
	s.GoodsUnverifiedHours = intOf(all, KeyGoodsUnverifiedHours, s.GoodsUnverifiedHours)
	s.ServiceUnassignedHours = intOf(all, KeyServiceUnassignedHours, s.ServiceUnassignedHours)
	s.EscortUnfinishedMinutes = intOf(all, KeyEscortUnfinishedMinutes, s.EscortUnfinishedMinutes)
	s.ServiceUnstartedMinutes = intOf(all, KeyServiceUnstartedMinutes, s.ServiceUnstartedMinutes)
	s.RentalOverdueHours = intOf(all, KeyRentalOverdueHours, s.RentalOverdueHours)
	s.RefundStuckHours = intOf(all, KeyRefundStuckHours, s.RefundStuckHours)
	s.ServiceAudioRetainDays = intOf(all, KeyServiceAudioRetainDays, s.ServiceAudioRetainDays)
	return s
}

// Save 将更新项持久化到通用系统配置。
func Save(updates map[string]any) error {
	return systemconfig.SetMany(updates, remark)
}

func intOf(all map[string]json.RawMessage, key string, def int) int {
	if raw, ok := all[key]; ok {
		var v int
		if json.Unmarshal(raw, &v) == nil {
			return v
		}
	}
	return def
}

func boolOf(all map[string]json.RawMessage, key string, def bool) bool {
	if raw, ok := all[key]; ok {
		var v bool
		if json.Unmarshal(raw, &v) == nil {
			return v
		}
	}
	return def
}