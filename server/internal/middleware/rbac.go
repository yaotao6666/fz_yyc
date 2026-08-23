package middleware

import (
	"errors"
	"sync"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 通用 RBAC 权限控制
// - owner 员工（merchant_staffs.role='owner'）为超管，直通所有权限校验
// - 普通员工按 员工->角色->菜单(权限标识) 判定，带 60s 内存缓存
// - 数据权限暂不处理（仅功能权限）
// ============================================================

type rbacCacheEntry struct {
	expiresAt time.Time
	perms     map[string]struct{}
}

var (
	rbacCacheMu sync.Mutex
	rbacCache   = map[uint64]rbacCacheEntry{}
	rbacCacheTTL = 60 * time.Second
)

// ClearRBACCache 清理全部权限缓存，角色/菜单/员工权限变更后调用
func ClearRBACCache() {
	rbacCacheMu.Lock()
	defer rbacCacheMu.Unlock()
	rbacCache = map[uint64]rbacCacheEntry{}
}

// GetCurrentStaff 获取当前登录的商家后台员工
func GetCurrentStaff(c *gin.Context) (*models.MerchantStaff, error) {
	// 优先按 token 中的员工ID 定位
	staffID := GetStaffID(c)
	if staffID > 0 {
		var staff models.MerchantStaff
		if err := database.DB.First(&staff, staffID).Error; err != nil {
			return nil, err
		}
		return &staff, nil
	}

	// 兼容旧 token（无 staff_id）：按 username 唯一索引定位
	username := GetUsername(c)
	if username == "" {
		return nil, errors.New("无法识别当前员工")
	}
	var staff models.MerchantStaff
	if err := database.DB.Where("username = ?", username).First(&staff).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// LoadStaffPermissions 加载员工的有效权限码集合（不含 owner 直通逻辑）
func LoadStaffPermissions(staffID uint64) (map[string]struct{}, error) {
	rbacCacheMu.Lock()
	if entry, ok := rbacCache[staffID]; ok && time.Now().Before(entry.expiresAt) {
		perms := entry.perms
		rbacCacheMu.Unlock()
		return perms, nil
	}
	rbacCacheMu.Unlock()

	var roleIDs []uint64
	if err := database.DB.Model(&models.MerchantStaffRole{}).
		Where("staff_id = ?", staffID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}

	perms := map[string]struct{}{}
	if len(roleIDs) == 0 {
		return perms, nil
	}

	var enabledRoleIDs []uint64
	if err := database.DB.Model(&models.SysRole{}).
		Where("id IN ? AND status = 1", roleIDs).
		Pluck("id", &enabledRoleIDs).Error; err != nil {
		return nil, err
	}
	if len(enabledRoleIDs) == 0 {
		return perms, nil
	}

	var menuIDs []uint64
	if err := database.DB.Model(&models.SysRoleMenu{}).
		Where("role_id IN ?", enabledRoleIDs).
		Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, err
	}
	if len(menuIDs) > 0 {
		var menus []models.SysMenu
		if err := database.DB.Where("id IN ? AND status = 1 AND permission <> ''", menuIDs).
			Find(&menus).Error; err != nil {
			return nil, err
		}
		for _, menu := range menus {
			if menu.Permission != "" {
				perms[menu.Permission] = struct{}{}
			}
		}
	}

	rbacCacheMu.Lock()
	rbacCache[staffID] = rbacCacheEntry{expiresAt: time.Now().Add(rbacCacheTTL), perms: perms}
	rbacCacheMu.Unlock()
	return perms, nil
}

// HasPermission 判断当前员工是否拥有指定权限（owner 超管直通）
func HasPermission(c *gin.Context, permission string) bool {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		return false
	}
	if staff.Role == "owner" {
		return true
	}
	perms, err := LoadStaffPermissions(staff.ID)
	if err != nil {
		return false
	}
	_, ok := perms[permission]
	return ok
}

// RBAC 权限校验中间件
func RBAC(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasPermission(c, permission) {
			c.Next()
			return
		}
		response.Forbidden(c, "无操作权限")
		c.Abort()
	}
}

// RBACAny 权限校验中间件（命中任一权限即放行）
func RBACAny(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, permission := range permissions {
			if HasPermission(c, permission) {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "无操作权限")
		c.Abort()
	}
}
