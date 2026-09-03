package service_staff

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/notify"
	"fz_yyc_api/internal/services/orderquery"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// PendingOrders 待接订单列表（已支付 + biz_status=1 待接单 + 未被接单）
// 阶段三：接单池按服务人员 service_region 与订单 delivery_district 区域过滤（区域为空=不限）
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

	// 区域过滤：区域为空视为不限；否则仅返回 delivery_district 匹配的订单
	regionCond := ""
	if regions := parseRegions(staff.ServiceRegion); len(regions) > 0 {
		regionCond = " AND (delivery_district = '' OR delivery_district IS NULL"
		for _, r := range regions {
			regionCond += " OR delivery_district LIKE '%" + strings.ReplaceAll(r, "'", "''") + "%'"
		}
		regionCond += ")"
	}

	// 分类过滤：1=实物 2=服务（缺省不过滤）
	baseSQL := "status = 2 AND biz_status = 1 AND assigned_staff_id IS NULL" + regionCond
	if cond := orderCategoryCond(c); cond != "" {
		baseSQL += " AND " + cond
	}

	sortMode := c.Query("sort")
	orderBy := "paid_at DESC"
	if sortMode == "scheduled_at" {
		// 按预约时间排序（无预约时间的排在后面）
		orderBy = "scheduled_at IS NULL, scheduled_at ASC"
	}

	var orders []models.Order
	query := database.DB.
		Where(baseSQL).
		Order(orderBy).
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Preload("Items").Preload("User").Find(&orders).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

	orderquery.FillServiceCustomerInfo(ordersSlice(orders)...)

	var total int64
	database.DB.Model(&models.Order{}).
		Where(baseSQL).Count(&total)

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
		Where("id = ? AND status = ? AND biz_status = ? AND assigned_staff_id IS NULL",
			orderID, 2, 1).
		Updates(map[string]interface{}{
			"assigned_staff_id": staff.ID,
			"assigned_at":       now, // 接单时间（已指派超时未签到预警依据）
			"biz_status":        2,   // 已接单
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

	// 异步尽力而为地通知下单用户已接单（吞错，不阻塞接单成功响应）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = notify.OrderAccepted(ctx, &order, staff.Name, staff.Phone)
	}()

	response.SuccessWithMessage(c, "接单成功", gin.H{
		"id":                order.ID,
		"assigned_staff_id": staff.ID,
		"accepted_at":       now,
	})
}

