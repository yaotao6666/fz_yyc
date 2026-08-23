package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListStaffAudits PC 后台审核列表（注册/变更/资质提交）
func ListStaffAudits(c *gin.Context) {
	auditType := c.Query("audit_type")
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

	query := database.DB.Table("service_staff_audit_records").
		Select("service_staff_audit_records.*, service_staffs.name AS staff_name, service_staffs.phone AS staff_phone, service_staffs.username AS staff_username")
	if auditType != "" {
		if t, err := strconv.Atoi(auditType); err == nil {
			query = query.Where("service_staff_audit_records.audit_type = ?", t)
		}
	}
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("service_staff_audit_records.status = ?", s)
		}
	}
	if keyword != "" {
		query = query.Joins("LEFT JOIN service_staffs ON service_staffs.id = service_staff_audit_records.staff_id").
			Where("service_staffs.name LIKE ? OR service_staffs.phone LIKE ? OR service_staffs.username LIKE ?",
				"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var records []map[string]interface{}
	query.Joins("LEFT JOIN service_staffs ON service_staffs.id = service_staff_audit_records.staff_id").
		Order("service_staff_audit_records.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Scan(&records)

	response.Success(c, gin.H{
		"list":  records,
		"total": total,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// StaffAuditDetail 审核详情（含资质材料）
func StaffAuditDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "审核记录ID错误")
		return
	}

	var record models.StaffAuditRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "审核记录不存在")
		return
	}

	var staff models.ServiceStaff
	database.DB.First(&staff, record.StaffID)

	// 资质材料规范化
	var quals []map[string]interface{}
	if len(record.Qualifications) > 0 {
		json.Unmarshal(record.Qualifications, &quals)
	}
	var beforeData, afterData map[string]interface{}
	if len(record.BeforeData) > 0 {
		json.Unmarshal(record.BeforeData, &beforeData)
	}
	if len(record.AfterData) > 0 {
		json.Unmarshal(record.AfterData, &afterData)
	}

	response.Success(c, gin.H{
		"record": record,
		"staff":  staff,
		"qualifications": quals,
		"before_data":    beforeData,
		"after_data":     afterData,
	})
}

// ProcessStaffAudit 审核处理通用逻辑（通过/驳回）
// auditType=1(注册申请)通过时启用账号；auditType=2(信息变更)通过时回写字段。
func ProcessStaffAudit(c *gin.Context, approve bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "审核记录ID错误")
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&req)

	var record models.StaffAuditRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "审核记录不存在")
		return
	}
	if record.Status != 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该记录已审核")
		return
	}

	reviewer, _ := middleware.GetCurrentStaff(c)
	now := time.Now()

	// 状态枚举定义
	status := uint8(2) // 驳回
	action := "已驳回"
	if approve {
		status = 1
		action = "已通过"
	}

	updates := map[string]interface{}{
		"status":        status,
		"reviewer_id":   reviewer.ID,
		"review_remark": req.Remark,
		"review_at":     now,
	}
	if err := database.DB.Model(&record).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "处理失败")
		return
	}

	if approve {
		staffUpdates := map[string]interface{}{
			"audit_status": 0, // 审核完成
		}
		if record.AuditType == 1 {
			// 注册申请通过 => 启用账号
			staffUpdates["status"] = 1
		}
		if record.AuditType == 2 {
			// 信息变更通过 => 回写 after_data 到正式字段
			if len(record.AfterData) > 0 {
				var after map[string]interface{}
				if err := json.Unmarshal(record.AfterData, &after); err == nil {
					if v, ok := after["name"].(string); ok && v != "" {
						staffUpdates["name"] = v
					}
					if v, ok := after["phone"].(string); ok && v != "" {
						staffUpdates["phone"] = v
					}
					if v, ok := after["avatar"].(string); ok {
						staffUpdates["avatar"] = v
					}
				}
			}
			// 回写资质材料（若有新提交）
			if len(record.Qualifications) > 0 {
				staffUpdates["qualifications"] = record.Qualifications
			}
		}
		database.DB.Model(&models.ServiceStaff{}).Where("id = ?", record.StaffID).Updates(staffUpdates)
	} else {
		// 驳回时清除待审核态
		database.DB.Model(&models.ServiceStaff{}).Where("id = ?", record.StaffID).
			Update("audit_status", 0)
	}

	response.SuccessWithMessage(c, action, gin.H{"id": record.ID, "status": status})
}

// ApproveStaffAudit 审核通过
func ApproveStaffAudit(c *gin.Context) {
	ProcessStaffAudit(c, true)
}

// RejectStaffAudit 审核驳回
func RejectStaffAudit(c *gin.Context) {
	ProcessStaffAudit(c, false)
}