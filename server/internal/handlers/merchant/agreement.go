package merchant

import (
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================
// 阶段三：协议管理（用户协议/隐私政策/录音定位授权）
// 同类型仅一份已发布生效；发布新版自动停用旧版
// ============================================

// agreementTypeValid 校验协议类型
func agreementTypeValid(t uint8) bool {
	return t == utils.AgreementTypeUser ||
		t == utils.AgreementTypePrivacy ||
		t == utils.AgreementTypeAuth
}

// agreementTypeText 协议类型中文
func agreementTypeText(t uint8) string {
	switch t {
	case utils.AgreementTypeUser:
		return "用户协议"
	case utils.AgreementTypePrivacy:
		return "隐私政策"
	case utils.AgreementTypeAuth:
		return "录音/定位授权协议"
	default:
		return "未知"
	}
}

// CreateAgreementRequest 新建协议请求
type CreateAgreementRequest struct {
	Type    uint8  `json:"type" binding:"required"`
	Title   string `json:"title" binding:"required,max=128"`
	Content string `json:"content"`
	Version string `json:"version" binding:"required,max=32"`
}

// CreateAgreement 新建协议（草稿）
func CreateAgreement(c *gin.Context) {
	var req CreateAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if !agreementTypeValid(req.Type) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议类型不合法")
		return
	}

	// 同类型版本号唯一
	var count int64
	database.DB.Model(&models.Agreement{}).
		Where("type = ? AND version = ?", req.Type, req.Version).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该类型下版本号已存在")
		return
	}

	agreement := models.Agreement{
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
		Version: req.Version,
		Status:  utils.AgreementStatusDraft,
	}
	if err := database.DB.Create(&agreement).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建失败")
		return
	}

	response.SuccessWithMessage(c, "创建成功", gin.H{"id": agreement.ID})
}

// GetAgreements 协议列表
func GetAgreements(c *gin.Context) {
	agreementType := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.Agreement{})
	if agreementType != "" {
		if t, err := strconv.Atoi(agreementType); err == nil {
			query = query.Where("type = ?", t)
		}
	}
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}

	var total int64
	query.Count(&total)

	var agreements []models.Agreement
	query.Order("type ASC, created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&agreements)

	list := make([]gin.H, 0, len(agreements))
	for _, a := range agreements {
		list = append(list, gin.H{
			"id":           a.ID,
			"type":         a.Type,
			"type_cn":      agreementTypeText(a.Type),
			"title":        a.Title,
			"version":      a.Version,
			"status":       a.Status,
			"status_cn":    map[uint8]string{0: "草稿", 1: "已发布"}[a.Status],
			"published_at": a.PublishedAt,
			"created_at":   a.CreatedAt,
			"updated_at":   a.UpdatedAt,
		})
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAgreementDetail 协议详情
func GetAgreementDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议ID错误")
		return
	}

	var agreement models.Agreement
	if err := database.DB.Where("id = ?", id).First(&agreement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAgreementNotFound, "协议不存在")
		return
	}

	response.Success(c, gin.H{
		"id":           agreement.ID,
		"type":         agreement.Type,
		"type_cn":      agreementTypeText(agreement.Type),
		"title":        agreement.Title,
		"content":      agreement.Content,
		"version":      agreement.Version,
		"status":       agreement.Status,
		"published_at": agreement.PublishedAt,
		"created_at":   agreement.CreatedAt,
		"updated_at":   agreement.UpdatedAt,
	})
}

// UpdateAgreementRequest 编辑协议请求
type UpdateAgreementRequest struct {
	Title   string `json:"title" binding:"required,max=128"`
	Content string `json:"content"`
}

// UpdateAgreement 编辑协议（仅草稿可编辑正文，已发布的只读）
func UpdateAgreement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议ID错误")
		return
	}

	var req UpdateAgreementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	var agreement models.Agreement
	if err := database.DB.Where("id = ?", id).First(&agreement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAgreementNotFound, "协议不存在")
		return
	}
	if agreement.Status == utils.AgreementStatusPublished {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "已发布协议不可编辑，请新建新版本")
		return
	}

	updates := map[string]interface{}{
		"title":        req.Title,
		"content":      req.Content,
		"updated_at":   time.Now(),
	}
	if err := database.DB.Model(&agreement).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新失败")
		return
	}

	response.SuccessWithMessage(c, "更新成功", gin.H{"id": agreement.ID})
}

// PublishAgreement 发布协议（同类型互斥：旧版自动下线）
func PublishAgreement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议ID错误")
		return
	}

	var agreement models.Agreement
	if err := database.DB.Where("id = ?", id).First(&agreement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAgreementNotFound, "协议不存在")
		return
	}
	if agreement.Status == utils.AgreementStatusPublished {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该协议已是发布状态")
		return
	}

	// 事务：当前协议置为发布，同类型其他已发布的全部下线
	txErr := database.DB.Transaction(func(db *gorm.DB) error {
		now := time.Now()
		if err := db.Model(&models.Agreement{}).
			Where("type = ? AND status = ? AND id != ?", agreement.Type, utils.AgreementStatusPublished, agreement.ID).
			Updates(map[string]interface{}{
				"status":     utils.AgreementStatusDraft,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return db.Model(&agreement).Updates(map[string]interface{}{
			"status":       utils.AgreementStatusPublished,
			"published_at": now,
			"updated_at":  now,
		}).Error
	})
	if txErr != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "发布失败")
		return
	}

	response.SuccessWithMessage(c, "发布成功", gin.H{
		"id":           agreement.ID,
		"type":         agreement.Type,
		"version":      agreement.Version,
		"published_at": time.Now(),
	})
}
