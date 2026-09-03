package merchant

import (
	"encoding/json"
	"errors"
	"fmt"
	"fz_yyc_api/internal/models"
	categorypkg "fz_yyc_api/internal/services/category"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductSpecOptionResponse struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock uint    `json:"stock,omitempty"`
}

type ProductSpecResponse struct {
	ID      uint64                      `json:"id"`
	Name    string                      `json:"name"`
	Options []ProductSpecOptionResponse `json:"options"`
}

type ProductResponse struct {
	ID                 uint64                `json:"id"`
	CategoryID         uint64                `json:"category_id"`
	Name               string                `json:"name"`
	Description        string                `json:"description"`
	Images             []string              `json:"images"`
	Price              float64               `json:"price"`
	OriginalPrice      float64               `json:"original_price"`
	Stock              uint                  `json:"stock"`
	Unit               string                `json:"unit"`
	ProductType        uint8                 `json:"product_type"`
	ServiceContent     interface{}           `json:"service_content"`
	SaleType           uint8                 `json:"sale_type"`
	RentalUnit         uint8                 `json:"rental_unit"`
	RentalPrice        float64               `json:"rental_price"`
	Deposit            float64               `json:"deposit"`
	MaxRentalDuration  uint                  `json:"max_rental_duration"`
	Sales              uint                  `json:"sales"`
	Sort               uint                  `json:"sort"`
	Status             uint8                 `json:"status"`
	CategoryName       string                `json:"category_name,omitempty"`
	Specs              []ProductSpecResponse `json:"specs"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
}

func parseStringArray(raw models.JSON) []string {
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

	trimmedRaw := strings.TrimSpace(string(raw))
	if trimmedRaw == "" {
		return []string{}
	}
	if strings.Contains(trimmedRaw, ",") {
		parts := strings.Split(trimmedRaw, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			normalized := strings.TrimSpace(strings.Trim(part, `"'`))
			if normalized != "" {
				result = append(result, normalized)
			}
		}
		if len(result) > 0 {
			return result
		}
	}

	normalized := strings.TrimSpace(strings.Trim(trimmedRaw, `"'`))
	if normalized != "" {
		return []string{normalized}
	}

	return []string{}
}

func buildAccessibleImages(images []string) []string {
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

func parseSpecOptions(raw models.JSON) []ProductSpecOptionResponse {
	if len(raw) == 0 {
		return []ProductSpecOptionResponse{}
	}

	var options []ProductSpecOptionResponse
	if err := json.Unmarshal(raw, &options); err == nil {
		return options
	}

	return []ProductSpecOptionResponse{}
}

// parseServiceContent 将 JSON 字段解析为 interface{}，为 nil 时返回 nil
func parseServiceContent(raw models.JSON) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var content interface{}
	if err := json.Unmarshal(raw, &content); err == nil {
		return content
	}
	return nil
}

func buildProductResponse(product models.Product) ProductResponse {
	categoryID := uint64(0)
	if product.CategoryID != nil {
		categoryID = *product.CategoryID
	}

	categoryName := ""
	if product.Category != nil {
		categoryName = product.Category.Name
	}

	specs := make([]ProductSpecResponse, 0, len(product.Specs))
	for _, spec := range product.Specs {
		specs = append(specs, ProductSpecResponse{
			ID:      spec.ID,
			Name:    spec.Name,
			Options: parseSpecOptions(spec.Options),
		})
	}

	return ProductResponse{
		ID:                 product.ID,
		CategoryID:         categoryID,
		Name:               product.Name,
		Description:        product.Description,
		Images:             buildAccessibleImages(parseStringArray(product.Images)),
		Price:              product.Price,
		OriginalPrice:      product.OriginalPrice,
		Stock:              product.Stock,
		Unit:               product.Unit,
		ProductType:        product.ProductType,
		ServiceContent:     parseServiceContent(product.ServiceContent),
		SaleType:           product.SaleType,
		RentalUnit:         product.RentalUnit,
		RentalPrice:        product.RentalPrice,
		Deposit:            product.Deposit,
		MaxRentalDuration:  product.MaxRentalDuration,
		Sales:              product.Sales,
		Sort:               product.Sort,
		Status:             product.Status,
		CategoryName:       categoryName,
		Specs:              specs,
		CreatedAt:          product.CreatedAt,
		UpdatedAt:          product.UpdatedAt,
	}
}

func isRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func respondProductQueryError(c *gin.Context, err error, notFoundMessage string, serverErrorMessage string) {
	if isRecordNotFoundError(err) {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, notFoundMessage)
		return
	}

	response.Fail(c, http.StatusInternalServerError, response.CodeServerError, serverErrorMessage)
}

func loadProductWithRelations(id uint64) (*models.Product, error) {
	var product models.Product
	err := database.DB.
		Where("id = ? AND deleted_at IS NULL", id).
		Preload("Category").
		First(&product).Error
	if err != nil {
		return nil, err
	}

	// 这里不用 GORM 的 Preload("Specs")，直接按 product_id 查询规格，避免运行时出现
	// “unsupported relations for schema Product” 并把内部错误误判成商品不存在。
	var specs []models.ProductSpec
	if err := database.DB.
		Where("product_id = ?", product.ID).
		Order("id ASC").
		Find(&specs).Error; err != nil {
		return nil, err
	}

	product.Specs = specs
	return &product, nil
}

func GetCategories(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)

	parentParam := c.Query("parent_id")

	// 不带 parent_id：返回以一级分类为根的分类树
	if parentParam == "" {
		tree, err := categorypkg.BuildTree(database.DB)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取分类列表失败")
			return
		}
		response.Success(c, tree)
		return
	}

	// 带 parent_id：返回该父分类下的直接子级平铺列表
	parentID, _ := strconv.ParseUint(parentParam, 10, 64)
	var children []models.Category
	if err := database.DB.Where("parent_id = ?", parentID).Order("sort ASC, id ASC").Find(&children).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取分类列表失败")
		return
	}

	result := make([]categorypkg.Node, 0, len(children))
	for _, cat := range children {
		count, _ := categorypkg.SubtreeProductCount(database.DB, cat.ID)
		result = append(result, categorypkg.Node{
			ID:           cat.ID,
			Name:         cat.Name,
			ParentID:     cat.ParentID,
			Level:        cat.Level,
			Sort:         cat.Sort,
			Status:       cat.Status,
			ProductCount: count,
			CreatedAt:    cat.CreatedAt,
			UpdatedAt:    cat.UpdatedAt,
		})
	}
	response.Success(c, result)
}

type CategoryRequest struct {
	Name     string  `json:"name" binding:"required"`
	ParentID *uint64 `json:"parent_id"`
	Sort     *uint   `json:"sort"`
	Status   uint8   `json:"status"`
}

// loadCategory 按 id 加载分类
func loadCategory(id uint64) (*models.Category, error) {
	var category models.Category
	if err := database.DB.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("分类不存在")
		}
		return nil, err
	}
	return &category, nil
}

// validateParentLevel 校验父分类并返回本节点应有的 level（空父=一级）
func validateParentLevel(parentID *uint64) (uint8, error) {
	if parentID == nil {
		return 1, nil
	}
	parent, err := loadCategory(*parentID)
	if err != nil {
		return 0, err
	}
	level := parent.Level + 1
	if level > 3 {
		return 0, errors.New("最多支持三级分类")
	}
	return level, nil
}

// countDirectChildren 统计直接子分类数量
func countDirectChildren(id uint64) (int64, error) {
	var cnt int64
	err := database.DB.Model(&models.Category{}).Where("parent_id = ?", id).Count(&cnt).Error
	return cnt, err
}

