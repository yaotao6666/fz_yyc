package wechatpay

import (
	"fmt"
	"strings"

	"fz_yyc_api/internal/config"
)

// AppIdentity 当前小程序应用身份（统一使用 WECHAT_APP_ID/WECHAT_APP_SECRET）。
type AppIdentity struct {
	AppID     string
	AppSecret string
}

// GetActiveAppIdentity 返回用户端(C端)小程序应用身份（WECHAT_APP_ID/WECHAT_APP_SECRET）。
// 用于 C 端登录、下单收单(sub_appid)、分账、生成 C 端小程序码等。
func GetActiveAppIdentity() (*AppIdentity, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	appID := strings.TrimSpace(config.Config.Wechat.AppID)
	appSecret := strings.TrimSpace(config.Config.Wechat.AppSecret)
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("用户端小程序配置不完整（缺少 WECHAT_APP_ID/WECHAT_APP_SECRET）")
	}
	return &AppIdentity{
		AppID:     appID,
		AppSecret: appSecret,
	}, nil
}

// GetStaffAppIdentity 返回服务人员端小程序应用身份（WECHAT_STAFF_APP_ID/WECHAT_STAFF_APP_SECRET）。
// 用于服务人员端小程序登录换取 openid，与用户端(C端)身份相互独立。
// 开发调试时若未单独配置服务人员端参数，则回退到用户端(C端)配置。
func GetStaffAppIdentity() (*AppIdentity, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	appID := strings.TrimSpace(config.Config.Wechat.StaffAppID)
	appSecret := strings.TrimSpace(config.Config.Wechat.StaffAppSecret)
	if appID == "" || appSecret == "" {
		return GetActiveAppIdentity()
	}
	return &AppIdentity{
		AppID:     appID,
		AppSecret: appSecret,
	}, nil
}
