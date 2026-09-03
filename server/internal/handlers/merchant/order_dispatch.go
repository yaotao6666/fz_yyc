package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/orderquery"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetDispatchableStaffList 返回可被派单的服务人员（已审核通过：status=1 且 audit_status=0）
func GetDispatchableStaffList(c *gin.Context) {
	var staffs []models.ServiceStaff
	if err := database.DB.
		Where("status = ? AND audit_status = ?", 1, 0).
		Order("id ASC").
		Find(&staffs).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取服务人员失败")
		return
	}

	list := make([]gin.H, 0, len(staffs))
	for _, s := range staffs {
		list = append(list, gin.H{
			"id":    s.ID,
			"name":  s.Name,
			"phone": s.Phone,
		})
	}
	response.Success(c, list)
}

// DispatchOrderRequest 派单请求
type DispatchOrderRequest struct {
	StaffID uint64 `json:"staff_id" binding:"required"`
}

// DispatchOrder 将订单指派给已审核服务人员（指派后 biz_status=2 待出发，服务人员工单可见）
func DispatchOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req DispatchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var staff models.ServiceStaff
	if err := database.DB.Where("id = ? AND status = ? AND audit_status = ?", req.StaffID, 1, 0).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该服务人员不可用或未通过审核")
		return
	}

	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}
	if order.Status != 2 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单须为已支付状态才能派单")
		return
	}
	// 已完成订单不允许再修改派单人员（服务单完成时仅 biz_status=5，orders.status 仍为 2）
	if order.Status == 3 || order.BizStatus == 5 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "已完成订单不允许修改派单人员")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"assigned_staff_id": req.StaffID,
		"assigned_at":       now, // 指派时间（已指派超时未签到预警依据）
		"biz_status":        2,   // 已接单/待出发
	}
	result := database.DB.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates)
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "派单失败")
		return
	}

	// 派单留痕（写审核记录）
	after, _ := json.Marshal(map[string]interface{}{
		"op":         "dispatch",
		"staff_id":   req.StaffID,
		"staff_name": staff.Name,
		"time":       now,
	})
	database.DB.Create(&models.StaffAuditRecord{
		StaffID:   req.StaffID,
		AuditType: 4, // 状态变更类
		ApplyType: 2, // 管理员操作
		AfterData: models.JSON(after),
		Status:    1,
	})

	// 重新加载订单
	database.DB.Preload("Items").First(&order, orderID)
	response.SuccessWithMessage(c, "派单成功", gin.H{
		"id":                order.ID,
		"assigned_staff_id": req.StaffID,
		"biz_status":        2,
	})
}

// RenewOrderRequest 续租请求
type RenewOrderRequest struct {
	Duration uint `json:"duration"` // 可选：续租时长（默认取原单最长租赁时长）
}

