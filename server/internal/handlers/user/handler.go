package user

import (
	"encoding/json"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	var user models.User
	openID := "mock_openid_" + req.Code

	result := database.DB.Where("openid = ?", openID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			OpenID:   openID,
			Nickname: "微信用户",
			Status:   1,
		}
		database.DB.Create(&user)
	}

	token := utils.GenerateToken(user.ID, "user", user.Nickname)
	response.Success(c, gin.H{
		"token": token,
		"user": user,
	})
}

func GetStoreHome(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商家不存在")
		return
	}

	if merchant.Status != 1 {
		response.Fail(c, http.StatusForbidden, response.Forbidden, "商家已停业")
		return
	}

	var categories []models.Category
	database.DB.Where("merchant_id = ? AND status = 1", id).Order("sort ASC").Find(&categories)

	var hotProducts []models.Product
	database.DB.Where("merchant_id = ? AND status = 1", id).Order("sales DESC").Limit(10).Find(&hotProducts)

	var deliverySettings models.MerchantDeliverySettings
	database.DB.Where("merchant_id = ?", id).First(&deliverySettings)

	response.Success(c, gin.H{
		"merchant":         merchant,
		"categories":       categories,
		"hot_products":     hotProducts,
		"delivery_settings": deliverySettings,
	})
}

func GetProducts(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	query := database.DB.Model(&models.Product{}).Where("merchant_id = ? AND status = 1", id)

	if categoryID != "" {
		catID, _ := strconv.ParseUint(categoryID, 10, 64)
		query = query.Where("category_id = ?", catID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&products).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "获取商品列表失败")
		return
	}

	var merchant models.Merchant
	database.DB.Select("id", "min_order_amount", "takeout_enabled", "dine_in_enabled").First(&merchant, id)

	response.Success(c, gin.H{
		"list": products,
		"merchant": gin.H{
			"min_order_amount": merchant.MinOrderAmount,
			"takeout_enabled": merchant.TakeoutEnabled,
			"dine_in_enabled": merchant.DineInEnabled,
		},
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

func GetProductDetail(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	productID := c.Param("product_id")
	mid, _ := strconv.ParseUint(merchantID, 10, 64)
	pid, _ := strconv.ParseUint(productID, 10, 64)

	var product models.Product
	if err := database.DB.Preload("Category").Preload("Specs").Where("id = ? AND merchant_id = ? AND status = 1", pid, mid).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商品不存在或已下架")
		return
	}

	response.Success(c, product)
}

func GetDeliveryRules(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var settings models.MerchantDeliverySettings
	if err := database.DB.Where("merchant_id = ?", id).First(&settings).Error; err != nil {
		response.Success(c, gin.H{
			"enabled": false,
		})
		return
	}

	var distanceRules []struct {
		Distance float64 `json:"distance"`
		Fee      float64 `json:"fee"`
	}
	if settings.DistanceRules != nil {
		json.Unmarshal(settings.DistanceRules, &distanceRules)
	}

	response.Success(c, gin.H{
		"enabled":             settings.Enabled,
		"base_fee":           settings.BaseFee,
		"free_delivery_amount": settings.FreeDeliveryAmount,
		"max_distance":       settings.MaxDistance,
		"distance_rules":     distanceRules,
	})
}

type CreateOrderRequest struct {
	MerchantID       uint64 `json:"merchant_id" binding:"required"`
	DeliveryType     uint8  `json:"delivery_type" binding:"required,oneof=1 2"`
	DeliveryDistance float64 `json:"delivery_distance"`
	DeliveryAddress  string `json:"delivery_address"`
	ContactName      string `json:"contact_name" binding:"required"`
	ContactPhone     string `json:"contact_phone" binding:"required"`
	Remark           string `json:"remark"`
	Items            []struct {
		ProductID uint64 `json:"product_id" binding:"required"`
		Quantity  uint   `json:"quantity" binding:"required,min=1"`
		SpecInfo  string `json:"spec_info"`
		Price     float64 `json:"price"`
	} `json:"items" binding:"required,min=1"`
}

func CreateOrder(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "用户未登录")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	var merchant models.Merchant
	if err := database.DB.First(&merchant, req.MerchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "商家不存在")
		return
	}

	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.InvalidParams, "商品不存在")
			return
		}

		if product.MerchantID != req.MerchantID {
			response.Fail(c, http.StatusBadRequest, response.InvalidParams, "商品不属于该商家")
			return
		}

		if product.Status != 1 {
			response.Fail(c, http.StatusBadRequest, response.InvalidParams, "商品已下架")
			return
		}

		var specInfo models.JSON
		if item.SpecInfo != "" {
			specInfo = models.JSON(item.SpecInfo)
		}

		price := product.Price
		if item.Price > 0 {
			price = item.Price
		}

		subtotal := price * float64(item.Quantity)
		totalAmount += subtotal

		var images []string
		if product.Images != nil {
			json.Unmarshal(product.Images, &images)
		}
		image := ""
		if len(images) > 0 {
			image = images[0]
		}

		orderItems = append(orderItems, models.OrderItem{
			MerchantID:  req.MerchantID,
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Image:       image,
			Price:       price,
			Quantity:    item.Quantity,
			SpecInfo:    specInfo,
			Subtotal:    subtotal,
		})
	}

	if totalAmount < merchant.MinOrderAmount {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单金额未达到最低消费")
		return
	}

	var deliveryFee float64
	if req.DeliveryType == 1 {
		var settings models.MerchantDeliverySettings
		if err := database.DB.Where("merchant_id = ?", req.MerchantID).First(&settings).Error; err == nil && settings.Enabled {
			deliveryFee = settings.BaseFee
			if req.DeliveryDistance > settings.MaxDistance {
				response.Fail(c, http.StatusBadRequest, response.InvalidParams, "超出配送范围")
				return
			}
		}
	}

	payAmount := totalAmount + deliveryFee

	orderNo := utils.GenerateOrderNo()
	verifyCode := utils.GenerateVerifyCode()

	tx := database.DB.Begin()

	order := models.Order{
		OrderNo:          orderNo,
		UserID:           userID,
		MerchantID:       req.MerchantID,
		TotalAmount:      totalAmount,
		DeliveryFee:      deliveryFee,
		PayAmount:        payAmount,
		DeliveryType:     req.DeliveryType,
		DeliveryDistance: req.DeliveryDistance,
		DeliveryAddress:  req.DeliveryAddress,
		ContactName:      req.ContactName,
		ContactPhone:     req.ContactPhone,
		Remark:           req.Remark,
		VerifyCode:       verifyCode,
		Status:           1,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "创建订单失败")
		return
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.ServerError, "创建订单商品失败")
			return
		}
	}

	// 扣减库存
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.ServerError, "扣减库存失败")
			return
		}
	}

	// 增加销量
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("sales", gorm.Expr("sales + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.ServerError, "更新销量失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "创建订单失败")
		return
	}

	database.DB.Preload("Items").First(&order, order.ID)
	response.Success(c, order)
}

