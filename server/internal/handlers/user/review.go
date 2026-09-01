package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	reviewpkg "fz_yyc_api/internal/services/review"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SubmitReviewRequest 提交评价请求
type SubmitReviewRequest struct {
	Score             uint8    `json:"score" binding:"required,min=1,max=5"`
	AttitudeScore     uint8    `json:"attitude_score" binding:"required,min=1,max=5"`
	ProfessionalScore uint8    `json:"professional_score" binding:"required,min=1,max=5"`
	PunctualScore     uint8    `json:"punctual_score" binding:"required,min=1,max=5"`
	Content           string   `json:"content" binding:"max=512"`
	Images            []string `json:"images" binding:"max=9"`
}

// SubmitReview 提交评价（校验：本人+服务订单+已完成+未评过），异步重算质量分
func SubmitReview(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	// 仅服务订单可评价
	if utils.OrderCategory(order.OrderType) != utils.OrderCategoryService {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "仅服务订单可评价")
		return
	}
	// 服务已完成（biz_status=5）
	if order.BizStatus != 5 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "服务未完成，暂不可评价")
		return
	}
	// 一单一评
	var count int64
	database.DB.Model(&models.ServiceReview{}).Where("order_id = ?", orderID).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeReviewExists, "该订单已评价")
		return
	}
	// 必须有服务人员
	if order.AssignedStaffID == nil || *order.AssignedStaffID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "订单无服务人员，暂不可评价")
		return
	}

	reviewItem := models.ServiceReview{
		OrderID:           orderID,
		UserID:            userID,
		StaffID:           *order.AssignedStaffID,
		Score:             req.Score,
		AttitudeScore:     req.AttitudeScore,
		ProfessionalScore: req.ProfessionalScore,
		PunctualScore:     req.PunctualScore,
		Content:           req.Content,
		Status:            utils.ReviewStatusVisible,
	}
	if len(req.Images) > 0 {
		if raw, err := json.Marshal(req.Images); err == nil {
			reviewItem.Images = models.JSON(raw)
		}
	}

	if err := database.DB.Create(&reviewItem).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "评价提交失败")
		return
	}

	// 异步重算质量分（不阻塞响应）
	staffID := *order.AssignedStaffID
	go func() { _ = reviewpkg.RecalculateQualityScore(staffID) }()

	response.SuccessWithMessage(c, "评价成功", gin.H{"id": reviewItem.ID})
}

// GetReview 查看订单评价（未评价返回 null）
func GetReview(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 64)
	if err != nil || orderID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var review models.ServiceReview
	if err := database.DB.Where("order_id = ? AND user_id = ?", orderID, userID).First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Success(c, nil)
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询失败")
		return
	}

	response.Success(c, review)
}