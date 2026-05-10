package user

import (
	"encoding/json"
	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
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

type StoreProductSpecOptionResponse struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock uint    `json:"stock,omitempty"`
}

type StoreProductSpecResponse struct {
	ID      uint64                          `json:"id"`
	Name    string                          `json:"name"`
	Options []StoreProductSpecOptionResponse `json:"options"`
}

type StoreProductResponse struct {
	ID            uint64                     `json:"id"`
	MerchantID    uint64                     `json:"merchant_id"`
	CategoryID    uint64                     `json:"category_id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Images        []string                   `json:"images"`
	Price         float64                    `json:"price"`
	OriginalPrice float64                    `json:"original_price"`
	Stock         uint                       `json:"stock"`
	Unit          string                     `json:"unit"`
	Sales         uint                       `json:"sales"`
	Sort          uint                       `json:"sort"`
	Status        uint8                      `json:"status"`
	Specs         []StoreProductSpecResponse `json:"specs"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

func parseProductImages(raw models.JSON) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return values
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil && single != "" {
		return []string{single}
	}

	return []string{}
}

func parseStoreSpecOptions(raw models.JSON) []StoreProductSpecOptionResponse {
	if len(raw) == 0 {
		return []StoreProductSpecOptionResponse{}
	}

	var options []StoreProductSpecOptionResponse
	if err := json.Unmarshal(raw, &options); err == nil {
		return options
	}

	return []StoreProductSpecOptionResponse{}
}

func buildStoreAccessibleImages(images []string) []string {
	service := qiniu.GetService()
	if service == nil {
		return images
	}

	result := make([]string, 0, len(images))
	for _, image := range images {
		result = append(result, service.BuildPrivateURL(image))
	}
	return result
}

func buildStoreProductResponse(product models.Product) StoreProductResponse {
	categoryID := uint64(0)
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
	}

	specs := make([]StoreProductSpecResponse, 0, len(product.Specs))
	for _, spec := range product.Specs {
		specs = append(specs, StoreProductSpecResponse{
			ID:   spec.ID,
			Name: spec.Name,
			Options: parseStoreSpecOptions(spec.Options),
		})
	}

	return StoreProductResponse{
		ID:            product.ID,
		MerchantID:    product.MerchantID,
		CategoryID:    categoryID,
		Name:          product.Name,
		Description:   product.Description,
		Images:        buildStoreAccessibleImages(parseProductImages(product.Images)),
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
		Stock:         product.Stock,
		Unit:          product.Unit,
		Sales:         product.Sales,
		Sort:          product.Sort,
		Status:        product.Status,
		Specs:         specs,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}
}

func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 使用code生成openid(实际项目中应调用微信API获取真实openid)
	openID := "wx_" + req.Code

	// 查找或创建用户
	var user models.User
	result := database.DB.Where("openid = ?", openID).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 首次登录,创建用户
		user = models.User{
			OpenID:   openID,
			Nickname: "微信用户",
			Status:   1,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建用户失败")
			return
		}
	} else if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "登录失败")
		return
	}

	// 生成JWT token
	token, _ := utils.GenerateToken(user.ID, "user", user.Nickname)

	// 返回完整用户信息
	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"openid":   user.OpenID,
			"nickname": user.Nickname,
		},
	})
}

