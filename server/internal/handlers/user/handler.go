package user

import (
	"context"
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/config"
	wsHandler "fz_yyc_api/internal/handlers/ws"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/fullreduction"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	ID      uint64                           `json:"id"`
	Name    string                           `json:"name"`
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

func getOrCreateStoreUser(openID string, now time.Time) (*models.User, bool, error) {
	var user models.User
	result := database.DB.Where("openid = ?", openID).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			OpenID:       openID,
			Nickname:     "微信用户",
			Status:       1,
			FirstVisitAt: &now,
			LastVisitAt:  &now,
			VisitCount:   1,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			return nil, false, err
		}
		return &user, true, nil
	}
	if result.Error != nil {
		return nil, false, result.Error
	}

	return &user, false, nil
}

func recordUserBehaviorEvent(merchantID uint64, userID uint64, openID string, eventType string, page string, productID *uint64, orderID *uint64, source string, payload map[string]interface{}) {
	var payloadJSON models.JSON
	if len(payload) > 0 {
		if raw, err := json.Marshal(payload); err == nil {
			payloadJSON = models.JSON(raw)
		}
	}

	event := models.UserBehaviorEvent{
		MerchantID: merchantID,
		UserID:     userID,
		OpenID:     openID,
		EventType:  eventType,
		Page:       page,
		ProductID:  productID,
		OrderID:    orderID,
		Source:     source,
		Payload:    payloadJSON,
	}
	database.DB.Create(&event)
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

func loadProductSpecs(productID uint64) ([]models.ProductSpec, error) {
	var specs []models.ProductSpec
	if err := database.DB.
		Where("product_id = ?", productID).
		Order("id ASC").
		Find(&specs).Error; err != nil {
		return nil, err
	}
	return specs, nil
}

func parseSelectedSpecNames(specInfo string) []string {
	if strings.TrimSpace(specInfo) == "" {
		return []string{}
	}

	rawItems := strings.Split(specInfo, "/")
	result := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		name := strings.TrimSpace(item)
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

func calculateOrderItemUnitPrice(product models.Product, specInfo string) (float64, error) {
	price := product.Price
	selectedSpecNames := parseSelectedSpecNames(specInfo)
	if len(selectedSpecNames) == 0 {
		return price, nil
	}

	specs, err := loadProductSpecs(product.ID)
	if err != nil {
		return 0, err
	}

	if len(specs) == 0 {
		return 0, fmt.Errorf("商品规格不存在或已变更")
	}

	if len(selectedSpecNames) > len(specs) {
		return 0, fmt.Errorf("商品规格信息无效")
	}

	for index, selectedName := range selectedSpecNames {
		options := parseStoreSpecOptions(specs[index].Options)
		matched := false
		for _, option := range options {
			if option.Name != selectedName {
				continue
			}
			price += option.Price
			matched = true
			break
		}
		if !matched {
			return 0, fmt.Errorf("商品规格已变更，请重新选择")
		}
	}

	return price, nil
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

func buildAccessibleOrderItemImage(image string) string {
	service := qiniu.GetService()
	if service == nil {
		return image
	}
	return service.BuildPrivateURL(image)
}

func buildAccessibleOrder(order models.Order) models.Order {
	if order.Merchant != nil {
		order.Merchant.Logo = buildAccessibleOrderItemImage(order.Merchant.Logo)
		order.Merchant.CoverImage = buildAccessibleOrderItemImage(order.Merchant.CoverImage)
	}

	for index := range order.Items {
		order.Items[index].Image = buildAccessibleOrderItemImage(order.Items[index].Image)
	}

	if order.VerifyCode != "" {
		verifyQRCodeURL, err := utils.BuildVerifyCodeQRCodeDataURL(order.VerifyCode)
		if err == nil {
			order.VerifyQRCodeURL = verifyQRCodeURL
		}
	}

	return order
}

func buildStoreProductResponse(product models.Product) StoreProductResponse {
	categoryID := uint64(0)
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
	}

	specs := make([]StoreProductSpecResponse, 0, len(product.Specs))
	for _, spec := range product.Specs {
		specs = append(specs, StoreProductSpecResponse{
			ID:      spec.ID,
			Name:    spec.Name,
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

	if req.Code == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "授权码不能为空")
		return
	}

	openID, unionID, err := getWechatOpenID(req.Code)
	if err != nil {
		log.Printf("获取微信openid失败: %v", err)
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "微信登录失败")
		return
	}

	var user models.User
	result := database.DB.Where("openid = ?", openID).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			OpenID:   openID,
			UnionID:  unionID,
			Nickname: "微信用户",
			Status:   1,
		}
		if createErr := database.DB.Create(&user).Error; createErr != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建用户失败")
			return
		}
	} else if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "登录失败")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"last_visit_at": now,
		"visit_count":   gorm.Expr("visit_count + 1"),
	}
	if user.FirstVisitAt == nil {
		updates["first_visit_at"] = now
	}
	database.DB.Model(&user).Updates(updates)

	token, _ := utils.GenerateToken(user.ID, "user", user.Nickname)
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token":    token,
		"app_mode": appIdentity.Mode,
		"app_id":   appIdentity.AppID,
		"user": gin.H{
			"id":       user.ID,
			"openid":   user.OpenID,
			"nickname": user.Nickname,
		},
	})
}

