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

// CreateAssessmentRequest 用户/服务人员提交评估的请求
type CreateAssessmentRequest struct {
	FormID      uint64            `json:"form_id" binding:"required"`
	RecordID    *uint64           `json:"record_id"`
	Answers     map[string]string `json:"answers"`
	SymptomDesc string            `json:"symptom_desc"`
}

// AssessmentFormUpsertRequest 管理端新增/编辑量表请求
type AssessmentFormUpsertRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Dimension   string                   `json:"dimension"`
	Description string                   `json:"description"`
	Questions   []map[string]interface{} `json:"questions"`
	ScoreRule   []map[string]interface{} `json:"score_rule"`
	Status      uint8                    `json:"status"`
}

// computeAssessmentScore 依据量表题目与评分规则计算总分/等级/结论。
// 分数为各题选中项 score 之和；等级取第一条命中 min<=总分<=max 的规则。
// 量表 JSON 字段先反序列化为 map 再取值，避免依赖固定结构体导致解析失败。
func computeAssessmentScore(form *models.HealthAssessmentForm, answers map[string]string) (totalScore float64, level, conclusion string) {
	if len(form.Questions) == 0 {
		return 0, "", ""
	}

	var questions []map[string]interface{}
	if err := json.Unmarshal(form.Questions, &questions); err != nil {
		return 0, "", ""
	}

	for _, question := range questions {
		key, _ := question["key"].(string)
		selectedLabel := answers[key]
		options, _ := question["options"].([]interface{})
		for _, opt := range options {
			optionMap, ok := opt.(map[string]interface{})
			if !ok {
				continue
			}
			label, _ := optionMap["label"].(string)
			if label != selectedLabel {
				continue
			}
			if score, ok := optionMap["score"].(float64); ok {
				totalScore += score
			}
			break
		}
	}

	if len(form.ScoreRule) == 0 {
		return totalScore, "", ""
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(form.ScoreRule, &rules); err != nil {
		return totalScore, "", ""
	}
	for _, rule := range rules {
		minScore, okMin := rule["min"].(float64)
		maxScore, okMax := rule["max"].(float64)
		if !okMin || !okMax {
			continue
		}
		if totalScore >= minScore && totalScore <= maxScore {
			level, _ = rule["level"].(string)
			conclusion, _ = rule["conclusion"].(string)
			return totalScore, level, conclusion
		}
	}
	return totalScore, "", ""
}

// saveAssessmentAndSyncLevel 保存评估记录并回写健康档案的评估等级。
// 回写优先落在 record_id 指定的档案；若未指定则回写用户最近档案。
func saveAssessmentAndSyncLevel(c *gin.Context, assessment *models.HealthAssessment) {
	if err := database.DB.Create(assessment).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存评估记录失败")
		return
	}
	// 回写评估等级：优先按 record_id；否则按 user_id 最近一条档案
	if assessment.RecordID != nil && *assessment.RecordID > 0 {
		database.DB.Model(&models.HealthRecord{}).
			Where("id = ?", *assessment.RecordID).
			Update("assessment_level", assessment.Level)
	} else {
		var lastRecord models.HealthRecord
		if database.DB.Where("user_id = ?", assessment.UserID).
			Order("updated_at DESC").First(&lastRecord).Error == nil {
			database.DB.Model(&models.HealthRecord{}).
				Where("id = ?", lastRecord.ID).
				Update("assessment_level", assessment.Level)
		}
	}

	response.Success(c, assessment)
}

// loadEnabledForm 加载启用的量表，未启用时返回 false，用于 C 端与服务人员端校验
func loadEnabledForm(formID uint64, form *models.HealthAssessmentForm) bool {
	if err := database.DB.First(form, formID).Error; err != nil {
		return false
	}
	return form.Status == utils.AssessmentFormStatusEnabled
}

// UserGetAssessmentForms C端用户获取启用的评估量表列表
func UserGetAssessmentForms(c *gin.Context) {
	var forms []models.HealthAssessmentForm
	database.DB.Where("status = ?", utils.AssessmentFormStatusEnabled).
		Order("id DESC").
		Find(&forms)
	response.Success(c, forms)
}

// UserListAssessments C端用户查看自己的评估记录（分页倒序，支持按 record_id 过滤）
func UserListAssessments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	recordIDStr := c.Query("record_id")
	baseQuery := database.DB.Model(&models.HealthAssessment{}).Where("user_id = ?", userID)
	if recordIDStr != "" {
		if rid, err := strconv.ParseUint(recordIDStr, 10, 64); err == nil && rid > 0 {
			baseQuery = baseQuery.Where("record_id = ?", rid)
		}
	}

	var total int64
	baseQuery.Count(&total)

	var list []models.HealthAssessment
	baseQuery.Preload("Form").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// UserCreateAssessment C端用户自助提交评估并回写档案等级
