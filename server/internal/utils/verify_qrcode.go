package utils

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func BuildVerifyCodeQRCodeDataURL(verifyCode string) (string, error) {
	if verifyCode == "" {
		return "", nil
	}

	content := fmt.Sprintf("verify_code=%s", verifyCode)
	pngBytes, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}