func getWechatOpenID(code string) (string, string, error) {
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return "", "", err
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appIdentity.AppID, appIdentity.AppSecret, code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("请求微信API失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("解析微信响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("微信API错误: code=%d, msg=%s", result.ErrCode, result.ErrMsg)
	}

	return result.OpenID, result.UnionID, nil
}

func GetStoreHome(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	id, _ := strconv.ParseUint(merchantID, 10, 64)

	var merchant models.Merchant
	if err := database.DB.First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	qiniuService := qiniu.GetService()
	if qiniuService != nil {
		merchant.Logo = qiniuService.BuildPrivateURL(merchant.Logo)
		merchant.CoverImage = qiniuService.BuildPrivateURL(merchant.CoverImage)
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
		"merchant":          merchant,
		"categories":        categories,
		"hot_products":      hotProductResponses,
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
	database.DB.Select("id", "min_order_amount", "takeout_enabled", "dine_in_enabled", "pickup_enabled").First(&merchant, id)

	list := make([]StoreProductResponse, 0, len(products))
	for _, product := range products {
		list = append(list, buildStoreProductResponse(product))
	}

	response.Success(c, gin.H{
		"list": list,
		"merchant": gin.H{
			"min_order_amount": merchant.MinOrderAmount,
			"takeout_enabled":  merchant.TakeoutEnabled,
			"dine_in_enabled":  merchant.DineInEnabled,
			"pickup_enabled":   merchant.PickupEnabled,
		},
		"pagination": gin.H{
			"total":     total,
			"page":      page,
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

	var merchant models.Merchant
	if err := database.DB.Select("id", "takeout_enabled", "dine_in_enabled", "pickup_enabled").First(&merchant, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
		return
	}

	var settings models.MerchantDeliverySettings
	if err := database.DB.Where("merchant_id = ?", id).First(&settings).Error; err != nil {
		response.Success(c, gin.H{
			"enabled":              false,
			"base_fee":             0,
			"free_delivery_amount": 0,
			"max_distance":         0,
			"distance_rules":       []struct{}{},
			"takeout_enabled":      merchant.TakeoutEnabled,
			"dine_in_enabled":      merchant.DineInEnabled,
			"pickup_enabled":       merchant.PickupEnabled,
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
		"enabled":              settings.Enabled,
		"base_fee":             settings.BaseFee,
		"free_delivery_amount": settings.FreeDeliveryAmount,
		"max_distance":         settings.MaxDistance,
		"distance_rules":       distanceRules,
		"takeout_enabled":      merchant.TakeoutEnabled,
		"dine_in_enabled":      merchant.DineInEnabled,
		"pickup_enabled":       merchant.PickupEnabled,
	})
}

func RecordUserVisit(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	mid, _ := strconv.ParseUint(merchantID, 10, 64)

	var req struct {
		OpenID string `json:"openid"`
		Source string `json:"source"`
	}
	_ = c.ShouldBindJSON(&req)

	if req.Source == "" {
		req.Source = "scan"
	}

	now := time.Now()

	userID := utils.GetUserID(c)
	var user *models.User
	if userID > 0 {
		var current models.User
		if err := database.DB.First(&current, userID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询用户失败")
			return
		}
		user = &current
	} else {
		if req.OpenID == "" {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "缺少openid")
			return
		}
		current, _, err := getOrCreateStoreUser(req.OpenID, now)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询用户失败")
			return
		}
		user = current
	}

	user.VisitCount += 1
	updates := map[string]interface{}{
		"last_visit_at": now,
		"visit_count":   gorm.Expr("visit_count + 1"),
	}
	if user.FirstVisitAt == nil {
		updates["first_visit_at"] = now
	}
	database.DB.Model(&user).Updates(updates)

	visit := models.UserVisit{
		UserID:     user.ID,
		MerchantID: mid,
		OpenID:     user.OpenID,
		VisitTime:  now,
		Source:     req.Source,
	}
	database.DB.Create(&visit)

	recordUserBehaviorEvent(mid, user.ID, user.OpenID, "store_visit", "store_home", nil, nil, req.Source, map[string]interface{}{
		"source": req.Source,
	})
	wsHandler.BroadcastStoreVisitNotify(mid, user.OpenID, req.Source)

	response.Success(c, gin.H{
		"user_id":     user.ID,
		"visit_count": user.VisitCount,
	})
}

func RecordBehaviorEvent(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	mid, _ := strconv.ParseUint(merchantID, 10, 64)

	var req struct {
		OpenID    string                 `json:"openid"`
		EventType string                 `json:"event_type" binding:"required"`
		Page      string                 `json:"page"`
		ProductID *uint64                `json:"product_id"`
		OrderID   *uint64                `json:"order_id"`
		Source    string                 `json:"source"`
		Payload   map[string]interface{} `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	now := time.Now()
	userID := utils.GetUserID(c)
	var user *models.User
	if userID > 0 {
		var current models.User
		if err := database.DB.First(&current, userID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询用户失败")
			return
		}
		user = &current
	} else {
		if req.OpenID == "" {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "缺少openid")
			return
		}
		current, _, err := getOrCreateStoreUser(req.OpenID, now)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询用户失败")
			return
		}
		user = current
	}

	switch req.EventType {
	case "page_view", "product_view", "submit_order", "pay_success":
	default:
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "不支持的事件类型")
		return
	}

	recordUserBehaviorEvent(mid, user.ID, user.OpenID, req.EventType, req.Page, req.ProductID, req.OrderID, req.Source, req.Payload)
	response.Success(c, gin.H{"message": "记录成功"})
}

type CreateOrderRequest struct {
	MerchantID       uint64  `json:"merchant_id"`
	DeliveryType     uint8   `json:"delivery_type" binding:"required,oneof=1 2 3"`
	DeliveryDistance float64 `json:"delivery_distance"`
	DeliveryAddress  string  `json:"delivery_address"`
	ContactName      string  `json:"contact_name"`
	ContactPhone     string  `json:"contact_phone"`
	Remark           string  `json:"remark"`
	Items            []struct {
		ProductID uint64  `json:"product_id" binding:"required"`
		Quantity  uint    `json:"quantity" binding:"required,min=1"`
		SpecInfo  string  `json:"spec_info"`
		Price     float64 `json:"price"`
	} `json:"items" binding:"required,min=1"`
}

func CreateOrder(c *gin.Context) {
	merchantID := c.Param("merchant_id")
	pathMerchantID, _ := strconv.ParseUint(merchantID, 10, 64)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	req.MerchantID = pathMerchantID

	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	var currentUser models.User
	_ = database.DB.First(&currentUser, userID).Error

	var merchant models.Merchant
	if err := database.DB.First(&merchant, req.MerchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeMerchantNotFound, "商家不存在")
		return
	}
	if merchant.Status != 1 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "商家休息中，暂不接单")
		return
	}

	switch req.DeliveryType {
	case 1:
		if !merchant.TakeoutEnabled {
			response.Fail(c, http.StatusBadRequest, response.CodeForbidden, "商家暂未开启配送")
			return
		}
		if req.DeliveryAddress == "" {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请输入收货地址")
			return
		}
		if req.ContactName == "" {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请输入联系人")
			return
		}
		if req.ContactPhone == "" {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请输入联系电话")
			return
		}
		if req.DeliveryDistance <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请选择配送距离档位")
			return
		}
	case 2:
		if !merchant.DineInEnabled {
			response.Fail(c, http.StatusBadRequest, response.CodeForbidden, "商家暂未开启堂食")
			return
		}
		req.DeliveryDistance = 0
		req.DeliveryAddress = ""
		req.ContactName = ""
		req.ContactPhone = ""
	case 3:
		if !merchant.PickupEnabled {
			response.Fail(c, http.StatusBadRequest, response.CodeForbidden, "商家暂未开启自提")
			return
		}
		req.DeliveryDistance = 0
		req.DeliveryAddress = ""
		req.ContactName = ""
		req.ContactPhone = ""
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
			if json.Valid([]byte(item.SpecInfo)) {
				specInfo = models.JSON(item.SpecInfo)
			} else {
				wrapped, _ := json.Marshal(item.SpecInfo)
				specInfo = models.JSON(wrapped)
			}
		}

		price, err := calculateOrderItemUnitPrice(product, item.SpecInfo)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
			return
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
		if err := database.DB.Where("merchant_id = ?", req.MerchantID).First(&settings).Error; err != nil || !settings.Enabled {
			response.Fail(c, http.StatusBadRequest, response.CodeForbidden, "商家暂未开启配送")
			return
		}
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

	// 满减按商品金额匹配，不包含配送费；最终支付金额再减去命中的优惠金额。
	fullReductionRules, err := fullreduction.GetActiveRulesByMerchantID(req.MerchantID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取满减规则失败")
		return
	}
	discountAmount, _ := fullreduction.CalculateDiscount(totalAmount, fullReductionRules)

	payAmount := totalAmount + deliveryFee - discountAmount
	if payAmount < 0 {
		payAmount = 0
	}

	orderNo := utils.GenerateOrderNo(req.MerchantID)
	verifyCode := utils.GenerateVerifyCode()

	tx := database.DB.Begin()

	order := models.Order{
		OrderNo:          orderNo,
		UserID:           userID,
		MerchantID:       req.MerchantID,
		TotalAmount:      totalAmount,
		DeliveryFee:      deliveryFee,
		DiscountAmount:   discountAmount,
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
		"has_ordered":  true,
		"total_orders": gorm.Expr("total_orders + 1"),
		"total_spent":  gorm.Expr("total_spent + ?", payAmount),
	})

	recordUserBehaviorEvent(req.MerchantID, userID, "", "submit_order", "store_confirm", nil, &order.ID, "store", map[string]interface{}{
		"delivery_type":     req.DeliveryType,
		"delivery_distance": req.DeliveryDistance,
		"pay_amount":        payAmount,
	})

	// 如果是支付金额大于0的订单，更新支付状态
	if payAmount > 0 {
		now := time.Now()
		database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"has_paid":      true,
			"first_paid_at": gorm.Expr("CASE WHEN first_paid_at IS NULL THEN ? ELSE first_paid_at END", now),
		})
	}

	// 如果需要支付，创建微信支付订单
	var payParams gin.H
	if payAmount > 0 {
		var merchant models.Merchant
		if err := database.DB.First(&merchant, order.MerchantID).Error; err == nil {
			client, clientErr := wechatpay.NewServiceProviderClient()
			if clientErr != nil {
				response.Fail(c, http.StatusInternalServerError, response.CodeServerError, clientErr.Error())
				return
			}

			payResponse, payErr := createWechatPayOrder(context.Background(), client, &merchant, order, int64(payAmount*100), currentUser.OpenID)
			if payErr != nil {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, payErr.Error())
				return
			}
			payParams = gin.H{
				"appId":     payResponse.AppID,
				"timeStamp": payResponse.TimeStamp,
				"nonceStr":  payResponse.NonceStr,
				"package":   payResponse.Package,
				"signType":  payResponse.SignType,
				"paySign":   payResponse.PaySign,
				"prepay_id": payResponse.PrepayID,
			}
		}
	}

	response.Success(c, gin.H{
		"order":      order,
		"pay_params": payParams,
	})
}

func createWechatPayOrder(
	ctx context.Context,
	client *wechatpay.ServiceProviderClient,
	merchant *models.Merchant,
	order models.Order,
	totalAmount int64,
	openID string,
) (*wechatpay.JSAPIPayResponse, error) {
	if merchant == nil {
		return nil, fmt.Errorf("商家不存在")
	}
	if merchant.SubMchID == "" {
		return nil, fmt.Errorf("商家尚未配置收款商户号")
	}
	if merchant.PaymentConfigStatus != 1 {
		return nil, fmt.Errorf("商家支付配置未完成，请联系服务商")
	}
	if openID == "" {
		return nil, fmt.Errorf("缺少用户支付标识")
	}
	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return nil, err
	}

	notifyURL := config.Config.WechatPay.CallbackURL
	if notifyURL == "" {
		return nil, fmt.Errorf("服务商支付回调地址未配置")
	}

	return client.CreatePartnerJSAPIPayOrder(ctx, wechatpay.JSAPIPayRequest{
		AppID:       appIdentity.AppID,
		OpenID:      openID,
		AppMode:     appIdentity.Mode,
		SubMchID:    merchant.SubMchID,
		OrderNo:     order.OrderNo,
		Description: fmt.Sprintf("%s订单支付", merchant.Name),
		TotalAmount: totalAmount,
		NotifyURL:   notifyURL,
	})
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

	accessibleOrders := make([]models.Order, 0, len(orders))
	for _, order := range orders {
		accessibleOrders = append(accessibleOrders, buildAccessibleOrder(order))
	}

	response.Success(c, gin.H{
		"list": accessibleOrders,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
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

	response.Success(c, buildAccessibleOrder(order))
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

	if order.Status == 6 {
		response.Fail(c, http.StatusBadRequest, response.CodeOrderCancelled, "订单已退款")
		return
	}

	if order.Status == 5 {
		var existing models.Refund
		if err := database.DB.Where("order_id = ? AND status IN (0, 1)", id).Order("created_at DESC").First(&existing).Error; err == nil {
			response.Success(c, existing)
			return
		}
	}

	refundNo := strconv.FormatInt(time.Now().UnixNano(), 10)
	refund := models.Refund{
		OrderID:      id,
		RefundNo:     refundNo,
		RefundAmount: order.PayAmount,
		RefundReason: req.Reason,
		Status:       0,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&refund).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "申请退款失败")
		return
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      5,
		"refunded_at": nil,
	}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新订单状态失败")
		return
	}

	tx.Commit()

	response.Success(c, refund)
}
