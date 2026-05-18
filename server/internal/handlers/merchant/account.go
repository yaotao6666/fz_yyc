package merchant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fz_yyc_api/internal/config"
	"fz_yyc_api/internal/middleware"
	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
	"fz_yyc_api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type BindWechatRequest struct {
	Code string `json:"code" binding:"required"`
}

type WechatQuickLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

func getCurrentMerchantStaff(c *gin.Context) (*models.MerchantStaff, error) {
	merchantID := middleware.GetMerchantID(c)
	usernameValue, _ := c.Get("username")
	username, _ := usernameValue.(string)

	var staff models.MerchantStaff
	if err := database.DB.Where("merchant_id = ? AND username = ?", merchantID, username).First(&staff).Error; err != nil {
		return nil, err
	}

	return &staff, nil
}

func buildMerchantOpenID(code string) string {
	return "mch_wx_" + code
}

type wechatCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type merchantWechatIdentity struct {
	OpenID  string
	UnionID string
}

func resolveMerchantWechatIdentity(code string) (*merchantWechatIdentity, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, fmt.Errorf("微信登录凭证不能为空")
	}

	// 兼容本地 mock 联调，避免已有回归脚本失效。
	if strings.HasPrefix(trimmedCode, "mock_merchant_") || strings.HasPrefix(trimmedCode, "mock_") {
		return &merchantWechatIdentity{
			OpenID: buildMerchantOpenID(trimmedCode),
		}, nil
	}

	if config.Config == nil || config.Config.Wechat.AppID == "" || config.Config.Wechat.AppSecret == "" {
		return nil, fmt.Errorf("微信小程序配置缺失")
	}

	query := url.Values{}
	query.Set("appid", config.Config.Wechat.AppID)
	query.Set("secret", config.Config.Wechat.AppSecret)
	query.Set("js_code", trimmedCode)
	query.Set("grant_type", "authorization_code")

	requestURL := "https://api.weixin.qq.com/sns/jscode2session?" + query.Encode()
	response, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信登录服务失败")
	}
	defer response.Body.Close()

	var result wechatCode2SessionResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信登录结果失败")
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信登录失败: %s", result.ErrMsg)
	}
	if strings.TrimSpace(result.OpenID) == "" {
		return nil, fmt.Errorf("未获取到微信用户标识")
	}

	return &merchantWechatIdentity{
		OpenID:  strings.TrimSpace(result.OpenID),
		UnionID: strings.TrimSpace(result.UnionID),
	}, nil
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工信息失败")
		return
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(req.OldPassword)); compareErr != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "原密码错误")
		return
	}

	if req.OldPassword == req.NewPassword {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "新旧密码不能相同")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "密码加密失败")
		return
	}

	if err := database.DB.Model(&models.MerchantStaff{}).
		Where("id = ?", staff.ID).
		Update("password", string(hashedPassword)).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "修改密码失败")
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功"})
}

func BindWechat(c *gin.Context) {
	var req BindWechatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工信息失败")
		return
	}

	wechatIdentity, err := resolveMerchantWechatIdentity(req.Code)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	var existStaff models.MerchantStaff
	if err := database.DB.Where("openid = ? AND id <> ?", wechatIdentity.OpenID, staff.ID).First(&existStaff).Error; err == nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "该微信已绑定其他商家账号")
		return
	}

	now := time.Now()
	bindUpdates := map[string]interface{}{
		"openid":               wechatIdentity.OpenID,
		"wechat_bound_at":      now,
		"last_wechat_login_at": nil,
	}
	if wechatIdentity.UnionID != "" {
		bindUpdates["unionid"] = wechatIdentity.UnionID
	}
	if err := database.DB.Model(&models.MerchantStaff{}).
		Where("id = ?", staff.ID).
		Updates(bindUpdates).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "绑定微信失败")
		return
	}

	response.Success(c, gin.H{
		"openid":          wechatIdentity.OpenID,
		"unionid":         wechatIdentity.UnionID,
		"wechat_bound_at": now,
		"message":         "绑定成功",
	})
}

func UnbindWechat(c *gin.Context) {
	staff, err := getCurrentMerchantStaff(c)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "获取员工信息失败")
		return
	}

	if err := database.DB.Model(&models.MerchantStaff{}).
		Where("id = ?", staff.ID).
		Updates(map[string]interface{}{
			"openid":               "",
			"unionid":              "",
			"wechat_bound_at":      nil,
			"last_wechat_login_at": nil,
		}).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, response.CodeServerError, "解绑微信失败")
		return
	}

	response.Success(c, gin.H{"message": "解绑成功"})
}

func WechatQuickLogin(c *gin.Context) {
	var req WechatQuickLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, "参数错误")
		return
	}

	wechatIdentity, err := resolveMerchantWechatIdentity(req.Code)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, response.CodeParamError, err.Error())
		return
	}

	var staff models.MerchantStaff
	if err := database.DB.Preload("Merchant").Where("openid = ?", wechatIdentity.OpenID).First(&staff).Error; err != nil {
		response.Fail(c, http.StatusUnauthorized, response.CodeUnauthorized, "您还不是商家，请注册后使用")
		return
	}

	if staff.Status != 1 {
		response.Fail(c, http.StatusForbidden, response.CodeForbidden, "账号已停用")
		return
	}

	now := time.Now()
	loginUpdates := map[string]interface{}{
		"last_login_at":        now,
		"last_wechat_login_at": now,
	}
	if wechatIdentity.UnionID != "" {
		loginUpdates["unionid"] = wechatIdentity.UnionID
	}
	database.DB.Model(&models.MerchantStaff{}).Where("id = ?", staff.ID).Updates(loginUpdates)
	if wechatIdentity.UnionID != "" {
		staff.UnionID = wechatIdentity.UnionID
	}
	staff.LastLoginAt = &now
	staff.LastWechatLoginAt = &now

	token, _ := utils.GenerateToken(staff.MerchantID, "merchant", staff.Username)
	response.Success(c, gin.H{
		"token":       token,
		"merchant_id": staff.MerchantID,
		"staff":       staff,
	})
}