func UserCreateAssessment(c *gin.Context) {
	var req CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var form models.HealthAssessmentForm
	if !loadEnabledForm(req.FormID, &form) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "量表不存在或未启用")
		return
	}

	userID := middleware.GetUserID(c)
	// 若传了 record_id：校验档案归属本人
	var recordID *uint64
	if req.RecordID != nil && *req.RecordID > 0 {
		var rec models.HealthRecord
		if err := database.DB.Where("id = ? AND user_id = ?", *req.RecordID, userID).First(&rec).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案不存在或无权使用")
			return
		}
		recordID = req.RecordID
	}

	answersRaw, _ := json.Marshal(req.Answers)
	totalScore, level, conclusion := computeAssessmentScore(&form, req.Answers)

	assessment := models.HealthAssessment{
		UserID:       userID,
		RecordID:     recordID,
		FormID:       form.ID,
		FormName:     form.Name,
		AssessorType: utils.AssessmentTypeSelf,
		Answers:      models.JSON(answersRaw),
		TotalScore:   totalScore,
		Level:        level,
		Conclusion:   conclusion,
		SymptomDesc:  req.SymptomDesc,
	}
	saveAssessmentAndSyncLevel(c, &assessment)
}

// StaffGetAssessmentForms 服务人员获取启用的评估量表列表
func StaffGetAssessmentForms(c *gin.Context) {
	var forms []models.HealthAssessmentForm
	database.DB.Where("status = ?", utils.AssessmentFormStatusEnabled).
		Order("id DESC").
		Find(&forms)
	response.Success(c, forms)
}

// StaffListResidentAssessments 服务人员查看客户的评估记录（需存在服务关系）
func StaffListResidentAssessments(c *gin.Context) {
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 按档案隔离评估记录：优先使用请求指定的 record_id，未指定时回退到该用户最近档案。
	// 无论哪种情况都必须严格校验 record_id 归属该 user_id，防止串档或越权拉取他人档案评估。
	rid := uint64(0)
	if ridStr := c.Query("record_id"); ridStr != "" {
		v, err2 := strconv.ParseUint(ridStr, 10, 64)
		if err2 != nil {
			response.Success(c, gin.H{"list": []models.HealthAssessment{}, "total": 0})
			return
		}
		rid = v
	}
	if rid == 0 {
		var rec models.HealthRecord
		if err := database.DB.Where("user_id = ?", userID).Order("updated_at DESC").First(&rec).Error; err != nil {
			response.Success(c, gin.H{"list": []models.HealthAssessment{}, "total": 0})
			return
		}
		rid = rec.ID
	}
	// 严格校验 record_id 必须属于该 user_id
	var owned models.HealthRecord
	if err := database.DB.Where("id = ? AND user_id = ?", rid, userID).First(&owned).Error; err != nil ||
		owned.ID != rid {
		response.Success(c, gin.H{"list": []models.HealthAssessment{}, "total": 0})
		return
	}

	query := database.DB.Model(&models.HealthAssessment{}).Where("user_id = ?", userID).Where("record_id = ?", rid)
	var total int64
	query.Count(&total)

	var list []models.HealthAssessment
	query.Preload("Form").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// StaffCreateAssessment 服务人员为客户手动登记评估并回写档案等级
func StaffCreateAssessment(c *gin.Context) {
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
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该客户登记评估")
		return
	}

	var req CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var form models.HealthAssessmentForm
	if !loadEnabledForm(req.FormID, &form) {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "量表不存在或未启用")
		return
	}

	// 若传了 record_id：校验档案属于该客户
	var recordID *uint64
	if req.RecordID != nil && *req.RecordID > 0 {
		var rec models.HealthRecord
		if err := database.DB.Where("id = ? AND user_id = ?", *req.RecordID, userID).First(&rec).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案不属于该客户")
			return
		}
		recordID = req.RecordID
	}

	answersRaw, _ := json.Marshal(req.Answers)
	totalScore, level, conclusion := computeAssessmentScore(&form, req.Answers)

	assessment := models.HealthAssessment{
		UserID:       userID,
		RecordID:     recordID,
		FormID:       form.ID,
		FormName:     form.Name,
		AssessorType: utils.AssessmentTypeStaff,
		StaffID:      &staff.ID,
		Answers:      models.JSON(answersRaw),
		TotalScore:   totalScore,
		Level:        level,
		Conclusion:   conclusion,
		SymptomDesc:  req.SymptomDesc,
	}
	saveAssessmentAndSyncLevel(c, &assessment)
}

