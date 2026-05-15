package wechatpay

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"fz_yyc_api/internal/config"
)

const baseURL = "https://api.mch.weixin.qq.com"

type ServiceProviderClient struct {
	httpClient *http.Client
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     config.WechatPay
}

func (c *ServiceProviderClient) GetSPMchID() string {
	return c.config.SPMchID
}

type JSAPIPayRequest struct {
	AppID       string
	OpenID      string
	SubMchID    string
	OrderNo     string
	Description string
	TotalAmount int64
	NotifyURL   string
}

type JSAPIPayResponse struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
	PrepayID  string `json:"prepay_id"`
}

type ProfitSharingReceiver struct {
	Type        string `json:"type"`
	Account     string `json:"account"`
	Name        string `json:"name,omitempty"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type ProfitSharingRequest struct {
	AppID         string
	SubMchID      string
	TransactionID string
	OrderNo       string
	Receivers     []ProfitSharingReceiver
}

type ProfitSharingResponse struct {
	OrderID    string `json:"order_id"`
	Status     string `json:"state"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type RefundRequest struct {
	SubMchID      string
	OrderNo       string
	RefundNo      string
	Reason        string
	NotifyURL     string
	RefundAmount  int64
	TotalAmount   int64
}

type RefundResponse struct {
	RefundID     string `json:"refund_id"`
	Status       string `json:"status"`
	SuccessTime  string `json:"success_time"`
}

type NotifyResult struct {
	EventType     string
	OrderNo       string
	TransactionID string
	TradeState    string
	SuccessTime   time.Time
	PayAmount     int64
	RawPayload    []byte
}

func NewServiceProviderClient() (*ServiceProviderClient, error) {
	cfg := config.Config.WechatPay
	if cfg.SPMchID == "" || cfg.APIV3Key == "" || cfg.CertSerialNo == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("服务商支付配置不完整")
	}

	privateKey, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解析服务商私钥失败: %w", err)
	}

	var publicKey *rsa.PublicKey
	if strings.TrimSpace(cfg.PublicKey) != "" {
		publicKey, err = parsePublicKey(cfg.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("解析微信支付公钥失败: %w", err)
		}
	}

	return &ServiceProviderClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     cfg,
	}, nil
}

func (c *ServiceProviderClient) CreatePartnerJSAPIPayOrder(ctx context.Context, req JSAPIPayRequest) (*JSAPIPayResponse, error) {
	payload := map[string]any{
		"sp_appid":    req.AppID,
		"sp_mchid":    c.config.SPMchID,
		"sub_mchid":   req.SubMchID,
		"description": req.Description,
		"out_trade_no": req.OrderNo,
		"notify_url":  req.NotifyURL,
		"amount": map[string]any{
			"total":    req.TotalAmount,
			"currency": "CNY",
		},
		"payer": map[string]any{
			"sp_openid": req.OpenID,
		},
	}

	var result struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := c.doJSONRequest(ctx, http.MethodPost, "/v3/pay/partner/transactions/jsapi", payload, &result); err != nil {
		return nil, err
	}
	if result.PrepayID == "" {
		return nil, fmt.Errorf("微信支付未返回 prepay_id")
	}

	nonceStr := randomString(32)
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	pkg := "prepay_id=" + result.PrepayID
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", req.AppID, timeStamp, nonceStr, pkg)
	signature, err := signWithPrivateKey(c.privateKey, message)
	if err != nil {
		return nil, fmt.Errorf("生成小程序支付签名失败: %w", err)
	}

	return &JSAPIPayResponse{
		AppID:     req.AppID,
		TimeStamp: timeStamp,
		NonceStr:  nonceStr,
		Package:   pkg,
		SignType:  "RSA",
		PaySign:   signature,
		PrepayID:  result.PrepayID,
	}, nil
}

