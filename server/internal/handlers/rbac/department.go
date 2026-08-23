package rbac

import (
	"net/http"
	"sort"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type DepartmentUpsertRequest struct {
	ParentID *uint64 `json:"parent_id"`
	Name     string  `json:"name" binding:"required"`
	Leader   string  `json:"leader"`
	Phone    string  `json:"phone"`
	Sort     uint    `json:"sort"`
	Status   *uint8  `json:"status"`
}

// buildDepartmentTree 构建部门树
func buildDepartmentTree(departments []models.SysDepartment) []models.SysDepartment {
	childrenMap := map[uint64][]models.SysDepartment{}
	for _, dept := range departments {
		childrenMap[dept.ParentID] = append(childrenMap[dept.ParentID], dept)
	}

	var build func(parentID uint64) []models.SysDepartment
	build = func(parentID uint64) []models.SysDepartment {
		nodes := childrenMap[parentID]
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Sort == nodes[j].Sort {
				return nodes[i].ID < nodes[j].ID
			}
			return nodes[i].Sort < nodes[j].Sort
		})
		result := make([]models.SysDepartment, 0, len(nodes))
		for index := range nodes {
			node := nodes[index]
			node.Children = build(node.ID)
			result = append(result, node)
		}
		return result
	}

	return build(0)
}

// GetDepartmentTree 获取部门树
func GetDepartmentTree(c *gin.Context) {
	var departments []models.SysDepartment
	if err := database.DB.Order("sort ASC").Find(&departments).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取部门列表失败")
		return
	}
	response.Success(c, buildDepartmentTree(departments))
}

// CreateDepartment 新增部门
func CreateDepartment(c *gin.Context) {
	var req DepartmentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	parentID := defaultUint64(req.ParentID, 0)
	if parentID != 0 {
		var parent models.SysDepartment
		if err := database.DB.First(&parent, parentID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父部门不存在")
			return
		}
	}

	department := models.SysDepartment{
		ParentID: parentID,
		Name:     req.Name,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Sort:     req.Sort,
		Status:   defaultUint8(req.Status, 1),
	}

	if err := database.DB.Create(&department).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建部门失败")
		return
	}

	response.Success(c, department)
}

// UpdateDepartment 更新部门
func UpdateDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var req DepartmentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var department models.SysDepartment
	if err := database.DB.First(&department, departmentID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "部门不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["leader"] = req.Leader
	updates["phone"] = req.Phone
	updates["sort"] = req.Sort
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := database.DB.Model(&department).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新部门失败")
		return
	}

	database.DB.First(&department, departmentID)
	response.Success(c, department)
}

// DeleteDepartment 删除部门（存在子部门或关联员工时拒绝）
func DeleteDepartment(c *gin.Context) {
	departmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var department models.SysDepartment
	if err := database.DB.First(&department, departmentID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "部门不存在")
		return
	}

	var childCount int64
	database.DB.Model(&models.SysDepartment{}).Where("parent_id = ?", departmentID).Count(&childCount)
	if childCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "存在子部门，无法删除")
		return
	}

	var staffCount int64
	database.DB.Model(&models.MerchantStaff{}).Where("department_id = ?", departmentID).Count(&staffCount)
	if staffCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "部门下存在员工，无法删除")
		return
	}

	if err := database.DB.Delete(&department).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除部门失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}
