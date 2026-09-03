package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fz_yyc_api/internal/config"
	wsHandler "fz_yyc_api/internal/handlers/ws"
	"fz_yyc_api/internal/models"
	categorypkg "fz_yyc_api/internal/services/category"
	couponpkg "fz_yyc_api/internal/services/coupon"
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

func recordUserBehaviorEvent(userID uint64, openID string, eventType string, page string, productID *uint64, orderID *uint64, source string, payload map[string]interface{}) {
	var payloadJSON models.JSON
	if len(payload) > 0 {
		if raw, err := json.Marshal(payload); err == nil {
			payloadJSON = models.JSON(raw)
		}
	}

	event := models.UserBehaviorEvent{
		UserID:    userID,
		OpenID:    openID,
		EventType: eventType,
		Page:      page,
		ProductID: productID,
		OrderID:   orderID,
		Source:    source,
		Payload:   payloadJSON,
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
	for index := range order.Items {
		order.Items[index].Image = buildAccessibleOrderItemImage(order.Items[index].Image)
	}

	// 指派服务人员姓名（gorm:"-" 瞬时字段，非查询返回）
	if order.AssignedStaffID != nil {
		var staff models.ServiceStaff
		if err := database.DB.Select("name").First(&staff, *order.AssignedStaffID).Error; err == nil {
			order.AssignedStaffName = staff.Name
		}
	}

	// 待评价标记：服务订单已完成(biz_status=5)且未评价
	if utils.OrderCategory(order.OrderType) == utils.OrderCategoryService && order.BizStatus == 5 {
		var reviewCount int64
		database.DB.Model(&models.ServiceReview{}).Where("order_id = ?", order.ID).Count(&reviewCount)
		order.CanReview = reviewCount == 0
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

// GetStoreHomeRecommends 查询商家首页推荐（仅启用，排序升序），附带商品/服务快照
func GetStoreHomeRecommends(c *gin.Context) {
	var recommends []models.HomeRecommend
	if err := database.DB.
		Preload("Product").
		Where("merchant_id = ? AND status = 1", utils.DefaultMerchantID).
		Order("sort ASC, id DESC").
		Find(&recommends).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询推荐失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(recommends))
	for _, r := range recommends {
		if r.Product == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":          r.ID,
			"product_id":  r.ProductID,
			"target_type": r.TargetType,
			"title":       r.Title,
			"product":     buildStoreProductResponse(*r.Product),
		})
	}
	response.Success(c, gin.H{"list": result})
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

	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "微信小程序配置缺失")
		return
	}

	openID, unionID, err := getWechatOpenID(req.Code, appIdentity)
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

	token, _ := utils.GenerateToken(user.ID, 0, "user", user.Nickname)

	response.Success(c, gin.H{
		"token":    token,
		"app_mode": "app",
		"app_id":   appIdentity.AppID,
		"user": gin.H{
			"id":       user.ID,
			"openid":   user.OpenID,
			"nickname": user.Nickname,
		},
	})
}

