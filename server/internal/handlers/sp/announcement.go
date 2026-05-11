package sp

import (
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getCurrentServiceProviderID(c *gin.Context) (uint64, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	adminID, ok := userIDValue.(uint64)
	if !ok {
		return 0, false
	}

	var admin models.ServiceProviderAdmin
	if err := database.DB.Select("service_provider_id").First(&admin, adminID).Error; err != nil {
		return 0, false
	}
	return admin.ServiceProviderID, true
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

	spID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var total int64
	database.DB.Model(&models.Announcement{}).Where("service_provider_id = ?", spID).Count(&total)

	var announcements []models.Announcement
	offset := (page - 1) * pageSize
	if err := database.DB.Where("service_provider_id = ?", spID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&announcements).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取公告列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": announcements,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Status  uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

func GetAnnouncementDetail(c *gin.Context) {
	id := c.Param("id")
	announcementID, _ := strconv.ParseUint(id, 10, 64)

	spID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var announcement models.Announcement
	if err := database.DB.Where("id = ? AND service_provider_id = ?", announcementID, spID).First(&announcement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	response.Success(c, announcement)
}

func CreateAnnouncement(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	spID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	status := uint8(1)
	if req.Status == 0 {
		status = 0
	}

	announcement := models.Announcement{
		ServiceProviderID: spID,
		Title:             req.Title,
		Content:           req.Content,
		Status:            status,
	}

	if err := database.DB.Create(&announcement).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建公告失败")
		return
	}

	response.Success(c, gin.H{"id": announcement.ID, "message": "创建成功"})
}

func UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")
	announcementID, _ := strconv.ParseUint(id, 10, 64)

	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	spID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var announcement models.Announcement
	if err := database.DB.Where("id = ? AND service_provider_id = ?", announcementID, spID).First(&announcement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	updates := map[string]interface{}{
		"title":   req.Title,
		"content": req.Content,
		"status":  req.Status,
	}

	if err := database.DB.Model(&announcement).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新公告失败")
		return
	}

	database.DB.First(&announcement, announcementID)
	response.Success(c, announcement)
}

func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	announcementID, _ := strconv.ParseUint(id, 10, 64)

	spID, ok := getCurrentServiceProviderID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "未登录")
		return
	}

	var announcement models.Announcement
	if err := database.DB.Where("id = ? AND service_provider_id = ?", announcementID, spID).First(&announcement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	if err := database.DB.Model(&announcement).Update("status", 0).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除公告失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}
