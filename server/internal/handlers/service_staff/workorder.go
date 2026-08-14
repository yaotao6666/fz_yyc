package service_staff

import (
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// PendingOrders 待接订单列表（已支付 + biz_status=1 待接单 + 未被接单）
func PendingOrders(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var orders []models.Order
	query := database.DB.
		Where("merchant_id = ? AND status = ? AND biz_status = ? AND assigned_staff_id IS NULL",
			staff.MerchantID, 2, 1). // status=2已支付, biz_status=1待接单
		Order("paid_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Preload("Items").Find(&orders).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

	var total int64
	database.DB.Model(&models.Order{}).
		Where("merchant_id = ? AND status = ? AND biz_status = ? AND assigned_staff_id IS NULL",
			staff.MerchantID, 2, 1).Count(&total)

	response.Success(c, gin.H{
		"list":  orders,
		"total": total,
	})
}

// AcceptOrder 接单（原子操作）
func AcceptOrder(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	// 原子接单：只有 biz_status=1 且 assigned_staff_id IS NULL 时才能接单
	now := time.Now()
	result := database.DB.Model(&models.Order{}).
		Where("id = ? AND merchant_id = ? AND status = ? AND biz_status = ? AND assigned_staff_id IS NULL",
			orderID, staff.MerchantID, 2, 1).
		Updates(map[string]interface{}{
			"assigned_staff_id": staff.ID,
			"biz_status":        2, // 已接单
			"actual_started_at": nil,
		})

	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "接单失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusConflict, response.CodeParamError, "订单已被接单或状态已变更")
		return
	}

	var order models.Order
	database.DB.First(&order, orderID)

	response.SuccessWithMessage(c, "接单成功", gin.H{
		"id":                order.ID,
		"assigned_staff_id": staff.ID,
		"accepted_at":       now,
	})
}

// AcceptedOrders 已接订单列表
func AcceptedOrders(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	bizStatus := c.Query("biz_status") // 可选筛选：2=已接单 3=服务中 5=已完成

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.
		Where("assigned_staff_id = ? AND merchant_id = ?", staff.ID, staff.MerchantID).
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if bizStatus != "" {
		if bs, err := strconv.Atoi(bizStatus); err == nil {
			query = query.Where("biz_status = ?", bs)
		}
	}

	var orders []models.Order
	if err := query.Preload("Items").Find(&orders).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

	var total int64
	countQuery := database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND merchant_id = ?", staff.ID, staff.MerchantID)
	if bizStatus != "" {
		if bs, err := strconv.Atoi(bizStatus); err == nil {
			countQuery = countQuery.Where("biz_status = ?", bs)
		}
	}
	countQuery.Count(&total)

	response.Success(c, gin.H{
		"list":  orders,
		"total": total,
	})
}

// OrderDetail 订单详情
func OrderDetail(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var order models.Order
	if err := database.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "订单不存在")
		return
	}

	// 只能查看自己接的或待接的订单
	if order.AssignedStaffID != nil && *order.AssignedStaffID != staff.ID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看此订单")
		return
	}
	if order.MerchantID != staff.MerchantID {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "无权查看此订单")
		return
	}

	response.Success(c, order)
}

// CheckInRequest 签到请求
type CheckInRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CheckIn 签到（开始服务）
func CheckIn(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req CheckInRequest
	_ = c.ShouldBindJSON(&req)

	now := time.Now()
	result := database.DB.Model(&models.Order{}).
		Where("id = ? AND assigned_staff_id = ? AND biz_status = ?",
			orderID, staff.ID, 2). // biz_status=2 已接单才能签到
		Updates(map[string]interface{}{
			"biz_status":        3, // 服务中
			"actual_started_at": now,
		})

	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "签到失败，订单状态不允许")
		return
	}

	response.SuccessWithMessage(c, "签到成功", gin.H{
		"id":               orderID,
		"actual_started_at": now,
	})
}

// CheckOutRequest 签退请求
type CheckOutRequest struct {
	Remark string   `json:"remark"`
	Images []string `json:"images"`
}

// CheckOut 签退（结束服务）
func CheckOut(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req CheckOutRequest
	_ = c.ShouldBindJSON(&req)

	now := time.Now()
	updates := map[string]interface{}{
		"biz_status":      5, // 已完成（服务完成）
		"actual_ended_at": now,
	}
	if req.Remark != "" {
		updates["rental_return_remark"] = req.Remark
	}

	result := database.DB.Model(&models.Order{}).
		Where("id = ? AND assigned_staff_id = ? AND biz_status = ?",
			orderID, staff.ID, 3). // biz_status=3 服务中才能签退
		Updates(updates)

	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "签退失败，订单状态不允许")
		return
	}

	// 如果是租赁订单，更新归还时间
	var order models.Order
	database.DB.First(&order, orderID)
	if order.OrderType == 2 {
		database.DB.Model(&order).Update("rental_returned_at", now)
	}

	response.SuccessWithMessage(c, "签退成功", gin.H{
		"id":             orderID,
		"actual_ended_at": now,
	})
}

// GetTodoList 待办列表（待出发 + 服务中）
func GetTodoList(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}

	var orders []models.Order
	database.DB.Where("assigned_staff_id = ? AND biz_status IN ?", staff.ID, []uint8{2, 3}).
		Preload("Items").
		Order("updated_at DESC").
		Find(&orders)

	// 统计
	var todayAssigned, todayCompleted int64
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND actual_started_at >= ? AND actual_started_at < ?",
			staff.ID, today, tomorrow).Count(&todayAssigned)
	database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ? AND biz_status = ? AND actual_ended_at >= ? AND actual_ended_at < ?",
			staff.ID, 5, today, tomorrow).Count(&todayCompleted)

	response.Success(c, gin.H{
		"list":  orders,
		"stats": gin.H{
			"today_assigned":  todayAssigned,
			"today_completed": todayCompleted,
		},
	})
}
