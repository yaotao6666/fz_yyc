package admin

import (
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetActivities(c *gin.Context) {
	activityType := c.Query("type")

	query := database.DB.Model(&models.Activity{})

	if activityType != "" {
		query = query.Where("type = ?", activityType)
	}

	var activities []models.Activity
	if err := query.Where("status = ? AND (start_time IS NULL OR start_time <= ?) AND (end_time IS NULL OR end_time >= ?)",
		1, time.Now(), time.Now()).
		Order("sort ASC, created_at DESC").
		Find(&activities).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取活动列表失败")
		return
	}

	var banners []models.Activity
	var announcements []models.Activity
	for _, activity := range activities {
		if activity.Type == "banner" {
			banners = append(banners, activity)
		} else if activity.Type == "announcement" {
			announcements = append(announcements, activity)
		}
	}

	response.Success(c, gin.H{
		"banners":       banners,
		"announcements": announcements,
	})
}

type ActivityRequest struct {
	Type      string     `json:"type" binding:"required"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Image     string     `json:"image"`
	LinkType  string     `json:"link_type"`
	LinkValue string     `json:"link_value"`
	Sort      uint       `json:"sort"`
	Status    uint8      `json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

func CreateActivity(c *gin.Context) {
	var req ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if req.Type == "banner" && req.Title == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "Banner标题不能为空")
		return
	}

	activity := models.Activity{
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Image:     req.Image,
		LinkType:  req.LinkType,
		LinkValue: req.LinkValue,
		Sort:      req.Sort,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if activity.Status == 0 {
		activity.Status = 1
	}

	if err := database.DB.Create(&activity).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建活动失败")
		return
	}

	response.Success(c, activity)
}

func UpdateActivity(c *gin.Context) {
	id := c.Param("id")
	activityID, _ := strconv.ParseUint(id, 10, 64)

	var req ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var activity models.Activity
	if err := database.DB.First(&activity, activityID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "活动不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Image != "" {
		updates["image"] = req.Image
	}
	if req.LinkType != "" {
		updates["link_type"] = req.LinkType
	}
	if req.LinkValue != "" {
		updates["link_value"] = req.LinkValue
	}
	updates["sort"] = req.Sort
	updates["status"] = req.Status
	updates["start_time"] = req.StartTime
	updates["end_time"] = req.EndTime

	if err := database.DB.Model(&activity).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新活动失败")
		return
	}

	database.DB.First(&activity, activityID)
	response.Success(c, activity)
}

func DeleteActivity(c *gin.Context) {
	id := c.Param("id")
	activityID, _ := strconv.ParseUint(id, 10, 64)

	if err := database.DB.Delete(&models.Activity{}, activityID).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除活动失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func GetInviteStats(c *gin.Context) {
	var totalInvites int64
	var completedInvites int64
	var pendingInvites int64
	var totalRewards int64

	database.DB.Model(&models.InviteRecord{}).Count(&totalInvites)
	database.DB.Model(&models.InviteRecord{}).Where("status = ?", 1).Count(&completedInvites)
	database.DB.Model(&models.InviteRecord{}).Where("status = ?", 0).Count(&pendingInvites)
	database.DB.Model(&models.InviteRecord{}).Where("reward_status = ?", 1).Count(&totalRewards)

	var inviteTrend []struct {
		Month   string `json:"month"`
		Invites int64  `json:"invites"`
	}

	database.DB.Table("invite_records").
		Select("DATE_FORMAT(created_at, '%Y-%m') as month, COUNT(*) as invites").
		Group("month").
		Order("month DESC").
		Limit(6).
		Scan(&inviteTrend)

	response.Success(c, gin.H{
		"total_invites":      totalInvites,
		"completed_invites":  completedInvites,
		"pending_invites":    pendingInvites,
		"total_rewards":      totalRewards,
		"invite_trend":       inviteTrend,
	})
}

func GetInviteRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.InviteRecord{}).Preload("Inviter").Preload("Invitee")

	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}

	var total int64
	query.Count(&total)

	var records []models.InviteRecord
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取邀请记录失败")
		return
	}

	response.Success(c, gin.H{
		"list": records,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type InviteRewardRequest struct {
	Enabled bool `json:"enabled"`
	Rewards []struct {
		Type        string `json:"type"`
		Condition   string `json:"condition"`
		Description string `json:"description"`
	} `json:"rewards"`
}

func UpdateInviteRewards(c *gin.Context) {
	var req InviteRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	tx := database.DB.Begin()

	database.DB.Where("1=1").Delete(&models.InviteReward{})

	for _, reward := range req.Rewards {
		inviteReward := models.InviteReward{
			Type:        reward.Type,
			Condition:   reward.Condition,
			Description: reward.Description,
			Enabled:     req.Enabled,
		}
		if err := tx.Create(&inviteReward).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存奖励规则失败")
			return
		}
	}

	tx.Commit()

	var rewards []models.InviteReward
	database.DB.Find(&rewards)

	response.Success(c, gin.H{
		"enabled": req.Enabled,
		"rewards": rewards,
	})
}

func GetInviteRewards(c *gin.Context) {
	var rewards []models.InviteReward
	if err := database.DB.Find(&rewards).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取奖励规则失败")
		return
	}

	enabled := true
	if len(rewards) > 0 {
		enabled = rewards[0].Enabled
	}

	response.Success(c, gin.H{
		"enabled": enabled,
		"rewards": rewards,
	})
}