func getWechatOpenID(code string, appIdentity *wechatpay.AppIdentity) (string, string, error) {
	// 开发环境 / H5 调试：客户端会传入 "dev_" 前缀的模拟 code，
	// 此时直接用 code 本身作为伪 openid，避免请求微信 jscode2session 失败无法登录。
	if strings.HasPrefix(code, "dev_") {
		pseudoOpenID := "o_" + code[4:]
		return pseudoOpenID, "", nil
	}

	if appIdentity == nil {
		return "", "", fmt.Errorf("应用身份配置缺失")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appIdentity.AppID, appIdentity.AppSecret, code,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("构造微信请求失败: %w", err)
	}
	resp, err := http.DefaultClient.Do(httpReq)
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
	var merchant models.Merchant
	if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商家不存在")
		return
	}

	qiniuService := qiniu.GetService()
	if qiniuService != nil {
		merchant.Logo = qiniuService.BuildPrivateURL(merchant.Logo)
		merchant.CoverImage = qiniuService.BuildPrivateURL(merchant.CoverImage)
	}

	var categories []models.Category
	database.DB.Where("status = 1").Order("sort ASC, id ASC").Find(&categories)

	var hotProducts []models.Product
	database.DB.Where("status = 1").Order("sales DESC").Limit(10).Find(&hotProducts)

	var deliverySettings models.MerchantDeliverySettings
	database.DB.First(&deliverySettings)

	// C 端小程序首页启用的轮播图
	var banners []models.MiniProgramBanner
	database.DB.
		Where("merchant_id = ? AND status = 1", utils.DefaultMerchantID).
		Order("sort ASC, id DESC").
		Find(&banners)

	bannerResponses := make([]map[string]interface{}, 0, len(banners))
	for _, banner := range banners {
		image := banner.Image
		if qiniuService != nil {
			image = qiniuService.BuildPrivateURL(image)
		}
		bannerResponses = append(bannerResponses, map[string]interface{}{
			"id":         banner.ID,
			"image":      image,
			"link_type":  banner.LinkType,
			"link_value": banner.LinkValue,
		})
	}

	hotProductResponses := make([]StoreProductResponse, 0, len(hotProducts))
	for _, product := range hotProducts {
		hotProductResponses = append(hotProductResponses, buildStoreProductResponse(product))
	}

	// C端平铺返回启用分类，product_count 为该分类自身+全部子孙分类直接挂载的商品数
	categoryResponses := make([]map[string]interface{}, 0, len(categories))
	for _, cat := range categories {
		count, _ := categorypkg.SubtreeProductCount(database.DB, cat.ID)
		categoryResponses = append(categoryResponses, map[string]interface{}{
			"id":            cat.ID,
			"name":          cat.Name,
			"parent_id":     cat.ParentID,
			"level":         cat.Level,
			"sort":          cat.Sort,
			"product_count": count,
		})
	}

	// 待评价数（首页红点，已登录用户的服务订单已完成且未评价）
	pendingReviewCount := int64(0)
	if uid := utils.GetUserID(c); uid > 0 {
		database.DB.Model(&models.Order{}).
			Where("user_id = ? AND biz_status = ? AND order_type BETWEEN ? AND ?",
				uid, 5, utils.OrderTypeServiceMin, utils.OrderTypeServiceMax).
			Where("id NOT IN (?)", database.DB.Model(&models.ServiceReview{}).Select("order_id")).
			Count(&pendingReviewCount)
	}

	response.Success(c, gin.H{
		"merchant":             merchant,
		"categories":           categoryResponses,
		"hot_products":         hotProductResponses,
		"delivery_settings":    deliverySettings,
		"banners":              bannerResponses,
		"pending_review_count": pendingReviewCount,
	})
}

// GetStoreDeliveryRules 返回商家配送费规则（C端确认订单页使用）
func GetStoreDeliveryRules(c *gin.Context) {
	var deliverySettings models.MerchantDeliverySettings
	if err := database.DB.First(&deliverySettings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 未配置时返回默认规则，避免 C 端下单流程被 404 阻断
			response.Success(c, gin.H{
				"enabled":              false,
				"base_fee":             0,
				"free_delivery_amount": 0,
				"max_distance":         10,
				"distance_rules":       []interface{}{},
			})
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取配送规则失败")
		return
	}
	response.Success(c, deliverySettings)
}