// categoryMaxDescendantDepth 返回该分类子树从自身算起的最大后代层级数（自身为0，子为1，孙为2）
func categoryMaxDescendantDepth(id uint64) uint8 {
	subtree, err := categorypkg.CollectSubtreeIDs(database.DB, id)
	if err != nil {
		return 0
	}
	self, err := loadCategory(id)
	if err != nil {
		return 0
	}
	var maxDepth uint8
	for _, sid := range subtree {
		if sid == id {
			continue
		}
		cat, err := loadCategory(sid)
		if err != nil {
			continue
		}
		depth := cat.Level - self.Level
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func CreateCategory(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	level, err := validateParentLevel(req.ParentID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	// 同级名称唯一校验
	if dupNameUnderParent(c, req.Name, req.ParentID, 0) {
		return
	}

	category := models.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
		Level:    level,
		Sort:     0,
		Status:   1,
	}
	if req.Sort != nil {
		category.Sort = *req.Sort
	}
	if req.Status > 0 {
		category.Status = req.Status
	}

	if err := database.DB.Create(&category).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建分类失败")
		return
	}

	response.Success(c, category)
}

// dupNameUnderParent 检测同级（同一父下）是否已存在同名分类；存在则返回 false 并写失败响应
func dupNameUnderParent(c *gin.Context, name string, parentID *uint64, excludeID uint64) bool {
	query := database.DB.Model(&models.Category{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}
	var count int64
	query.Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "同级已存在同名分类")
		return true
	}
	return false
}

func UpdateCategory(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	categoryID := c.Param("category_id")
	id, _ := strconv.ParseUint(categoryID, 10, 64)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	category, err := loadCategory(id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" && req.Name != category.Name {
		if dupNameUnderParent(c, req.Name, category.ParentID, id) {
			return
		}
		updates["name"] = req.Name
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	// 父级变更（仅支持显式指定新的父 id；如需恢复一级可另行支持，此处保证向下分层正确）
	if req.ParentID != nil {
		newParent := *req.ParentID
		if newParent == category.ID {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父分类不能是自身")
			return
		}
		if category.ParentID == nil || *category.ParentID != newParent {
			// 新父不能是自己子孙（防环）
			subtree, err := categorypkg.CollectSubtreeIDs(database.DB, category.ID)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "校验分类层级失败")
				return
			}
			for _, sid := range subtree {
				if sid == newParent {
					response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父分类不能是自己的子分类")
					return
				}
			}
			level, err := validateParentLevel(req.ParentID)
			if err != nil {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
				return
			}
			// 迁移后自身子树最大深度仍 ≤3
			if uint8(level)+categoryMaxDescendantDepth(category.ID) > 3 {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, "移动后子分类将超过三级")
				return
			}
			updates["parent_id"] = newParent
			updates["level"] = level
		}
	}

	if len(updates) == 0 {
		response.Success(c, category)
		return
	}

	if err := database.DB.Model(category).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新分类失败")
		return
	}

	database.DB.First(category, id)
	response.Success(c, category)
}

func DeleteCategory(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	categoryID := c.Param("category_id")
	id, _ := strconv.ParseUint(categoryID, 10, 64)

	childCount, err := countDirectChildren(id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除分类失败")
		return
	}
	if childCount > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "请先删除该分类的子分类")
		return
	}

	if err := database.DB.Where("id = ?", id).Delete(&models.Category{}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除分类失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

type SortCategoriesRequest struct {
	Categories []struct {
		ID   uint64 `json:"id"`
		Sort uint   `json:"sort"`
	} `json:"categories" binding:"required"`
}

func SortCategories(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)

	var req SortCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	tx := database.DB.Begin()
	for _, item := range req.Categories {
		if err := tx.Model(&models.Category{}).Where("id = ?", item.ID).Update("sort", item.Sort).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "排序失败")
			return
		}
	}
	tx.Commit()

	response.Success(c, gin.H{"message": "排序成功"})
}

func GetProduct(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	product, err := loadProductWithRelations(id)
	if err != nil {
		respondProductQueryError(c, err, "商品不存在", "获取商品详情失败")
		return
	}

	response.Success(c, buildProductResponse(*product))
}

