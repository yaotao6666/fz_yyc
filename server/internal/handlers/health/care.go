package health

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	serviceStaffHandler "fz_yyc_api/internal/handlers/service_staff"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// CareVisitCreateRequest 服务人员录入上门照护记录的请求。
// plan_id / order_id 至少提供一个；JSON 复杂字段用切片/对象绑定，写入时再序列化为 models.JSON，
// 避免硬编码结构体导致解析失败。
type CareVisitCreateRequest struct {
	PlanID         *uint64                  `json:"plan_id"`
	OrderID        *uint64                  `json:"order_id"`
	VisitAt        *time.Time               `json:"visit_at"`
	NursingItems   []map[string]interface{} `json:"nursing_items"`
	Vitals         map[string]interface{}   `json:"vitals"`
	Photos         []string                 `json:"photos"`
	Remark         string                   `json:"remark"`
	FollowUpAdvice string                   `json:"follow_up_advice"`
}

// CarePlanUpsertRequest 管理端新增/编辑照护计划的请求（创建与编辑共用）。
// Status 用指针以区分「未传」与「显式置为草稿(0)」，配合 map 更新实现局部修改。
type CarePlanUpsertRequest struct {
	UserID          uint64                   `json:"user_id"`
	Name            string                   `json:"name"`
	PlanType        uint8                    `json:"plan_type"`
	StartDate       string                   `json:"start_date"`
	EndDate         string                   `json:"end_date"`
	Frequency       string                   `json:"frequency"`
	Goals           string                   `json:"goals"`
	Items           []map[string]interface{} `json:"items"`
	AssignedStaffID *uint64                  `json:"assigned_staff_id"`
	OrderID         *uint64                  `json:"order_id"`
	Status          *uint8                   `json:"status"`
}

