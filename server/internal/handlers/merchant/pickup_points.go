package merchant

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type pickupPointRequest struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	IsDefault *bool    `json:"is_default"`
	Status    *uint8   `json:"status"`
	Sort      *uint    `json:"sort"`
}

func GetPickupPoints(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var points []models.MerchantPickupPoint
	if err := database.DB.Where("merchant_id = ?", merchantID).
		Order("is_default DESC, sort ASC, id DESC").
		Find(&points).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取自提点失败")
		return
	}

	response.Success(c, points)
}

func CreatePickupPoint(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req pickupPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	point, err := buildPickupPointFromRequest(merchantID, 0, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	tx := database.DB.Begin()
	if point.IsDefault {
		if err := tx.Model(&models.MerchantPickupPoint{}).
			Where("merchant_id = ?", merchantID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认自提点失败")
			return
		}
	}

	if err := tx.Create(&point).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建自提点失败")
		return
	}

	if err := ensurePickupPointHasDefault(tx, merchantID); err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建自提点失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建自提点失败")
		return
	}

	response.Success(c, point)
}

func UpdatePickupPoint(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	pointID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if pointID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "自提点不存在")
		return
	}

	var existing models.MerchantPickupPoint
	if err := database.DB.Where("id = ? AND merchant_id = ?", pointID, merchantID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "自提点不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询自提点失败")
		return
	}

	var req pickupPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	point, err := buildPickupPointFromRequest(merchantID, existing.ID, req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	updates := map[string]any{
		"name":    point.Name,
		"address": point.Address,
		"lat":     point.Lat,
		"lng":     point.Lng,
		"status":  point.Status,
		"sort":    point.Sort,
	}
	if req.IsDefault != nil {
		updates["is_default"] = point.IsDefault
	}

	tx := database.DB.Begin()
	if req.IsDefault != nil && point.IsDefault {
		if err := tx.Model(&models.MerchantPickupPoint{}).
			Where("merchant_id = ? AND id <> ?", merchantID, pointID).
			Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认自提点失败")
			return
		}
	}

	if err := tx.Model(&models.MerchantPickupPoint{}).
		Where("id = ? AND merchant_id = ?", pointID, merchantID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新自提点失败")
		return
	}

	if err := ensurePickupPointHasDefault(tx, merchantID); err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新自提点失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新自提点失败")
		return
	}

	database.DB.Where("id = ?", pointID).First(&existing)
	response.Success(c, existing)
}

func DeletePickupPoint(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	pointID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if pointID == 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "自提点不存在")
		return
	}

	var existing models.MerchantPickupPoint
	if err := database.DB.Where("id = ? AND merchant_id = ?", pointID, merchantID).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Fail(c, http.StatusNotFound, response.CodeNotFound, "自提点不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "查询自提点失败")
		return
	}

	tx := database.DB.Begin()
	if err := tx.Delete(&models.MerchantPickupPoint{}, "id = ? AND merchant_id = ?", pointID, merchantID).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除自提点失败")
		return
	}

	if existing.IsDefault {
		if err := ensurePickupPointHasDefault(tx, merchantID); err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除自提点失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除自提点失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

func buildPickupPointFromRequest(merchantID uint64, existingID uint64, req pickupPointRequest) (models.MerchantPickupPoint, error) {
	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	if name == "" {
		return models.MerchantPickupPoint{}, fmt.Errorf("请输入自提点名称")
	}
	if address == "" {
		return models.MerchantPickupPoint{}, fmt.Errorf("请输入自提点地址")
	}
	if req.Lat == nil || req.Lng == nil {
		return models.MerchantPickupPoint{}, fmt.Errorf("请选择自提点位置")
	}
	lat := *req.Lat
	lng := *req.Lng
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return models.MerchantPickupPoint{}, fmt.Errorf("自提点经纬度不正确")
	}

	point := models.MerchantPickupPoint{
		ID:         existingID,
		MerchantID: merchantID,
		Name:       name,
		Address:    address,
		Lat:        lat,
		Lng:        lng,
		IsDefault:  false,
		Status:     1,
		Sort:       0,
	}

	if req.IsDefault != nil {
		point.IsDefault = *req.IsDefault
	}
	if req.Status != nil {
		point.Status = *req.Status
	}
	if req.Sort != nil {
		point.Sort = *req.Sort
	}

	return point, nil
}

func ensurePickupPointHasDefault(tx *gorm.DB, merchantID uint64) error {
	var count int64
	if err := tx.Model(&models.MerchantPickupPoint{}).
		Where("merchant_id = ? AND is_default = ?", merchantID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var point models.MerchantPickupPoint
	if err := tx.Where("merchant_id = ? AND status = 1", merchantID).
		Order("sort ASC, id DESC").
		First(&point).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	return tx.Model(&models.MerchantPickupPoint{}).
		Where("id = ? AND merchant_id = ?", point.ID, merchantID).
		Update("is_default", true).Error
}
