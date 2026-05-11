package user

import (
	"net/http"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetMerchantStatus(c *gin.Context) {
	phone := strings.TrimSpace(c.Query("phone"))
	if phone == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "phone不能为空")
		return
	}

	var merchant models.Merchant
	if err := database.DB.
		Select("id, name, applyment_status, audit_status, audit_remark, sub_mch_status, created_at").
		Where("contact_phone = ?", phone).
		Order("id DESC").
		First(&merchant).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "未找到入驻记录")
		return
	}

	status := "audit_pending"
	statusText := "资质审核中"
	progress := 25

	switch merchant.AuditStatus {
	case 2:
		status = "audit_rejected"
		statusText = "审核驳回"
		progress = 0
	case 1:
		switch merchant.ApplymentStatus {
		case 2:
			status = "wechat_approved"
			statusText = "微信审核通过"
			progress = 100
		case 0, 1:
			status = "wechat_pending"
			statusText = "微信审核中"
			progress = 75
		default:
			status = "wechat_failed"
			statusText = "微信审核失败"
			progress = 60
		}
	}

	applyTime := merchant.CreatedAt.Format("2006-01-02 15:04:05")
	if merchant.CreatedAt.IsZero() {
		applyTime = time.Now().Format("2006-01-02 15:04:05")
	}

	response.Success(c, gin.H{
		"status":        status,
		"status_text":   statusText,
		"progress":      progress,
		"merchant_name": merchant.Name,
		"apply_time":    applyTime,
		"audit_remark":  merchant.AuditRemark,
	})
}
