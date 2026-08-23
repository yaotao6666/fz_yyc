package health

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	serviceStaffHandler "fz_yyc_api/internal/handlers/service_staff"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ============================================================
// 随访任务
// ============================================================

// FollowUpResult 随访结果内容（写入 result JSON 字段）。
// contact_method 同时回写任务列 contact_method，education_article_ids 为宣教文章ID数组。
type FollowUpResult struct {
	ContactMethod       uint8    `json:"contact_method"`
	Content             string   `json:"content"`
	EducationArticleIDs []uint64 `json:"education_article_ids"`
	Satisfaction        uint8    `json:"satisfaction"`
	Remark              string   `json:"remark"`
}

// FollowUpCompleteRequest 完成随访任务的请求（result 整体绑定为结构体，写入时再序列化为 JSON）
type FollowUpCompleteRequest struct {
	Result FollowUpResult `json:"result"`
}

// FollowUpCreateRequest 管理端手动创建随访任务的请求（source_type 固定为手动）
type FollowUpCreateRequest struct {
	UserID         uint64     `json:"user_id"`
	TaskType       uint8      `json:"task_type"`
	SourceID       *uint64    `json:"source_id"`
	PlanFollowTime *time.Time `json:"plan_follow_time"`
	StaffID        *uint64    `json:"staff_id"`
	ContactMethod  uint8      `json:"contact_method"`
	Remark         string     `json:"remark"`
}