func GetOrders(c *gin.Context) {
	userID := utils.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Order{}).Where("user_id = ?", userID)

	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	offset := (page - 1) * pageSize
	if err := query.Preload("Merchant").Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "获取订单列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": orders,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

func GetOrderDetail(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var order models.Order
	if err := database.DB.Preload("Merchant").Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	response.Success(c, order)
}

func CancelOrder(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	if order.Status != 1 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单状态不正确，无法取消")
		return
	}

	now := time.Now()
	if err := database.DB.Model(&order).Updates(map[string]interface{}{
		"status":       4,
		"cancelled_at": now,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "取消订单失败")
		return
	}

	response.Success(c, gin.H{"message": "订单已取消"})
}

type ApplyRefundRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func ApplyRefund(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderID := c.Param("order_id")
	id, _ := strconv.ParseUint(orderID, 10, 64)

	var req ApplyRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "参数错误")
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.NotFound, "订单不存在")
		return
	}

	if order.Status < 2 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单未支付，无法退款")
		return
	}

	if order.Status >= 5 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParams, "订单已退款")
		return
	}

	refundNo := strconv.FormatInt(time.Now().UnixNano(), 10)
	refund := models.Refund{
		OrderID:      id,
		RefundNo:     refundNo,
		RefundAmount: order.PayAmount,
		RefundReason: req.Reason,
		Status:       0,
	}

	if err := database.DB.Create(&refund).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.ServerError, "申请退款失败")
		return
	}

	response.Success(c, refund)
}
