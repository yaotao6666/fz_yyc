package service_staff

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
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

	baseSQL := "status = 2 AND biz_status = 1 AND assigned_staff_id IS NULL" + regionCond

	var orders []models.Order
	query := database.DB.
		Where(baseSQL).
		Order("paid_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize)

	if err := query.Preload("Items").Find(&orders).Error; err != nil {
		response.Success(c, gin.H{"list": []interface{}{}, "total": 0})
		return
	}

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
		Where("assigned_staff_id = ?", staff.ID).
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
		Where("assigned_staff_id = ?", staff.ID)
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
		"id":               orderID,
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
			OrderID:         orderID,
			StaffID:         staff.ID,
			StartTime:       order.ActualStartedAt,
			EndTime:         &now,
			AudioURL:        req.AudioURL,
			Status:          utils.ServiceRecordStatusNormal,
		}
		if req.AudioURL != "" {
			newRecord.AudioUploadedAt = &now
		}
		_ = database.DB.Create(&newRecord).Error
	} else {
		_ = database.DB.Model(&record).Updates(recordUpdates).Error
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
