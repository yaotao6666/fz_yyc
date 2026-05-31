package wechatpay

import (
	"fmt"
	"strings"

	"fz_yyc_api/internal/config"
)

const (
	AppModeSPApp  = "sp_app"
	AppModeSubApp = "sub_app"
)

type AppIdentity struct {
	Mode            string
	AppID           string
	AppSecret       string
	OpenIDFieldName string
}

func GetActiveAppIdentity() (*AppIdentity, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}

	mode := normalizeAppMode(config.Config.WechatPay.AppMode)
	switch mode {
	case AppModeSubApp:
		appID := strings.TrimSpace(config.Config.Wechat.SubAppID)
		appSecret := strings.TrimSpace(config.Config.Wechat.SubAppSecret)
		if appID == "" || appSecret == "" {
			return nil, fmt.Errorf("非服务商主体小程序配置不完整")
		}
		return &AppIdentity{
			Mode:            mode,
			AppID:           appID,
			AppSecret:       appSecret,
			OpenIDFieldName: "sub_openid",
		}, nil
	default:
		appID := strings.TrimSpace(config.Config.Wechat.AppID)
		appSecret := strings.TrimSpace(config.Config.Wechat.AppSecret)
		if appID == "" || appSecret == "" {
			return nil, fmt.Errorf("服务商主体小程序配置不完整")
		}
		return &AppIdentity{
			Mode:            AppModeSPApp,
			AppID:           appID,
			AppSecret:       appSecret,
			OpenIDFieldName: "sp_openid",
		}, nil
	}
}

func normalizeAppMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case AppModeSubApp:
		return AppModeSubApp
	default:
		return AppModeSPApp
	}
}
