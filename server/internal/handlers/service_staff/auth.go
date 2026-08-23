package service_staff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/services/wechatpay"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 注册申请请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
}

// LoginRequest 账号密码登录
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// WechatLoginRequest 微信快捷登录
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// Register 服务人员注册申请（待审核）
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&models.ServiceStaff{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "用户名已存在")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	staff := models.ServiceStaff{
		Username: req.Username,
		Password: string(hashedPassword),
		Name:     req.Name,
		Phone:    req.Phone,
		Status:   0, // 待审核
	}

	if err := database.DB.Create(&staff).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "注册失败")
		return
	}

	response.SuccessWithMessage(c, "注册申请已提交，请等待审核", gin.H{
		"id":     staff.ID,
		"status": staff.Status,
	})
}

// Login 账号密码登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	var staff models.ServiceStaff
	if err := database.DB.Where("username = ?", req.Username).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.Password)); err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	if staff.Status == 0 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "账号待审核，请等待管理员审核")
		return
	}
	if staff.Status == 2 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "账号已禁用")
		return
	}

	now := time.Now()
	database.DB.Model(&models.ServiceStaff{}).Where("id = ?", staff.ID).Update("last_login_at", now)

	token, _ := utils.GenerateToken(staff.ID, 0, "service_staff", staff.Username)
	response.Success(c, gin.H{
		"token": token,
		"staff": staff,
	})
}

// WechatLogin 微信快捷登录
func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	openID, err := resolveWechatOpenID(req.Code)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	var staff models.ServiceStaff
	if err := database.DB.Where("openid = ?", openID).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "您还不是服务人员，请先注册")
		return
	}

	if staff.Status == 0 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "账号待审核，请等待管理员审核")
		return
	}
	if staff.Status == 2 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "账号已禁用")
		return
	}

	now := time.Now()
	database.DB.Model(&models.ServiceStaff{}).Where("id = ?", staff.ID).Update("last_login_at", now)

	token, _ := utils.GenerateToken(staff.ID, 0, "service_staff", staff.Username)
	response.Success(c, gin.H{
		"token": token,
		"staff": staff,
	})
}

// GetCurrentStaff 从上下文获取当前服务人员
func GetCurrentStaff(c *gin.Context) (*models.ServiceStaff, error) {
	staffID := utils.GetUserID(c)
	if staffID == 0 {
		return nil, fmt.Errorf("未获取到服务人员ID")
	}

	var staff models.ServiceStaff
	if err := database.DB.First(&staff, staffID).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

// resolveWechatOpenID 通过微信 code 换取 openID
func resolveWechatOpenID(code string) (string, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return "", fmt.Errorf("微信登录凭证不能为空")
	}

	// 兼容本地 mock 联调
	if strings.HasPrefix(trimmedCode, "mock_staff_") || strings.HasPrefix(trimmedCode, "mock_") {
		return "staff_wx_" + trimmedCode, nil
	}

	appIdentity, err := wechatpay.GetActiveAppIdentity()
	if err != nil {
		return "", fmt.Errorf("微信小程序配置缺失")
	}

	query := url.Values{}
	query.Set("appid", appIdentity.AppID)
	query.Set("secret", appIdentity.AppSecret)
	query.Set("js_code", trimmedCode)
	query.Set("grant_type", "authorization_code")

	requestURL := "https://api.weixin.qq.com/sns/jscode2session?" + query.Encode()
	resp, err := http.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("请求微信登录服务失败")
	}
	defer resp.Body.Close()

	var result struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析微信登录结果失败")
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信登录失败: %s", result.ErrMsg)
	}
	if strings.TrimSpace(result.OpenID) == "" {
		return "", fmt.Errorf("未获取到微信用户标识")
	}

	return strings.TrimSpace(result.OpenID), nil
}

// GetProfile 获取服务人员个人信息
func GetProfile(c *gin.Context) {
	staff, err := GetCurrentStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取信息失败")
		return
	}
	response.Success(c, staff)
}
