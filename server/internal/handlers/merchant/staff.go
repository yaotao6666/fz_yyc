package merchant

import (
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetStaffList(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	database.DB.Model(&models.MerchantStaff{}).Where("merchant_id = ?", merchantID).Count(&total)

	var staffList []models.MerchantStaff
	offset := (page - 1) * pageSize
	if err := database.DB.Where("merchant_id = ?", merchantID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&staffList).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取员工列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": staffList,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type CreateStaffRequest struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

func CreateStaff(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var existCount int64
	database.DB.Model(&models.MerchantStaff{}).Where("merchant_id = ? AND (username = ? OR phone = ?)", merchantID, req.Username, req.Phone).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "用户名或手机号已存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	role := req.Role
	if role == "" {
		role = "staff"
	}

	staff := models.MerchantStaff{
		MerchantID:          merchantID,
		Name:                req.Name,
		Phone:               req.Phone,
		Username:            req.Username,
		Password:            string(hashedPassword),
		Role:                role,
		NotifyEnabled:       true,
		BrowseNotifyEnabled: true,
		Status:              1,
	}

	if err := database.DB.Create(&staff).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建员工失败")
		return
	}

	response.Success(c, gin.H{"id": staff.ID, "message": "创建成功"})
}

type UpdateStaffRequest struct {
	Name                string `json:"name"`
	Phone               string `json:"phone"`
	Role                string `json:"role"`
	NotifyEnabled       *bool  `json:"notify_enabled"`
	BrowseNotifyEnabled *bool  `json:"browse_notify_enabled"`
	Status              *uint8 `json:"status"`
}

func UpdateStaff(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	staffID := c.Param("id")
	id, _ := strconv.ParseUint(staffID, 10, 64)

	var req UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var staff models.MerchantStaff
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "员工不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.NotifyEnabled != nil {
		updates["notify_enabled"] = *req.NotifyEnabled
	}
	if req.BrowseNotifyEnabled != nil {
		updates["browse_notify_enabled"] = *req.BrowseNotifyEnabled
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.Role != "" && staff.Role == "owner" && req.Role != "owner" {
		var ownerCount int64
		database.DB.Model(&models.MerchantStaff{}).Where("merchant_id = ? AND role = ? AND status = ?", merchantID, "owner", 1).Count(&ownerCount)
		if ownerCount <= 1 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "至少保留一个店铺负责人")
			return
		}
	}

	if err := database.DB.Model(&staff).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新员工失败")
		return
	}

	database.DB.First(&staff, id)
	response.Success(c, staff)
}

func DeleteStaff(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	usernameValue, _ := c.Get("username")
	currentUsername, _ := usernameValue.(string)
	staffID := c.Param("id")
	id, _ := strconv.ParseUint(staffID, 10, 64)

	var staff models.MerchantStaff
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "员工不存在")
		return
	}

	if staff.Username == currentUsername {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "不能删除当前登录账号")
		return
	}

	if staff.Role == "owner" {
		var ownerCount int64
		database.DB.Model(&models.MerchantStaff{}).Where("merchant_id = ? AND role = ? AND status = ?", merchantID, "owner", 1).Count(&ownerCount)
		if ownerCount <= 1 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "至少保留一个店铺负责人")
			return
		}
	}

	if err := database.DB.Delete(&staff).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除员工失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

type ResetStaffPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func ResetStaffPassword(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	staffID := c.Param("id")
	id, _ := strconv.ParseUint(staffID, 10, 64)

	var req ResetStaffPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var staff models.MerchantStaff
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "员工不存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	if err := database.DB.Model(&models.MerchantStaff{}).Where("id = ?", staff.ID).Update("password", string(hashedPassword)).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "重置密码失败")
		return
	}

	response.Success(c, gin.H{"message": "重置密码成功"})
}

func GetAnnouncements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	database.DB.Model(&models.Announcement{}).Where("status = ?", 1).Count(&total)

	var announcements []models.Announcement
	offset := (page - 1) * pageSize
	if err := database.DB.Where("status = ?", 1).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&announcements).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取公告列表失败")
		return
	}

	var sp models.ServiceProvider
	database.DB.First(&sp)

	response.Success(c, gin.H{
		"list": announcements,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func GetAnnouncementDetail(c *gin.Context) {
	announcementID := c.Param("id")
	id, _ := strconv.ParseUint(announcementID, 10, 64)

	var announcement models.Announcement
	if err := database.DB.Where("id = ? AND status = ?", id, 1).First(&announcement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	response.Success(c, announcement)
}
