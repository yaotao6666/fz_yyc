package user

import (
	"net/http"
	"strconv"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type addressRequest struct {
	Name      string   `json:"name" binding:"required"`
	Phone     string   `json:"phone" binding:"required"`
	Province  string   `json:"province"`
	City      string   `json:"city"`
	District  string   `json:"district"`
	Address   string   `json:"address" binding:"required"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	IsDefault *bool    `json:"is_default"`
}

func GetAddresses(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var addresses []models.UserAddress
	if err := database.DB.Where("user_id = ?", userID).Order("is_default DESC, id DESC").Find(&addresses).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取收货地址失败")
		return
	}
	response.Success(c, addresses)
}

func CreateAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	address := models.UserAddress{
		UserID:    userID,
		Name:      req.Name,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Address:   req.Address,
		IsDefault: req.IsDefault != nil && *req.IsDefault,
	}
	if req.Lat != nil {
		address.Lat = *req.Lat
	}
	if req.Lng != nil {
		address.Lng = *req.Lng
	}

	tx := database.DB.Begin()
	if address.IsDefault {
		if err := tx.Model(&models.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认地址失败")
			return
		}
	}
	if err := tx.Create(&address).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建收货地址失败")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建收货地址失败")
		return
	}

	response.Success(c, address)
}

func UpdateAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	addressID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var address models.UserAddress
	if err := database.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "地址不存在")
		return
	}

	updates := map[string]interface{}{
		"name":     req.Name,
		"phone":    req.Phone,
		"province": req.Province,
		"city":     req.City,
		"district": req.District,
		"address":  req.Address,
	}
	if req.Lat != nil {
		updates["lat"] = *req.Lat
	}
	if req.Lng != nil {
		updates["lng"] = *req.Lng
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	tx := database.DB.Begin()
	if req.IsDefault != nil && *req.IsDefault {
		if err := tx.Model(&models.UserAddress{}).Where("user_id = ? AND id <> ?", userID, address.ID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认地址失败")
			return
		}
	}
	if err := tx.Model(&address).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新收货地址失败")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新收货地址失败")
		return
	}

	var updated models.UserAddress
	database.DB.Where("id = ?", address.ID).First(&updated)
	response.Success(c, updated)
}

func DeleteAddress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	addressID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if err := database.DB.Where("id = ? AND user_id = ?", addressID, userID).Delete(&models.UserAddress{}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除收货地址失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
