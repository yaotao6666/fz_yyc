package merchant

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type prdSpecResponse struct {
	ID     uint64   `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func GetProductSpecs(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)

	var product models.Product
	if err := database.DB.Select("id").Where("id = ? AND merchant_id = ? AND deleted_at IS NULL", productID, merchantID).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	var specs []models.ProductSpec
	if err := database.DB.Where("product_id = ?", product.ID).Order("id ASC").Find(&specs).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取规格失败")
		return
	}

	result := make([]prdSpecResponse, 0, len(specs))
	for _, spec := range specs {
		options := parseSpecOptions(spec.Options)
		values := make([]string, 0, len(options))
		for _, opt := range options {
			if opt.Name != "" {
				values = append(values, opt.Name)
			}
		}
		result = append(result, prdSpecResponse{
			ID:     spec.ID,
			Name:   spec.Name,
			Values: values,
		})
	}

	response.Success(c, gin.H{
		"specs": result,
		"skus":  []any{},
	})
}

type prdUpdateProductSpecsRequest struct {
	Specs []struct {
		ID     uint64   `json:"id"`
		Name   string   `json:"name" binding:"required"`
		Values []string `json:"values"`
	} `json:"specs" binding:"required"`
	Skus []any `json:"skus"`
}

func UpdateProductSpecs(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)

	var req prdUpdateProductSpecsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var product models.Product
	if err := database.DB.Select("id").Where("id = ? AND merchant_id = ? AND deleted_at IS NULL", productID, merchantID).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	tx := database.DB.Begin()
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductSpec{}).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存失败")
		return
	}

	for _, spec := range req.Specs {
		opts := make([]ProductSpecOptionResponse, 0, len(spec.Values))
		for _, value := range spec.Values {
			if value == "" {
				continue
			}
			opts = append(opts, ProductSpecOptionResponse{Name: value})
		}
		raw, _ := json.Marshal(opts)
		productSpec := models.ProductSpec{
			ProductID: product.ID,
			Name:      spec.Name,
			Options:   models.JSON(raw),
		}
		if err := tx.Create(&productSpec).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "保存失败")
		return
	}

	response.SuccessWithMessage(c, "保存成功", nil)
}

type prdDeleteProductSpecsRequest struct {
	SpecIDs []uint64 `json:"spec_ids" binding:"required"`
}

func DeleteProductSpecs(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)

	var req prdDeleteProductSpecsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if len(req.SpecIDs) == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "spec_ids不能为空")
		return
	}

	var product models.Product
	if err := database.DB.Select("id").Where("id = ? AND merchant_id = ? AND deleted_at IS NULL", productID, merchantID).First(&product).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "商品不存在")
		return
	}

	if err := database.DB.Where("product_id = ? AND id IN ?", product.ID, req.SpecIDs).Delete(&models.ProductSpec{}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除失败")
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
