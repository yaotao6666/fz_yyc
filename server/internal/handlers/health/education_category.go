package health

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 权限码占位（实际由路由中间件通过 sys_menus.permission 校验）
const (
	PermEduCategoryView   = "education-categories:view"
	PermEduCategoryCreate = "education-categories:create"
	PermEduCategoryUpdate = "education-categories:update"
	PermEduCategoryDelete = "education-categories:delete"
)

// ---- 后台分类 CRUD ----

type merchantCategoryReq struct {
	ParentID uint64 `json:"parent_id"`
	Name     string `json:"name" binding:"required,max=64"`
	Sort     int32  `json:"sort"`
	Status   uint8  `json:"status" binding:"oneof=0 1"`
}

// MerchantListEducationCategories 后台列表（全部，按 parent_id + sort 排序）
func MerchantListEducationCategories(c *gin.Context) {
	var list []models.HealthEducationCategory
	if err := database.DB.Order("parent_id ASC, sort ASC, id ASC").Find(&list).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询失败")
		return
	}
	response.Success(c, list)
}

// MerchantCreateEducationCategory 后台新建分类（支持两级：parent_id=0 为一级）
func MerchantCreateEducationCategory(c *gin.Context) {
	var req merchantCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if req.ParentID > 0 {
		var parent models.HealthEducationCategory
		if err := database.DB.First(&parent, req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父分类不存在")
				return
			}
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询失败")
			return
		}
		// 仅支持两级：若父分类已有父，则不再允许作为父
		if parent.ParentID != 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "仅支持两级分类，无法在二级分类下创建子分类")
			return
		}
	}
	cat := models.HealthEducationCategory{
		ParentID: req.ParentID,
		Name:     req.Name,
		Sort:     req.Sort,
		Status:   req.Status,
	}
	if err := database.DB.Create(&cat).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建失败")
		return
	}
	response.Success(c, cat)
}

// MerchantUpdateEducationCategory 后台更新分类
func MerchantUpdateEducationCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "分类ID错误")
		return
	}
	var req merchantCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	var cat models.HealthEducationCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "分类不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询失败")
		return
	}
	// 调整父分类校验：不能把自己挂到自己或自己的子下
	if req.ParentID > 0 {
		if req.ParentID == cat.ID {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父分类不能指向自身")
			return
		}
		var parent models.HealthEducationCategory
		if err := database.DB.First(&parent, req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusBadRequest, response.CodeParamError, "父分类不存在")
				return
			}
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询失败")
			return
		}
		if parent.ParentID != 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "仅支持两级分类，无法挂到二级分类下")
			return
		}
	}
	updates := map[string]any{
		"parent_id":  req.ParentID,
		"name":       req.Name,
		"sort":       req.Sort,
		"status":     req.Status,
		"updated_at": time.Now(),
	}
	if err := database.DB.Model(&cat).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新失败")
		return
	}
	response.Success(c, nil)
}

// MerchantDeleteEducationCategory 后台删除分类（有子分类或有关联文章时拒绝）
func MerchantDeleteEducationCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "分类ID错误")
		return
	}
	var count int64
	// 有子分类
	database.DB.Model(&models.HealthEducationCategory{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "存在子分类，无法删除")
		return
	}
	// 有关联文章
	database.DB.Model(&models.HealthEducationArticle{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该分类下存在文章，无法删除")
		return
	}
	res := database.DB.Delete(&models.HealthEducationCategory{}, id)
	if res.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除失败")
		return
	}
	if res.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "分类不存在")
		return
	}
	response.Success(c, nil)
}

// ---- C 端公开接口 ----

// StoreListEducationCategories C 端：启用中的一级+二级分类列表（按 sort 升序）
func StoreListEducationCategories(c *gin.Context) {
	var list []models.HealthEducationCategory
	database.DB.Where("status = ?", 1).Order("parent_id ASC, sort ASC, id ASC").Find(&list)
	response.Success(c, list)
}

// StoreListEducationArticles C 端：按分类查已发布文章（category_id 可选，空=全部已发布）
func StoreListEducationArticles(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	keyword := c.Query("keyword")

	db := database.DB.Model(&models.HealthEducationArticle{}).Where("status = ?", 1)
	if categoryIDStr != "" {
		if catID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil && catID > 0 {
			db = db.Where("category_id = ?", catID)
		}
	}
	if keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	var total int64
	db.Count(&total)

	var list []models.HealthEducationArticle
	db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	response.Success(c, gin.H{"list": list, "total": total})
}