func GetProducts(c *gin.Context) {
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

	query := database.DB.Model(&models.Product{}).Where("status = 1")

	if categoryID != "" {
		catID, _ := strconv.ParseUint(categoryID, 10, 64)
		// 查询该分类及其全部子孙分类下的商品（商品可挂任意层）
		categoryIDs, err := categorypkg.CollectSubtreeIDs(database.DB, catID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取商品列表失败")
			return
		}
		query = query.Where("category_id IN ?", categoryIDs)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	// 按商品类型过滤（逗号分隔多值，如 product_types=3,4 取康养套餐+陪诊服务）
	if ptStr := c.Query("product_types"); ptStr != "" {
		var ptValues []int
		for _, part := range strings.Split(ptStr, ",") {
			if v, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && v > 0 {
				ptValues = append(ptValues, v)
			}
		}
		if len(ptValues) > 0 {
			query = query.Where("product_type IN ?", ptValues)
		}
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
	productID := c.Param("product_id")
	pid, _ := strconv.ParseUint(productID, 10, 64)

	var product models.Product
	if err := database.DB.Preload("Category").Preload("Specs").Where("id = ? AND status = 1", pid).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeProductNotFound, "商品不存在或已下架")
		return
	}

	response.Success(c, buildStoreProductResponse(product))
}

func RecordUserVisit(c *gin.Context) {
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
		UserID:    user.ID,
		OpenID:    user.OpenID,
		VisitTime: now,
		Source:    req.Source,
	}
	database.DB.Create(&visit)

	recordUserBehaviorEvent(user.ID, user.OpenID, "store_visit", "store_home", nil, nil, req.Source, map[string]interface{}{
		"source": req.Source,
	})
	wsHandler.BroadcastStoreVisitNotify(user.OpenID, req.Source)

	response.Success(c, gin.H{
		"user_id":     user.ID,
		"visit_count": user.VisitCount,
	})
}

func RecordBehaviorEvent(c *gin.Context) {
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

	recordUserBehaviorEvent(user.ID, user.OpenID, req.EventType, req.Page, req.ProductID, req.OrderID, req.Source, req.Payload)
	response.Success(c, gin.H{"message": "记录成功"})
}

