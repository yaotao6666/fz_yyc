// Package agreement 协议接口（阶段三：服务过程安全）。
// C端用户与服务人员共用：服务开始前动态拉取当前生效协议并确认留痕。
package agreement

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================
// 协议拉取 + 同意留痕（C端用户 / 服务人员共用）
// ============================================

// userTypeToConsentType 将 token 用户类型映射为留痕用户类型
func userTypeToConsentType(userType string) (uint8, bool) {
	switch userType {
	case "user":
		return utils.ConsentUserTypeUser, true
	case "service_staff":
		return utils.ConsentUserTypeStaff, true
	default:
		return 0, false
	}
}

// agreementTypeValid 校验协议类型
func agreementTypeValid(t uint8) bool {
	return t == utils.AgreementTypeUser ||
		t == utils.AgreementTypePrivacy ||
		t == utils.AgreementTypeAuth
}

// GetActiveAgreement 获取当前生效协议（?type=1|2|3，默认3=录音/定位授权）
func GetActiveAgreement(c *gin.Context) {
	agreementType, err := strconv.Atoi(c.DefaultQuery("type", "3"))
	if err != nil || !agreementTypeValid(uint8(agreementType)) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议类型不合法")
		return
	}

	var agreement models.Agreement
	if err := database.DB.Where("type = ? AND status = ?", uint8(agreementType), utils.AgreementStatusPublished).
		Order("published_at DESC").First(&agreement).Error; err != nil {
		// 无生效协议时返回空对象（前端按未配置处理，不阻断主流程）
		response.Success(c, gin.H{"agreement": nil})
		return
	}

	// 若为登录用户，附带其是否已同意当前版本
	userType, _ := userTypeToConsentType(middleware.GetUserType(c))
	consented := false
	if userType != 0 {
		userID := middleware.GetUserID(c)
		if userID > 0 {
			var count int64
			database.DB.Model(&models.AgreementConsent{}).
				Where("agreement_id = ? AND user_type = ? AND user_id = ? AND version = ?",
					agreement.ID, userType, userID, agreement.Version).
				Count(&count)
			consented = count > 0
		}
	}

	response.Success(c, gin.H{
		"agreement": gin.H{
			"id":           agreement.ID,
			"type":         agreement.Type,
			"title":        agreement.Title,
			"content":      agreement.Content,
			"version":      agreement.Version,
			"published_at": agreement.PublishedAt,
		},
		"consented": consented,
	})
}

// ConsentAgreement 同意协议（留痕：记录用户/版本/时间）
func ConsentAgreement(c *gin.Context) {
	consentType, ok := userTypeToConsentType(middleware.GetUserType(c))
	if !ok {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权操作")
		return
	}
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	agreementID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || agreementID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "协议ID错误")
		return
	}

	var agreement models.Agreement
	if err := database.DB.Where("id = ? AND status = ?", agreementID, utils.AgreementStatusPublished).
		First(&agreement).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAgreementNotFound, "协议不存在或未发布")
		return
	}

	// 同一协议同版本重复留痕幂等（先查再写）
	var count int64
	database.DB.Model(&models.AgreementConsent{}).
		Where("agreement_id = ? AND user_type = ? AND user_id = ? AND version = ?",
			agreement.ID, consentType, userID, agreement.Version).
		Count(&count)
	if count > 0 {
		response.SuccessWithMessage(c, "已确认过该版本", gin.H{"agreement_id": agreement.ID})
		return
	}

	consent := models.AgreementConsent{
		AgreementID: agreement.ID,
		UserType:    consentType,
		UserID:      userID,
		Version:     agreement.Version,
	}
	if err := database.DB.Create(&consent).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "确认失败")
		return
	}

	response.SuccessWithMessage(c, "确认成功", gin.H{
		"agreement_id": agreement.ID,
		"version":      agreement.Version,
		"consented_at": consent.CreatedAt,
	})
}