// MerchantListAssessmentForms 管理端量表列表（含草稿，按 id 倒序）
func MerchantListAssessmentForms(c *gin.Context) {
	var forms []models.HealthAssessmentForm
	database.DB.Order("id DESC").Find(&forms)
	response.Success(c, forms)
}

// MerchantCreateAssessmentForm 管理端新增评估量表
func MerchantCreateAssessmentForm(c *gin.Context) {
	var req AssessmentFormUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	questionsRaw, err := json.Marshal(req.Questions)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "题目格式错误")
		return
	}
	scoreRuleRaw, err := json.Marshal(req.ScoreRule)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "评分规则格式错误")
		return
	}

	form := models.HealthAssessmentForm{
		Name:        req.Name,
		Dimension:   req.Dimension,
		Description: req.Description,
		Questions:   models.JSON(questionsRaw),
		ScoreRule:   models.JSON(scoreRuleRaw),
		Version:     1,
		Status:      req.Status,
	}
	if err := database.DB.Create(&form).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建量表失败")
		return
	}
	response.Success(c, form)
}

// MerchantUpdateAssessmentForm 管理端编辑评估量表
func MerchantUpdateAssessmentForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "量表ID错误")
		return
	}

	var req AssessmentFormUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var form models.HealthAssessmentForm
	if err := database.DB.First(&form, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "量表不存在")
		return
	}

	questionsRaw, err := json.Marshal(req.Questions)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "题目格式错误")
		return
	}
	scoreRuleRaw, err := json.Marshal(req.ScoreRule)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "评分规则格式错误")
		return
	}

	form.Name = req.Name
	form.Dimension = req.Dimension
	form.Description = req.Description
	form.Questions = models.JSON(questionsRaw)
	form.ScoreRule = models.JSON(scoreRuleRaw)
	form.Status = req.Status
	if err := database.DB.Save(&form).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存量表失败")
		return
	}
	response.Success(c, form)
}

// MerchantDeleteAssessmentForm 管理端删除评估量表
func MerchantDeleteAssessmentForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "量表ID错误")
		return
	}
	if err := database.DB.Delete(&models.HealthAssessmentForm{}, id).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除量表失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}

// SetAssessmentFormStatusRequest 管理端启用/停用量表请求
type SetAssessmentFormStatusRequest struct {
	Status uint8 `json:"status"`
}

// MerchantSetAssessmentFormStatus 管理端启用/停用评估量表（仅更新状态，不要求其他字段）
func MerchantSetAssessmentFormStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "量表ID错误")
		return
	}

	var req SetAssessmentFormStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.Status != utils.AssessmentFormStatusDraft && req.Status != utils.AssessmentFormStatusEnabled {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "状态值错误")
		return
	}

	var form models.HealthAssessmentForm
	if err := database.DB.First(&form, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "量表不存在")
		return
	}

	if err := database.DB.Model(&form).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存量表状态失败")
		return
	}
	response.SuccessWithMessage(c, "更新成功", gin.H{"id": id, "status": req.Status})
}

// MerchantListHealthAssessments 管理端评估记录列表（支持按用户昵称/手机号与量表筛选）
func MerchantListHealthAssessments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	keyword := c.Query("keyword")
	formID := c.Query("form_id")

	countQuery := database.DB.Model(&models.HealthAssessment{})
	listQuery := database.DB.Model(&models.HealthAssessment{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN users ON users.id = health_assessments.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
		listQuery = listQuery.
			Joins("LEFT JOIN users ON users.id = health_assessments.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
	}
	if formID != "" {
		if fid, err := strconv.ParseUint(formID, 10, 64); err == nil {
			countQuery = countQuery.Where("health_assessments.form_id = ?", fid)
			listQuery = listQuery.Where("health_assessments.form_id = ?", fid)
		}
	}

	var total int64
	countQuery.Count(&total)

	var list []models.HealthAssessment
	listQuery.Preload("User").Preload("Form").
		Order("health_assessments.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// MerchantListRecordAssessments 管理端查看指定健康档案的评估记录。
// 按 record_id 严格归属该档案；历史未挂档案的评估记录需经数据回填归档案，避免跨档案串数据。
func MerchantListRecordAssessments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "档案ID错误")
		return
	}
	var record models.HealthRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "档案不存在")
		return
	}
	var list []models.HealthAssessment
	database.DB.
		Where("record_id = ?", record.ID).
		Preload("Form").
		Order("created_at DESC").
		Find(&list)
	if list == nil {
		list = []models.HealthAssessment{}
	}
	response.Success(c, list)
}
