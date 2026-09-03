package merchant

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 系统配置模型：config_key(value-key + JSON 键标识) -> config_value(JSON 字符串) + remark(备注)
// 全局用量：keyPattern 约束 config_key 合法字符集（字母数字及 _ . -），供读写删共用
var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9_.\-]{1,100}$`)

// ============================================
// 通用系统配置(system_configs) — key-value + JSON + 备注
// 用途：PC 后台「系统管理 > 系统配置」页面对 expose 通用配置项，
// 支持查看、编辑、新增（按 config_key 幂等 upsert）。
// ============================================

// SystemConfigItem 系统配置项（config_value 为前端序列化好的 JSON 字符串）
type SystemConfigItem struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	Remark      string `json:"remark"`
}

// ListSystemConfigs 获取全部系统配置（按 config_key 升序）
func ListSystemConfigs(c *gin.Context) {
	var list []models.SystemConfig
	if err := database.DB.Order("config_key ASC").Find(&list).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "读取系统配置失败")
		return
	}
	response.Success(c, list)
}

// UpdateSystemConfigRequest 批量更新/新增系统配置
// dive 使校验递归进入每个元素，强制 config_key/config_value 非空
type UpdateSystemConfigRequest struct {
	Items []SystemConfigItem `json:"items" binding:"required,min=1,dive"`
}

// UpdateSystemConfig 批量更新/新增系统配置（config_key 幂等 upsert，事务原子）
func UpdateSystemConfig(c *gin.Context) {
	var req UpdateSystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// config_key 仅允许字母数字及 _ . - ，长度 <= 100（keyPattern 为包级定义）
	now := time.Now()

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, it := range req.Items {
			if !keyPattern.MatchString(it.ConfigKey) {
				return gorm.ErrInvalidField
			}
			var c models.SystemConfig
			e := tx.Where("config_key = ?", it.ConfigKey).First(&c).Error
			if e == nil {
				if err := tx.Model(&c).Updates(map[string]any{
					"config_value": it.ConfigValue,
					"remark":       it.Remark,
					"updated_at":   now,
				}).Error; err != nil {
					return err
				}
				continue
			}
			if !errors.Is(e, gorm.ErrRecordNotFound) {
				return e
			}
			if err := tx.Create(&models.SystemConfig{
				ConfigKey:   it.ConfigKey,
				ConfigValue: it.ConfigValue,
				Remark:      it.Remark,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "config_key 仅允许字母数字及 _ . -")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存系统配置失败")
		return
	}

	response.SuccessWithMessage(c, "保存成功", nil)
}

// DeleteSystemConfig 删除指定配置项（config_key）
func DeleteSystemConfig(c *gin.Context) {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" || !keyPattern.MatchString(key) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "无效的配置键")
		return
	}
	res := database.DB.Where("config_key = ?", key).Delete(&models.SystemConfig{})
	if res.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除系统配置失败")
		return
	}
	if res.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "配置项不存在")
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}