func (c *ServiceProviderClient) CreateProfitSharingOrder(ctx context.Context, req ProfitSharingRequest) (*ProfitSharingResponse, error) {
	payload := map[string]any{
		"appid":          req.AppID,
		"sub_mchid":      req.SubMchID,
		"transaction_id": req.TransactionID,
		"out_order_no":   req.OrderNo,
		"receivers":      req.Receivers,
		"unfreeze_unsplit": false,
	}

	var result ProfitSharingResponse
	if err := c.doJSONRequest(ctx, http.MethodPost, "/v3/profitsharing/orders", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ServiceProviderClient) CreatePartnerRefund(ctx context.Context, req RefundRequest) (*RefundResponse, error) {
	if req.OrderNo == "" || req.RefundNo == "" {
		return nil, fmt.Errorf("缺少退款单号")
	}
	if req.TotalAmount <= 0 || req.RefundAmount <= 0 {
		return nil, fmt.Errorf("退款金额不正确")
	}
	if req.NotifyURL == "" {
		return nil, fmt.Errorf("缺少退款回调地址")
	}

	payload := map[string]any{
		"out_trade_no":  req.OrderNo,
		"out_refund_no": req.RefundNo,
		"notify_url":    req.NotifyURL,
		"amount": map[string]any{
			"refund":   req.RefundAmount,
			"total":    req.TotalAmount,
			"currency": "CNY",
		},
	}
	if strings.TrimSpace(req.Reason) != "" {
		payload["reason"] = req.Reason
	}
	if strings.TrimSpace(req.SubMchID) != "" {
		payload["sub_mchid"] = req.SubMchID
	}

	var result RefundResponse
	if err := c.doJSONRequest(ctx, http.MethodPost, "/v3/refund/domestic/refunds", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *ServiceProviderClient) ParseAndVerifyNotifyEvent(headers http.Header, body []byte) (string, []byte, error) {
	if c.publicKey != nil {
		if err := c.verifyNotifySignature(headers, body); err != nil {
			return "", nil, err
		}
	}

	var notifyReq struct {
		EventType string `json:"event_type"`
		Resource  struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &notifyReq); err != nil {
		return "", nil, fmt.Errorf("解析微信支付回调失败: %w", err)
	}

	plaintext, err := decryptResource(c.config.APIV3Key, notifyReq.Resource.AssociatedData, notifyReq.Resource.Nonce, notifyReq.Resource.Ciphertext)
	if err != nil {
		return "", nil, err
	}

	return notifyReq.EventType, plaintext, nil
}

func (c *ServiceProviderClient) ParseAndVerifyNotify(headers http.Header, body []byte) (*NotifyResult, error) {
	eventType, plaintext, err := c.ParseAndVerifyNotifyEvent(headers, body)
	if err != nil {
		return nil, err
	}

	var resource struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        struct {
			PayerTotal int64 `json:"payer_total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plaintext, &resource); err != nil {
		return nil, fmt.Errorf("解析支付回调资源失败: %w", err)
	}

	successTime, _ := time.Parse(time.RFC3339, resource.SuccessTime)
	return &NotifyResult{
		EventType:     eventType,
		OrderNo:       resource.OutTradeNo,
		TransactionID: resource.TransactionID,
		TradeState:    resource.TradeState,
		SuccessTime:   successTime,
		PayAmount:     resource.Amount.PayerTotal,
		RawPayload:    plaintext,
	}, nil
}

func (c *ServiceProviderClient) doJSONRequest(ctx context.Context, method, path string, payload any, target any) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("编码请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建微信支付请求失败: %w", err)
	}

	nonceStr := randomString(32)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, path, timestamp, nonceStr, string(bodyBytes))
	signature, err := signWithPrivateKey(c.privateKey, message)
	if err != nil {
		return fmt.Errorf("生成微信支付签名失败: %w", err)
	}

	auth := fmt.Sprintf(
		`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		c.config.SPMchID,
		nonceStr,
		timestamp,
		c.config.CertSerialNo,
		signature,
	)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "fz_yyc_api/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求微信支付失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取微信支付响应失败: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("微信支付请求失败: %s", strings.TrimSpace(string(respBody)))
	}
	if target == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("解析微信支付响应失败: %w", err)
	}
	return nil
}

func (c *ServiceProviderClient) verifyNotifySignature(headers http.Header, body []byte) error {
	signature := headers.Get("Wechatpay-Signature")
	timestamp := headers.Get("Wechatpay-Timestamp")
	nonce := headers.Get("Wechatpay-Nonce")
	if signature == "" || timestamp == "" || nonce == "" {
		return fmt.Errorf("微信支付回调签名头缺失")
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("微信支付回调签名解码失败: %w", err)
	}

	message := fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, string(body))
	hash := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(c.publicKey, crypto.SHA256, hash[:], signatureBytes); err != nil {
		return fmt.Errorf("微信支付回调验签失败: %w", err)
	}
	return nil
}

func decryptResource(apiV3Key, associatedData, nonce, ciphertext string) ([]byte, error) {
	key := []byte(apiV3Key)
	if len(key) != 32 {
		return nil, fmt.Errorf("API V3 Key 长度必须为32字节")
	}

	cipherTextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("解码回调密文失败: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建 AES Cipher 失败: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 AES GCM 失败: %w", err)
	}

	plaintext, err := aesgcm.Open(nil, []byte(nonce), cipherTextBytes, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("解密微信支付回调失败: %w", err)
	}
	return plaintext, nil
}

func signWithPrivateKey(privateKey *rsa.PrivateKey, message string) (string, error) {
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

func parsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	content, err := readPEMContent(raw)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("未找到私钥 PEM 块")
	}

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}

	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥类型不是 RSA")
	}
	return privateKey, nil
}

func parsePublicKey(raw string) (*rsa.PublicKey, error) {
	content, err := readPEMContent(raw)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("未找到公钥 PEM 块")
	}

	if strings.Contains(block.Type, "CERTIFICATE") {
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("证书中的公钥不是 RSA")
		}
		return publicKey, nil
	}

	keyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := keyAny.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥类型不是 RSA")
	}
	return publicKey, nil
}

func readPEMContent(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("PEM 内容为空")
	}
	if strings.Contains(trimmed, "BEGIN") {
		return []byte(trimmed), nil
	}
	return os.ReadFile(trimmed)
}

func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buffer := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	for i := range buffer {
		buffer[i] = letters[int(randomBytes[i])%len(letters)]
	}
	return string(buffer)
}
