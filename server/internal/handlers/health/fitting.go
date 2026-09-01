package health

import (
	"encoding/json"
	"errors"
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

// RecommendedProductItem 推荐商品项（请求入参：商品ID + 推荐理由）
type RecommendedProductItem struct {
	ProductID uint64 `json:"product_id"`
	Reason    string `json:"reason"`
}

// FittingRecommendationCreateRequest 服务人员生成适配建议请求
type FittingRecommendationCreateRequest struct {
	RecordID            *uint64                  `json:"record_id"`
	AssessmentID        *uint64                  `json:"assessment_id"`
	SymptomDesc         string                   `json:"symptom_desc"`
	FittingResult       string                   `json:"fitting_result"`
	RecommendedProducts []RecommendedProductItem `json:"recommended_products"`
}

// FittingRecommendationUpdateRequest 管理端编辑适配建议请求（仅更新传入字段）
type FittingRecommendationUpdateRequest struct {
	FittingResult       string                   `json:"fitting_result"`
	RecommendedProducts []RecommendedProductItem `json:"recommended_products"`
	Status              *uint8                   `json:"status"`
	OrderID             *uint64                  `json:"order_id"`
}

// fittingRecommendationVO 适配建议视图：
// 推荐商品解析为数组返回（兼容字符串/数组两种存储），管理端列表可携带 User，C端/服务人员端不携带
type fittingRecommendationVO struct {
	ID                  uint64        `json:"id"`
	UserID              uint64        `json:"user_id"`
	RecordID            *uint64       `json:"record_id"`
	AssessmentID        *uint64       `json:"assessment_id"`
	SymptomDesc         string        `json:"symptom_desc"`
	FittingResult       string        `json:"fitting_result"`
	RecommendedProducts []interface{} `json:"recommended_products"`
	StaffID             *uint64       `json:"staff_id"`
	Status              uint8         `json:"status"`
	OrderID             *uint64       `json:"order_id"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	User                *models.User  `json:"user,omitempty"`
}

// normalizeRecommendedProducts 校验推荐商品并生成快照数组。
// 对每个 product_id 查询商品，不存在或未上架返回错误；
// 快照字段为 {product_id,name,reason,sale_type}，统一序列化为 models.JSON 存储，
// 供服务人员创建与管理端更新复用，保持两端写入逻辑一致。
func normalizeRecommendedProducts(items []RecommendedProductItem) (models.JSON, error) {
	snapshots := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil || product.Status != 1 {
			return nil, errors.New("推荐商品不可用")
		}
		snapshots = append(snapshots, map[string]interface{}{
			"product_id": product.ID,
			"name":       product.Name,
			"reason":     item.Reason,
			"sale_type":  product.SaleType,
		})
	}
	raw, err := json.Marshal(snapshots)
	if err != nil {
		return nil, err
	}
	return models.JSON(raw), nil
}

// parseRecommendedProducts 将推荐商品 JSON 字段解析为数组返回。
// 兼容字符串存储（JSON 字符串内部再包一层数组），解析失败或为空时返回空数组。
func parseRecommendedProducts(raw models.JSON) []interface{} {
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

// toFittingRecommendationVO 将建议记录转换为视图，推荐商品解析为数组，user 为空时自动省略
func toFittingRecommendationVO(record *models.FittingRecommendation, user *models.User) fittingRecommendationVO {
	return fittingRecommendationVO{
		ID:                  record.ID,
		UserID:              record.UserID,
		RecordID:            record.RecordID,
		AssessmentID:        record.AssessmentID,
		SymptomDesc:         record.SymptomDesc,
		FittingResult:       record.FittingResult,
		RecommendedProducts: parseRecommendedProducts(record.RecommendedProducts),
		StaffID:             record.StaffID,
		Status:              record.Status,
		OrderID:             record.OrderID,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
		User:                user,
	}
}

// UserListFittingRecommendations C端用户查看自己的适配建议列表（分页倒序）
func UserListFittingRecommendations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	database.DB.Model(&models.FittingRecommendation{}).
		Where("user_id = ?", userID).Count(&total)

	var records []models.FittingRecommendation
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	list := make([]fittingRecommendationVO, 0, len(records))
	for i := range records {
		list = append(list, toFittingRecommendationVO(&records[i], nil))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// UserGetFittingRecommendation C端用户查看自己的适配建议详情
func UserGetFittingRecommendation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "建议ID错误")
		return
	}

	var record models.FittingRecommendation
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "建议不存在")
		return
	}
	if record.UserID != userID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该建议")
		return
	}
	response.Success(c, toFittingRecommendationVO(&record, nil))
}

// UserConfirmFittingRecommendation C端用户确认适配建议（仅草稿状态可确认，确认后状态置为1）
func UserConfirmFittingRecommendation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "建议ID错误")
		return
	}

	var record models.FittingRecommendation
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "建议不存在")
		return
	}
	if record.UserID != userID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权操作该建议")
		return
	}
	if record.Status != utils.FittingStatusDraft {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "当前状态不可确认")
		return
	}

	record.Status = utils.FittingStatusConfirmed
	if err := database.DB.Save(&record).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "确认建议失败")
		return
	}
	response.Success(c, toFittingRecommendationVO(&record, nil))
}

// StaffListResidentFittingRecommendations 服务人员查看客户的适配建议列表（需存在服务关系）
func StaffListResidentFittingRecommendations(c *gin.Context) {
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
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看该客户建议")
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

	var total int64
	database.DB.Model(&models.FittingRecommendation{}).
		Where("user_id = ?", userID).Count(&total)

	var records []models.FittingRecommendation
	database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	list := make([]fittingRecommendationVO, 0, len(records))
	for i := range records {
		list = append(list, toFittingRecommendationVO(&records[i], nil))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// StaffCreateFittingRecommendation 服务人员为客户生成适配建议（需存在服务关系）
func StaffCreateFittingRecommendation(c *gin.Context) {
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
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权为该客户生成建议")
		return
	}

	var req FittingRecommendationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
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

	recommendedRaw, err := normalizeRecommendedProducts(req.RecommendedProducts)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	record := models.FittingRecommendation{
		UserID:              userID,
		RecordID:            recordID,
		AssessmentID:        req.AssessmentID,
		SymptomDesc:         req.SymptomDesc,
		FittingResult:       req.FittingResult,
		RecommendedProducts: recommendedRaw,
		StaffID:             &staff.ID,
		Status:              utils.FittingStatusDraft,
	}
	if err := database.DB.Create(&record).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "生成建议失败")
		return
	}
	response.Success(c, toFittingRecommendationVO(&record, nil))
}

// MerchantListFittingRecommendations 管理端适配建议列表（支持按用户昵称/手机号与状态筛选）
func MerchantListFittingRecommendations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	keyword := c.Query("keyword")
	status := c.Query("status")

	countQuery := database.DB.Model(&models.FittingRecommendation{})
	listQuery := database.DB.Model(&models.FittingRecommendation{})
	if keyword != "" {
		like := "%" + keyword + "%"
		countQuery = countQuery.
			Joins("LEFT JOIN users ON users.id = fitting_recommendations.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
		listQuery = listQuery.
			Joins("LEFT JOIN users ON users.id = fitting_recommendations.user_id").
			Where("users.nickname LIKE ? OR users.phone LIKE ?", like, like)
	}
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			countQuery = countQuery.Where("fitting_recommendations.status = ?", s)
			listQuery = listQuery.Where("fitting_recommendations.status = ?", s)
		}
	}

	var total int64
	countQuery.Count(&total)

	var records []models.FittingRecommendation
	listQuery.Preload("User").
		Order("fitting_recommendations.created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	list := make([]fittingRecommendationVO, 0, len(records))
	for i := range records {
		list = append(list, toFittingRecommendationVO(&records[i], records[i].User))
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// MerchantGetFittingRecommendation 管理端适配建议详情（含用户信息）
func MerchantGetFittingRecommendation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "建议ID错误")
		return
	}

	var record models.FittingRecommendation
	if err := database.DB.Preload("User").First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "建议不存在")
		return
	}
	response.Success(c, toFittingRecommendationVO(&record, record.User))
}

// MerchantUpdateFittingRecommendation 管理端编辑适配建议（仅更新传入字段，推荐商品重新校验并快照）
func MerchantUpdateFittingRecommendation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "建议ID错误")
		return
	}

	var req FittingRecommendationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var record models.FittingRecommendation
	if err := database.DB.First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "建议不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.FittingResult != "" {
		updates["fitting_result"] = req.FittingResult
	}
	if req.RecommendedProducts != nil {
		recommendedRaw, err := normalizeRecommendedProducts(req.RecommendedProducts)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
			return
		}
		updates["recommended_products"] = recommendedRaw
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.OrderID != nil {
		updates["order_id"] = *req.OrderID
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&record).Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存建议失败")
			return
		}
	}

	if err := database.DB.Preload("User").First(&record, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "建议不存在")
		return
	}
	response.Success(c, toFittingRecommendationVO(&record, record.User))
}

// MerchantDeleteFittingRecommendation 管理端删除适配建议
func MerchantDeleteFittingRecommendation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "建议ID错误")
		return
	}
	if err := database.DB.Delete(&models.FittingRecommendation{}, id).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除建议失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", gin.H{"id": id})
}