func GetProducts(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID := c.Query("category_id")
	status := c.Query("status")
	keyword := c.Query("keyword")
	saleType := c.Query("sale_type")
	productType := c.Query("product_type")
	productTypes := c.Query("product_types") // 逗号分隔多值：如 3,4（服务管理）/1,2,5（商品管理）

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Product{}).Where("deleted_at IS NULL")

	if categoryID != "" {
		id, _ := strconv.ParseUint(categoryID, 10, 64)
		if id > 0 {
			// 按一级分类查询时，同时命中其下全部子孙分类的商品（分类最多三级）
			if ids, err := categorypkg.CollectSubtreeIDs(database.DB, id); err == nil && len(ids) > 0 {
				query = query.Where("category_id IN ?", ids)
			} else {
				query = query.Where("category_id = ?", id)
			}
		}
	}
	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
	}
	if saleType != "" {
		saleTypeInt, _ := strconv.Atoi(saleType)
		query = query.Where("sale_type = ?", saleTypeInt)
	}
	if productType != "" {
		productTypeInt, _ := strconv.Atoi(productType)
		query = query.Where("product_type = ?", productTypeInt)
	} else if productTypes != "" {
		typeList := make([]int, 0)
		for _, part := range strings.Split(productTypes, ",") {
			if value, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && value > 0 {
				typeList = append(typeList, value)
			}
		}
		if len(typeList) > 0 {
			query = query.Where("product_type IN ?", typeList)
		}
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	offset := (page - 1) * pageSize
	if err := query.Preload("Category").Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&products).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取商品列表失败")
		return
	}

	result := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		result = append(result, buildProductResponse(product))
	}

	response.Success(c, gin.H{
		"list": result,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

type ProductRequest struct {
	CategoryID        *uint64     `json:"category_id"`
	Name              string      `json:"name" binding:"required"`
	Description       string      `json:"description"`
	Images            []string    `json:"images"`
	Price             float64     `json:"price" binding:"required"`
	OriginalPrice     float64     `json:"original_price"`
	Stock             uint        `json:"stock"`
	Unit              string      `json:"unit"`
	ProductType       uint8       `json:"product_type"`
	ServiceContent    interface{} `json:"service_content"`
	SaleType          uint8       `json:"sale_type"`
	RentalUnit        uint8       `json:"rental_unit"`
	RentalPrice       float64     `json:"rental_price"`
	Deposit           float64     `json:"deposit"`
	MaxRentalDuration uint        `json:"max_rental_duration"`
	Sort              uint        `json:"sort"`
	Sales             *uint       `json:"sales"`
	Specs             []struct {
		Name    string `json:"name" binding:"required"`
		Options []struct {
			Name  string  `json:"name" binding:"required"`
			Price float64 `json:"price"`
		} `json:"options"`
	} `json:"specs"`
}

// normalizeProductType 根据 product_type 和 sale_type 自动归一化，保持业务语义一致
// product_type: 1=辅具零售 2=辅具租赁 3=康养套餐 4=陪诊服务
// sale_type:    1=一口价 2=租赁
// 规则：辅具租赁固定 sale_type=2；其他类型固定 sale_type=1
func normalizeProductType(productType uint8, saleType uint8) (uint8, uint8) {
	if productType == 0 {
		// 未指定 product_type 时，按 sale_type 反推：sale_type=2 则为辅具租赁，否则默认辅具零售
		if saleType == 2 {
			productType = 2
		} else {
			productType = 1
		}
	}
	switch productType {
	case 2: // 辅具租赁
		saleType = 2
	default: // 辅具零售/康养套餐/陪诊
		saleType = 1
	}
	return productType, saleType
}

func CreateProduct(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 商品类型与销售类型归一化
	productType, saleType := normalizeProductType(req.ProductType, req.SaleType)

	// 租赁类型校验
	if saleType == 2 {
		if req.RentalUnit == 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁商品必须选择计费周期")
			return
		}
		if req.RentalPrice <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁商品单位租金必须大于0")
			return
		}
	}

	imagesJSON, _ := json.Marshal(req.Images)

	// service_content JSON 序列化
	var serviceContentJSON models.JSON
	if req.ServiceContent != nil {
		if raw, err := json.Marshal(req.ServiceContent); err == nil {
			serviceContentJSON = models.JSON(raw)
		}
	}

	product := models.Product{
		CategoryID:         req.CategoryID,
		Name:               req.Name,
		Description:        req.Description,
		Images:             models.JSON(imagesJSON),
		Price:              req.Price,
		OriginalPrice:      req.OriginalPrice,
		Stock:              req.Stock,
		Unit:               req.Unit,
		ProductType:        productType,
		ServiceContent:     serviceContentJSON,
		SaleType:           saleType,
		RentalUnit:         req.RentalUnit,
		RentalPrice:        req.RentalPrice,
		Deposit:            req.Deposit,
		MaxRentalDuration:  req.MaxRentalDuration,
		Sort:               req.Sort,
		Status:             1,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建商品失败")
		return
	}

	for _, spec := range req.Specs {
		optionsJSON, _ := json.Marshal(spec.Options)
		productSpec := models.ProductSpec{
			ProductID: product.ID,
			Name:      spec.Name,
			Options:   models.JSON(optionsJSON),
		}
		if err := tx.Create(&productSpec).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建规格失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建商品失败")
		return
	}

	productWithRelations, err := loadProductWithRelations(product.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "读取商品详情失败")
		return
	}

	response.Success(c, buildProductResponse(*productWithRelations))
}

