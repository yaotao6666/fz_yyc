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

// GetActiveAppIdentity 返回当前生效的小程序应用身份（WECHAT_APP_ID/WECHAT_APP_SECRET）。
func GetActiveAppIdentity() (*AppIdentity, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	appID := strings.TrimSpace(config.Config.Wechat.AppID)
	appSecret := strings.TrimSpace(config.Config.Wechat.AppSecret)
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("微信小程序配置不完整（缺少 WECHAT_APP_ID/WECHAT_APP_SECRET）")
	}
	return &AppIdentity{
		AppID:     appID,
		AppSecret: appSecret,
	}, nil
}
