package merchant

import (
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/qiniu"
	"fz_yyc_api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 小程序轮播图管理（商家端 CRUD）
// ============================================================

type BannerCreateRequest struct {
	Title     string `json:"title"`
	Image     string `json:"image" binding:"required"`
	LinkType  string `json:"link_type" binding:"required,oneof=none product category url"`
	LinkValue string `json:"link_value"`
	Sort      uint   `json:"sort"`
	Status    uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

type BannerUpdateRequest struct {
	Title     *string `json:"title"`
	Image     *string `json:"image"`
	LinkType  *string `json:"link_type" binding:"omitempty,oneof=none product category url"`
	LinkValue *string `json:"link_value"`
	Sort      *uint   `json:"sort"`
	Status    *uint8  `json:"status" binding:"omitempty,oneof=0 1"`
}

func buildBannerAccessibleURL(banner models.MiniProgramBanner) map[string]interface{} {
	qiniuService := qiniu.GetService()
	image := banner.Image
	if qiniuService != nil {
		image = qiniuService.BuildPrivateURL(image)
	}
	return map[string]interface{}{
		"id":         banner.ID,
		"merchant_id": banner.MerchantID,
		"title":      banner.Title,
		"image":      image,
		"link_type":  banner.LinkType,
		"link_value": banner.LinkValue,
		"sort":       banner.Sort,
		"status":     banner.Status,
		"created_at": banner.CreatedAt,
		"updated_at": banner.UpdatedAt,
	}
}

// GetBanners 查询商家轮播图列表（全部，含启用/禁用）
func GetBanners(c *gin.Context) {
	var banners []models.MiniProgramBanner
	err := database.DB.
		Where("merchant_id = ?", utils.DefaultMerchantID).
		Order("sort ASC, id DESC").
		Find(&banners).Error
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询轮播图失败")
		return
	}

	result := make([]map[string]interface{}, 0, len(banners))
	for _, b := range banners {
		result = append(result, buildBannerAccessibleURL(b))
	}
	response.Success(c, gin.H{"list": result})
}

// GetBanner 查询单个轮播图详情
func GetBanner(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var banner models.MiniProgramBanner
	if err := database.DB.
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		First(&banner).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "轮播图不存在")
		return
	}
	response.Success(c, buildBannerAccessibleURL(banner))
}

// CreateBanner 新增轮播图
func CreateBanner(c *gin.Context) {
	var req BannerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	banner := models.MiniProgramBanner{
		MerchantID: utils.DefaultMerchantID,
		Title:      req.Title,
		Image:      req.Image,
		LinkType:   req.LinkType,
		LinkValue:  req.LinkValue,
		Sort:       req.Sort,
		Status:     1,
	}
	if req.Status != 0 {
		banner.Status = req.Status
	}

	if err := database.DB.Create(&banner).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建轮播图失败")
		return
	}
	response.Success(c, buildBannerAccessibleURL(banner))
}

// UpdateBanner 更新轮播图
func UpdateBanner(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	var banner models.MiniProgramBanner
	if err := database.DB.
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		First(&banner).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "轮播图不存在")
		return
	}

	var req BannerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if req.Title != nil {
		banner.Title = *req.Title
	}
	if req.Image != nil {
		banner.Image = *req.Image
	}
	if req.LinkType != nil {
		banner.LinkType = *req.LinkType
	}
	if req.LinkValue != nil {
		banner.LinkValue = *req.LinkValue
	}
	if req.Sort != nil {
		banner.Sort = *req.Sort
	}
	if req.Status != nil {
		banner.Status = *req.Status
	}

	if err := database.DB.Save(&banner).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新轮播图失败")
		return
	}
	response.Success(c, buildBannerAccessibleURL(banner))
}

// UpdateBannerStatus 启用/禁用轮播图（专用接口，避免全量更新的必填校验）
func UpdateBannerStatus(c *gin.Context) {
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
		Model(&models.MiniProgramBanner{}).
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Update("status", *req.Status)
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新状态失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "轮播图不存在")
		return
	}
	response.Success(c, nil)
}

// DeleteBanner 删除轮播图
func DeleteBanner(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数无效")
		return
	}

	result := database.DB.
		Where("id = ? AND merchant_id = ?", id, utils.DefaultMerchantID).
		Delete(&models.MiniProgramBanner{})
	if result.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除失败")
		return
	}
	if result.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "轮播图不存在")
		return
	}
	response.Success(c, nil)
}