func UpdateProduct(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 商品类型与销售类型归一化
	productType, saleType := normalizeProductType(req.ProductType, req.SaleType)

	// 租赁类型校验
	if saleType == 2 {
		if req.RentalUnit == 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁商品必须选择计费周期")
			return
		}
		if req.RentalPrice <= 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "租赁商品单位租金必须大于0")
			return
		}
	}

	var product models.Product
	if err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&product).Error; err != nil {
		respondProductQueryError(c, err, "商品不存在", "查询商品失败")
		return
	}

	imagesJSON, _ := json.Marshal(req.Images)

	// service_content JSON 序列化
	var serviceContentJSON models.JSON
	if req.ServiceContent != nil {
		if raw, err := json.Marshal(req.ServiceContent); err == nil {
			serviceContentJSON = models.JSON(raw)
		}
	}

	updates := map[string]interface{}{
		"category_id":         req.CategoryID,
		"name":                req.Name,
		"description":         req.Description,
		"images":              models.JSON(imagesJSON),
		"price":               req.Price,
		"original_price":      req.OriginalPrice,
		"stock":               req.Stock,
		"unit":                req.Unit,
		"product_type":        productType,
		"service_content":     serviceContentJSON,
		"sale_type":           saleType,
		"rental_unit":         req.RentalUnit,
		"rental_price":        req.RentalPrice,
		"deposit":             req.Deposit,
		"max_rental_duration": req.MaxRentalDuration,
		"sort":                req.Sort,
	}
	if req.Sales != nil {
		updates["sales"] = *req.Sales
	}

	tx := database.DB.Begin()
	if err := tx.Model(&product).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商品失败")
		return
	}

	if err := tx.Where("product_id = ?", id).Delete(&models.ProductSpec{}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "清理旧规格失败")
		return
	}

	for _, spec := range req.Specs {
		optionsJSON, _ := json.Marshal(spec.Options)
		productSpec := models.ProductSpec{
			ProductID: id,
			Name:      spec.Name,
			Options:   models.JSON(optionsJSON),
		}
		if err := tx.Create(&productSpec).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新规格失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商品失败")
		return
	}

	productWithRelations, err := loadProductWithRelations(id)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "读取商品详情失败")
		return
	}

	response.Success(c, buildProductResponse(*productWithRelations))
}

func ProductOnSale(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	result := database.DB.Model(&models.Product{}).Where("id = ? AND deleted_at IS NULL", id).Update("status", 1)
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "上架成功"})
}

func ProductOffSale(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	result := database.DB.Model(&models.Product{}).Where("id = ? AND deleted_at IS NULL", id).Update("status", 2)
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "下架成功"})
}

type BatchStatusRequest struct {
	ProductIDs []uint64 `json:"product_ids" binding:"required"`
	Status     uint8    `json:"status" binding:"required"`
}

func BatchUpdateProductStatus(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)

	var req BatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if err := database.DB.Model(&models.Product{}).Where("id IN ? AND deleted_at IS NULL", req.ProductIDs).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "批量更新状态失败")
		return
	}

	response.Success(c, gin.H{"message": "批量更新成功"})
}

func DeleteProduct(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	now := time.Now()
	result := database.DB.Model(&models.Product{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"status":     2,
		})
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除商品失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

type StockRequest struct {
	Stock uint `json:"stock" binding:"required"`
}

func UpdateStock(c *gin.Context) {
	_, _ = resolveTargetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	var req StockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	result := database.DB.Model(&models.Product{}).Where("id = ? AND deleted_at IS NULL", id).Update("stock", req.Stock)
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "库存更新成功"})
}
