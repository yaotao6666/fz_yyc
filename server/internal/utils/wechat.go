package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"fz_yyc_api/internal/config"
)

// 获取微信access_token
func GetAccessToken() (string, error) {
	appID := config.Config.Wechat.AppID
	appSecret := config.Config.Wechat.AppSecret

	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("微信AppID或AppSecret未配置")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求access_token失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		errmsg, _ := result["errmsg"].(string)
		return "", fmt.Errorf("获取access_token失败: %d - %s", int(errcode), errmsg)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token不存在: %s", string(body))
	}

	return accessToken, nil
}

// 生成不限数量的小程序码（使用 getwxacodeunlimit 接口，支持 scene + page）
func CreateWXACode(scene string, page string, width int) ([]byte, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", accessToken)

	request := map[string]interface{}{
		"scene": scene,
		"page":  page,
		"width": float64(width),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求生成二维码失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查是否返回错误（JSON格式）
	var errResult map[string]interface{}
	if err := json.Unmarshal(body, &errResult); err == nil {
		if errcode, ok := errResult["errcode"].(float64); ok && errcode != 0 {
			errmsg, _ := errResult["errmsg"].(string)
			return nil, fmt.Errorf("生成二维码失败: %d - %s", int(errcode), errmsg)
		}
	}

	// 如果返回的是JSON而不是图片，说明出错了
	if len(body) > 0 && body[0] == '{' {
		return nil, fmt.Errorf("生成二维码失败，返回了错误信息: %s", string(body))
	}

	return body, nil
}

// 生成小程序二维码（使用 createwxaqrcode 接口）
func CreateWXAQRCode(scene string, page string, width int) ([]byte, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=%s", accessToken)

	request := map[string]interface{}{
		"scene": scene,
		"path":  page,
		"width": float64(width),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求生成二维码失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查是否返回错误（JSON格式）
	var errResult map[string]interface{}
	if err := json.Unmarshal(body, &errResult); err == nil {
		if errcode, ok := errResult["errcode"].(float64); ok && errcode != 0 {
			errmsg, _ := errResult["errmsg"].(string)
			return nil, fmt.Errorf("生成二维码失败: %d - %s", int(errcode), errmsg)
		}
	}

	// 如果返回的是JSON而不是图片，说明出错了
	if len(body) > 0 && body[0] == '{' {
		return nil, fmt.Errorf("生成二维码失败，返回了错误信息: %s", string(body))
	}

	return body, nil
}
