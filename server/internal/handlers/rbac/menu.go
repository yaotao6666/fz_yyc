package rbac

import (
	"net/http"
	"sort"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// buildMenuTreeFull 构建完整菜单树（管理/分配权限用，不剔除空目录）
func buildMenuTreeFull(menus []models.SysMenu) []models.SysMenu {
	childrenMap := map[uint64][]models.SysMenu{}
	for _, menu := range menus {
		childrenMap[menu.ParentID] = append(childrenMap[menu.ParentID], menu)
	}

	var build func(parentID uint64) []models.SysMenu
	build = func(parentID uint64) []models.SysMenu {
		nodes := childrenMap[parentID]
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Sort == nodes[j].Sort {
				return nodes[i].ID < nodes[j].ID
			}
			return nodes[i].Sort < nodes[j].Sort
		})
		result := make([]models.SysMenu, 0, len(nodes))
		for index := range nodes {
			node := nodes[index]
			node.Children = build(node.ID)
			result = append(result, node)
		}
		return result
	}

	return build(0)
}

type MenuUpsertRequest struct {
	ParentID   *uint64 `json:"parent_id"`
	MenuType   *uint8  `json:"menu_type"`
	Name       string  `json:"name" binding:"required"`
	Path       string  `json:"path"`
	Icon       string  `json:"icon"`
	Sort       uint    `json:"sort"`
	Status     *uint8  `json:"status"`
	Visible    *uint8  `json:"visible"`
	Permission string  `json:"permission"`
}

func defaultUint8(value *uint8, fallback uint8) uint8 {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultUint64(value *uint64, fallback uint64) uint64 {
	if value == nil {
		return fallback
	}
	return *value
}

// GetMenuTree 获取完整菜单树（菜单管理/角色分配权限用）
func GetMenuTree(c *gin.Context) {
	var menus []models.SysMenu
	if err := database.DB.Order("sort ASC").Find(&menus).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取菜单失败")
		return
	}
	response.Success(c, buildMenuTreeFull(menus))
}

// CreateMenu 新增菜单
func CreateMenu(c *gin.Context) {
	var req MenuUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	parentID := defaultUint64(req.ParentID, 0)
	if parentID != 0 {
		var parent models.SysMenu
		if err := database.DB.First(&parent, parentID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父菜单不存在")
			return
		}
	}

	menu := models.SysMenu{
		ParentID:   parentID,
		MenuType:   defaultUint8(req.MenuType, 1),
		Name:       req.Name,
		Path:       req.Path,
		Icon:       req.Icon,
		Sort:       req.Sort,
		Status:     defaultUint8(req.Status, 1),
		Visible:    defaultUint8(req.Visible, 1),
		Permission: req.Permission,
	}

	if err := database.DB.Create(&menu).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建菜单失败")
		return
	}

	middleware.ClearRBACCache()
	response.Success(c, menu)
}

// UpdateMenu 更新菜单
func UpdateMenu(c *gin.Context) {
	menuID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var req MenuUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var menu models.SysMenu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "菜单不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.MenuType != nil {
		updates["menu_type"] = *req.MenuType
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["path"] = req.Path
	updates["icon"] = req.Icon
	updates["sort"] = req.Sort
	updates["permission"] = req.Permission
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Visible != nil {
		updates["visible"] = *req.Visible
	}

	if err := database.DB.Model(&menu).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新菜单失败")
		return
	}

	middleware.ClearRBACCache()
	database.DB.First(&menu, menuID)
	response.Success(c, menu)
}

// DeleteMenu 删除菜单（存在子菜单或被角色引用时拒绝）
func DeleteMenu(c *gin.Context) {
	menuID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var menu models.SysMenu
	if err := database.DB.First(&menu, menuID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "菜单不存在")
		return
	}

	var childCount int64
	database.DB.Model(&models.SysMenu{}).Where("parent_id = ?", menuID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "存在子菜单，无法删除")
		return
	}

	var refCount int64
	database.DB.Model(&models.SysRoleMenu{}).Where("menu_id = ?", menuID).Count(&refCount)
	if refCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "菜单已被角色引用，无法删除")
		return
	}

	if err := database.DB.Delete(&menu).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除菜单失败")
		return
	}

	middleware.ClearRBACCache()
	response.Success(c, gin.H{"message": "删除成功"})
}
