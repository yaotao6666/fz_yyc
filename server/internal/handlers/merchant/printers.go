package merchant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
)

type createPrinterRequest struct {
	Name       string          `json:"name" binding:"required"`
	Type       string          `json:"type" binding:"required"`
	DeviceNo   string          `json:"device_no" binding:"required"`
	APIKey     string          `json:"api_key"`
	APIURL     string          `json:"api_url"`
	FeieUser   string          `json:"feie_user"`
	FeieUKey   string          `json:"feie_ukey"`
	FeieSN     string          `json:"feie_sn"`
	PrintTypes json.RawMessage `json:"print_types"`
	Status     *uint8          `json:"status"`
	AutoPrint  *bool           `json:"auto_print"`
	IsDefault  *bool           `json:"is_default"`
}

type printerResponse struct {
	ID           uint64      `json:"id"`
	MerchantID   uint64      `json:"merchant_id"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DeviceNo     string      `json:"device_no"`
	APIURL       string      `json:"api_url"`
	FeieUser     string      `json:"feie_user"`
	FeieSN       string      `json:"feie_sn"`
	PrintTypes   models.JSON `json:"print_types"`
	Status       uint8       `json:"status"`
	AutoPrint    bool        `json:"auto_print"`
	IsDefault    bool        `json:"is_default"`
	PrintCount   int         `json:"print_count"`
	LastPrintAt  *time.Time  `json:"last_print_at"`
	HasAPIKey    bool        `json:"has_api_key"`
	HasFeieUKey  bool        `json:"has_feie_ukey"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func GetPrinters(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var printers []models.CloudPrinter
	if err := database.DB.Where("merchant_id = ?", merchantID).Order("id DESC").Find(&printers).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取云打印机列表失败")
		return
	}

	list := make([]printerResponse, 0, len(printers))
	for _, printer := range printers {
		list = append(list, buildPrinterResponse(printer))
	}

	response.Success(c, list)
}

func CreatePrinter(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)

	var req createPrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}
	if err := validatePrinterPayload(req.Type, req.DeviceNo, req.FeieUser, req.FeieUKey, req.FeieSN); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	printer := models.CloudPrinter{
		MerchantID: merchantID,
		Name:       req.Name,
		Brand:      req.Type,
		DeviceNo:   req.DeviceNo,
		APIKey:     req.APIKey,
		APIURL:     req.APIURL,
		FeieUser:   req.FeieUser,
		FeieUKey:   req.FeieUKey,
		FeieSN:     req.FeieSN,
		Status:     1,
	}
	if len(req.PrintTypes) > 0 {
		printer.PrintTypes = models.JSON(req.PrintTypes)
	}
	if req.Status != nil {
		printer.Status = *req.Status
	}
	if req.AutoPrint != nil {
		printer.AutoPrint = *req.AutoPrint
	}
	if req.IsDefault != nil {
		printer.IsDefault = *req.IsDefault
	}

	tx := database.DB.Begin()
	if printer.IsDefault {
		if err := tx.Model(&models.CloudPrinter{}).Where("merchant_id = ?", merchantID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认打印机失败")
			return
		}
	}
	if err := tx.Create(&printer).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建云打印机失败")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建云打印机失败")
		return
	}

	response.Success(c, buildPrinterResponse(printer))
}

type updatePrinterRequest struct {
	Name       *string         `json:"name"`
	Type       *string         `json:"type"`
	DeviceNo   *string         `json:"device_no"`
	APIKey     *string         `json:"api_key"`
	APIURL     *string         `json:"api_url"`
	FeieUser   *string         `json:"feie_user"`
	FeieUKey   *string         `json:"feie_ukey"`
	FeieSN     *string         `json:"feie_sn"`
	PrintTypes json.RawMessage `json:"print_types"`
	Status     *uint8          `json:"status"`
	AutoPrint  *bool           `json:"auto_print"`
	IsDefault  *bool           `json:"is_default"`
}

