package service_staff

import (
	"encoding/json"
	"net/http"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// QualificationItem 资质材料条目
type QualificationItem struct {
	Type string `json:"type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// RequestProfileChangeRequest 服务人员端上资料变更申请
type RequestProfileChangeRequest struct {
	Name          string              `json:"name"`
	Phone         string              `json:"phone"`
	Avatar        string              `json:"avatar"`
	Qualifications []QualificationItem `json:"qualifications"`
}

// RequestProfileChange 服务人员提交资料变更（进入待审核，不直接改正式字段）
func RequestProfileChange(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	var req RequestProfileChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 校验姓名必填
	if req.Name == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "姓名不能为空")
		return
	}
	if req.Phone == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "手机号不能为空")
		return
	}

	// 校验变更是否有效（与当前值全同则无需审核）
	if req.Name == staff.Name && req.Phone == staff.Phone && req.Avatar == staff.Avatar {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "未检测到信息变更")
		return
	}

	// 变更前快照
	beforeData, _ := json.Marshal(map[string]interface{}{
		"name":   staff.Name,
		"phone":  staff.Phone,
		"avatar": staff.Avatar,
	})
	// 变更后快照
	afterData, _ := json.Marshal(map[string]interface{}{
		"name":   req.Name,
		"phone":  req.Phone,
		"avatar": req.Avatar,
	})
	qualsRaw, _ := json.Marshal(req.Qualifications)

	record := models.StaffAuditRecord{
		StaffID:        staff.ID,
		AuditType:      2, // 2=信息变更
		ApplyType:      3, // 3=端上变更
		BeforeData:     models.JSON(beforeData),
		AfterData:      models.JSON(afterData),
		Qualifications: models.JSON(qualsRaw),
		Status:         0, // 待审
	}
	if err := database.DB.Create(&record).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "提交变更失败")
		return
	}

	// 置服务人员审核态为待审（不影响 status 启用/禁用）
	if err := database.DB.Model(&models.ServiceStaff{}).Where("id = ?", staff.ID).
		Update("audit_status", 1).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新审核状态失败")
		return
	}

	response.SuccessWithMessage(c, "资料变更已提交，等待审核", gin.H{"audit_id": record.ID})
}

// GetMyAuditList 服务人员查看本人审核记录
func GetMyAuditList(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	var records []models.StaffAuditRecord
	if err := database.DB.Where("staff_id = ?", staff.ID).Order("id DESC").Find(&records).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}})
		return
	}

	response.Success(c, gin.H{"list": records, "total": len(records)})
}