// GiveUpOrder 放弃工单（已接单退回待接单池）：仅待出发工单可放弃
func GiveUpOrder(c *gin.Context) {
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

	// 原子放弃：仅本人已接单(biz_status=2)的工单可退回待接单池
	result := database.DB.Model(&models.Order{}).
		Where("id = ? AND assigned_staff_id = ? AND biz_status = ?",
			orderID, staff.ID, 2). // biz_status=2 已接单(待出发)
		Updates(map[string]interface{}{
			"biz_status":        1, // 退回待接单
			"assigned_staff_id": nil,
			"actual_started_at": nil,
		})

	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "放弃工单失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "仅待出发工单可放弃，或工单已变更")
		return
	}

	response.SuccessWithMessage(c, "已放弃工单，退回待接单池", gin.H{
		"id": orderID,
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
		Where("assigned_staff_id = ?", staff.ID).
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if bizStatus != "" {
		if bs, err := strconv.Atoi(bizStatus); err == nil {
			query = query.Where("biz_status = ?", bs)
		}
	}
	// 分类过滤：1=实物 2=服务（缺省不过滤）
	if cond := orderCategoryCond(c); cond != "" {
		query = query.Where(cond)
	}

	var orders []models.Order
	if err := query.Preload("Items").Preload("User").Find(&orders).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

	orderquery.FillServiceCustomerInfo(ordersSlice(orders)...)

	var total int64
	countQuery := database.DB.Model(&models.Order{}).
		Where("assigned_staff_id = ?", staff.ID)
	if bizStatus != "" {
		if bs, err := strconv.Atoi(bizStatus); err == nil {
			countQuery = countQuery.Where("biz_status = ?", bs)
		}
	}
	if cond := orderCategoryCond(c); cond != "" {
		countQuery = countQuery.Where(cond)
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
	if err := database.DB.Preload("User").Preload("Items").First(&order, orderID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "订单不存在")
		return
	}
	orderquery.FillServiceCustomerInfo(&order)

	// 只能查看自己接的或待接的订单
	if order.AssignedStaffID != nil && *order.AssignedStaffID != staff.ID {
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

	// 阶段三：签到时创建服务记录（start_time 快照）
	var record models.ServiceRecord
	if err := database.DB.Where("order_id = ?", orderID).First(&record).Error; err != nil {
		record = models.ServiceRecord{
			OrderID:   orderID,
			StaffID:   staff.ID,
			StartTime: &now,
			Status:    utils.ServiceRecordStatusNormal,
		}
		_ = database.DB.Create(&record).Error
	} else if record.StartTime == nil {
		_ = database.DB.Model(&record).Update("start_time", now).Error
	}

	response.SuccessWithMessage(c, "签到成功", gin.H{
		"id":                orderID,
		"actual_started_at": now,
	})
}

// CheckOutRequest 签退请求
type CheckOutRequest struct {
	Remark   string   `json:"remark"`
	Images   []string `json:"images"`
	AudioURL string   `json:"audio_url"` // 阶段三：服务录音URL（客户端七牛直传后随签退提交）
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

	// 阶段三：签退时写入服务记录（end_time + 录音快照）
	recordUpdates := map[string]interface{}{
		"end_time": now,
		"staff_id": staff.ID,
	}
	if req.AudioURL != "" {
		recordUpdates["audio_url"] = req.AudioURL
		recordUpdates["audio_uploaded_at"] = now
	}
	var record models.ServiceRecord
	if err := database.DB.Where("order_id = ?", orderID).First(&record).Error; err != nil {
		// 无记录时兜底创建（含签到缺失场景）
		newRecord := models.ServiceRecord{
			OrderID:   orderID,
			StaffID:   staff.ID,
			StartTime: order.ActualStartedAt,
			EndTime:   &now,
			AudioURL:  req.AudioURL,
			Status:    utils.ServiceRecordStatusNormal,
		}
		if req.AudioURL != "" {
			newRecord.AudioUploadedAt = &now
		}
		_ = database.DB.Create(&newRecord).Error
	} else {
		_ = database.DB.Model(&record).Updates(recordUpdates).Error
	}

	response.SuccessWithMessage(c, "签退成功", gin.H{
		"id":              orderID,
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
		Preload("User").
		Order("updated_at DESC").
		Find(&orders)
	orderquery.FillServiceCustomerInfo(ordersSlice(orders)...)

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
		"list": orders,
		"stats": gin.H{
			"today_assigned":  todayAssigned,
			"today_completed": todayCompleted,
		},
	})
}

// ordersSlice 将值切片转为指针切片，用于批量填充 customer 增强展示信息
func ordersSlice(orders []models.Order) []*models.Order {
	result := make([]*models.Order, len(orders))
	for index := range orders {
		result[index] = &orders[index]
	}
	return result
}

// orderCategoryCond 解析 category 查询参数（1=实物 2=服务，缺省/非法返回空串=不过滤）
// 返回可直接用于 WHERE 拼接的 order_type 范围条件字符串。
func orderCategoryCond(c *gin.Context) string {
	category := c.Query("category")
	if category == "" {
		return ""
	}
	switch category {
	case "1":
		return "order_type >= " + strconv.Itoa(int(utils.OrderTypeGoodsMin)) +
			" AND order_type <= " + strconv.Itoa(int(utils.OrderTypeGoodsMax))
	case "2":
		return "order_type >= " + strconv.Itoa(int(utils.OrderTypeServiceMin)) +
			" AND order_type <= " + strconv.Itoa(int(utils.OrderTypeServiceMax))
	}
	return ""
}
