package rbac

import (
	"net/http"
	"sort"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// buildMenuTree 将扁平菜单列表构建为树形结构，并剔除无子节点的空目录
func buildMenuTree(menus []models.SysMenu) []models.SysMenu {
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
			// 目录节点（无权限、菜单类型）若无可见子菜单则隐藏
			if node.MenuType == 1 && node.Permission == "" && len(node.Children) == 0 {
				continue
			}
			result = append(result, node)
		}
		return result
	}

	return build(0)
}

// BuildStaffMenusAndPermissions 计算员工可见菜单树与权限码集合（owner 返回全部）
func BuildStaffMenusAndPermissions(staff *models.MerchantStaff) ([]models.SysMenu, []string, error) {
	var all []models.SysMenu
	if err := database.DB.Where("status = 1").Order("sort ASC").Find(&all).Error; err != nil {
		return nil, nil, err
	}

	granted := map[string]struct{}{}
	if staff.Role != "owner" {
		perms, err := middleware.LoadStaffPermissions(staff.ID)
		if err != nil {
			return nil, nil, err
		}
		granted = perms
	}

	visible := make([]models.SysMenu, 0, len(all))
	permCodes := make([]string, 0, len(all))
	for _, menu := range all {
		if menu.Visible != 1 {
			continue
		}
		if staff.Role == "owner" {
			visible = append(visible, menu)
			if menu.Permission != "" {
				permCodes = append(permCodes, menu.Permission)
			}
			continue
		}
		// 非 owner：目录（无权限标识）或有权限的菜单才可见
		if menu.MenuType == 1 && (menu.Permission == "" || hasPermission(granted, menu.Permission)) {
			visible = append(visible, menu)
		}
		if menu.Permission != "" && hasPermission(granted, menu.Permission) {
			permCodes = append(permCodes, menu.Permission)
		}
	}

	tree := buildMenuTree(visible)
	sort.Strings(permCodes)
	return tree, permCodes, nil
}

func hasPermission(perms map[string]struct{}, code string) bool {
	_, ok := perms[code]
	return ok
}

// GetMyMenus 获取当前登录员工可见菜单树与权限码集合（登录后刷新权限用）
func GetMyMenus(c *gin.Context) {
	staff, err := middleware.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工身份失败")
		return
	}

	menus, permissions, err := BuildStaffMenusAndPermissions(staff)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "加载权限失败")
		return
	}

	response.Success(c, gin.H{
		"menus":       menus,
		"permissions": permissions,
	})
}
