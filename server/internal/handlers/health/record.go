package health

import (
	"encoding/json"
	"net/http"
	"strconv"

	serviceStaffHandler "fz_yyc_api/internal/handlers/service_staff"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthRecordUpsertRequest 健康档案提交字段（用户端/管理端共用，不含 id/user_id/status）。
// JSON 数组字段用 []string 绑定，写入时再序列化为 models.JSON，避免硬编码结构体导致解析失败。
type HealthRecordUpsertRequest struct {
	RealName         string    `json:"real_name"`
	Gender           uint8     `json:"gender"`
	BirthDate        string    `json:"birth_date"`
	IDCard           string    `json:"id_card"`
	Phone            string    `json:"phone"`
	EmergencyContact string    `json:"emergency_contact"`
	EmergencyPhone   string    `json:"emergency_phone"`
	Address          string    `json:"address"`
	HeightCm         *float64  `json:"height_cm"`
	WeightKg         *float64  `json:"weight_kg"`
	BloodType        string    `json:"blood_type"`
	PastHistory      []string  `json:"past_history"`
	AllergyHistory   []string  `json:"allergy_history"`
	FamilyHistory    []string  `json:"family_history"`
	SurgeryHistory   []string  `json:"surgery_history"`
	MedicationList   []string  `json:"medication_list"`
	ChronicTags      []string  `json:"chronic_tags"`
	Smoking          string    `json:"smoking"`
	Drinking         string    `json:"drinking"`
	Remark           string    `json:"remark"`
}

// marshalJSONList 将切片序列化为 models.JSON 列值，
// 避免空数组/普通字符串被 GORM 以错误类型写入 JSON 列。
func marshalJSONList(list []string) models.JSON {
	if list == nil {
		return nil
	}
	raw, err := json.Marshal(list)
	if err != nil {
		return nil
	}
	return models.JSON(raw)
}

// toHealthRecord 将请求字段转换为档案模型，供创建/更新复用，保持两端写入逻辑一致。
func toHealthRecord(req HealthRecordUpsertRequest) models.HealthRecord {
	record := models.HealthRecord{
		RealName:         req.RealName,
		Gender:           req.Gender,
		BirthDate:        req.BirthDate,
		IDCard:           req.IDCard,
		Phone:            req.Phone,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		Address:          req.Address,
		BloodType:        req.BloodType,
		PastHistory:      marshalJSONList(req.PastHistory),
		AllergyHistory:   marshalJSONList(req.AllergyHistory),
		FamilyHistory:    marshalJSONList(req.FamilyHistory),
		SurgeryHistory:   marshalJSONList(req.SurgeryHistory),
		MedicationList:   marshalJSONList(req.MedicationList),
		ChronicTags:      marshalJSONList(req.ChronicTags),
		Smoking:          req.Smoking,
		Drinking:         req.Drinking,
		Remark:           req.Remark,
	}
	if req.HeightCm != nil {
		record.HeightCm = *req.HeightCm
	}
	if req.WeightKg != nil {
		record.WeightKg = *req.WeightKg
	}
	return record
}

// staffCanAccessResident 校验服务人员与该客户的绑定关系：
// 只要存在指派给当前服务人员的该客户订单，即视为存在服务关系，允许访问其健康数据。
func staffCanAccessResident(staff *models.ServiceStaff, userID uint64) bool {
	var count int64
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND user_id = ?", staff.ID, userID).
		Count(&count)
	return count > 0
}

// UserGetHealthRecord C端用户查看自己的健康档案（未建档返回 data=nil）
func UserGetHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var record models.HealthRecord
	if err := database.DB.Where("user_id = ?", userID).First(&record).Error; err != nil {
		response.Success(c, nil)
		return
	}
	response.Success(c, record)
}

// UserUpsertHealthRecord C端用户新增或更新自己的健康档案
func UserUpsertHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req HealthRecordUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	updated := toHealthRecord(req)
	updated.UserID = userID

	var existing models.HealthRecord
	if database.DB.Where("user_id = ?", userID).First(&existing).Error == nil {
		// 已有档案则保留原状态与创建时间，整份覆盖
		updated.ID = existing.ID
		updated.Status = existing.Status
		updated.CreatedAt = existing.CreatedAt
		if err := database.DB.Save(&updated).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存档案失败")
			return
		}
	} else {
		updated.Status = utils.HealthRecordStatusNormal
		if err := database.DB.Create(&updated).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存档案失败")
			return
		}
	}
	response.Success(c, updated)
}

// StaffGetResidentHealthRecord 服务人员查看客户健康档案（需存在服务关系）
func StaffGetResidentHealthRecord(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "客户ID错误")
		return
	}
	if !staffCanAccessResident(staff, userID) {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该客户档案")
		return
	}

	var record models.HealthRecord
	if err := database.DB.Where("user_id = ?", userID).First(&record).Error; err != nil {
		response.Success(c, nil)
		return
	}
	response.Success(c, record)
}

// MerchantListHealthRecords 管理端健康档案列表（支持姓名/手机号/身份证模糊与评估等级筛选）
func MerchantListHealthRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	keyword := c.Query("keyword")
	assessmentLevel := c.Query("assessment_level")

	countQuery := database.DB.Model(&models.HealthRecord{})
	listQuery := database.DB.Model(&models.HealthRecord{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.Where("real_name LIKE ? OR phone LIKE ? OR id_card LIKE ?", like, like, like)
		listQuery = listQuery.Where("real_name LIKE ? OR phone LIKE ? OR id_card LIKE ?", like, like, like)
	}
	if assessmentLevel != "" {
		countQuery = countQuery.Where("assessment_level = ?", assessmentLevel)
		listQuery = listQuery.Where("assessment_level = ?", assessmentLevel)
	}

	var total int64
	countQuery.Count(&total)

	var records []models.HealthRecord
	listQuery.Preload("User").
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	response.Success(c, gin.H{
		"list":  records,
		"total": total,
	})
}

// MerchantGetHealthRecord 管理端健康档案详情（含用户信息与该用户全部评估记录）
func MerchantGetHealthRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案ID错误")
		return
	}

	var record models.HealthRecord
	if err := database.DB.Preload("User").First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在")
		return
	}

	var assessments []models.HealthAssessment
	database.DB.Where("user_id = ?", record.UserID).
		Order("created_at DESC").
		Find(&assessments)

	response.Success(c, gin.H{
		"record":      record,
		"assessments": assessments,
	})
}

// MerchantUpdateHealthRecord 管理端编辑健康档案
func MerchantUpdateHealthRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案ID错误")
		return
	}

	var req HealthRecordUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var existing models.HealthRecord
	if err := database.DB.First(&existing, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在")
		return
	}

	updated := toHealthRecord(req)
	updated.ID = existing.ID
	updated.UserID = existing.UserID
	updated.Status = existing.Status
	updated.CreatedAt = existing.CreatedAt
	if err := database.DB.Save(&updated).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存档案失败")
		return
	}
	response.Success(c, updated)
}
