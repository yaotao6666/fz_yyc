package merchant

import (
	"encoding/json"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var categories []models.Category
	if err := database.DB.Where("merchant_id = ?", merchantID).Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取分类列表失败")
		return
	}

	var result []map[string]interface{}
	for _, cat := range categories {
		var count int64
		database.DB.Model(&models.Product{}).Where("category_id = ? AND merchant_id = ?", cat.ID, merchantID).Count(&count)
		
		catMap := map[string]interface{}{
			"id":             cat.ID,
			"merchant_id":    cat.MerchantID,
			"name":           cat.Name,
			"sort":           cat.Sort,
			"status":         cat.Status,
			"created_at":     cat.CreatedAt,
			"updated_at":     cat.UpdatedAt,
			"product_count":  count,
		}
		result = append(result, catMap)
	}

	response.Success(c, result)
}

type CategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	Sort   *uint  `json:"sort"`
	Status uint8  `json:"status"`
}

func CreateCategory(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	category := models.Category{
		MerchantID: merchantID,
		Name:       req.Name,
		Sort:       0,
		Status:     1,
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

func UpdateCategory(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	categoryID := c.Param("category_id")
	id, _ := strconv.ParseUint(categoryID, 10, 64)

	var req CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var category models.Category
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&category).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "分类不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	if err := database.DB.Model(&category).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新分类失败")
		return
	}

	database.DB.First(&category, id)
	response.Success(c, category)
}

func DeleteCategory(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	categoryID := c.Param("category_id")
	id, _ := strconv.ParseUint(categoryID, 10, 64)

	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).Delete(&models.Category{}).Error; err != nil {
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
	merchantID := middleware.GetMerchantID(c)

	var req SortCategoriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	tx := database.DB.Begin()
	for _, item := range req.Categories {
		if err := tx.Model(&models.Category{}).Where("id = ? AND merchant_id = ?", item.ID, merchantID).Update("sort", item.Sort).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "排序失败")
			return
		}
	}
	tx.Commit()

	response.Success(c, gin.H{"message": "排序成功"})
}

func GetProduct(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	var product models.Product
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).Preload("Category").Preload("Specs").First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, product)
}

func GetProducts(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	categoryID := c.Query("category_id")
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.DB.Model(&models.Product{}).Where("merchant_id = ?", merchantID)

	if categoryID != "" {
		id, _ := strconv.ParseUint(categoryID, 10, 64)
		query = query.Where("category_id = ?", id)
	}
	if status != "" {
		statusInt, _ := strconv.Atoi(status)
		query = query.Where("status = ?", statusInt)
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

	response.Success(c, gin.H{
		"list": products,
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"page_size": pageSize,
		},
	})
}

type ProductRequest struct {
	CategoryID    *uint64  `json:"category_id"`
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	Images        []string `json:"images"`
	Price         float64  `json:"price" binding:"required"`
	OriginalPrice float64  `json:"original_price"`
	Stock         uint     `json:"stock"`
	Unit          string   `json:"unit"`
	Sort          uint     `json:"sort"`
	Specs         []struct {
		Name    string `json:"name" binding:"required"`
		Options []struct {
			Name  string  `json:"name" binding:"required"`
			Price float64 `json:"price"`
		} `json:"options"`
	} `json:"specs"`
}

func CreateProduct(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	imagesJSON, _ := json.Marshal(req.Images)

	product := models.Product{
		MerchantID:    merchantID,
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Description:   req.Description,
		Images:        models.JSON(imagesJSON),
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Unit:          req.Unit,
		Sort:          req.Sort,
		Status:        1,
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

	tx.Commit()

	database.DB.Preload("Category").Preload("Specs").First(&product, product.ID)
	response.Success(c, product)
}

func UpdateProduct(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	var req ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var product models.Product
	if err := database.DB.Where("id = ? AND merchant_id = ?", id, merchantID).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	imagesJSON, _ := json.Marshal(req.Images)

	updates := map[string]interface{}{
		"category_id":     req.CategoryID,
		"name":            req.Name,
		"description":     req.Description,
		"images":          models.JSON(imagesJSON),
		"price":           req.Price,
		"original_price":  req.OriginalPrice,
		"stock":           req.Stock,
		"unit":            req.Unit,
		"sort":            req.Sort,
	}

	tx := database.DB.Begin()
	if err := tx.Model(&product).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新商品失败")
		return
	}

	if len(req.Specs) > 0 {
		tx.Model(&models.ProductSpec{}).Where("product_id = ?", id).Delete("")
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
	}

	tx.Commit()

	database.DB.Preload("Category").Preload("Specs").First(&product, id)
	response.Success(c, product)
}

func ProductOnSale(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	result := database.DB.Model(&models.Product{}).Where("id = ? AND merchant_id = ?", id, merchantID).Update("status", 1)
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "上架成功"})
}

func ProductOffSale(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	result := database.DB.Model(&models.Product{}).Where("id = ? AND merchant_id = ?", id, merchantID).Update("status", 2)
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
	merchantID := middleware.GetMerchantID(c)

	var req BatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	if err := database.DB.Model(&models.Product{}).Where("id IN ? AND merchant_id = ?", req.ProductIDs, merchantID).Update("status", req.Status).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "批量更新状态失败")
		return
	}

	response.Success(c, gin.H{"message": "批量更新成功"})
}

func DeleteProduct(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	tx := database.DB.Begin()
	if err := tx.Where("id = ? AND merchant_id = ?", id, merchantID).Delete(&models.Product{}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除商品失败")
		return
	}

	tx.Where("product_id = ?", id).Delete(&models.ProductSpec{})
	tx.Commit()

	response.Success(c, gin.H{"message": "删除成功"})
}

type StockRequest struct {
	Stock uint `json:"stock" binding:"required"`
}

func UpdateStock(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID := c.Param("product_id")
	id, _ := strconv.ParseUint(productID, 10, 64)

	var req StockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	result := database.DB.Model(&models.Product{}).Where("id = ? AND merchant_id = ?", id, merchantID).Update("stock", req.Stock)
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	response.Success(c, gin.H{"message": "库存更新成功"})
}