func GetStoreHome(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	if merchant.Status != 1 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "商家已停业")
		return
	}

	var categories []models.Category
	database.DB.Where("merchant_id = ? AND status = 1", id).Order("sort ASC").Find(&categories)

	var hotProducts []models.Product
	database.DB.Where("merchant_id = ? AND status = 1", id).Order("sales DESC").Limit(10).Find(&hotProducts)

	var deliverySettings models.MerchantDeliverySettings
	database.DB.Where("merchant_id = ?", id).First(&deliverySettings)

	hotProductResponses := make([]StoreProductResponse, 0, len(hotProducts))
	for _, product := range hotProducts {
		hotProductResponses = append(hotProductResponses, buildStoreProductResponse(product))
	}

	response.Success(c, gin.H{
		"merchant":         merchant,
		"categories":       categories,
		"hot_products":     hotProductResponses,
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
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取商品列表失败")
		return
	}

	var merchant models.Merchant
	database.DB.Select("id", "min_order_amount", "takeout_enabled", "dine_in_enabled").First(&merchant, id)

	list := make([]StoreProductResponse, 0, len(products))
	for _, product := range products {
		list = append(list, buildStoreProductResponse(product))
	}

	response.Success(c, gin.H{
		"list": list,
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
		response.Fail(c, http.StatusNotFound, response.CodeProductNotFound, "商品不存在或已下架")
		return
	}

	response.Success(c, buildStoreProductResponse(product))
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
		MinDistance float64 `json:"min_distance"`
		MaxDistance float64 `json:"max_distance"`
		Fee         float64 `json:"fee"`
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

func RecordUserVisit(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	mid, _ := strconv.ParseUint(merchantID, 10, 64)

	var req struct {
		OpenID string `json:"openid" binding:"required"`
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "缺少openid")
		return
	}

	if req.Source == "" {
		req.Source = "scan"
	}

	var user models.User
	result := database.DB.Where("openid = ?", req.OpenID).First(&user)

	now := time.Now()

	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			OpenID:       req.OpenID,
			Nickname:     "微信用户",
			Status:       1,
			FirstVisitAt: &now,
			LastVisitAt:  &now,
			VisitCount:   1,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建用户失败")
			return
		}
	} else if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询用户失败")
		return
	} else {
		updates := map[string]interface{}{
			"last_visit_at": now,
			"visit_count":   gorm.Expr("visit_count + 1"),
		}
		if user.FirstVisitAt == nil {
			updates["first_visit_at"] = now
		}
		database.DB.Model(&user).Updates(updates)
	}

	visit := models.UserVisit{
		UserID:     user.ID,
		MerchantID: mid,
		OpenID:     req.OpenID,
		VisitTime:  now,
		Source:     req.Source,
	}
	database.DB.Create(&visit)

	response.Success(c, gin.H{
		"user_id":     user.ID,
		"visit_count": user.VisitCount,
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
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if code, ok := body["code"].(string); ok && code != "" {
				openID := "mock_openid_" + code
				var user models.User
				result := database.DB.Where("openid = ?", openID).First(&user)
				if result.Error == gorm.ErrRecordNotFound {
					user = models.User{
						OpenID:   openID,
						Nickname: "微信用户",
						Status:   1,
					}
					database.DB.Create(&user)
				}
				userID = user.ID
			}
		}

		if userID == 0 {
			userID = 1
		}
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var merchant models.Merchant
	if err := database.DB.First(&merchant, req.MerchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
		return
	}

	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeProductNotFound, "商品不存在")
			return
		}

		if product.MerchantID != req.MerchantID {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "商品不属于该商家")
			return
		}

		if product.Status != 1 {
			response.Fail(c, http.StatusBadRequest, response.CodeProductOffSale, "商品已下架")
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
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单金额未达到最低消费")
		return
	}

	var deliveryFee float64
	if req.DeliveryType == 1 {
		var settings models.MerchantDeliverySettings
		if err := database.DB.Where("merchant_id = ?", req.MerchantID).First(&settings).Error; err == nil && settings.Enabled {
			if req.DeliveryDistance > float64(settings.MaxDistance) {
				response.Fail(c, http.StatusBadRequest, response.CodeOutOfRange, "超出配送范围")
				return
			}

			var rules []map[string]interface{}
			if settings.DistanceRules != nil {
				_ = json.Unmarshal(settings.DistanceRules, &rules)
			}
			deliveryFee = utils.CalculateDeliveryFee(totalAmount, settings.BaseFee, settings.FreeDeliveryAmount, req.DeliveryDistance, rules)
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
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建订单失败")
		return
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建订单商品失败")
			return
		}
	}

	// 扣减库存
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "扣减库存失败")
			return
		}
	}

	// 增加销量
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("sales", gorm.Expr("sales + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新销量失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建订单失败")
		return
	}

	database.DB.Preload("Items").First(&order, order.ID)

	// 更新用户下单统计
	database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"has_ordered":   true,
		"total_orders":  gorm.Expr("total_orders + 1"),
		"total_spent":   gorm.Expr("total_spent + ?", payAmount),
	})

	// 如果是支付金额大于0的订单，更新支付状态
	if payAmount > 0 {
		now := time.Now()
		database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"has_paid":       true,
			"first_paid_at":  gorm.Expr("CASE WHEN first_paid_at IS NULL THEN ? ELSE first_paid_at END", now),
		})
	}

	// 如果需要支付，创建微信支付订单
	var payParams gin.H
	if payAmount > 0 {
		var merchant models.Merchant
		if err := database.DB.First(&merchant, order.MerchantID).Error; err == nil {
			payParams = createWechatPayOrder(merchant.SubMchID, order.OrderNo, int64(payAmount*100), config.Config.Wechat.AppID)
		}
	}

	response.Success(c, gin.H{
		"order":      order,
		"pay_params": payParams,
	})
}

func createWechatPayOrder(subMchID, orderNo string, totalAmount int64, appID string) gin.H {
	// 服务商模式微信支付统一下单
	// 实际项目中需要调用微信支付API
	// 这里返回模拟支付参数用于测试

	return gin.H{
		"appId":     appID,
		"timeStamp": strconv.FormatInt(time.Now().Unix(), 10),
		"nonceStr":  strconv.FormatInt(time.Now().UnixNano(), 10),
		"package":   "prepay_id=wx" + orderNo,
		"signType":  "MD5",
		"paySign":   "",
	}
}

func GetOrders(c *gin.Context) {
	userID := utils.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	merchantID := c.Query("merchant_id")

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

	if merchantID != "" {
		query = query.Where("merchant_id = ?", merchantID)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	offset := (page - 1) * pageSize
	if err := query.Preload("Merchant").Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取订单列表失败")
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
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
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
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	if order.Status != 1 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "订单状态不正确，无法取消")
		return
	}

	now := time.Now()
	if err := database.DB.Model(&order).Updates(map[string]interface{}{
		"status":       4,
		"cancelled_at": now,
	}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "取消订单失败")
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
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var order models.Order
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	if order.Status < 2 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderStatusError, "订单未支付，无法退款")
		return
	}

	if order.Status >= 5 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderCancelled, "订单已退款")
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
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "申请退款失败")
		return
	}

	response.Success(c, refund)
}
