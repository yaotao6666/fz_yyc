package merchant

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// parseRecommendImagesJSON 解析商品图片 JSON 数组
func parseRecommendImagesJSON(raw models.JSON) []string {
	var images []string
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &images)
	}
	return images
}

// ============================================================
// 小程序首页推荐管理（商家端 CRUD + 启停）
// ============================================================

type HomeRecommendCreateRequest struct {
	ProductID  uint64  `json:"product_id" binding:"required"`
	TargetType uint8   `json:"target_type" binding:"required,oneof=1 2"`
	Title      *string `json:"title"`
	Sort       uint    `json:"sort"`
	Status     *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type HomeRecommendUpdateRequest struct {
	ProductID  *uint64 `json:"product_id"`
	TargetType *uint8  `json:"target_type" binding:"omitempty,oneof=1 2"`
	Title      *string `json:"title"`
	Sort       *uint   `json:"sort"`
	Status     *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// productTypeMatchesTarget 校验目标商品/服务的类型是否与推荐对象类型匹配
func productTypeMatchesTarget(productType uint8, targetType uint8) bool {
	switch targetType {
	case 1:
		return productType == 1 || productType == 2
	case 2:
		return productType == 3 || productType == 4
	}
	return false
}

// validateRecommendProduct 校验推荐目标商品/服务存在、上架且类型匹配
func validateRecommendProduct(productID uint64, targetType uint8) (*models.Product, error) {
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("目标商品/服务不存在")
	}
	if product.Status != 1 {
		return nil, errors.New("目标商品/服务未上架")
	}
	if !productTypeMatchesTarget(product.ProductType, targetType) {
		return nil, errors.New("目标商品/服务类型与推荐对象类型不匹配")
	}
	return &product, nil
}

// buildHomeRecommendResponse 构造推荐项响应（附商品名与封面）
func buildHomeRecommendResponse(r models.HomeRecommend) map[string]interface{} {
	qiniuService := qiniu.GetService()
	var productName string
	var productPrice float64
	var productImage string
	if r.Product != nil {
		productName = r.Product.Name
		productPrice = r.Product.Price
		images := parseRecommendImagesJSON(r.Product.Images)
		if len(images) > 0 {
			image := images[0]
			if qiniuService != nil {
				image = qiniuService.BuildPrivateURL(image)
			}
			productImage = image
		}
	}
	return map[string]interface{}{
		"id":            r.ID,
		"merchant_id":   r.MerchantID,
		"product_id":    r.ProductID,
		"target_type":   r.TargetType,
		"title":         r.Title,
		"sort":          r.Sort,
		"status":        r.Status,
		"product_name":  productName,
		"product_price": productPrice,
		"product_image": productImage,
		"created_at":    r.CreatedAt,
		"updated_at":    r.UpdatedAt,
	}
}

// GetHomeRecommends 查询商家首页推荐列表（全部，含启用/禁用）
func GetHomeRecommends(c *gin.Context) {
	var recommends []models.HomeRecommend
	err := database.DB.
		Preload("Product").
		Where("merchant_id = ?", utils.DefaultMerchantID).
		Order("sort ASC, id DESC").
		Find(&recommends).Error
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询推荐失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(recommends))
	for _, r := range recommends {
		result = append(result, buildHomeRecommendResponse(r))
	}
	response.Success(c, gin.H{"list": result})
}

// CreateHomeRecommend 新增首页推荐
func CreateHomeRecommend(c *gin.Context) {
	var req HomeRecommendCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	product, err := validateRecommendProduct(req.ProductID, req.TargetType)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	r := models.HomeRecommend{
		MerchantID: utils.DefaultMerchantID,
		ProductID:  req.ProductID,
		TargetType: req.TargetType,
		Sort:       req.Sort,
		Status:     1,
	}
	if req.Title != nil {
		r.Title = *req.Title
	}
	if req.Status != nil {
		r.Status = *req.Status
	}
	if r.Title == "" {
		r.Title = product.Name
	}

	if err := database.DB.Create(&r).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建推荐失败")
		return
	}
	response.Success(c, buildHomeRecommendResponse(r))
}

// UpdateHomeRecommend 更新首页推荐
func UpdateHomeRecommend(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var r models.HomeRecommend
	if err := database.DB.
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		First(&r).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "推荐不存在")
		return
	}

	var req HomeRecommendUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	// 若变更了目标商品/类型，需重新校验
	needValidate := (req.ProductID != nil && *req.ProductID != r.ProductID) ||
		(req.TargetType != nil && *req.TargetType != r.TargetType)
	targetProductID := r.ProductID
	targetType := r.TargetType
	if req.ProductID != nil {
		targetProductID = *req.ProductID
	}
	if req.TargetType != nil {
		targetType = *req.TargetType
	}
	if needValidate {
		product, verr := validateRecommendProduct(targetProductID, targetType)
		if verr != nil {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, verr.Error())
			return
		}
		r.Product = product
	}

	if req.ProductID != nil {
		r.ProductID = *req.ProductID
	}
	if req.TargetType != nil {
		r.TargetType = *req.TargetType
	}
	if req.Title != nil {
		r.Title = *req.Title
	}
	if req.Sort != nil {
		r.Sort = *req.Sort
	}
	if req.Status != nil {
		r.Status = *req.Status
	}

	// 标题置空时回退用商品名
	if r.Title == "" && r.Product != nil {
		r.Title = r.Product.Name
	}

	if err := database.DB.Save(&r).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新推荐失败")
		return
	}
	response.Success(c, buildHomeRecommendResponse(r))
}

// UpdateHomeRecommendStatus 启用/禁用首页推荐
func UpdateHomeRecommendStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var req struct {
		Status *uint8 `json:"status" binding:"required,oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result := database.DB.
		Model(&models.HomeRecommend{}).
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Update("status", *req.Status)
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新状态失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "推荐不存在")
		return
	}
	response.Success(c, nil)
}

// DeleteHomeRecommend 删除首页推荐
func DeleteHomeRecommend(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	result := database.DB.
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Delete(&models.HomeRecommend{})
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "推荐不存在")
		return
	}
	response.Success(c, nil)
}