type CreateOrderRequest struct {
	OrderType       uint8   `json:"order_type"`
	BizStatus       uint8   `json:"biz_status"`
	ScheduledAt     string  `json:"scheduled_at"`
	AssignedStaffID uint64  `json:"assigned_staff_id"`
	RecordID        uint64  `json:"record_id"`
	AddressID       uint64  `json:"address_id"`
	UserCouponID    uint64  `json:"user_coupon_id"`
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
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	userID := utils.GetUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
		return
	}

	var currentUser models.User
	_ = database.DB.First(&currentUser, userID).Error

	var merchant models.Merchant
	if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err != nil {
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
	productIDs := make([]uint64, 0, len(req.Items))
	categoryIDs := make([]uint64, 0, len(req.Items))
	hasGoods := false // 是否含实物商品（product_type 1/2）
	hasService := false // 是否含服务商品（product_type 3/4）

	for _, item := range req.Items {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeProductNotFound, "商品不存在")
			return
		}

		if product.ProductType > maxProductType {
			maxProductType = product.ProductType
		}

		// 判定商品归属：实物(1/2)与服务(3/4)互斥，禁止混单支付
		if product.ProductType == 3 || product.ProductType == 4 {
			hasService = true
		} else {
			hasGoods = true
		}

		// 记录商品/分类ID集合，供优惠券适用范围匹配
		productIDs = append(productIDs, product.ID)
		if product.CategoryID != nil && *product.CategoryID > 0 {
			categoryIDs = append(categoryIDs, *product.CategoryID)
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

	// 禁止实物商品与服务商品混单支付（在创建订单事务前拦截）
	if hasGoods && hasService {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单不能同时包含实物商品与服务商品，请拆分下单")
		return
	}

	discountAmount := 0.0

	// 优惠券抵扣（PRD V2.0 阶段二）：提前校验并预计算抵扣金额，
	// 押金与配送费不参与抵扣；核销在下单事务内二次校验后执行
	if req.UserCouponID > 0 {
		var uc models.UserCoupon
		if err := database.DB.Where("id = ? AND user_id = ?", req.UserCouponID, userID).
			First(&uc).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeCouponNotUsable, "优惠券不存在")
			return
		}
		var tmpl models.CouponTemplate
		if err := database.DB.First(&tmpl, uc.TemplateID).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeCouponNotUsable, "优惠券不存在")
			return
		}
		d, err := couponpkg.UsableCheck(uc, tmpl, totalAmount, productIDs, categoryIDs, time.Now())
		if err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeCouponNotUsable, err.Error())
			return
		}
		discountAmount = d
	}

	payAmount := totalAmount + deliveryFee - discountAmount + totalDeposit
	if payAmount < 0 {
		payAmount = 0
	}

	orderNo := utils.GenerateOrderNo(utils.DefaultMerchantID)

	// 根据 product_type 推断 order_type 和 biz_status
	orderType := req.OrderType
	bizStatus := req.BizStatus

	if orderType == 0 {
		switch maxProductType {
		case 2: // 辅具租赁
			orderType = 2 // 租赁商品
		case 3: // 康养套餐
			orderType = 3 // 康养上门
		case 4: // 陪诊服务
			orderType = 4 // 陪诊服务
		default: // 零售
			orderType = 1 // 普通商品
		}
	}

	// 服务类订单（康养/陪诊）支付后需要派工，初始 biz_status=1（待接单）
	if bizStatus == 0 && (orderType == 3 || orderType == 4 || orderType == 5) {
		bizStatus = 1 // 待接单
	}

	// 服务订单不计配送费，且不受商家配送参数/配送距离约束：
	// 当前 handler 未引入配送费计算，deliveryFee 恒为 0，此处显式兜底，并忽略任何配送距离/配送参数。
	if utils.OrderCategory(orderType) == utils.OrderCategoryService {
		deliveryFee = 0
	}

	// 服务订单硬性规则：必须绑定 1 个本人账号下的健康档案（PRD V2.0 订单口径）
	var recordID *uint64
	if utils.OrderCategory(orderType) == utils.OrderCategoryService {
		if req.RecordID == 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "服务订单必须指定服务对象（健康档案）")
			return
		}
		var record models.HealthRecord
		if err := database.DB.Where("id = ? AND user_id = ?", req.RecordID, userID).First(&record).Error; err != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeNotFound, "健康档案不存在或不属于当前用户，请先建档")
			return
		}
		if record.Status == utils.HealthRecordStatusArchived {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "服务对象档案已归档，请重新选择")
			return
		}
		recordID = &req.RecordID
	}

	// 从收货地址提取区县，供服务订单区域匹配使用
	var deliveryDistrict string
	if req.AddressID > 0 {
		var address models.UserAddress
		if err := database.DB.Where("id = ? AND user_id = ?", req.AddressID, userID).First(&address).Error; err == nil {
			deliveryDistrict = address.District
		}
	}

	tx := database.DB.Begin()

	order := models.Order{
		OrderNo:         orderNo,
		UserID:          userID,
		OrderType:       orderType,
		BizStatus:       bizStatus,
		ScheduledAt:     scheduledAt,
		AssignedStaffID: assignedStaffID,
		RecordID:        recordID,
		DeliveryDistrict: deliveryDistrict,
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

	// 优惠券核销（事务内二次校验，防止并发重复使用）
	if req.UserCouponID > 0 {
		if _, err := couponpkg.Redeem(tx, userID, req.UserCouponID, totalAmount, productIDs, categoryIDs, order.ID, orderNo); err != nil {
			tx.Rollback()
			if errors.Is(err, couponpkg.ErrCouponNotUsable) || errors.Is(err, couponpkg.ErrCouponNotFound) {
				response.Fail(c, http.StatusBadRequest, response.CodeCouponNotUsable, err.Error())
			} else {
				response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "优惠券核销失败")
			}
			return
		}
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

	recordUserBehaviorEvent(userID, "", "submit_order", "store_confirm", nil, &order.ID, "store", map[string]interface{}{
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
		if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err == nil {
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
	category := c.Query("category") // 1=实物订单 2=服务订单（不传=全部）

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
	// 订单分类过滤：实物 order_type 1-2，服务 order_type 3-6
	if category == strconv.Itoa(int(utils.OrderCategoryGoods)) {
		query = query.Where("order_type >= ? AND order_type <= ?", utils.OrderTypeGoodsMin, utils.OrderTypeGoodsMax)
	} else if category == strconv.Itoa(int(utils.OrderCategoryService)) {
		query = query.Where("order_type >= ? AND order_type <= ?", utils.OrderTypeServiceMin, utils.OrderTypeServiceMax)
	}

	var total int64
	query.Count(&total)

	var orders []models.Order
	offset := (page - 1) * pageSize
	if err := query.Preload("Items").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取订单列表失败")
		return
	}

	accessibleOrders := make([]models.Order, 0, len(orders))
	for _, order := range orders {
		accessibleOrders = append(accessibleOrders, buildAccessibleOrder(order))
	}

	// 填充服务对象（健康档案）名称 record_name，供服务订单卡片展示
	for i := range accessibleOrders {
		orderquery.FillOrderRecordInfo(&accessibleOrders[i])
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
	if err := database.DB.Preload("Items").Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeOrderNotFound, "订单不存在")
		return
	}

	orderquery.FillOrderRecordInfo(&order)
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

	// 未支付订单取消：返还已核销的优惠券（过期时间不变，仅恢复未使用状态）
	database.DB.Model(&models.UserCoupon{}).
		Where("order_id = ? AND status = ?", order.ID, utils.UserCouponStatusUsed).
		Updates(map[string]interface{}{
			"status":   utils.UserCouponStatusUnused,
			"used_at":  nil,
			"order_id": nil,
			"order_no": "",
		})

	response.Success(c, gin.H{"message": "订单已取消"})
}

// RenewOrderRequest 用户端续租请求
type RenewOrderRequest struct {
	Duration uint `json:"duration"` // 可选：续租时长（默认取原单最长租赁时长）
}

// RenewOrder 用户端续租：基于本人已支付的租赁订单生成关联新订单（parent_order_id=原ID, renew_flag=1, 待支付），并创建支付参数
func RenewOrder(c *gin.Context) {
	userID := utils.GetUserID(c)
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "订单ID错误")
		return
	}

	var req RenewOrderRequest
	_ = c.ShouldBindJSON(&req)

	var src models.Order
	if err := database.DB.Preload("Items").Where("id = ? AND user_id = ?", orderID, userID).First(&src).Error; err != nil {
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

	// 存在待支付的续租单时避免重复生成
	var pendingRenew int64
	database.DB.Model(&models.Order{}).
		Where("parent_order_id = ? AND status = 1", orderID).Count(&pendingRenew)
	if pendingRenew > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该订单已存在待支付的续租单")
		return
	}

	duration := req.Duration
	if duration == 0 {
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

	// 复制订单费用：按续租时长重算租金，押金不变
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
		UserID:          userID,
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

	// 创建微信支付参数（沿用下单逻辑）
	var payParams gin.H
	var payHint string
	if payAmount > 0 {
		var merchant models.Merchant
		var currentUser models.User
		if err := database.DB.First(&currentUser, userID).Error; err == nil {
			if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err == nil {
				client, clientErr := wechatpay.NewServiceProviderClient()
				merchantReady := merchant.SubMchID != "" && merchant.PaymentConfigStatus == 1
				if clientErr != nil || !merchantReady {
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
				} else {
					payResponse, payErr := createWechatPayOrder(context.Background(), client, &merchant, newOrder, int64(payAmount*100), currentUser.OpenID)
					if payErr != nil {
						payHint = fmt.Sprintf("创建支付单失败: %s", payErr.Error())
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
	}

	respBody := gin.H{
		"order":      buildAccessibleOrder(newOrder),
		"pay_params": payParams,
	}
	if payHint != "" {
		respBody["pay_hint"] = payHint
	}
	response.SuccessWithMessage(c, "续租订单已生成", respBody)
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
		if err := database.DB.First(&merchant, utils.DefaultMerchantID).Error; err == nil && merchant.SubMchID != "" {
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