func UpdatePrinter(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	printerID, _ := strconv.ParseUint(c.Param("printer_id"), 10, 64)

	var req updatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var printer models.CloudPrinter
	if err := database.DB.Where("id = ? AND merchant_id = ?", printerID, merchantID).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	targetType := printer.Brand
	if req.Type != nil {
		targetType = *req.Type
	}
	targetDeviceNo := printer.DeviceNo
	if req.DeviceNo != nil {
		targetDeviceNo = *req.DeviceNo
	}
	targetFeieUser := printer.FeieUser
	if req.FeieUser != nil {
		targetFeieUser = *req.FeieUser
	}
	targetFeieUKey := printer.FeieUKey
	if req.FeieUKey != nil {
		targetFeieUKey = *req.FeieUKey
	}
	targetFeieSN := printer.FeieSN
	if req.FeieSN != nil {
		targetFeieSN = *req.FeieSN
	}
	if err := validatePrinterPayload(targetType, targetDeviceNo, targetFeieUser, targetFeieUKey, targetFeieSN); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["brand"] = *req.Type
	}
	if req.DeviceNo != nil {
		updates["device_no"] = *req.DeviceNo
	}
	if req.APIKey != nil {
		updates["api_key"] = *req.APIKey
	}
	if req.APIURL != nil {
		updates["api_url"] = *req.APIURL
	}
	if req.FeieUser != nil {
		updates["feie_user"] = *req.FeieUser
	}
	if req.FeieUKey != nil {
		updates["feie_ukey"] = *req.FeieUKey
	}
	if req.FeieSN != nil {
		updates["feie_sn"] = *req.FeieSN
	}
	if len(req.PrintTypes) > 0 {
		updates["print_types"] = models.JSON(req.PrintTypes)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AutoPrint != nil {
		updates["auto_print"] = *req.AutoPrint
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	tx := database.DB.Begin()
	if req.IsDefault != nil && *req.IsDefault {
		if err := tx.Model(&models.CloudPrinter{}).Where("merchant_id = ? AND id <> ?", merchantID, printer.ID).Update("is_default", false).Error; err != nil {
			tx.Rollback()
			response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "设置默认打印机失败")
			return
		}
	}
	if err := tx.Model(&printer).Updates(updates).Error; err != nil {
		tx.Rollback()
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新云打印机失败")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新云打印机失败")
		return
	}

	var updated models.CloudPrinter
	database.DB.Where("id = ?", printer.ID).First(&updated)
	response.Success(c, buildPrinterResponse(updated))
}

func DeletePrinter(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	printerID, _ := strconv.ParseUint(c.Param("printer_id"), 10, 64)

	if err := database.DB.Where("id = ? AND merchant_id = ?", printerID, merchantID).Delete(&models.CloudPrinter{}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除云打印机失败")
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}

func TestPrinter(c *gin.Context) {
	merchantID := middleware.GetMerchantID(c)
	printerID, _ := strconv.ParseUint(c.Param("printer_id"), 10, 64)

	var printer models.CloudPrinter
	if err := database.DB.Where("id = ? AND merchant_id = ?", printerID, merchantID).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	now := time.Now()
	logItem := models.PrintLog{
		MerchantID: merchantID,
		PrinterID:  printer.ID,
		Type:       "test",
		Status:     1,
		CreatedAt:  now,
	}
	if err := database.DB.Create(&logItem).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "打印测试失败")
		return
	}

	database.DB.Model(&models.CloudPrinter{}).
		Where("id = ?", printer.ID).
		Updates(map[string]interface{}{
			"last_print_at": now,
			"print_count":   printer.PrintCount + 1,
		})

	response.Success(c, gin.H{
		"success": true,
		"message": "打印测试成功",
	})
}

func validatePrinterPayload(printerType string, deviceNo string, feieUser string, feieUKey string, feieSN string) error {
	if deviceNo == "" {
		return fmt.Errorf("打印机设备编号不能为空")
	}
	if printerType == "feie" {
		if feieUser == "" {
			return fmt.Errorf("请输入飞鹅账号")
		}
		if feieUKey == "" {
			return fmt.Errorf("请输入飞鹅 UKey")
		}
		if feieSN == "" {
			return fmt.Errorf("请输入飞鹅打印机终端号")
		}
	}
	return nil
}

func buildPrinterResponse(printer models.CloudPrinter) printerResponse {
	return printerResponse{
		ID:          printer.ID,
		MerchantID:  printer.MerchantID,
		Name:        printer.Name,
		Type:        printer.Brand,
		DeviceNo:    printer.DeviceNo,
		APIURL:      printer.APIURL,
		FeieUser:    printer.FeieUser,
		FeieSN:      printer.FeieSN,
		PrintTypes:  printer.PrintTypes,
		Status:      printer.Status,
		AutoPrint:   printer.AutoPrint,
		IsDefault:   printer.IsDefault,
		PrintCount:  printer.PrintCount,
		LastPrintAt: printer.LastPrintAt,
		HasAPIKey:   printer.APIKey != "",
		HasFeieUKey: printer.FeieUKey != "",
		CreatedAt:   printer.CreatedAt,
		UpdatedAt:   printer.UpdatedAt,
	}
}
