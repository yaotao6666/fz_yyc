package wechatpay

import (
	"fmt"
	"strings"

	"fz_yyc_api/internal/config"
)

// AppIdentity 当前小程序应用身份。
// 本项目固定使用特约商户主体小程序（sub_app 模式），不再支持 sp_app/sub_app 切换。
type AppIdentity struct {
	Mode            string
	AppID           string
	AppSecret       string
	OpenIDFieldName string
}

// GetActiveAppIdentity 返回当前生效的应用身份（固定 sub_app，即特约商户主体小程序）。
func GetActiveAppIdentity() (*AppIdentity, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	appID := strings.TrimSpace(config.Config.Wechat.SubAppID)
	appSecret := strings.TrimSpace(config.Config.Wechat.SubAppSecret)
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("特约商户主体小程序配置不完整（缺少 SubAppID/SubAppSecret）")
	}
	return &AppIdentity{
		Mode:            "sub_app",
		AppID:           appID,
		AppSecret:       appSecret,
		OpenIDFieldName: "sub_openid",
	}, nil
}