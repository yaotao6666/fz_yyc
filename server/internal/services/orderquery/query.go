package orderquery

import (
	"context"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"

	"gorm.io/gorm"
)

type ListOptions struct {
	Status          *int
	OrderType       *int
	BizStatus       *int
	AssignedStaffID *uint64
	StartDate       string
	EndDate         string
	Keyword         string
	Page            int
	PageSize        int
	IncludeMerchant bool
}

type DetailOptions struct {
	IncludeMerchant bool
}

type ListResult struct {
	List     []models.Order
	Total    int64
	Page     int
	PageSize int
}

func LoadMerchantOrderByID(orderID uint64) (*models.Order, error) {
	return GetOrderDetail(orderID, DetailOptions{})
}

func GetOrderList(ctx context.Context, options ListOptions) (*ListResult, error) {
	page, pageSize := normalizePagination(options.Page, options.PageSize)
	baseQuery := applyOrderScopes(database.DB.Model(&models.Order{}), options)

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	query := preloadOrderAssociations(applyOrderScopes(database.DB, options), options.IncludeMerchant)

	var orders []models.Order
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("orders.created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	refreshOrderListRefundStatus(ctx, options, orders)

	accessibleOrders := make([]models.Order, 0, len(orders))
	for _, order := range orders {
		accessibleOrders = append(accessibleOrders, BuildAccessibleOrder(order))
	}

	return &ListResult{
		List:     accessibleOrders,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func GetOrderDetail(orderID uint64, options DetailOptions) (*models.Order, error) {
	var order models.Order
	query := preloadOrderAssociations(database.DB, options.IncludeMerchant).
		Where("orders.id = ?", orderID)
	query = applyDetailScopes(query, options)

	if err := query.First(&order).Error; err != nil {
		return nil, err
	}

	client, err := wechatpay.NewServiceProviderClient()
	if err == nil {
		refreshSingleOrderRefundStatus(context.Background(), client, &order, options)
	}

	accessibleOrder := BuildAccessibleOrder(order)
	return &accessibleOrder, nil
}

func BuildAccessibleOrder(order models.Order) models.Order {
	for index := range order.Items {
		order.Items[index].Image = BuildAccessibleOrderItemImage(order.Items[index].Image)
	}
	return order
}

func BuildAccessibleOrderItemImage(image string) string {
	return buildAccessibleOrderItemImage(image)
}

func SyncRefundAndOrderStatus(
	tx *gorm.DB,
	order *models.Order,
	refund *models.Refund,
	refundStatus string,
	refundID string,
	successTime string,
) error {
	if tx == nil || order == nil || refund == nil {
		return nil
	}

	normalizedStatus := strings.ToUpper(strings.TrimSpace(refundStatus))
	trimmedRefundID := strings.TrimSpace(refundID)
	refundUpdates := map[string]any{}
	if trimmedRefundID != "" {
		refundUpdates["refund_id"] = trimmedRefundID
	}

	orderUpdates := map[string]any{}
	switch normalizedStatus {
	case "SUCCESS":
		refundedAt := parseRefundSuccessTime(successTime)
		refundUpdates["status"] = 2
		refundUpdates["refunded_at"] = refundedAt
		orderUpdates["status"] = 6
		orderUpdates["refunded_at"] = refundedAt
	case "CLOSED", "ABNORMAL":
		refundUpdates["status"] = 3
		orderUpdates["status"] = buildOrderStatusAfterRefundFailure(order)
		orderUpdates["refunded_at"] = nil
	default:
		refundUpdates["status"] = 1
		orderUpdates["status"] = 5
	}

	if len(refundUpdates) > 0 {
		if err := tx.Model(refund).Updates(refundUpdates).Error; err != nil {
			return err
		}
	}
	if len(orderUpdates) > 0 {
		if err := tx.Model(order).Updates(orderUpdates).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}

func applyOrderScopes(query *gorm.DB, options ListOptions) *gorm.DB {
	scopedQuery := query
	if options.Status != nil {
		scopedQuery = scopedQuery.Where("orders.status = ?", *options.Status)
	}
	if options.OrderType != nil {
		scopedQuery = scopedQuery.Where("orders.order_type = ?", *options.OrderType)
	}
	if options.BizStatus != nil {
		scopedQuery = scopedQuery.Where("orders.biz_status = ?", *options.BizStatus)
	}
	if options.AssignedStaffID != nil {
		scopedQuery = scopedQuery.Where("orders.assigned_staff_id = ?", *options.AssignedStaffID)
	}
	if trimmedKeyword := strings.TrimSpace(options.Keyword); trimmedKeyword != "" {
		scopedQuery = scopedQuery.Where("orders.order_no LIKE ?", "%"+trimmedKeyword+"%")
	}
	if options.StartDate != "" {
		scopedQuery = scopedQuery.Where("orders.created_at >= ?", options.StartDate)
	}
	if options.EndDate != "" {
		if endDateTime, err := time.Parse("2006-01-02", options.EndDate); err == nil {
			scopedQuery = scopedQuery.Where("orders.created_at <= ?", endDateTime.Add(24*time.Hour))
		}
	}
	return scopedQuery
}

func applyDetailScopes(query *gorm.DB, options DetailOptions) *gorm.DB {
	return query
}

func preloadOrderAssociations(query *gorm.DB, includeMerchant bool) *gorm.DB {
	return query.Preload("User").Preload("Items")
}

func refreshOrderListRefundStatus(ctx context.Context, options ListOptions, orders []models.Order) {
	client, err := wechatpay.NewServiceProviderClient()
	if err != nil {
		return
	}
	for index := range orders {
		refreshSingleOrderRefundStatus(ctx, client, &orders[index], DetailOptions{
			IncludeMerchant: options.IncludeMerchant,
		})
	}
}

func refreshSingleOrderRefundStatus(ctx context.Context, client *wechatpay.ServiceProviderClient, order *models.Order, options DetailOptions) {
	if client == nil || order == nil || order.Status != 5 {
		return
	}

	var refund models.Refund
	if err := database.DB.
		Where("order_id = ? AND status IN (0, 1)", order.ID).
		Order("created_at DESC").
		First(&refund).Error; err != nil {
		return
	}
	if strings.TrimSpace(refund.RefundNo) == "" {
		return
	}

	refundStatus, queryErr := client.QueryPartnerRefundByRefundNo(ctx, refund.RefundNo)
	if queryErr != nil {
		return
	}
	if err := SyncRefundAndOrderStatus(database.DB, order, &refund, refundStatus.Status, refundStatus.RefundID, refundStatus.SuccessTime); err != nil {
		return
	}

	refreshedOrder, err := reloadOrderDetail(order.ID, options)
	if err != nil {
		return
	}
	*order = *refreshedOrder
}

func reloadOrderDetail(orderID uint64, options DetailOptions) (*models.Order, error) {
	var order models.Order
	query := preloadOrderAssociations(database.DB, options.IncludeMerchant).
		Where("orders.id = ?", orderID)
	query = applyDetailScopes(query, options)
	if err := query.First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func parseRefundSuccessTime(successTime string) time.Time {
	now := time.Now()
	if strings.TrimSpace(successTime) == "" {
		return now
	}
	if parsed, err := time.Parse(time.RFC3339, successTime); err == nil {
		return parsed
	}
	return now
}

func buildOrderStatusAfterRefundFailure(order *models.Order) uint8 {
	if order == nil {
		return 2
	}
	if order.CompletedAt != nil {
		return 3
	}
	if order.PaidAt != nil || strings.TrimSpace(order.TransactionID) != "" {
		return 2
	}
	return order.Status
}

func buildAccessibleOrderItemImage(image string) string {
	service := qiniu.GetService()
	if service == nil {
		return image
	}
	return service.BuildPrivateURL(image)
}
