package merchant

import (
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============================================
// 阶段三：预警中心（SOS求助 / 服务超时未结束）
// ============================================

// fillAlertStaffInfo 批量填充预警事件的服务人员信息
func fillAlertStaffInfo(events []models.ServiceAlertEvent) map[uint64]gin.H {
	staffMap := make(map[uint64]gin.H)
	if len(events) == 0 {
		return staffMap
	}
	staffIDs := make([]uint64, 0, len(events))
	for _, e := range events {
		// 订单级预警可能无服务人员（staff_id 可空）
		if e.StaffID == nil {
			continue
		}
		staffIDs = append(staffIDs, *e.StaffID)
	}
	var staffs []models.ServiceStaff
	database.DB.Select("id, name, phone").Where("id IN ?", staffIDs).Find(&staffs)
	for _, s := range staffs {
		staffMap[s.ID] = gin.H{"id": s.ID, "name": s.Name, "phone": s.Phone}
	}
	return staffMap
}

// fillAlertOrderInfo 批量填充预警事件的订单信息
func fillAlertOrderInfo(events []models.ServiceAlertEvent) map[uint64]gin.H {
	orderMap := make(map[uint64]gin.H)
	for _, e := range events {
		if e.OrderID == nil || *e.OrderID == 0 {
			continue
		}
		var order models.Order
		if err := database.DB.Select("id, order_no, delivery_address, biz_status, status").
			Where("id = ?", *e.OrderID).First(&order).Error; err == nil {
			orderMap[*e.OrderID] = gin.H{
				"id":         order.ID,
				"order_no":   order.OrderNo,
				"address":    order.DeliveryAddress,
				"biz_status": order.BizStatus,
				"status":     order.Status,
			}
		}
	}
	return orderMap
}

// GetAlertEvents 预警事件列表
func GetAlertEvents(c *gin.Context) {
	alertType := c.Query("alert_type")
	status := c.Query("status")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&models.ServiceAlertEvent{})
	if alertType != "" {
		if t, err := strconv.Atoi(alertType); err == nil {
			query = query.Where("alert_type = ?", t)
		}
	}
	if status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}
	if keyword != "" {
		// 按服务人员姓名/手机号模糊匹配
		query = query.Where("staff_id IN (?)",
			database.DB.Model(&models.ServiceStaff{}).Select("id").
				Where("name LIKE ? OR phone LIKE ?", "%"+keyword+"%", "%"+keyword+"%"))
	}

	var total int64
	query.Count(&total)

	var events []models.ServiceAlertEvent
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&events)

	staffMap := fillAlertStaffInfo(events)
	orderMap := fillAlertOrderInfo(events)

	list := make([]gin.H, 0, len(events))
	for _, e := range events {
		var staffID uint64
		if e.StaffID != nil {
			staffID = *e.StaffID
		}
		item := gin.H{
			"id":            e.ID,
			"order_id":      e.OrderID,
			"staff_id":      e.StaffID,
			"staff":         staffMap[staffID],
			"alert_type":    e.AlertType,
			"alert_type_cn": alertTypeText(e.AlertType),
			"lat":           e.Lat,
			"lng":           e.Lng,
			"address":       e.Address,
			"summary":       e.Summary,
			"status":        e.Status,
			"status_cn":     alertStatusText(e.Status),
			"handler_name":  e.HandlerName,
			"handled_at":    e.HandledAt,
			"created_at":    e.CreatedAt,
		}
		if e.OrderID != nil {
			item["order"] = orderMap[*e.OrderID]
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

// GetAlertEventDetail 预警事件详情
func GetAlertEventDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "预警事件ID错误")
		return
	}

	var event models.ServiceAlertEvent
	if err := database.DB.Where("id = ?", id).First(&event).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAlertEventNotFound, "预警事件不存在")
		return
	}

	staffMap := fillAlertStaffInfo([]models.ServiceAlertEvent{event})
	orderMap := fillAlertOrderInfo([]models.ServiceAlertEvent{event})

	var staffID uint64
	if event.StaffID != nil {
		staffID = *event.StaffID
	}
	item := gin.H{
		"id":            event.ID,
		"order_id":      event.OrderID,
		"staff_id":      event.StaffID,
		"staff":         staffMap[staffID],
		"alert_type":    event.AlertType,
		"alert_type_cn": alertTypeText(event.AlertType),
		"lat":           event.Lat,
		"lng":           event.Lng,
		"address":       event.Address,
		"summary":       event.Summary,
		"status":        event.Status,
		"status_cn":     alertStatusText(event.Status),
		"handler_id":    event.HandlerID,
		"handler_name":  event.HandlerName,
		"handle_remark": event.HandleRemark,
		"handled_at":    event.HandledAt,
		"created_at":    event.CreatedAt,
	}
	if event.OrderID != nil {
		item["order"] = orderMap[*event.OrderID]
	}

	response.Success(c, item)
}

// HandleAlertEventRequest 处理预警请求
type HandleAlertEventRequest struct {
	Status uint8  `json:"status" binding:"required,oneof=2 3"` // 2=处理中 3=已处理
	Remark string `json:"remark" binding:"max=512"`
}

// HandleAlertEvent 处理预警（状态流转 + 留痕）
func HandleAlertEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "预警事件ID错误")
		return
	}

	var req HandleAlertEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	var event models.ServiceAlertEvent
	if err := database.DB.Where("id = ?", id).First(&event).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeAlertEventNotFound, "预警事件不存在")
		return
	}
	// 已处理的事件不允许再流转
	if event.Status == utils.AlertStatusHandled {
		response.Fail(c, http.StatusBadRequest, response.CodeAlertEventHandled, "该预警已处理，不可再操作")
		return
	}

	handlerID := middleware.GetStaffID(c)
	handlerName := middleware.GetUsername(c)

	updates := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_remark": req.Remark,
		"updated_at":    time.Now(),
	}
	if req.Status == utils.AlertStatusHandled {
		updates["handled_at"] = time.Now()
	} else {
		updates["handled_at"] = nil
	}

	if err := database.DB.Model(&event).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "处理失败")
		return
	}

	response.SuccessWithMessage(c, "处理成功", gin.H{
		"id":           event.ID,
		"status":       req.Status,
		"handler_id":   handlerID,
		"handler_name": handlerName,
	})
}

// alertTypeText 预警类型中文
func alertTypeText(t uint8) string {
	switch t {
	case utils.AlertTypeSOS:
		return "SOS求助"
	case utils.AlertTypeTimeout:
		return "服务超时未结束"
	default:
		return "未知"
	}
}

// alertStatusText 预警状态中文
func alertStatusText(s uint8) string {
	switch s {
	case utils.AlertStatusPending:
		return "待处理"
	case utils.AlertStatusProcessing:
		return "处理中"
	case utils.AlertStatusHandled:
		return "已处理"
	default:
		return "未知"
	}
}