// RenewOrder 基于原租赁订单生成关联新订单（parent_order_id=原ID, renew_flag=1, 待支付）
func RenewOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req RenewOrderRequest
	_ = c.ShouldBindJSON(&req)

	var src models.Order
	if err := database.DB.Preload("Items").First(&src, orderID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}
	if src.OrderType != 2 || len(src.Items) == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "仅租赁订单可续租")
		return
	}
	if src.Status == 1 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "原订单尚未支付，无法续租")
		return
	}

	// 若有续租中的待支付单，提示
	var pendingRenew int64
	database.DB.Model(&models.Order{}).
		Where("parent_order_id = ? AND status = 1", orderID).Count(&pendingRenew)
	if pendingRenew > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该订单已存在待支付的续租单")
		return
	}

	duration := req.Duration
	if duration == 0 {
		// 默认取原单最长租赁时长
		for _, it := range src.Items {
			if it.SaleType == 2 && it.RentalDuration > duration {
				duration = it.RentalDuration
			}
		}
	}
	if duration == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "续租时长无效")
		return
	}

	// 复制订单费用（按续租时长重算租金，押金不变）
	var newTotalAmount float64
	var newItems []models.OrderItem
	for _, it := range src.Items {
		renewalUnits := duration
		switch it.RentalUnit {
		case 2: // 周
			renewalUnits = duration * 7
		case 3: // 月
			renewalUnits = duration * 30
		}
		subtotal := it.UnitRentalPrice * float64(renewalUnits) * float64(it.Quantity)
		newTotalAmount += subtotal
		newItems = append(newItems, models.OrderItem{
			ProductID:       it.ProductID,
			ProductName:     it.ProductName,
			Image:           it.Image,
			Price:           it.UnitRentalPrice,
			Quantity:        it.Quantity,
			SpecInfo:        it.SpecInfo,
			Subtotal:        subtotal,
			SaleType:        2,
			RentalUnit:      it.RentalUnit,
			RentalDuration:  duration,
			UnitRentalPrice: it.UnitRentalPrice,
			RentalSubtotal:  subtotal,
			Deposit:         it.Deposit,
		})
	}

	payAmount := newTotalAmount + src.DeliveryFee - src.DiscountAmount + src.TotalDeposit
	if payAmount < 0 {
		payAmount = 0
	}

	orderNo := utils.GenerateOrderNo(utils.DefaultMerchantID)

	tx := database.DB.Begin()
	newOrder := models.Order{
		OrderNo:         orderNo,
		UserID:          src.UserID,
		OrderType:       2, // 租赁
		BizStatus:       0,
		AssignedStaffID: src.AssignedStaffID,
		TotalAmount:     newTotalAmount,
		DeliveryFee:     src.DeliveryFee,
		DiscountAmount:  src.DiscountAmount,
		PayAmount:       payAmount,
		TotalDeposit:    src.TotalDeposit,
		DeliveryAddress: src.DeliveryAddress,
		ContactName:     src.ContactName,
		ContactPhone:    src.ContactPhone,
		Remark:          src.Remark,
		Status:          1,
		ParentOrderID:   &src.ID,
		RenewFlag:       1,
		ScheduledAt:     src.ScheduledAt,
	}
	if newOrder.TotalDeposit > 0 {
		newOrder.DepositStatus = 1
	}
	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建续租订单失败")
		return
	}
	for i := range newItems {
		newItems[i].OrderID = newOrder.ID
		if err := tx.Create(&newItems[i]).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建续租商品失败")
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建续租订单失败")
		return
	}

	database.DB.Preload("Items").First(&newOrder, newOrder.ID)
	accessibleOrder := buildAccessibleMerchantOrder(newOrder)
	response.SuccessWithMessage(c, "续租订单已生成", gin.H{
		"order": accessibleOrder,
	})
}

// ListRentalDueOrders 租赁到期提醒列表
// 未归还 = order_type=2 且 status IN(2,3) 且 deposit_status IN(0,1)（尚未归还退押金）
func ListRentalDueOrders(c *gin.Context) {
	dueRange := c.Query("due_range") // 可选: soon(将到期7天内) overdue(已逾期) all(默认)
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	now := time.Now()
	query := database.DB.Model(&models.Order{}).
		Where("order_type = ? AND status IN ? AND deposit_status IN ? AND rental_end_at IS NOT NULL", 2, []uint8{2, 3}, []uint8{0, 1})

	if dueRange == "soon" {
		// 将到期：rental_end_at 在 [now, now+7天]
		query = query.Where("rental_end_at >= ? AND rental_end_at <= ?", now, now.AddDate(0, 0, 7))
	} else if dueRange == "overdue" {
		query = query.Where("rental_end_at < ?", now)
	}

	if keyword != "" {
		query = query.Where("order_no LIKE ? OR contact_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	query.Preload("Items").Preload("User").
		Order("rental_end_at ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&orders)

	nowDate := time.Now()
	list := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		isOverdue := false
		daysLeft := 0
		if o.RentalEndAt != nil {
			remain := o.RentalEndAt.Sub(nowDate)
			if remain < 0 {
				isOverdue = true
			} else {
				daysLeft = int(remain.Hours() / 24)
			}
		}
		acc := orderquery.BuildAccessibleOrder(o)
		list = append(list, gin.H{
			"order":         acc,
			"rental_end_at": o.RentalEndAt,
			"is_overdue":    isOverdue,
			"days_left":     daysLeft,
		})
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