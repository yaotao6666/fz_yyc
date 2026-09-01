package health

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---------- C端 ----------

// UserListEducationArticles C端用户：启用的宣教文章列表（按 category_id/关键词筛选）
func UserListEducationArticles(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	// 兼容旧字符串 category 参数：category 存在且 category_id 未传时，按字符串匹配
	category := c.Query("category")
	keyword := c.Query("keyword")

	db := database.DB.Model(&models.HealthEducationArticle{}).
		Where("status = ?", utils.EducationStatusPublished)
	if categoryIDStr != "" {
		if catID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil && catID > 0 {
			db = db.Where("category_id = ?", catID)
		}
	} else if category != "" {
		db = db.Where("category = ?", category)
	}
	if keyword != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var list []models.HealthEducationArticle
	db.Order("id DESC").Find(&list)

	response.Success(c, list)
}

// UserGetEducationArticle C端用户：文章详情（含阅读量+1）
func UserGetEducationArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}

	var article models.HealthEducationArticle
	if err := database.DB.Where("id = ? AND status = ?", id, utils.EducationStatusPublished).First(&article).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}

	database.DB.Model(&article).UpdateColumn("views", article.Views+1)
	article.Views++

	response.Success(c, article)
}

// ---------- 服务人员端 ----------

// StaffListEducationArticles 服务人员端：启用的宣教文章列表
func StaffListEducationArticles(c *gin.Context) {
	var list []models.HealthEducationArticle
	database.DB.Model(&models.HealthEducationArticle{}).
		Where("status = ?", utils.EducationStatusPublished).
		Order("id DESC").
		Find(&list)
	response.Success(c, list)
}

// ---------- Merchant 后台 ----------

// MerchantListEducationArticles 后台：宣教文章列表（全部状态），支持按 category_id/关键词/状态筛选
func MerchantListEducationArticles(c *gin.Context) {
	statusStr := c.Query("status")
	categoryIDStr := c.Query("category_id")
	category := c.Query("category")
	keyword := c.Query("keyword")

	db := database.DB.Model(&models.HealthEducationArticle{})
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			db = db.Where("status = ?", uint8(s))
		}
	}
	if categoryIDStr != "" {
		if catID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil && catID > 0 {
			db = db.Where("category_id = ?", catID)
		}
	} else if category != "" {
		db = db.Where("category = ?", category)
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
	db.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list)

	response.Success(c, gin.H{"list": list, "total": total})
}

// MerchantCreateEducationArticleRequest 后台创建/更新文章请求
type MerchantCreateEducationArticleRequest struct {
	Title      string   `json:"title" binding:"required,max=128"`
	CategoryID uint64   `json:"category_id"` // 必填（发布时）；草稿允许为 0
	Category   string   `json:"category" binding:"max=32"`
	Cover      string   `json:"cover" binding:"max=512"`
	Content    string   `json:"content" binding:"required"`
	Tags       []string `json:"tags"`
	Status     uint8    `json:"status" binding:"oneof=0 1"`
}

// MerchantCreateEducationArticle 后台：创建文章
func MerchantCreateEducationArticle(c *gin.Context) {
	var req MerchantCreateEducationArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	// 发布状态必须指定有效分类
	if req.Status == utils.EducationStatusPublished && req.CategoryID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "发布前必须选择宣教分类")
		return
	}
	if req.CategoryID > 0 {
		var catCount int64
		database.DB.Model(&models.HealthEducationCategory{}).Where("id = ?", req.CategoryID).Count(&catCount)
		if catCount == 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "所选宣教分类不存在")
			return
		}
	}

	article := models.HealthEducationArticle{
		Title:      req.Title,
		CategoryID: req.CategoryID,
		Category:   req.Category,
		Cover:      req.Cover,
		Content:    req.Content,
		Status:     req.Status,
	}
	if len(req.Tags) > 0 {
		if raw, err := jsonMarshal(req.Tags); err == nil {
			article.Tags = models.JSON(raw)
		}
	}
	if req.Status == utils.EducationStatusPublished {
		now := time.Now()
		article.PublishAt = &now
	}

	if err := database.DB.Create(&article).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建失败")
		return
	}
	response.Success(c, article)
}

// MerchantUpdateEducationArticle 后台：更新文章
func MerchantUpdateEducationArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}

	var req MerchantCreateEducationArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if req.Status == utils.EducationStatusPublished && req.CategoryID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "发布前必须选择宣教分类")
		return
	}
	if req.CategoryID > 0 {
		var catCount int64
		database.DB.Model(&models.HealthEducationCategory{}).Where("id = ?", req.CategoryID).Count(&catCount)
		if catCount == 0 {
			response.Fail(c, http.StatusBadRequest, response.CodeParamError, "所选宣教分类不存在")
			return
		}
	}

	var article models.HealthEducationArticle
	if err := database.DB.First(&article, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}

	updates := map[string]any{
		"title":       req.Title,
		"category_id": req.CategoryID,
		"category":    req.Category,
		"cover":       req.Cover,
		"content":     req.Content,
		"status":      req.Status,
	}
	if len(req.Tags) > 0 {
		if raw, err := jsonMarshal(req.Tags); err == nil {
			updates["tags"] = models.JSON(raw)
		}
	} else {
		updates["tags"] = models.JSON("null")
	}
	if req.Status == utils.EducationStatusPublished && article.PublishAt == nil {
		now := time.Now()
		updates["publish_at"] = &now
	}

	if err := database.DB.Model(&article).Updates(updates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新失败")
		return
	}
	response.Success(c, nil)
}

// MerchantDeleteEducationArticle 后台：删除文章
func MerchantDeleteEducationArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "文章ID错误")
		return
	}

	res := database.DB.Delete(&models.HealthEducationArticle{}, id)
	if res.Error != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除失败")
		return
	}
	if res.RowsAffected == 0 {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "文章不存在")
		return
	}
	response.Success(c, nil)
}

// jsonMarshal 对任意值序列化为 json 字节切片（复用 JSON 字段写入）
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}