package utils

import (
	"encoding/base64"
)

const MerchantStoreHomePage = "pages/store/home"

type MerchantStoreQRCodeResult struct {
	QRCodeURL string
	Scene     string
	Page      string
}

// GenerateMerchantStoreQRCode 统一生成商家店铺首页微信小程序码（单商户模式，scene 固定为 store=1）。
func GenerateMerchantStoreQRCode(width int) (MerchantStoreQRCodeResult, error) {
	scene := "store=1"
	result := MerchantStoreQRCodeResult{
		Scene: scene,
		Page:  MerchantStoreHomePage,
	}

	qrCodeBytes, err := CreateWXACode(scene, result.Page, width)
	if err != nil {
		return MerchantStoreQRCodeResult{}, err
	}

	result.QRCodeURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCodeBytes)
	return result, nil
}
