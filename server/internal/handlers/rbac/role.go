package rbac

import (
	"net/http"
	"strconv"
	"strings"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleUpsertRequest struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Remark string `json:"remark"`
	Status *uint8 `json:"status"`
}

// GetRoleList 分页获取角色列表
func GetRoleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	status := strings.TrimSpace(c.Query("status"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.SysRole{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}
	if status == "0" || status == "1" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var roles []models.SysRole
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&roles).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取角色列表失败")
		return
	}

	// 附带每个角色的菜单绑定数量
	roleMenuCounts := map[uint64]int64{}
	var roleIDs []uint64
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	if len(roleIDs) > 0 {
		type countRow struct {
			RoleID uint64
			Cnt    int64
		}
		var rows []countRow
		database.DB.Model(&models.SysRoleMenu{}).
			Select("role_id, COUNT(*) AS cnt").
			Where("role_id IN ?", roleIDs).
			Group("role_id").
			Scan(&rows)
		for _, row := range rows {
			roleMenuCounts[row.RoleID] = row.Cnt
		}
	}

	type roleItem struct {
		models.SysRole
		MenuCount int64 `json:"menu_count"`
	}
	list := make([]roleItem, 0, len(roles))
	for _, role := range roles {
		list = append(list, roleItem{SysRole: role, MenuCount: roleMenuCounts[role.ID]})
	}

	response.Success(c, gin.H{
		"list": list,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAllRoles 获取全部启用角色（员工管理下拉选择用）
func GetAllRoles(c *gin.Context) {
	var roles []models.SysRole
	if err := database.DB.Where("status = 1").Order("id ASC").Find(&roles).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取角色列表失败")
		return
	}
	response.Success(c, roles)
}

// CreateRole 新增角色
func CreateRole(c *gin.Context) {
	var req RoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	role := models.SysRole{
		Name:   req.Name,
		Code:   strings.TrimSpace(req.Code),
		Remark: req.Remark,
		Status: defaultUint8(req.Status, 1),
	}

	var existCount int64
	database.DB.Model(&models.SysRole{}).Where("code = ?", role.Code).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "角色编码已存在")
		return
	}

	if err := database.DB.Create(&role).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建角色失败")
		return
	}

	middleware.ClearRBACCache()
	response.Success(c, role)
}

// UpdateRole 更新角色
func UpdateRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var req RoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var role models.SysRole
	if err := database.DB.First(&role, roleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "角色不存在")
		return
	}

	newCode := strings.TrimSpace(req.Code)
	if newCode != "" && newCode != role.Code {
		var existCount int64
		database.DB.Model(&models.SysRole{}).Where("code = ? AND id <> ?", newCode, roleID).Count(&existCount)
		if existCount > 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "角色编码已存在")
			return
		}
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if newCode != "" {
		updates["code"] = newCode
	}
	updates["remark"] = req.Remark
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新角色失败")
		return
	}

	middleware.ClearRBACCache()
	database.DB.First(&role, roleID)
	response.Success(c, role)
}

// DeleteRole 删除角色（被员工引用时拒绝）
func DeleteRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var role models.SysRole
	if err := database.DB.First(&role, roleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "角色不存在")
		return
	}

	var staffRefCount int64
	database.DB.Model(&models.MerchantStaffRole{}).Where("role_id = ?", roleID).Count(&staffRefCount)
	if staffRefCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "角色已被员工绑定，无法删除")
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	}); err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除角色失败")
		return
	}

	middleware.ClearRBACCache()
	response.Success(c, gin.H{"message": "删除成功"})
}

// GetRoleMenus 获取角色已绑定的菜单ID列表（分配权限回显）
func GetRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var role models.SysRole
	if err := database.DB.First(&role, roleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "角色不存在")
		return
	}

	var menuIDs []uint64
	database.DB.Model(&models.SysRoleMenu{}).
		Where("role_id = ?", roleID).
		Order("menu_id ASC").
		Pluck("menu_id", &menuIDs)

	response.Success(c, gin.H{"role_id": roleID, "menu_ids": menuIDs})
}

type AssignRoleMenusRequest struct {
	MenuIDs []uint64 `json:"menu_ids"`
}

// AssignRoleMenus 覆盖式分配角色菜单
func AssignRoleMenus(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var req AssignRoleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var role models.SysRole
	if err := database.DB.First(&role, roleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "角色不存在")
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&models.SysRoleMenu{}).Error; err != nil {
			return err
		}
		if len(req.MenuIDs) == 0 {
			return nil
		}
		rows := make([]models.SysRoleMenu, 0, len(req.MenuIDs))
		seen := map[uint64]struct{}{}
		for _, menuID := range req.MenuIDs {
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			rows = append(rows, models.SysRoleMenu{RoleID: roleID, MenuID: menuID})
		}
		return tx.Create(&rows).Error
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "分配权限失败")
		return
	}

	middleware.ClearRBACCache()
	response.Success(c, gin.H{"message": "分配成功"})
}