// followUpTaskVO 随访任务视图：result 解析为对象，管理端/服务人员端列表可携带 User，服务人员姓名转义为 StaffName
type followUpTaskVO struct {
	ID             uint64                 `json:"id"`
	UserID         uint64                 `json:"user_id"`
	TaskType       uint8                  `json:"task_type"`
	SourceType     uint8                  `json:"source_type"`
	SourceID       *uint64                `json:"source_id"`
	PlanFollowTime *time.Time             `json:"plan_follow_time"`
	StaffID        *uint64                `json:"staff_id"`
	ContactMethod  uint8                  `json:"contact_method"`
	Status         uint8                  `json:"status"`
	Result         map[string]interface{} `json:"result"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Remark         string                 `json:"remark"`
	CreatedAt      time.Time              `json:"created_at"`
	User           *models.User           `json:"user,omitempty"`
	StaffName      string                 `json:"staff_name,omitempty"`
}

// toFollowUpTaskVO 将随访任务转换为视图，result JSON 做防御解析
func toFollowUpTaskVO(task *models.FollowUpTask) followUpTaskVO {
	vo := followUpTaskVO{
		ID:             task.ID,
		UserID:         task.UserID,
		TaskType:       task.TaskType,
		SourceType:     task.SourceType,
		SourceID:       task.SourceID,
		PlanFollowTime: task.PlanFollowTime,
		StaffID:        task.StaffID,
		ContactMethod:  task.ContactMethod,
		Status:         task.Status,
		Result:         parseJSONObject(task.Result),
		CompletedAt:    task.CompletedAt,
		Remark:         task.Remark,
		CreatedAt:      task.CreatedAt,
		User:           task.User,
	}
	if task.Staff != nil {
		vo.StaffName = task.Staff.Name
	}
	return vo
}

// canOperateFollowUpTask 校验服务人员对该随访任务的操作权限：
// 任务须属于当前服务人员，或处于待认领（staff_id IS NULL）状态。
func canOperateFollowUpTask(task *models.FollowUpTask, staffID uint64) bool {
	return task.StaffID == nil || *task.StaffID == staffID
}

// claimFollowUpTask 若任务处于待认领状态则认领给当前服务人员，返回认领后 staff_id 是否变更
func claimFollowUpTask(task *models.FollowUpTask, staffID uint64) {
	if task.StaffID == nil {
		database.DB.Model(task).Update("staff_id", staffID)
		task.StaffID = &staffID
	}
}

// UserListFollowUpTasks C端用户查看自己的随访任务（分页倒序，可按状态筛选）
func UserListFollowUpTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p := utils.GetPagination(c)
	status := c.Query("status")

	countQuery := database.DB.Model(&models.FollowUpTask{}).Where("user_id = ?", userID)
	listQuery := database.DB.Model(&models.FollowUpTask{}).Where("user_id = ?", userID)
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("status = ?", st)
			listQuery = listQuery.Where("status = ?", st)
		}
	}

	var total int64
	countQuery.Count(&total)

	var tasks []models.FollowUpTask
	listQuery.Preload("User").Preload("Staff").
		Order("created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&tasks)

	list := make([]followUpTaskVO, 0, len(tasks))
	for i := range tasks {
		list = append(list, toFollowUpTaskVO(&tasks[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// StaffListFollowUpTasks 服务人员随访任务列表：可见「指派给自己」或「待认领」的任务，支持状态筛选，分页倒序
func StaffListFollowUpTasks(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	p := utils.GetPagination(c)
	status := c.Query("status")

	countQuery := database.DB.Model(&models.FollowUpTask{}).
		Where("staff_id = ? OR staff_id IS NULL", staff.ID)
	listQuery := database.DB.Model(&models.FollowUpTask{}).
		Where("staff_id = ? OR staff_id IS NULL", staff.ID)
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("status = ?", st)
			listQuery = listQuery.Where("status = ?", st)
		}
	}

	var total int64
	countQuery.Count(&total)

	var tasks []models.FollowUpTask
	listQuery.Preload("User").Preload("Staff").
		Order("created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&tasks)

	list := make([]followUpTaskVO, 0, len(tasks))
	for i := range tasks {
		list = append(list, toFollowUpTaskVO(&tasks[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// StaffGetFollowUpTask 服务人员查看随访任务详情（须属于当前服务人员或待认领）
func StaffGetFollowUpTask(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务ID错误")
		return
	}

	var task models.FollowUpTask
	if err := database.DB.Preload("User").Preload("Staff").First(&task, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "任务不存在")
		return
	}
	if !canOperateFollowUpTask(&task, staff.ID) {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该任务")
		return
	}
	response.Success(c, toFollowUpTaskVO(&task))
}

// StaffCompleteFollowUpTask 服务人员完成随访任务。
// 权限校验：任务属于当前服务人员或待认领；若为待认领则认领为当前人员；
// status→已完成、completed_at=now、contact_method 回写列、result 存 JSON。
func StaffCompleteFollowUpTask(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务ID错误")
		return
	}

	var req FollowUpCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var task models.FollowUpTask
	if err := database.DB.First(&task, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "任务不存在")
		return
	}
	if !canOperateFollowUpTask(&task, staff.ID) {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权执行该任务")
		return
	}
	claimFollowUpTask(&task, staff.ID)

	resultRaw, err := json.Marshal(req.Result)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "随访结果格式错误")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{
		"status":         utils.FollowUpStatusCompleted,
		"completed_at":   now,
		"contact_method": req.Result.ContactMethod,
		"result":         models.JSON(resultRaw),
		"staff_id":       task.StaffID,
	}
	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存随访结果失败")
		return
	}
	database.DB.Preload("User").Preload("Staff").First(&task, id)
	response.Success(c, toFollowUpTaskVO(&task))
}

// StaffSkipFollowUpTask 服务人员跳过随访任务。权限校验同上；status→已跳过。
func StaffSkipFollowUpTask(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务ID错误")
		return
	}

	var task models.FollowUpTask
	if err := database.DB.First(&task, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "任务不存在")
		return
	}
	if !canOperateFollowUpTask(&task, staff.ID) {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权操作该任务")
		return
	}
	claimFollowUpTask(&task, staff.ID)

	if err := database.DB.Model(&task).Update("status", utils.FollowUpStatusSkipped).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "跳过任务失败")
		return
	}
	database.DB.Preload("User").Preload("Staff").First(&task, id)
	response.Success(c, toFollowUpTaskVO(&task))
}

// MerchantListFollowUpTasks 管理端随访任务列表（支持用户昵称/手机号、状态、任务类型筛选，分页倒序）
func MerchantListFollowUpTasks(c *gin.Context) {
	p := utils.GetPagination(c)
	keyword := c.Query("keyword")
	status := c.Query("status")
	taskType := c.Query("task_type")

	countQuery := database.DB.Model(&models.FollowUpTask{})
	listQuery := database.DB.Model(&models.FollowUpTask{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN users ON users.id = follow_up_tasks.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
		listQuery = listQuery.
			Joins("LEFT JOIN users ON users.id = follow_up_tasks.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
	}
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("follow_up_tasks.status = ?", st)
			listQuery = listQuery.Where("follow_up_tasks.status = ?", st)
		}
	}
	if taskType != "" {
		if tt, err := strconv.Atoi(taskType); err == nil {
			countQuery = countQuery.Where("follow_up_tasks.task_type = ?", tt)
			listQuery = listQuery.Where("follow_up_tasks.task_type = ?", tt)
		}
	}

	var total int64
	countQuery.Count(&total)

	var tasks []models.FollowUpTask
	listQuery.Preload("User").Preload("Staff").
		Order("follow_up_tasks.created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&tasks)

	list := make([]followUpTaskVO, 0, len(tasks))
	for i := range tasks {
		list = append(list, toFollowUpTaskVO(&tasks[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// MerchantGetFollowUpTask 管理端随访任务详情（含用户与服务人员信息）
func MerchantGetFollowUpTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务ID错误")
		return
	}

	var task models.FollowUpTask
	if err := database.DB.Preload("User").Preload("Staff").First(&task, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "任务不存在")
		return
	}
	response.Success(c, toFollowUpTaskVO(&task))
}

// MerchantCreateFollowUpTask 管理端手动创建随访任务（来源固定为手动）
func MerchantCreateFollowUpTask(c *gin.Context) {
	var req FollowUpCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.UserID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请选择居民用户")
		return
	}
	if req.TaskType == 0 || req.TaskType > utils.FollowUpTypeAssessment {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务类型错误")
		return
	}
	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "居民用户不存在")
		return
	}

	planFollowTime := time.Now().Add(72 * time.Hour)
	if req.PlanFollowTime != nil {
		planFollowTime = *req.PlanFollowTime
	}

	task := models.FollowUpTask{
		UserID:         req.UserID,
		TaskType:       req.TaskType,
		SourceType:     utils.FollowUpSourceManual,
		SourceID:       req.SourceID,
		PlanFollowTime: &planFollowTime,
		StaffID:        req.StaffID,
		ContactMethod:  req.ContactMethod,
		Status:         utils.FollowUpStatusPending,
		Remark:         req.Remark,
	}
	if err := database.DB.Create(&task).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建随访任务失败")
		return
	}
	database.DB.Preload("User").Preload("Staff").First(&task, task.ID)
	response.Success(c, toFollowUpTaskVO(&task))
}

// MerchantCompleteFollowUpTask 管理端代执行随访任务（与服务人员完成逻辑一致，不做认领）
func MerchantCompleteFollowUpTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "任务ID错误")
		return
	}

	var req FollowUpCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var task models.FollowUpTask
	if err := database.DB.First(&task, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "任务不存在")
		return
	}

	resultRaw, err := json.Marshal(req.Result)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "随访结果格式错误")
		return
	}
	now := time.Now()
	updates := map[string]interface{}{
		"status":         utils.FollowUpStatusCompleted,
		"completed_at":   now,
		"contact_method": req.Result.ContactMethod,
		"result":         models.JSON(resultRaw),
	}
	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存随访结果失败")
		return
	}
	database.DB.Preload("User").Preload("Staff").First(&task, id)
	response.Success(c, toFollowUpTaskVO(&task))
}

// ============================================================
// 生命体征监测
// ============================================================

// MonitoringCreateRequest 生命体征录入请求（用户/服务人员共用）
type MonitoringCreateRequest struct {
	RecordType uint8                  `json:"record_type"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`
	Extra      map[string]interface{} `json:"extra"`
	RecordedAt *time.Time             `json:"recorded_at"`
	Remark     string                 `json:"remark"`
}

// monitoringVO 生命体征监测视图：extra 解析为对象，管理端列表可携带 User
type monitoringVO struct {
	ID         uint64                 `json:"id"`
	UserID     uint64                 `json:"user_id"`
	RecordType uint8                  `json:"record_type"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`
	Extra      map[string]interface{} `json:"extra"`
	RecordedBy uint64                 `json:"recorded_by"`
	RecordedAt *time.Time             `json:"recorded_at"`
	Remark     string                 `json:"remark"`
	CreatedAt  time.Time              `json:"created_at"`
	User       *models.User           `json:"user,omitempty"`
}

// toMonitoringVO 将监测记录转换为视图，extra JSON 做防御解析
func toMonitoringVO(m *models.HealthMonitoring) monitoringVO {
	return monitoringVO{
		ID:         m.ID,
		UserID:     m.UserID,
		RecordType: m.RecordType,
		Value:      m.Value,
		Unit:       m.Unit,
		Extra:      parseJSONObject(m.Extra),
		RecordedBy: m.RecordedBy,
		RecordedAt: m.RecordedAt,
		Remark:     m.Remark,
		CreatedAt:  m.CreatedAt,
		User:       m.User,
	}
}

// buildMonitoring 由请求构造监测记录，校验测量值与类型，extra/recorded_at 做默认兜底
func buildMonitoring(c *gin.Context, req *MonitoringCreateRequest, userID, recordedBy uint64) (*models.HealthMonitoring, bool) {
	if req.RecordType == 0 || req.RecordType > utils.MonitoringTypeWeight {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "监测类型错误")
		return nil, false
	}
	if req.Value <= 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请填写测量值")
		return nil, false
	}

	extraRaw := []byte("{}")
	if req.Extra != nil {
		if raw, err := json.Marshal(req.Extra); err == nil {
			extraRaw = raw
		}
	}
	recordedAt := time.Now()
	if req.RecordedAt != nil {
		recordedAt = *req.RecordedAt
	}

	return &models.HealthMonitoring{
		UserID:     userID,
		RecordType: req.RecordType,
		Value:      req.Value,
		Unit:       req.Unit,
		Extra:      models.JSON(extraRaw),
		RecordedBy: recordedBy,
		RecordedAt: &recordedAt,
		Remark:     req.Remark,
	}, true
}

// UserListMonitoring C端用户查看自己的生命体征记录（分页倒序，record_type 筛选可选）
func UserListMonitoring(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p := utils.GetPagination(c)
	recordType := c.Query("record_type")

	countQuery := database.DB.Model(&models.HealthMonitoring{}).Where("user_id = ?", userID)
	listQuery := database.DB.Model(&models.HealthMonitoring{}).Where("user_id = ?", userID)
	if recordType != "" {
		if rt, err := strconv.Atoi(recordType); err == nil {
			countQuery = countQuery.Where("record_type = ?", rt)
			listQuery = listQuery.Where("record_type = ?", rt)
		}
	}

	var total int64
	countQuery.Count(&total)

	var list []models.HealthMonitoring
	listQuery.
		Order("recorded_at DESC, id DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&list)

	voList := make([]monitoringVO, 0, len(list))
	for i := range list {
		voList = append(voList, toMonitoringVO(&list[i]))
	}
	response.Success(c, gin.H{
		"list":  voList,
		"total": total,
	})
}

// UserCreateMonitoring C端用户自助录入生命体征（recorded_by=0 表示用户本人）
func UserCreateMonitoring(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req MonitoringCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	mon, ok := buildMonitoring(c, &req, userID, 0)
	if !ok {
		return
	}
	if err := database.DB.Create(mon).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存监测记录失败")
		return
	}
	response.Success(c, toMonitoringVO(mon))
}

// StaffListResidentMonitoring 服务人员查看客户的生命体征记录（需存在服务关系，分页倒序）
func StaffListResidentMonitoring(c *gin.Context) {
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
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该客户监测数据")
		return
	}

	p := utils.GetPagination(c)
	var total int64
	database.DB.Model(&models.HealthMonitoring{}).Where("user_id = ?", userID).Count(&total)

	var list []models.HealthMonitoring
	database.DB.Where("user_id = ?", userID).
		Order("recorded_at DESC, id DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&list)

	voList := make([]monitoringVO, 0, len(list))
	for i := range list {
		voList = append(voList, toMonitoringVO(&list[i]))
	}
	response.Success(c, gin.H{
		"list":  voList,
		"total": total,
	})
}

// StaffCreateResidentMonitoring 服务人员为客户录入生命体征（需存在服务关系，recorded_by=当前服务人员）
func StaffCreateResidentMonitoring(c *gin.Context) {
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
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该客户录入监测数据")
		return
	}

	var req MonitoringCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	mon, ok := buildMonitoring(c, &req, userID, staff.ID)
	if !ok {
		return
	}
	if err := database.DB.Create(mon).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存监测记录失败")
		return
	}
	response.Success(c, toMonitoringVO(mon))
}

// MerchantListMonitoring 管理端生命体征记录列表（支持用户昵称/手机号、监测类型筛选，分页倒序）
func MerchantListMonitoring(c *gin.Context) {
	p := utils.GetPagination(c)
	keyword := c.Query("keyword")
	recordType := c.Query("record_type")

	countQuery := database.DB.Model(&models.HealthMonitoring{})
	listQuery := database.DB.Model(&models.HealthMonitoring{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN users ON users.id = health_monitoring.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
		listQuery = listQuery.
			Joins("LEFT JOIN users ON users.id = health_monitoring.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
	}
	if recordType != "" {
		if rt, err := strconv.Atoi(recordType); err == nil {
			countQuery = countQuery.Where("health_monitoring.record_type = ?", rt)
			listQuery = listQuery.Where("health_monitoring.record_type = ?", rt)
		}
	}

	var total int64
	countQuery.Count(&total)

	var list []models.HealthMonitoring
	listQuery.Preload("User").
		Order("health_monitoring.recorded_at DESC, health_monitoring.id DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&list)

	voList := make([]monitoringVO, 0, len(list))
	for i := range list {
		voList = append(voList, toMonitoringVO(&list[i]))
	}
	response.Success(c, gin.H{
		"list":  voList,
		"total": total,
	})
}

// ============================================================
// 健康宣教
// ============================================================

// EducationArticleUpsertRequest 健康宣教文章新增/编辑请求（Status 用指针区分草稿与未传）
type EducationArticleUpsertRequest struct {
	Title     string     `json:"title"`
	Category  string     `json:"category"`
	Cover     string     `json:"cover"`
	Content   string     `json:"content"`
	Tags      []string   `json:"tags"`
	Status    *uint8     `json:"status"`
	PublishAt *time.Time `json:"publish_at"`
}

// educationArticleVO 健康宣教文章视图：tags 解析为字符串数组
type educationArticleVO struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Category  string     `json:"category"`
	Cover     string     `json:"cover"`
	Content   string     `json:"content"`
	Tags      []string   `json:"tags"`
	Status    uint8      `json:"status"`
	PublishAt *time.Time `json:"publish_at"`
	Views     uint       `json:"views"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// parseJSONStringArray 将 JSON 字段解析为字符串数组，解析失败或为空时返回空数组
func parseJSONStringArray(raw models.JSON) []string {
	list := parseJSONArray(raw)
	out := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// toEducationArticleVO 将文章转换为视图，tags JSON 做防御解析
func toEducationArticleVO(a *models.HealthEducationArticle) educationArticleVO {
	return educationArticleVO{
		ID:        a.ID,
		Title:     a.Title,
		Category:  a.Category,
		Cover:     a.Cover,
		Content:   a.Content,
		Tags:      parseJSONStringArray(a.Tags),
		Status:    a.Status,
		PublishAt: a.PublishAt,
		Views:     a.Views,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// countStringIntersection 计算两个字符串数组的交集数，用于定向排序
func countStringIntersection(a, b []string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	count := 0
	for _, s := range b {
		if _, ok := set[s]; ok {
			count++
		}
	}
	return count
}

// UserListEducationArticles C端用户查看已发布的健康宣教文章，支持分类筛选。
// 定向排序：取当前用户档案慢病标签与文章标签的交集数，交集多的排前；交集相同按发布时间倒序。
// 无档案时取全量按发布时间倒序。
func UserListEducationArticles(c *gin.Context) {
	userID := middleware.GetUserID(c)
	category := c.Query("category")

	query := database.DB.Model(&models.HealthEducationArticle{}).
		Where("status = ?", utils.EducationStatusPublished)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	var articles []models.HealthEducationArticle
	query.Order("publish_at DESC, id DESC").Find(&articles)

	// 取当前用户慢病标签
	userTags := []string{}
	var record models.HealthRecord
	if err := database.DB.Where("user_id = ?", userID).First(&record).Error; err == nil {
		userTags = parseJSONStringArray(record.ChronicTags)
	}

	// 定向排序：交集数多者优先，交集相同按 publish_at 倒序
	sort.SliceStable(articles, func(i, j int) bool {
		ci := countStringIntersection(userTags, parseJSONStringArray(articles[i].Tags))
		cj := countStringIntersection(userTags, parseJSONStringArray(articles[j].Tags))
		if ci != cj {
			return ci > cj
		}
		ti, tj := articles[i].PublishAt, articles[j].PublishAt
		if ti == nil {
			return false
		}
		if tj == nil {
			return true
		}
		if !ti.Equal(*tj) {
			return ti.After(*tj)
		}
		return articles[i].ID > articles[j].ID
	})

	list := make([]educationArticleVO, 0, len(articles))
	for i := range articles {
		list = append(list, toEducationArticleVO(&articles[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": len(list),
	})
}

// UserGetEducationArticle C端用户查看文章详情（仅已发布可见，浏览量 +1 异步执行忽略错误）
func UserGetEducationArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}

	var article models.HealthEducationArticle
	if err := database.DB.First(&article, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}
	if article.Status != utils.EducationStatusPublished {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}

	// 浏览量 +1（异步执行，忽略错误）
	go func() {
		database.DB.Model(&models.HealthEducationArticle{}).
			Where("id = ?", id).
			UpdateColumn("views", gorm.Expr("views + 1"))
	}()

	response.Success(c, toEducationArticleVO(&article))
}

// StaffListEducationArticles 服务人员获取已发布的健康宣教文章（供随访选用，按发布时间倒序）
func StaffListEducationArticles(c *gin.Context) {
	var articles []models.HealthEducationArticle
	database.DB.Where("status = ?", utils.EducationStatusPublished).
		Order("publish_at DESC, id DESC").
		Find(&articles)

	list := make([]educationArticleVO, 0, len(articles))
	for i := range articles {
		list = append(list, toEducationArticleVO(&articles[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": len(list),
	})
}

// MerchantListEducationArticles 管理端宣教文章列表（含草稿，支持标题、分类、状态筛选，分页倒序）
func MerchantListEducationArticles(c *gin.Context) {
	p := utils.GetPagination(c)
	keyword := c.Query("keyword")
	category := c.Query("category")
	status := c.Query("status")

	countQuery := database.DB.Model(&models.HealthEducationArticle{})
	listQuery := database.DB.Model(&models.HealthEducationArticle{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.Where("title LIKE ?", like)
		listQuery = listQuery.Where("title LIKE ?", like)
	}
	if category != "" {
		countQuery = countQuery.Where("category = ?", category)
		listQuery = listQuery.Where("category = ?", category)
	}
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("status = ?", st)
			listQuery = listQuery.Where("status = ?", st)
		}
	}

	var total int64
	countQuery.Count(&total)

	var articles []models.HealthEducationArticle
	listQuery.
		Order("updated_at DESC, id DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&articles)

	list := make([]educationArticleVO, 0, len(articles))
	for i := range articles {
		list = append(list, toEducationArticleVO(&articles[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// MerchantCreateEducationArticle 管理端新增健康宣教文章
func MerchantCreateEducationArticle(c *gin.Context) {
	var req EducationArticleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.Title == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请填写标题")
		return
	}

	tagsRaw, err := json.Marshal(req.Tags)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "标签格式错误")
		return
	}

	status := utils.EducationStatusDraft
	if req.Status != nil {
		status = *req.Status
	}
	article := models.HealthEducationArticle{
		Title:    req.Title,
		Category: req.Category,
		Cover:    req.Cover,
		Content:  req.Content,
		Tags:     models.JSON(tagsRaw),
		Status:   status,
	}
	// 发布且未指定发布时间时回填当前时间，保证用户端按发布时间倒序展示
	if status == utils.EducationStatusPublished && req.PublishAt == nil {
		now := time.Now()
		article.PublishAt = &now
	} else {
		article.PublishAt = req.PublishAt
	}

	if err := database.DB.Create(&article).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建文章失败")
		return
	}
	response.Success(c, toEducationArticleVO(&article))
}

// MerchantUpdateEducationArticle 管理端编辑健康宣教文章（仅更新传入字段）
func MerchantUpdateEducationArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}

	var req EducationArticleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var article models.HealthEducationArticle
	if err := database.DB.First(&article, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Cover != "" {
		updates["cover"] = req.Cover
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Tags != nil {
		tagsRaw, err := json.Marshal(req.Tags)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "标签格式错误")
			return
		}
		updates["tags"] = models.JSON(tagsRaw)
	}
	if req.PublishAt != nil {
		updates["publish_at"] = req.PublishAt
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		// 发布且未指定发布时间时回填当前时间，保证用户端按发布时间倒序展示
		if *req.Status == utils.EducationStatusPublished && req.PublishAt == nil {
			now := time.Now()
			updates["publish_at"] = now
		}
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存文章失败")
			return
		}
	}
	database.DB.First(&article, id)
	response.Success(c, toEducationArticleVO(&article))
}

// MerchantDeleteEducationArticle 管理端删除健康宣教文章
func MerchantDeleteEducationArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}
	if err := database.DB.Delete(&models.HealthEducationArticle{}, id).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除文章失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}
