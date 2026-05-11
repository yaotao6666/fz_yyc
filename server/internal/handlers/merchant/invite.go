package merchant

import (
	"fmt"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateInviteCode(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var existingRecord models.InviteRecord
	if err := database.DB.Where("inviter_id = ? AND status = ?", merchantID, 0).First(&existingRecord).Error; err == nil {
		response.Success(c, gin.H{
			"invite_code": existingRecord.InviteCode,
			"qrcode_url":  fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=TOKEN&path=pages/invite/invite?code=%s", existingRecord.InviteCode),
			"share_link":  fmt.Sprintf("https://example.com/invite/%s", existingRecord.InviteCode),
			"expire_time": nil,
		})
		return
	}

	inviteCode := fmt.Sprintf("XM%s%s", time.Now().Format("20060102"), generateRandomString(6))

	record := models.InviteRecord{
		InviterID:  merchantID,
		InviteCode: inviteCode,
		Status:     0,
	}

	if err := database.DB.Create(&record).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "生成邀请码失败")
		return
	}

	response.Success(c, gin.H{
		"invite_code": inviteCode,
		"qrcode_url":  fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=TOKEN&path=pages/invite/invite?code=%s", inviteCode),
		"share_link":  fmt.Sprintf("https://example.com/invite/%s", inviteCode),
		"expire_time": nil,
	})
}

func generateRandomString(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[int(time.Now().UnixNano()+int64(i))%len(chars)]
	}
	return string(result)
}

func GetInviteInfo(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var record models.InviteRecord
	if err := database.DB.Where("inviter_id = ? AND status = ?", merchantID, 0).First(&record).Error; err != nil {
		record.InviteCode = ""
	}

	var totalInvites int64
	var completedInvites int64
	var pendingInvites int64

	database.DB.Model(&models.InviteRecord{}).Where("inviter_id = ?", merchantID).Count(&totalInvites)
	database.DB.Model(&models.InviteRecord{}).Where("inviter_id = ? AND status = ?", merchantID, 1).Count(&completedInvites)
	database.DB.Model(&models.InviteRecord{}).Where("inviter_id = ? AND status = ?", merchantID, 0).Count(&pendingInvites)

	var freeYearCount int64
	database.DB.Model(&models.InviteRecord{}).
		Where("inviter_id = ? AND reward_type = ? AND reward_status = ?", merchantID, "free_year", 1).
		Count(&freeYearCount)

	var lowestRateQualified bool
	if completedInvites > 0 {
		lowestRateQualified = true
	}

	response.Success(c, gin.H{
		"invite_code":       record.InviteCode,
		"total_invites":     totalInvites,
		"completed_invites": completedInvites,
		"pending_invites":   pendingInvites,
		"rewards": gin.H{
			"free_year_count":       freeYearCount,
			"lowest_rate_qualified": lowestRateQualified,
		},
	})
}

func GetInviteRecords(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.InviteRecord{}).Where("inviter_id = ?", merchantID)

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

	var result []gin.H
	for _, record := range records {
		item := gin.H{
			"id":           record.ID,
			"invite_code":  record.InviteCode,
			"status":       record.Status,
			"reward_type":  record.RewardType,
			"created_at":   record.CreatedAt,
			"completed_at": record.CompletedAt,
		}

		if record.InviteeID > 0 {
			var invitee models.Merchant
			if err := database.DB.Select("name", "contact_phone").First(&invitee, record.InviteeID).Error; err == nil {
				item["invitee_name"] = invitee.Name
				item["invitee_phone"] = maskPhone(invitee.ContactPhone)
			}
		}

		result = append(result, item)
	}

	response.Success(c, gin.H{
		"list": result,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func maskPhone(phone string) string {
	if len(phone) < 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
