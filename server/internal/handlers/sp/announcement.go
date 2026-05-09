package sp

import (
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
	database.DB.Model(&models.Announcement{}).Count(&total)

	var announcements []models.Announcement
	offset := (page - 1) * pageSize
	if err := database.DB.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&announcements).Error; err != nil {
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
}

func CreateAnnouncement(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var sp models.ServiceProvider
	if err := database.DB.First(&sp).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务商信息失败")
		return
	}

	announcement := models.Announcement{
		ServiceProviderID: sp.ID,
		Title:          req.Title,
		Content:        req.Content,
		Status:         1,
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

	var announcement models.Announcement
	if err := database.DB.First(&announcement, announcementID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	updates := map[string]interface{}{
		"title":   req.Title,
		"content": req.Content,
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

	var announcement models.Announcement
	if err := database.DB.First(&announcement, announcementID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "公告不存在")
		return
	}

	if err := database.DB.Model(&announcement).Update("status", 0).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除公告失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}
