// Package systemconfig 通用系统配置服务。
//
// 承载 system_configs 表（key-value + JSON + 备注），供任意业务域复用。
// 它取代了原先 alert_settings 单行多列的结构：每个配置项独立成行
// （config_key + 序列化为 JSON 的 config_value + remark）。
package systemconfig

import (
	"encoding/json"
	"errors"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"

	"gorm.io/gorm"
)

// All 读取全量配置，返回 map[config_key]json.RawMessage。
func All() map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	var rows []models.SystemConfig
	if err := database.DB.Find(&rows).Error; err != nil {
		return m
	}
	for _, r := range rows {
		m[r.ConfigKey] = json.RawMessage(r.ConfigValue)
	}
	return m
}

// Get 读取单个 key 并反序列化到 dest；key 不存在或反序列化失败返回 false。
func Get(key string, dest any) bool {
	raw, ok := All()[key]
	if !ok {
		return false
	}
	return json.Unmarshal(raw, dest) == nil
}

// SetMany 批量写入/更新（按 config_key upsert，事务保证原子性）。
//   - values: 配置键 -> 任意可 JSON 序列化值
//   - remark: 新增行的备注（不覆盖已有行的备注）
//
// 空 map 时直接返回，不落库。
func SetMany(values map[string]any, remark string) error {
	if len(values) == 0 {
		return nil
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for key, val := range values {
			raw, err := json.Marshal(val)
			if err != nil {
				return err
			}
			value := string(raw)
			var c models.SystemConfig
			err = tx.Where("config_key = ?", key).First(&c).Error
			if err == nil {
				// 已存在 -> 更新值
				if err := tx.Model(&c).Updates(map[string]any{
					"config_value": value,
					"updated_at":   now,
				}).Error; err != nil {
					return err
				}
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// 不存在 -> 新增
			if err := tx.Create(&models.SystemConfig{
				ConfigKey:   key,
				ConfigValue: value,
				Remark:      remark,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}