// carePlanVO 照护计划视图：护理项解析为数组，附带指派人员姓名与上门照护次数；
// 管理端列表/详情可携带 User，C端/服务人员端不携带。
type carePlanVO struct {
	ID              uint64        `json:"id"`
	UserID          uint64        `json:"user_id"`
	Name            string        `json:"name"`
	PlanType        uint8         `json:"plan_type"`
	StartDate       string        `json:"start_date"`
	EndDate         string        `json:"end_date"`
	Frequency       string        `json:"frequency"`
	Goals           string        `json:"goals"`
	Items           []interface{} `json:"items"`
	AssignedStaffID *uint64       `json:"assigned_staff_id"`
	AssignedStaff   string        `json:"assigned_staff,omitempty"` // 指派服务人员姓名
	OrderID         *uint64       `json:"order_id"`
	Status          uint8         `json:"status"`
	VisitCount      int64         `json:"visit_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	User            *models.User  `json:"user,omitempty"`
}

// careVisitVO 上门照护记录视图：护理项/照片解析为数组、生命体征解析为对象，附带录入人员姓名；
// 管理端列表/详情可携带 User，C端/服务人员端不携带。
type careVisitVO struct {
	ID             uint64                 `json:"id"`
	PlanID         *uint64                `json:"plan_id"`
	OrderID        *uint64                `json:"order_id"`
	UserID         uint64                 `json:"user_id"`
	StaffID        uint64                 `json:"staff_id"`
	StaffName      string                 `json:"staff_name,omitempty"` // 录入服务人员姓名
	VisitAt        *time.Time             `json:"visit_at"`
	NursingItems   []interface{}          `json:"nursing_items"`
	Vitals         map[string]interface{} `json:"vitals"`
	Photos         []interface{}          `json:"photos"`
	Remark         string                 `json:"remark"`
	FollowUpAdvice string                 `json:"follow_up_advice"`
	CreatedAt      time.Time              `json:"created_at"`
	User           *models.User           `json:"user,omitempty"`
}

// parseJSONArray 将 JSON 字段解析为数组返回，解析失败或为空时返回空数组。
// 兼容字符串存储（JSON 字符串内部再包一层数组），与 fitting.go 的解析思路一致。
func parseJSONArray(raw models.JSON) []interface{} {
	if len(raw) == 0 {
		return []interface{}{}
	}
	var list []interface{}
	if err := json.Unmarshal(raw, &list); err == nil {
		return list
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		var inner []interface{}
		if err := json.Unmarshal([]byte(str), &inner); err == nil {
			return inner
		}
	}
	return []interface{}{}
}

// parseJSONObject 将 JSON 字段解析为对象返回，解析失败或为空时返回空对象。
func parseJSONObject(raw models.JSON) map[string]interface{} {
	if len(raw) == 0 {
		return map[string]interface{}{}
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return map[string]interface{}{}
	}
	return obj
}

// toCarePlanVO 将计划记录转换为视图，护理项解析为数组，user 为空时自动省略
func toCarePlanVO(plan *models.CarePlan, visitCount int64) carePlanVO {
	vo := carePlanVO{
		ID:              plan.ID,
		UserID:          plan.UserID,
		Name:            plan.Name,
		PlanType:        plan.PlanType,
		StartDate:       plan.StartDate,
		EndDate:         plan.EndDate,
		Frequency:       plan.Frequency,
		Goals:           plan.Goals,
		Items:           parseJSONArray(plan.Items),
		AssignedStaffID: plan.AssignedStaffID,
		OrderID:         plan.OrderID,
		Status:          plan.Status,
		VisitCount:      visitCount,
		CreatedAt:       plan.CreatedAt,
		UpdatedAt:       plan.UpdatedAt,
		User:            plan.User,
	}
	if plan.AssignedStaff != nil {
		vo.AssignedStaff = plan.AssignedStaff.Name
	}
	return vo
}

// toCareVisitVO 将照护记录转换为视图，JSON 字段做防御解析，user/staff 为空时自动省略
func toCareVisitVO(visit *models.CareVisit) careVisitVO {
	vo := careVisitVO{
		ID:             visit.ID,
		PlanID:         visit.PlanID,
		OrderID:        visit.OrderID,
		UserID:         visit.UserID,
		StaffID:        visit.StaffID,
		VisitAt:        visit.VisitAt,
		NursingItems:   parseJSONArray(visit.NursingItems),
		Vitals:         parseJSONObject(visit.Vitals),
		Photos:         parseJSONArray(visit.Photos),
		Remark:         visit.Remark,
		FollowUpAdvice: visit.FollowUpAdvice,
		CreatedAt:      visit.CreatedAt,
		User:           visit.User,
	}
	if visit.Staff != nil {
		vo.StaffName = visit.Staff.Name
	}
	return vo
}

// collectPlanIDs 提取计划ID列表，用于批量统计上门照护次数，避免逐条查询
func collectPlanIDs(plans []models.CarePlan) []uint64 {
	ids := make([]uint64, 0, len(plans))
	for _, plan := range plans {
		ids = append(ids, plan.ID)
	}
	return ids
}

// loadCarePlanVisitCounts 批量统计各计划的上门照护次数，返回 plan_id -> count 映射
func loadCarePlanVisitCounts(planIDs []uint64) map[uint64]int64 {
	counts := map[uint64]int64{}
	if len(planIDs) == 0 {
		return counts
	}
	var rows []struct {
		PlanID uint64
		Count  int64
	}
	database.DB.Model(&models.CareVisit{}).
		Select("plan_id, COUNT(*) AS count").
		Where("plan_id IN ?", planIDs).
		Group("plan_id").
		Scan(&rows)
	for _, row := range rows {
		counts[row.PlanID] = row.Count
	}
	return counts
}

// countCareVisitByPlan 统计单个计划的上门照护次数
func countCareVisitByPlan(planID uint64) int64 {
	var count int64
	database.DB.Model(&models.CareVisit{}).Where("plan_id = ?", planID).Count(&count)
	return count
}

// UserListCarePlans C端用户查看自己的照护计划列表（分页倒序）
func UserListCarePlans(c *gin.Context) {
	userID := middleware.GetUserID(c)
	p := utils.GetPagination(c)

	var total int64
	database.DB.Model(&models.CarePlan{}).Where("user_id = ?", userID).Count(&total)

	var plans []models.CarePlan
	database.DB.Preload("AssignedStaff").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&plans)

	visitCounts := loadCarePlanVisitCounts(collectPlanIDs(plans))
	list := make([]carePlanVO, 0, len(plans))
	for i := range plans {
		list = append(list, toCarePlanVO(&plans[i], visitCounts[plans[i].ID]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// UserGetCarePlan C端用户查看自己的照护计划详情（含上门照护记录）
func UserGetCarePlan(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "计划ID错误")
		return
	}

	var plan models.CarePlan
	if err := database.DB.Preload("AssignedStaff").First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}
	if plan.UserID != userID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该计划")
		return
	}

	var visits []models.CareVisit
	database.DB.Where("plan_id = ?", id).Order("visit_at DESC, id DESC").Find(&visits)

	visitList := make([]careVisitVO, 0, len(visits))
	for i := range visits {
		visitList = append(visitList, toCareVisitVO(&visits[i]))
	}
	response.Success(c, gin.H{
		"plan":   toCarePlanVO(&plan, countCareVisitByPlan(plan.ID)),
		"visits": visitList,
	})
}

// StaffListCarePlans 服务人员查看指派给自己的照护计划列表（分页倒序）
func StaffListCarePlans(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	p := utils.GetPagination(c)

	var total int64
	database.DB.Model(&models.CarePlan{}).Where("assigned_staff_id = ?", staff.ID).Count(&total)

	var plans []models.CarePlan
	database.DB.Preload("AssignedStaff").
		Where("assigned_staff_id = ?", staff.ID).
		Order("created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&plans)

	visitCounts := loadCarePlanVisitCounts(collectPlanIDs(plans))
	list := make([]carePlanVO, 0, len(plans))
	for i := range plans {
		list = append(list, toCarePlanVO(&plans[i], visitCounts[plans[i].ID]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// StaffGetCarePlan 服务人员查看指派给自己的照护计划详情（含上门照护记录）
func StaffGetCarePlan(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "计划ID错误")
		return
	}

	var plan models.CarePlan
	if err := database.DB.Preload("AssignedStaff").First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}
	if plan.AssignedStaffID == nil || *plan.AssignedStaffID != staff.ID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该计划")
		return
	}

	var visits []models.CareVisit
	database.DB.Where("plan_id = ?", id).Order("visit_at DESC, id DESC").Find(&visits)

	visitList := make([]careVisitVO, 0, len(visits))
	for i := range visits {
		visitList = append(visitList, toCareVisitVO(&visits[i]))
	}
	response.Success(c, gin.H{
		"plan":   toCarePlanVO(&plan, countCareVisitByPlan(plan.ID)),
		"visits": visitList,
	})
}

// StaffCreateCareVisit 服务人员录入上门照护记录。
// 校验 plan_id/order_id 至少提供一个且均须指派给当前服务人员；
// 居民取计划或订单的 user_id，两者同时提供时须为同一居民。
func StaffCreateCareVisit(c *gin.Context) {
	staff, err := serviceStaffHandler.GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	var req CareVisitCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if req.PlanID == nil && req.OrderID == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请关联照护计划或服务订单")
		return
	}
	if req.VisitAt == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请填写到访时间")
		return
	}

	// 归属校验：计划/订单必须指派给当前服务人员，并确定居民 user_id
	var userID uint64
	if req.PlanID != nil {
		var plan models.CarePlan
		if err := database.DB.First(&plan, *req.PlanID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "照护计划不存在")
			return
		}
		if plan.AssignedStaffID == nil || *plan.AssignedStaffID != staff.ID {
			response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该计划录入记录")
			return
		}
		userID = plan.UserID
	}
	if req.OrderID != nil {
		var order models.Order
		if err := database.DB.First(&order, *req.OrderID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "服务订单不存在")
			return
		}
		if order.AssignedStaffID == nil || *order.AssignedStaffID != staff.ID {
			response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该订单录入记录")
			return
		}
		if req.PlanID != nil && userID != order.UserID {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "照护计划与订单不属于同一居民")
			return
		}
		if req.PlanID == nil {
			userID = order.UserID
		}
	}

	// JSON 字段序列化前补空默认值，保证写入合法 JSON（而非 null）
	if req.NursingItems == nil {
		req.NursingItems = []map[string]interface{}{}
	}
	if req.Vitals == nil {
		req.Vitals = map[string]interface{}{}
	}
	if req.Photos == nil {
		req.Photos = []string{}
	}
	nursingRaw, err := json.Marshal(req.NursingItems)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "护理项格式错误")
		return
	}
	vitalsRaw, err := json.Marshal(req.Vitals)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "生命体征格式错误")
		return
	}
	photosRaw, err := json.Marshal(req.Photos)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "照片格式错误")
		return
	}

	visit := models.CareVisit{
		PlanID:         req.PlanID,
		OrderID:        req.OrderID,
		UserID:         userID,
		StaffID:        staff.ID,
		VisitAt:        req.VisitAt,
		NursingItems:   models.JSON(nursingRaw),
		Vitals:         models.JSON(vitalsRaw),
		Photos:         models.JSON(photosRaw),
		Remark:         req.Remark,
		FollowUpAdvice: req.FollowUpAdvice,
	}
	if err := database.DB.Create(&visit).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存照护记录失败")
		return
	}
	response.Success(c, toCareVisitVO(&visit))
}

// MerchantListCarePlans 管理端照护计划列表（支持用户昵称/手机号、计划类型、状态筛选，分页倒序）
func MerchantListCarePlans(c *gin.Context) {
	p := utils.GetPagination(c)
	keyword := c.Query("keyword")
	planType := c.Query("plan_type")
	status := c.Query("status")

	countQuery := database.DB.Model(&models.CarePlan{})
	listQuery := database.DB.Model(&models.CarePlan{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN users ON users.id = care_plans.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
		listQuery = listQuery.
			Joins("LEFT JOIN users ON users.id = care_plans.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
	}
	if planType != "" {
		if pt, err := strconv.Atoi(planType); err == nil {
			countQuery = countQuery.Where("care_plans.plan_type = ?", pt)
			listQuery = listQuery.Where("care_plans.plan_type = ?", pt)
		}
	}
	if status != "" {
		if st, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("care_plans.status = ?", st)
			listQuery = listQuery.Where("care_plans.status = ?", st)
		}
	}

	var total int64
	countQuery.Count(&total)

	var plans []models.CarePlan
	listQuery.Preload("User").Preload("AssignedStaff").
		Order("care_plans.created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&plans)

	visitCounts := loadCarePlanVisitCounts(collectPlanIDs(plans))
	list := make([]carePlanVO, 0, len(plans))
	for i := range plans {
		list = append(list, toCarePlanVO(&plans[i], visitCounts[plans[i].ID]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// MerchantGetCarePlan 管理端照护计划详情（含用户信息与上门照护记录）
func MerchantGetCarePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "计划ID错误")
		return
	}

	var plan models.CarePlan
	if err := database.DB.Preload("User").Preload("AssignedStaff").First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}

	var visits []models.CareVisit
	database.DB.Preload("Staff").
		Where("plan_id = ?", id).
		Order("visit_at DESC, id DESC").
		Find(&visits)

	visitList := make([]careVisitVO, 0, len(visits))
	for i := range visits {
		visitList = append(visitList, toCareVisitVO(&visits[i]))
	}
	response.Success(c, gin.H{
		"plan":   toCarePlanVO(&plan, countCareVisitByPlan(plan.ID)),
		"visits": visitList,
	})
}

// MerchantCreateCarePlan 管理端新增照护计划
func MerchantCreateCarePlan(c *gin.Context) {
	var req CarePlanUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if req.UserID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请选择居民用户")
		return
	}
	if req.Name == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请填写计划名称")
		return
	}

	itemsRaw, err := json.Marshal(req.Items)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "护理项格式错误")
		return
	}

	status := utils.CarePlanStatusDraft
	if req.Status != nil {
		status = *req.Status
	}

	plan := models.CarePlan{
		UserID:          req.UserID,
		Name:            req.Name,
		PlanType:        req.PlanType,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		Frequency:       req.Frequency,
		Goals:           req.Goals,
		Items:           models.JSON(itemsRaw),
		AssignedStaffID: req.AssignedStaffID,
		OrderID:         req.OrderID,
		Status:          status,
	}
	if err := database.DB.Create(&plan).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建计划失败")
		return
	}
	response.Success(c, toCarePlanVO(&plan, 0))
}

// MerchantUpdateCarePlan 管理端编辑照护计划（仅更新传入字段）
func MerchantUpdateCarePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "计划ID错误")
		return
	}

	var req CarePlanUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var plan models.CarePlan
	if err := database.DB.First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.PlanType != 0 {
		updates["plan_type"] = req.PlanType
	}
	if req.StartDate != "" {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		updates["end_date"] = req.EndDate
	}
	if req.Frequency != "" {
		updates["frequency"] = req.Frequency
	}
	if req.Goals != "" {
		updates["goals"] = req.Goals
	}
	if req.Items != nil {
		itemsRaw, err := json.Marshal(req.Items)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "护理项格式错误")
			return
		}
		updates["items"] = models.JSON(itemsRaw)
	}
	if req.AssignedStaffID != nil {
		updates["assigned_staff_id"] = *req.AssignedStaffID
	}
	if req.OrderID != nil {
		updates["order_id"] = *req.OrderID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&plan).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存计划失败")
			return
		}
	}

	if err := database.DB.Preload("User").Preload("AssignedStaff").First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}
	response.Success(c, toCarePlanVO(&plan, countCareVisitByPlan(plan.ID)))
}

// MerchantDeleteCarePlan 管理端删除照护计划。
// 已有关联上门照护记录的计划拒绝删除：照护记录是服务闭环的历史凭证，
// 仅删计划会破坏记录溯源，因此提示先处理关联记录。
func MerchantDeleteCarePlan(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "计划ID错误")
		return
	}

	var plan models.CarePlan
	if err := database.DB.First(&plan, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "计划不存在")
		return
	}

	var visitCount int64
	database.DB.Model(&models.CareVisit{}).Where("plan_id = ?", id).Count(&visitCount)
	if visitCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该计划已有关联的上门照护记录，无法删除")
		return
	}

	if err := database.DB.Delete(&models.CarePlan{}, id).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除计划失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}

// MerchantListCareVisits 管理端上门照护记录列表（支持按计划/居民筛选，分页倒序）
func MerchantListCareVisits(c *gin.Context) {
	p := utils.GetPagination(c)
	planID := c.Query("plan_id")
	userID := c.Query("user_id")

	countQuery := database.DB.Model(&models.CareVisit{})
	listQuery := database.DB.Model(&models.CareVisit{})
	if planID != "" {
		if pid, err := strconv.ParseUint(planID, 10, 64); err == nil {
			countQuery = countQuery.Where("plan_id = ?", pid)
			listQuery = listQuery.Where("plan_id = ?", pid)
		}
	}
	if userID != "" {
		if uid, err := strconv.ParseUint(userID, 10, 64); err == nil {
			countQuery = countQuery.Where("user_id = ?", uid)
			listQuery = listQuery.Where("user_id = ?", uid)
		}
	}

	var total int64
	countQuery.Count(&total)

	var visits []models.CareVisit
	listQuery.Preload("User").Preload("Staff").
		Order("care_visits.created_at DESC").
		Offset(p.GetOffset()).
		Limit(p.PageSize).
		Find(&visits)

	list := make([]careVisitVO, 0, len(visits))
	for i := range visits {
		list = append(list, toCareVisitVO(&visits[i]))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}
