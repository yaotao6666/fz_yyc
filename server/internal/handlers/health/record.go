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

// HealthRecordUpsertRequest 健康档案提交字段（用户端/员工端/管理端共用）。
// 新增 create/按 id update 流程支持 relation；relation 取值：1本人 2父母 3其他亲属。
type HealthRecordUpsertRequest struct {
	Relation         *uint8    `json:"relation"`
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
	AssessmentLevel  string    `json:"assessment_level"`
	Remark           string    `json:"remark"`
}

// marshalJSONList 将切片序列化为 models.JSON 列值，避免空数组误写入字符串。
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

// toHealthRecord 将请求字段转为档案模型（不含 id/user_id/status）。
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
	if req.Relation != nil && *req.Relation >= utils.HealthRecordRelationSelf && *req.Relation <= utils.HealthRecordRelationOtherKin {
		record.Relation = *req.Relation
	} else {
		// 未传或非法：默认本人，兼容存量单档案前端
		record.Relation = utils.HealthRecordRelationSelf
	}
	// 评估等级仅由评估流程回写，编辑时未传则保留原值，避免被覆盖为空
	if req.AssessmentLevel != "" {
		record.AssessmentLevel = req.AssessmentLevel
	}
	return record
}

// staffCanAccessResident 校验服务人员与客户是否存在服务关系（指派过订单即可）。
func staffCanAccessResident(staff *models.ServiceStaff, userID uint64) bool {
	var count int64
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND user_id = ?", staff.ID, userID).
		Count(&count)
	return count > 0
}

// recordReferencedByOrders 档案是否被订单引用（若被引用则禁止删除）。
func recordReferencedByOrders(recordID uint64) bool {
	var count int64
	database.DB.Model(&models.Order{}).
		Where("record_id = ?", recordID).
		Count(&count)
	return count > 0
}

// ---- 兼容存量单档案接口（暂不强制移除） ----

// UserGetHealthRecord C端用户获取当前账号下最近更新的健康档案（兼容单档案前端）。
func UserGetHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var record models.HealthRecord
	if err := database.DB.Where("user_id = ?", userID).Order("updated_at DESC").First(&record).Error; err != nil {
		response.Success(c, nil)
		return
	}
	response.Success(c, record)
}

// UserUpsertHealthRecord C端用户新增或更新"最近档案"（兼容单档案前端）。
// 若已有档案则按 id 覆盖；若无则新建一条本人档案。
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
	if database.DB.Where("user_id = ?", userID).Order("updated_at DESC").First(&existing).Error == nil {
		updated.ID = existing.ID
		updated.Status = existing.Status
		updated.CreatedAt = existing.CreatedAt
		// 兼容存量：若原档案无 relation，保留本次默认值（本人）
		if updated.Relation == 0 {
			updated.Relation = utils.HealthRecordRelationSelf
		}
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

// ---- 多档案新接口 ----

// UserListHealthRecords C端用户获取本人账号下所有档案列表（按更新时间倒序）。
func UserListHealthRecords(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var records []models.HealthRecord
	database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	if records == nil {
		records = []models.HealthRecord{}
	}
	response.Success(c, records)
}

// UserCreateHealthRecord C端用户新建档案（支持 relation 选择）。
func UserCreateHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req HealthRecordUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	var count int64
	database.DB.Model(&models.HealthRecord{}).Where("user_id = ?", userID).Count(&count)
	if count >= int64(utils.MaxHealthRecordsPerUser) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "最多可创建 5 个健康档案")
		return
	}
	record := toHealthRecord(req)
	record.UserID = userID
	record.Status = utils.HealthRecordStatusNormal
	if err := database.DB.Create(&record).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建档案失败")
		return
	}
	response.Success(c, record)
}

// UserUpdateHealthRecord C端用户按 id 更新档案（仅允许操作本人账号下的档案）。
func UserUpdateHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
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
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&existing).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在或无权操作")
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

// UserDeleteHealthRecord C端用户按 id 删除档案；若被服务订单引用则拒绝。
func UserDeleteHealthRecord(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案ID错误")
		return
	}
	var existing models.HealthRecord
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&existing).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在或无权操作")
		return
	}
	if recordReferencedByOrders(id) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "存在关联服务订单，档案不可删除")
		return
	}
	if err := database.DB.Delete(&existing).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除档案失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}

// ---- 服务人员端（多档案列表/详情/登记健康数值） ----

// StaffListResidentHealthRecords 服务人员查看客户的全部健康档案。
func StaffListResidentHealthRecords(c *gin.Context) {
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
	var records []models.HealthRecord
	database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&records)
	if records == nil {
		records = []models.HealthRecord{}
	}
	response.Success(c, records)
}

// StaffGetResidentHealthRecord 服务人员查看客户最近更新的档案（兼容旧接口）。
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
	if err := database.DB.Where("user_id = ?", userID).Order("updated_at DESC").First(&record).Error; err != nil {
		response.Success(c, nil)
		return
	}
	response.Success(c, record)
}

// StaffUpdateResidentHealthValues 服务人员更新客户指定档案的健康数值。
// 权限边界：仅允许更新身高、体重、慢病标签、长期用药等健康数值字段；
// 基础信息（姓名/性别/出生日期/身份证/关系/联系方式/地址等）不允许改动。
func StaffUpdateResidentHealthValues(c *gin.Context) {
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
	recordID, err := strconv.ParseUint(c.Param("record_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案ID错误")
		return
	}
	if !staffCanAccessResident(staff, userID) {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权操作该客户档案")
		return
	}

	var req HealthRecordUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var existing models.HealthRecord
	if err := database.DB.Where("id = ? AND user_id = ?", recordID, userID).First(&existing).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在")
		return
	}

	// 仅允许更新健康数值字段：按白名单赋值，杜绝基础信息被改动。
	updates := map[string]interface{}{}
	if req.HeightCm != nil {
		updates["height_cm"] = *req.HeightCm
	}
	if req.WeightKg != nil {
		updates["weight_kg"] = *req.WeightKg
	}
	if req.BloodType != "" {
		updates["blood_type"] = req.BloodType
	}
	updates["past_history"] = marshalJSONList(req.PastHistory)
	updates["allergy_history"] = marshalJSONList(req.AllergyHistory)
	updates["family_history"] = marshalJSONList(req.FamilyHistory)
	updates["surgery_history"] = marshalJSONList(req.SurgeryHistory)
	updates["medication_list"] = marshalJSONList(req.MedicationList)
	updates["chronic_tags"] = marshalJSONList(req.ChronicTags)
	if req.Smoking != "" {
		updates["smoking"] = req.Smoking
	}
	if req.Drinking != "" {
		updates["drinking"] = req.Drinking
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if err := database.DB.Model(&existing).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存失败")
		return
	}
	if err := database.DB.First(&existing, recordID).Error; err == nil {
		response.Success(c, existing)
		return
	}
	response.Success(c, existing)
}

// ---- 管理端 ----

// MerchantListHealthRecords 管理端健康档案列表（支持姓名/手机号/身份证模糊与评估等级筛选）。
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

// MerchantGetHealthRecord 管理端健康档案详情（含用户信息与该档案的评估记录）。
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
	database.DB.Where("record_id = ?", record.ID).
		Order("created_at DESC").
		Find(&assessments)

	response.Success(c, gin.H{
		"record":      record,
		"assessments": assessments,
	})
}

// MerchantUpdateHealthRecord 管理端编辑健康档案（允许更新关系+所有字段）。
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
