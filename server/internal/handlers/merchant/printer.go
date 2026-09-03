package merchant

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// feiEAPIUrl 飞鹅云开放平台接口地址（测试打印 best-effort 调用）
const feiEAPIUrl = "http://api.feieyun.cn/Api/Open"

// printerToMap 将打印机模型转换为返回体，额外计算 has_feie_ukey（飞鹅 UKey 是否已配置）
func printerToMap(p models.Printer) gin.H {
	return gin.H{
		"id":            p.ID,
		"name":          p.Name,
		"type":          p.Type,
		"feie_user":     p.FeieUser,
		"feie_ukey":     p.FeieUKey,
		"feie_sn":       p.FeieSN,
		"status":        p.Status,
		"auto_print":    p.AutoPrint,
		"is_default":    p.IsDefault,
		"print_count":   p.PrintCount,
		"last_print_at": p.LastPrintAt,
		"remark":        p.Remark,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
		"has_feie_ukey": p.FeieUKey != "",
	}
}

func GetPrinters(c *gin.Context) {
	var printers []models.Printer
	if err := database.DB.Order("is_default DESC, id ASC").Find(&printers).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "获取打印机列表失败")
		return
	}

	// 空数组兜底，保证前端拿到的始终是数组
	list := make([]gin.H, 0, len(printers))
	for _, p := range printers {
		list = append(list, printerToMap(p))
	}

	response.Success(c, gin.H{"list": list})
}

func GetPrinter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var printer models.Printer
	if err := database.DB.Where("id = ?", id).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	response.Success(c, printerToMap(printer))
}

type CreatePrinterRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      uint8  `json:"type"`
	FeieUser  string `json:"feie_user"`
	FeieUKey  string `json:"feie_ukey"`
	FeieSN    string `json:"feie_sn"`
	Status    *uint8 `json:"status"`
	AutoPrint *bool  `json:"auto_print"`
	Remark    string `json:"remark"`
}

func CreatePrinter(c *gin.Context) {
	var req CreatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	printer := models.Printer{
		Name:     req.Name,
		Type:     req.Type,
		FeieUser: req.FeieUser,
		FeieUKey: req.FeieUKey,
		FeieSN:   req.FeieSN,
		Remark:   req.Remark,
	}
	// 合法类型仅 1=飞鹅 2=通用云打印，非法值回退为飞鹅
	if printer.Type != 1 && printer.Type != 2 {
		printer.Type = 1
	}
	if req.Status != nil {
		printer.Status = *req.Status
	} else {
		printer.Status = 1
	}
	if req.AutoPrint != nil {
		printer.AutoPrint = *req.AutoPrint
	}

	// 若当前没有任何打印机，则自动将首台设为默认打印机
	var count int64
	if err := database.DB.Model(&models.Printer{}).Count(&count).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建打印机失败")
		return
	}
	printer.IsDefault = count == 0

	if err := database.DB.Create(&printer).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "创建打印机失败")
		return
	}

	response.Success(c, gin.H{"id": printer.ID, "message": "创建成功"})
}

type UpdatePrinterRequest struct {
	Name      string `json:"name"`
	Type      *uint8 `json:"type"`
	FeieUser  string `json:"feie_user"`
	FeieUKey  string `json:"feie_ukey"`
	FeieSN    string `json:"feie_sn"`
	Status    *uint8 `json:"status"`
	AutoPrint *bool  `json:"auto_print"`
	IsDefault *bool  `json:"is_default"`
	Remark    string `json:"remark"`
}

func UpdatePrinter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req UpdatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var printer models.Printer
	if err := database.DB.Where("id = ?", id).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.FeieUser != "" {
		updates["feie_user"] = req.FeieUser
	}
	if req.FeieUKey != "" {
		updates["feie_ukey"] = req.FeieUKey
	}
	if req.FeieSN != "" {
		updates["feie_sn"] = req.FeieSN
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AutoPrint != nil {
		updates["auto_print"] = *req.AutoPrint
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	// 设置默认打印机时先取消原有默认，保证任一时刻至多一台默认
	setDefault := req.IsDefault != nil && *req.IsDefault && !printer.IsDefault

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if setDefault {
			if err := tx.Model(&models.Printer{}).
				Where("is_default = ?", true).
				Update("is_default", false).Error; err != nil {
				return err
			}
			updates["is_default"] = true
		}
		if len(updates) > 0 {
			if err := tx.Model(&printer).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "更新打印机失败")
		return
	}

	database.DB.First(&printer, id)
	response.Success(c, printerToMap(printer))
}

func DeletePrinter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var printer models.Printer
	if err := database.DB.Where("id = ?", id).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&printer).Error; err != nil {
			return err
		}
		// 若删除的是默认打印机，删除后如仍存在打印机则把其中一台设为默认
		if printer.IsDefault {
			var remain int64
			if err := tx.Model(&models.Printer{}).Count(&remain).Error; err != nil {
				return err
			}
			if remain > 0 {
				var first models.Printer
				if err := tx.Order("id ASC").First(&first).Error; err != nil {
					return err
				}
				if err := tx.Model(&first).Update("is_default", true).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "删除打印机失败")
		return
	}

	response.Success(c, gin.H{"message": "删除成功"})
}

// TestPrinter 测试打印：best-effort 调用飞鹅云打印接口
func TestPrinter(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var printer models.Printer
	if err := database.DB.Where("id = ?", id).First(&printer).Error; err != nil {
		response.Fail(c, http.StatusNotFound, response.CodeNotFound, "打印机不存在")
		return
	}

	if printer.FeieUser == "" || printer.FeieUKey == "" || printer.FeieSN == "" {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "打印机未配置飞鹅账号/UKey/编号，无法测试打印")
		return
	}

	if err := callFeiePrint(printer.FeieUser, printer.FeieUKey, printer.FeieSN); err != nil {
		response.Fail(c, http.StatusBadGateway, response.CodeServerError, "测试打印失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "打印任务已发送"})
}

// callFeiePrint 调用飞鹅云 Open_printMsg 接口发送一次性打印任务，超时 10s
func callFeiePrint(feieUser, feieUKey, feieSN string) error {
	stime := strconv.FormatInt(time.Now().Unix(), 10)
	// sig = md5(user + ukey + stime)，ukey 即飞鹅云 API 密钥
	hash := md5.Sum([]byte(feieUser + feieUKey + stime))
	sig := hex.EncodeToString(hash[:])

	form := url.Values{}
	form.Set("user", feieUser)
	form.Set("ukey", feieUKey)
	form.Set("sn", feieSN)
	form.Set("apiname", "Open_printMsg")
	form.Set("content", "测试打印\n你好，这是一条测试打印。\n")
	form.Set("times", "1")
	form.Set("stime", stime)
	form.Set("sig", sig)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(feiEAPIUrl, form)
	if err != nil {
		return fmt.Errorf("请求飞鹅云接口失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取飞鹅云响应失败: %v", err)
	}

	var result struct {
		Ret  int    `json:"ret"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析飞鹅云响应失败: %v (响应: %s)", err, string(body))
	}
	if result.Ret != 0 {
		return fmt.Errorf("飞鹅云返回错误: %s", result.Msg)
	}
	return nil
}
