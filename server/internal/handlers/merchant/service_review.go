package merchant

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	reviewpkg "fz_yyc_api/internal/services/review"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// fillReviewStaffInfo 批量填充评价的服务人员信息
func fillReviewStaffInfo(reviews []models.ServiceReview) map[uint64]gin.H {
	staffMap := make(map[uint64]gin.H)
	if len(reviews) == 0 {
		return staffMap
	}
	staffIDs := make([]uint64, 0, len(reviews))
	for _, r := range reviews {
		staffIDs = append(staffIDs, r.StaffID)
	}
	var staffs []models.ServiceStaff
	database.DB.Select("id, name, phone").Where("id IN ?", staffIDs).Find(&staffs)
	for _, s := range staffs {
		staffMap[s.ID] = gin.H{"id": s.ID, "name": s.Name, "phone": s.Phone}
	}
	return staffMap
}

// fillReviewOrderInfo 批量填充评价的订单信息
func fillReviewOrderInfo(reviews []models.ServiceReview) map[uint64]gin.H {
	orderMap := make(map[uint64]gin.H)
	if len(reviews) == 0 {
		return orderMap
	}
	orderIDs := make([]uint64, 0, len(reviews))
	for _, r := range reviews {
		orderIDs = append(orderIDs, r.OrderID)
	}
	var orders []models.Order
	database.DB.Select("id, order_no").Where("id IN ?", orderIDs).Find(&orders)
	for _, o := range orders {
		orderMap[o.ID] = gin.H{"id": o.ID, "order_no": o.OrderNo}
	}
	return orderMap
}

// GetServiceReviews 服务评价列表
func GetServiceReviews(c *gin.Context) {
	status := c.Query("status")
	score := c.Query("score")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.ServiceReview{})
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}
	if score != "" {
		if s, err := strconv.Atoi(score); err == nil {
			query = query.Where("score = ?", s)
		}
	}
	if keyword != "" {
		query = query.Where("staff_id IN (?) OR user_id IN (?)",
			database.DB.Model(&models.ServiceStaff{}).Select("id").
				Where("name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%"),
			database.DB.Model(&models.User{}).Select("id").
				Where("nickname LIKE ?", "%"+keyword+"%"))
	}

	var total int64
	query.Count(&total)

	var reviews []models.ServiceReview
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&reviews)

	staffMap := fillReviewStaffInfo(reviews)
	orderMap := fillReviewOrderInfo(reviews)

	list := make([]gin.H, 0, len(reviews))
	for _, r := range reviews {
		item := gin.H{
			"id":                 r.ID,
			"order_id":           r.OrderID,
			"order":              orderMap[r.OrderID],
			"staff_id":           r.StaffID,
			"staff":              staffMap[r.StaffID],
			"score":              r.Score,
			"attitude_score":     r.AttitudeScore,
			"professional_score": r.ProfessionalScore,
			"punctual_score":     r.PunctualScore,
			"content":            r.Content,
			"images":             r.Images,
			"status":             r.Status,
			"status_cn":          reviewStatusText(r.Status),
			"created_at":         r.CreatedAt,
		}
		list = append(list, item)
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

// HideServiceReview 隐藏违规评价（status=0），并重算质量分
func HideServiceReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "评价ID错误")
		return
	}

	var review models.ServiceReview
	if err := database.DB.Where("id = ?", id).First(&review).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeReviewNotFound, "评价不存在")
		return
	}
	if review.Status == utils.ReviewStatusHidden {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该评价已隐藏")
		return
	}

	if err := database.DB.Model(&review).Update("status", utils.ReviewStatusHidden).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "隐藏失败")
		return
	}

	// 异步重算质量分（剔除被隐藏评价）
	go func() { _ = reviewpkg.RecalculateQualityScore(review.StaffID) }()

	response.SuccessWithMessage(c, "已隐藏", gin.H{"id": review.ID})
}

// reviewStatusText 评价状态中文
func reviewStatusText(s uint8) string {
	switch s {
	case utils.ReviewStatusVisible:
		return "正常展示"
	case utils.ReviewStatusHidden:
		return "已隐藏"
	default:
		return "未知"
	}
}