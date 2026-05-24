package utils

import (
	"encoding/base64"
	"fmt"
)

const MerchantStoreHomePage = "pages/store/home"

type MerchantStoreQRCodeResult struct {
	QRCodeURL string
	Scene     string
	Page      string
}

// GenerateMerchantStoreQRCode 统一生成商家店铺首页微信小程序码。
func GenerateMerchantStoreQRCode(merchantID uint64, width int) (MerchantStoreQRCodeResult, error) {
	scene := fmt.Sprintf("merchant_id=%d", merchantID)
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
