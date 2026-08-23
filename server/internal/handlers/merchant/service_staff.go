package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateServiceStaffRequest PC 后台添加服务人员请求
type CreateServiceStaffRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

// CreateServiceStaff PC 后台直接添加服务人员（默认启用，区别于小程序自注册的待审核）
func CreateServiceStaff(c *gin.Context) {
	var req CreateServiceStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var count int64
	database.DB.Model(&models.ServiceStaff{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "用户名已存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	staff := models.ServiceStaff{
		Username: req.Username,
		Password: string(hashedPassword),
		Name:     req.Name,
		Phone:    req.Phone,
		Status:   1, // 后台直接添加默认为启用
	}

	if err := database.DB.Create(&staff).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "添加失败")
		return
	}

	// PC 添加留痕：写一条已通过的注册申请记录（admin 直接添加默认启用）
	after, _ := json.Marshal(map[string]interface{}{
		"username": req.Username,
		"name":     req.Name,
		"phone":    req.Phone,
	})
	database.DB.Create(&models.StaffAuditRecord{
		StaffID:   staff.ID,
		AuditType: 1, // 注册申请
		ApplyType: 2, // PC 添加
		AfterData: models.JSON(after),
		Status:    1, // 直接通过
	})

	response.SuccessWithMessage(c, "添加成功", gin.H{"id": staff.ID})
}

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

	// 状态变更留痕（审核记录）
	after, _ := json.Marshal(map[string]interface{}{
		"status": req.Status,
	})
	database.DB.Model(&models.StaffAuditRecord{}).Create(&models.StaffAuditRecord{
		StaffID:   id,
		AuditType: 4, // 状态变更
		ApplyType: 2, // 管理员操作
		AfterData: models.JSON(after),
		Status:    1, // 管理员直接操作，视为已生效
	})

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
