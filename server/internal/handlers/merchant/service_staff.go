package merchant

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetServiceStaffList 服务人员列表（PC 后台管理）
func GetServiceStaffList(c *gin.Context) {
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.ServiceStaff{})
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR phone LIKE ? OR username LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var staffs []models.ServiceStaff
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&staffs)

	response.Success(c, gin.H{
		"list":  staffs,
		"total": total,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateServiceStaffStatusRequest 更新服务人员状态请求
type UpdateServiceStaffStatusRequest struct {
	Status uint8 `json:"status" binding:"required"`
}

// UpdateServiceStaffStatus 审核通过(0→1) / 启用(2→1) / 禁用(1→2)
func UpdateServiceStaffStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "服务人员ID错误")
		return
	}

	var req UpdateServiceStaffStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if req.Status != 1 && req.Status != 2 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "状态值不合法")
		return
	}

	result := database.DB.Model(&models.ServiceStaff{}).
		Where("id = ?", id).
		Update("status", req.Status)
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "服务人员不存在")
		return
	}

	response.SuccessWithMessage(c, "状态更新成功", gin.H{"id": id, "status": req.Status})
}

// ResetServiceStaffPasswordRequest 重置密码请求
type ResetServiceStaffPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetServiceStaffPassword 商家重置服务人员密码
func ResetServiceStaffPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "服务人员ID错误")
		return
	}

	var req ResetServiceStaffPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	result := database.DB.Model(&models.ServiceStaff{}).
		Where("id = ?", id).
		Update("password", string(hashedPassword))
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "服务人员不存在")
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", gin.H{"id": id})
}

// DeleteServiceStaff 删除服务人员
func DeleteServiceStaff(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "服务人员ID错误")
		return
	}

	result := database.DB.Where("id = ?", id).Delete(&models.ServiceStaff{})
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "服务人员不存在")
		return
	}

	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}
