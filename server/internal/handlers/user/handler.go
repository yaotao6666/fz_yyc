package user

import (
	"context"
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/config"
	wsHandler "fz_yyc_api/internal/handlers/ws"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/orderquery"
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
	ID                uint64                     `json:"id"`
	MerchantID        uint64                     `json:"merchant_id"`
	CategoryID        uint64                     `json:"category_id"`
	Name              string                     `json:"name"`
	Description       string                     `json:"description"`
	Images            []string                   `json:"images"`
	Price             float64                    `json:"price"`
	OriginalPrice     float64                    `json:"original_price"`
	Stock             uint                       `json:"stock"`
	Unit              string                     `json:"unit"`
	ProductType       uint8                      `json:"product_type"`
	ServiceContent    interface{}                `json:"service_content"`
	SaleType          uint8                      `json:"sale_type"`
	RentalUnit        uint8                      `json:"rental_unit"`
	RentalPrice       float64                    `json:"rental_price"`
	Deposit           float64                    `json:"deposit"`
	MaxRentalDuration uint                       `json:"max_rental_duration"`
	Sales             uint                       `json:"sales"`
	Sort              uint                       `json:"sort"`
	Status            uint8                      `json:"status"`
	Specs             []StoreProductSpecResponse `json:"specs"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
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

func parseServiceContent(raw models.JSON) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var result interface{}
	if err := json.Unmarshal(raw, &result); err == nil {
		return result
	}
	return nil
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
		ID:                product.ID,
		MerchantID:        product.MerchantID,
		CategoryID:        categoryID,
		Name:              product.Name,
		Description:       product.Description,
		Images:            buildStoreAccessibleImages(parseProductImages(product.Images)),
		Price:             product.Price,
		OriginalPrice:     product.OriginalPrice,
		Stock:             product.Stock,
		Unit:              product.Unit,
		ProductType:       product.ProductType,
		ServiceContent:    parseServiceContent(product.ServiceContent),
		SaleType:          product.SaleType,
		RentalUnit:        product.RentalUnit,
		RentalPrice:       product.RentalPrice,
		Deposit:           product.Deposit,
		MaxRentalDuration: product.MaxRentalDuration,
		Sales:             product.Sales,
		Sort:              product.Sort,
		Status:            product.Status,
		Specs:             specs,
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
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

	list := make([]StoreProductResponse, 0, len(products))
	for _, product := range products {
		list = append(list, buildStoreProductResponse(product))
	}

	response.Success(c, gin.H{
		"list": list,
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
	MerchantID      uint64  `json:"merchant_id"`
	OrderType       uint8   `json:"order_type"`
	BizStatus       uint8   `json:"biz_status"`
	ScheduledAt     string  `json:"scheduled_at"`
	AssignedStaffID uint64  `json:"assigned_staff_id"`
	DeliveryAddress string  `json:"delivery_address"`
	ContactName     string  `json:"contact_name"`
	ContactPhone    string  `json:"contact_phone"`
	Remark          string  `json:"remark"`
	Items           []struct {
		ProductID      uint64  `json:"product_id" binding:"required"`
		Quantity       uint    `json:"quantity" binding:"required,min=1"`
		SpecInfo       string  `json:"spec_info"`
		Price          float64 `json:"price"`
		RentalDuration uint    `json:"rental_duration"`
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

	var scheduledAt *time.Time
	if req.ScheduledAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ScheduledAt); err == nil {
			scheduledAt = &parsed
		} else if parsed, err := time.Parse("2006-01-02 15:04:05", req.ScheduledAt); err == nil {
			scheduledAt = &parsed
		}
	}

	var assignedStaffID *uint64
	if req.AssignedStaffID > 0 {
		assignedStaffID = &req.AssignedStaffID
	}

	var totalAmount float64
	var totalDeposit float64
	var orderItems []models.OrderItem
	var maxProductType uint8 // 跟踪最高商品类型以推断 order_type

	for _, item := range req.Items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeProductNotFound, "商品不存在")
			return
		}

		if product.ProductType > maxProductType {
			maxProductType = product.ProductType
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

		var images []string
		if product.Images != nil {
			json.Unmarshal(product.Images, &images)
		}
		image := ""
		if len(images) > 0 {
			image = images[0]
		}

		if product.SaleType == 2 {
			if item.RentalDuration == 0 {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁商品必须选择租赁时长")
				return
			}
			if product.MaxRentalDuration > 0 && item.RentalDuration > product.MaxRentalDuration {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁时长超出限制")
				return
			}
			rentalSubtotal := product.RentalPrice * float64(item.RentalDuration) * float64(item.Quantity)
			itemDeposit := product.Deposit * float64(item.Quantity)
			totalAmount += rentalSubtotal
			totalDeposit += itemDeposit

			orderItems = append(orderItems, models.OrderItem{
				MerchantID:      req.MerchantID,
				ProductID:       item.ProductID,
				ProductName:     product.Name,
				Image:           image,
				Price:           product.RentalPrice,
				Quantity:        item.Quantity,
				SpecInfo:        specInfo,
				Subtotal:        rentalSubtotal,
				SaleType:        2,
				RentalUnit:      product.RentalUnit,
				RentalDuration:  item.RentalDuration,
				UnitRentalPrice: product.RentalPrice,
				RentalSubtotal:  rentalSubtotal,
				Deposit:         itemDeposit,
			})
		} else {
			price, err := calculateOrderItemUnitPrice(product, item.SpecInfo)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
				return
			}

			subtotal := price * float64(item.Quantity)
			totalAmount += subtotal

			orderItems = append(orderItems, models.OrderItem{
				MerchantID:  req.MerchantID,
				ProductID:   item.ProductID,
				ProductName: product.Name,
				Image:       image,
				Price:       price,
				Quantity:    item.Quantity,
				SpecInfo:    specInfo,
				Subtotal:    subtotal,
				SaleType:    1,
			})
		}
	}

	var deliveryFee float64

	discountAmount := 0.0

	payAmount := totalAmount + deliveryFee - discountAmount + totalDeposit
	if payAmount < 0 {
		payAmount = 0
	}

	orderNo := utils.GenerateOrderNo(req.MerchantID)

	// 根据 product_type 推断 order_type 和 biz_status
	orderType := req.OrderType
	bizStatus := req.BizStatus

	if orderType == 0 {
		switch maxProductType {
		case 2: // 辅具租赁
			orderType = 2 // 租赁商品
		case 3: // 康养套餐
			orderType = 5 // 上门服务
		case 4: // 陪诊服务
			orderType = 4 // 预约服务
		default: // 零售/资讯
			orderType = 1 // 普通商品
		}
	}

	// 服务类订单（康养/陪诊）支付后需要派工，初始 biz_status=1（待接单）
	if bizStatus == 0 && (orderType == 3 || orderType == 4 || orderType == 5) {
		bizStatus = 1 // 待接单
	}

	tx := database.DB.Begin()

	order := models.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		MerchantID:      req.MerchantID,
		OrderType:       orderType,
		BizStatus:       bizStatus,
		ScheduledAt:     scheduledAt,
		AssignedStaffID: assignedStaffID,
		TotalAmount:     totalAmount,
		DeliveryFee:     deliveryFee,
		DiscountAmount:  discountAmount,
		PayAmount:       payAmount,
		TotalDeposit:    totalDeposit,
		DeliveryAddress: req.DeliveryAddress,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		Remark:          req.Remark,
		Status:          1,
	}
	if totalDeposit > 0 {
		order.DepositStatus = 1
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

	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "扣减库存失败")
			return
		}
	}

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

	database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"has_ordered":  true,
		"total_orders": gorm.Expr("total_orders + 1"),
		"total_spent":  gorm.Expr("total_spent + ?", payAmount),
	})

	recordUserBehaviorEvent(req.MerchantID, userID, "", "submit_order", "store_confirm", nil, &order.ID, "store", map[string]interface{}{
		"order_type": orderType,
		"pay_amount": payAmount,
	})

	if payAmount > 0 {
		now := time.Now()
		database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"has_paid":      true,
			"first_paid_at": gorm.Expr("CASE WHEN first_paid_at IS NULL THEN ? ELSE first_paid_at END", now),
		})
	}

	var payParams gin.H
	var payHint string
	if payAmount > 0 {
		var merchant models.Merchant
		if err := database.DB.First(&merchant, order.MerchantID).Error; err == nil {
			client, clientErr := wechatpay.NewServiceProviderClient()
			merchantReady := merchant.SubMchID != "" && merchant.PaymentConfigStatus == 1
			if clientErr != nil || !merchantReady {
				// 降级：不阻断下单，提示支付凭证未配置
				var reasons []string
				if clientErr != nil {
					reasons = append(reasons, fmt.Sprintf("服务商支付凭证未配置: %s", clientErr.Error()))
				}
				if merchant.SubMchID == "" {
					reasons = append(reasons, "商家尚未绑定收款商户号")
				}
				if merchant.PaymentConfigStatus != 1 {
					reasons = append(reasons, "商家支付配置未完成")
				}
				payHint = strings.Join(reasons, "；")
				log.Printf("[CREATE-ORDER-DEGRADE] order_no=%s, hint=%s", order.OrderNo, payHint)
			} else {
				payResponse, payErr := createWechatPayOrder(context.Background(), client, &merchant, order, int64(payAmount*100), currentUser.OpenID)
				if payErr != nil {
					payHint = fmt.Sprintf("创建支付单失败: %s", payErr.Error())
					log.Printf("[CREATE-ORDER-DEGRADE] order_no=%s, pay_error=%s", order.OrderNo, payHint)
				} else {
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
		}
	}

	respBody := gin.H{
		"order":      order,
		"pay_params": payParams,
	}
	if payHint != "" {
		respBody["pay_hint"] = payHint
	}
	response.Success(c, respBody)
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
		return nil, fmt.Errorf("商家支付配置未完成，请在 PC 后台完成支付配置")
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

	// 同步调用微信退款接口（仅当订单已完成支付回调、商家已配置子商户号时）
	if strings.TrimSpace(order.TransactionID) != "" && order.PayAmount > 0 {
		var merchant models.Merchant
		if err := database.DB.First(&merchant, order.MerchantID).Error; err == nil && merchant.SubMchID != "" {
			notifyURL := config.Config.WechatPay.CallbackURL
			client, cliErr := wechatpay.NewServiceProviderClient()
			if cliErr == nil && notifyURL != "" {
				refundResp, callErr := client.CreatePartnerRefund(context.Background(), wechatpay.RefundRequest{
					SubMchID:     merchant.SubMchID,
					OrderNo:      order.OrderNo,
					RefundNo:     refund.RefundNo,
					Reason:       req.Reason,
					NotifyURL:    notifyURL,
					RefundAmount: int64(order.PayAmount * 100),
					TotalAmount:  int64(order.PayAmount * 100),
				})
				if callErr == nil {
					refundID := strings.TrimSpace(refundResp.RefundID)
					_ = orderquery.SyncRefundAndOrderStatus(database.DB, &order, &refund, refundResp.Status, refundID, refundResp.SuccessTime)
					refundStatus := strings.ToUpper(strings.TrimSpace(refundResp.Status))
					if refundStatus != "SUCCESS" {
						if refreshed, queryErr := client.QueryPartnerRefundByRefundNo(context.Background(), refund.RefundNo); queryErr == nil {
							_ = orderquery.SyncRefundAndOrderStatus(database.DB, &order, &refund, refreshed.Status, refreshed.RefundID, refreshed.SuccessTime)
						}
					}
				} else {
					_ = database.DB.Model(&refund).Updates(map[string]any{
						"status": 2,
					}).Error
					log.Printf("[ApplyRefund] 调用微信退款失败: refund_no=%s, err=%v", refund.RefundNo, callErr)
				}
			}
		}
	}

	database.DB.First(&refund, refund.ID)
	response.Success(c, refund)
}
