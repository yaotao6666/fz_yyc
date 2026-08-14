package wechatpay

import (
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
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/config"

	wxpay "github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/partnerpayments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
)

type JSAPIPayRequest struct {
	AppID       string
	OpenID      string
	AppMode     string
	SubMchID    string
	OrderNo     string
	Description string
	TotalAmount int64
	NotifyURL   string
}

type JSAPIPayResponse struct {
	AppID     string
	TimeStamp string
	NonceStr  string
	Package   string
	SignType  string
	PaySign   string
	PrepayID  string
}

type PartnerRefundResult struct {
	Status      string
	RefundID    string
	SuccessTime string
}

type RefundRequest struct {
	SubMchID     string
	OrderNo      string
	RefundNo     string
	Reason       string
	NotifyURL    string
	RefundAmount int64
	TotalAmount  int64
}

type RefundResponse struct {
	RefundID    string
	Status      string
	SuccessTime string
}

type ServiceProviderClient struct {
	spMchID        string
	apiv3Key       string
	certSerialNo   string
	privateKey     *rsa.PrivateKey
	wechatPubKey   string
	callbackURL    string
	client         *wxpay.Client
	notifyHandler  *notify.Handler
}

func loadRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = strings.TrimSpace(pemStr)
	var raw []byte
	if strings.Contains(pemStr, "-----BEGIN") {
		block, _ := pem.Decode([]byte(pemStr))
		if block == nil {
			return nil, errors.New("私钥PEM解析失败")
		}
		raw = block.Bytes
	} else {
		b, err := base64.StdEncoding.DecodeString(pemStr)
		if err != nil {
			return nil, fmt.Errorf("私钥Base64解码失败: %w", err)
		}
		raw = b
	}

	key, err := x509.ParsePKCS8PrivateKey(raw)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("私钥不是RSA类型")
	}
	rsaKey, err2 := x509.ParsePKCS1PrivateKey(raw)
	if err2 != nil {
		return nil, fmt.Errorf("解析RSA私钥失败: %w / %w", err, err2)
	}
	return rsaKey, nil
}

func loadX509Cert(pemOrBase64 string) (*x509.Certificate, error) {
	pemStr := strings.TrimSpace(pemOrBase64)
	var raw []byte
	if strings.Contains(pemStr, "-----BEGIN") {
		block, _ := pem.Decode([]byte(pemStr))
		if block == nil {
			return nil, errors.New("证书PEM解析失败")
		}
		raw = block.Bytes
	} else {
		b, err := base64.StdEncoding.DecodeString(pemStr)
		if err != nil {
			return nil, fmt.Errorf("证书Base64解码失败: %w", err)
		}
		raw = b
	}
	cert, err := x509.ParseCertificate(raw)
	if err != nil {
		return nil, fmt.Errorf("解析X509证书失败: %w", err)
	}
	return cert, nil
}

func NewServiceProviderClient() (*ServiceProviderClient, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	wc := config.Config.WechatPay

	spMchID := strings.TrimSpace(wc.SPMchID)
	apiv3Key := strings.TrimSpace(wc.APIV3Key)
	certSerialNo := strings.TrimSpace(wc.CertSerialNo)
	privateKeyStr := strings.TrimSpace(wc.PrivateKey)
	callbackURL := strings.TrimSpace(wc.CallbackURL)

	if spMchID == "" || apiv3Key == "" || certSerialNo == "" || privateKeyStr == "" {
		return nil, fmt.Errorf("微信支付服务商凭证未配置（mch_id/api_v3_key/cert/private_key）")
	}

	privKey, err := loadRSAPrivateKey(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("解析支付私钥失败: %w", err)
	}

	httpClient := &nethttp.Client{Timeout: 15 * time.Second}

	wxClient, err := wxpay.NewClient(
		context.Background(),
		option.WithMerchantCredential(spMchID, certSerialNo, privKey),
		option.WithHTTPClient(httpClient),
		option.WithoutValidator(),
	)
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}

	client := &ServiceProviderClient{
		spMchID:      spMchID,
		apiv3Key:     apiv3Key,
		certSerialNo: certSerialNo,
		privateKey:   privKey,
		wechatPubKey: strings.TrimSpace(wc.PublicKey),
		callbackURL:  callbackURL,
		client:       wxClient,
	}

	if err := client.initNotifyHandler(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ServiceProviderClient) initNotifyHandler() error {
	var v wxpay.CertificateGetter
	if c.wechatPubKey != "" {
		cert, err := loadX509Cert(c.wechatPubKey)
		if err == nil {
			v = newStaticCertGetter(cert)
		}
	}
	if v == nil {
		d, err := downloader.NewCertificateDownloader(
			context.Background(),
			c.spMchID,
			c.privateKey,
			c.certSerialNo,
			c.apiv3Key,
		)
		if err != nil {
			return fmt.Errorf("初始化微信平台证书下载器失败: %w", err)
		}
		v = d
	}
	handler, err := notify.NewRSANotifyHandler(c.apiv3Key, verifiers.NewSHA256WithRSAVerifier(v))
	if err != nil {
		return fmt.Errorf("初始化通知处理器失败: %w", err)
	}
	c.notifyHandler = handler
	return nil
}

type staticCertGetter struct {
	cert *x509.Certificate
}

func newStaticCertGetter(cert *x509.Certificate) *staticCertGetter { return &staticCertGetter{cert: cert} }

func (g *staticCertGetter) Get(_ context.Context, serialNumber string) (*x509.Certificate, bool) {
	_ = serialNumber
	if g.cert == nil {
		return nil, false
	}
	return g.cert, true
}

func (g *staticCertGetter) GetAll(_ context.Context) map[string]*x509.Certificate {
	if g.cert == nil {
		return map[string]*x509.Certificate{}
	}
	serial := strings.ToUpper(fmt.Sprintf("%X", g.cert.SerialNumber.Bytes()))
	return map[string]*x509.Certificate{serial: g.cert}
}

func (g *staticCertGetter) GetNewestSerial(_ context.Context) string {
	if g.cert == nil {
		return ""
	}
	return strings.ToUpper(fmt.Sprintf("%X", g.cert.SerialNumber.Bytes()))
}

func (c *ServiceProviderClient) CreatePartnerJSAPIPayOrder(
	ctx context.Context,
	req JSAPIPayRequest,
) (*JSAPIPayResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	req.NotifyURL = c.resolveNotifyURL(req.NotifyURL)
	if req.NotifyURL == "" {
		return nil, fmt.Errorf("支付回调地址未配置")
	}

	svc := jsapi.JsapiApiService{Client: c.client}

	spAppid := strings.TrimSpace(config.Config.Wechat.AppID)
	if spAppid == "" {
		spAppid = req.AppID
	}
	var subAppid *string
	if strings.EqualFold(req.AppMode, AppModeSubApp) {
		sub := strings.TrimSpace(config.Config.Wechat.SubAppID)
		if sub == "" {
			sub = req.AppID
		}
		subAppid = wxpay.String(sub)
	}

	r := jsapi.PrepayRequest{
		SpAppid:     wxpay.String(spAppid),
		SpMchid:     wxpay.String(c.spMchID),
		SubAppid:    subAppid,
		SubMchid:    wxpay.String(req.SubMchID),
		Description: wxpay.String(req.Description),
		OutTradeNo:  wxpay.String(req.OrderNo),
		NotifyUrl:   wxpay.String(req.NotifyURL),
		Amount: &jsapi.Amount{
			Total:    wxpay.Int64(req.TotalAmount),
			Currency: wxpay.String("CNY"),
		},
		Payer: &jsapi.Payer{
			SpOpenid:  wxpay.String(req.OpenID),
			SubOpenid: subAppidPayerOpenID(req.AppMode, req.OpenID),
		},
	}
	expireAt := time.Now().Add(30 * time.Minute)
	r.TimeExpire = &expireAt

	resp, _, err := svc.Prepay(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("微信JSAPI下单失败: %w", err)
	}

	prepayID := derefString(resp.PrepayId)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := randomString(32)
	pkg := "prepay_id=" + prepayID

	appID := spAppid
	if strings.EqualFold(req.AppMode, AppModeSubApp) && subAppid != nil {
		appID = *subAppid
	}

	message := appID + "\n" + timestamp + "\n" + nonce + "\n" + pkg + "\n"
	signature, err := signRSA(message, c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("生成支付参数签名失败: %w", err)
	}

	return &JSAPIPayResponse{
		AppID:     appID,
		TimeStamp: timestamp,
		NonceStr:  nonce,
		Package:   pkg,
		SignType:  "RSA",
		PaySign:   signature,
		PrepayID:  prepayID,
	}, nil
}

func subAppidPayerOpenID(mode, openID string) *string {
	if strings.EqualFold(mode, AppModeSubApp) {
		return wxpay.String(openID)
	}
	return nil
}

func (c *ServiceProviderClient) resolveNotifyURL(raw string) string {
	if raw != "" {
		return raw
	}
	return c.callbackURL
}

func (c *ServiceProviderClient) CreatePartnerRefund(
	ctx context.Context,
	req RefundRequest,
) (*RefundResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	notifyURL := c.resolveNotifyURL(req.NotifyURL)
	if notifyURL != "" {
		idx := strings.LastIndex(notifyURL, "/")
		if idx >= 0 {
			notifyURL = notifyURL[:idx] + "/refund"
		}
	}
	svc := refunddomestic.RefundsApiService{Client: c.client}
	var reason *string
	if strings.TrimSpace(req.Reason) != "" {
		reason = wxpay.String(req.Reason)
	}
	r := refunddomestic.CreateRequest{
		SubMchid:    wxpay.String(req.SubMchID),
		OutTradeNo:  wxpay.String(req.OrderNo),
		OutRefundNo: wxpay.String(req.RefundNo),
		Reason:      reason,
		Amount: &refunddomestic.AmountReq{
			Refund:   wxpay.Int64(req.RefundAmount),
			Total:    wxpay.Int64(req.TotalAmount),
			Currency: wxpay.String("CNY"),
		},
	}
	if notifyURL != "" {
		r.NotifyUrl = wxpay.String(notifyURL)
	}
	resp, _, err := svc.Create(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("微信退款申请失败: %w", err)
	}
	return &RefundResponse{
		RefundID:    derefString(resp.RefundId),
		Status:      refundStatusString(resp.Status),
		SuccessTime: formatTimePtr(resp.SuccessTime),
	}, nil
}

func (c *ServiceProviderClient) QueryPartnerRefundByRefundNo(
	ctx context.Context,
	refundNo string,
) (*PartnerRefundResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	svc := refunddomestic.RefundsApiService{Client: c.client}
	resp, _, err := svc.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{
		OutRefundNo: wxpay.String(refundNo),
	})
	if err != nil {
		return nil, fmt.Errorf("微信退款查询失败: %w", err)
	}
	return &PartnerRefundResult{
		Status:      refundStatusString(resp.Status),
		RefundID:    derefString(resp.RefundId),
		SuccessTime: formatTimePtr(resp.SuccessTime),
	}, nil
}

// ============================================
// 通知解析（验签 + 解密）
// ============================================

type NotifyEncryptResource struct {
	Algorithm      string `json:"algorithm"`
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	Ciphertext     string `json:"ciphertext"`
	OriginalType   string `json:"original_type"`
}

type NotifyRequest struct {
	ID           string                `json:"id"`
	EventType    string                `json:"event_type"`
	ResourceType string                `json:"resource_type"`
	Resource     NotifyEncryptResource `json:"resource"`
	Summary      string                `json:"summary"`
}

type PayTransactionResource struct {
	SpAppid           string `json:"sp_appid"`
	SpMchid           string `json:"sp_mchid"`
	SubAppid          string `json:"sub_appid,omitempty"`
	SubMchid          string `json:"sub_mchid"`
	OutTradeNo        string `json:"out_trade_no"`
	TransactionID     string `json:"transaction_id"`
	TradeType         string `json:"trade_type"`
	TradeState        string `json:"trade_state"`
	TradeStateDesc    string `json:"trade_state_desc"`
	BankType          string `json:"bank_type"`
	Attach            string `json:"attach"`
	SuccessTime       string `json:"success_time"`
	PayerSpOpenid     string `json:"-"`
	PayerSubOpenid    string `json:"-"`
	AmountTotal       int64  `json:"-"`
	AmountPayerTotal  int64  `json:"-"`
}

type refundNotifyRawPayer struct {
	SpOpenid  string `json:"sp_openid"`
	SubOpenid string `json:"sub_openid,omitempty"`
}

type refundNotifyRawAmount struct {
	Total         int64  `json:"total"`
	PayerTotal    int64  `json:"payer_total"`
	Currency      string `json:"currency"`
	PayerCurrency string `json:"payer_currency"`
}

type RefundNotifyResource struct {
	SpMchid             string  `json:"sp_mchid"`
	SubMchid            string  `json:"sub_mchid"`
	OutTradeNo          string  `json:"out_trade_no"`
	TransactionID       string  `json:"transaction_id"`
	OutRefundNo         string  `json:"out_refund_no"`
	RefundID            string  `json:"refund_id"`
	RefundStatus        string  `json:"refund_status"`
	SuccessTime         string  `json:"success_time"`
	Amount              float64 `json:"amount"`
	UserReceivedAccount string  `json:"user_received_account"`
}

// ParseNotifyRequest 从 *http.Request 解析并验签解密通知
// content 接收解密后的明文结构体（可选）
func (c *ServiceProviderClient) ParseNotifyRequest(
	ctx context.Context,
	req *nethttp.Request,
	content interface{},
) (*notify.Request, error) {
	if c.notifyHandler == nil {
		if err := c.initNotifyHandler(); err != nil {
			return nil, err
		}
	}
	if c.notifyHandler == nil {
		return nil, fmt.Errorf("通知处理器未初始化")
	}
	return c.notifyHandler.ParseNotifyRequest(ctx, req, content)
}

// ParseNotifyRequestRaw 基于 headers + body 解析并验签解密通知（Gin/未直接使用 net/http.Request 时）
func (c *ServiceProviderClient) ParseNotifyRequestRaw(
	ctx context.Context,
	headers map[string][]string,
	body []byte,
	content interface{},
) (*notify.Request, error) {
	if c.notifyHandler == nil {
		if err := c.initNotifyHandler(); err != nil {
			return nil, err
		}
	}
	if c.notifyHandler == nil {
		return nil, fmt.Errorf("通知处理器未初始化")
	}
	fakeReq, err := buildFakeHTTPRequest(headers, body)
	if err != nil {
		return nil, err
	}
	return c.notifyHandler.ParseNotifyRequest(ctx, fakeReq, content)
}

func buildFakeHTTPRequest(headers map[string][]string, body []byte) (*nethttp.Request, error) {
	req, err := nethttp.NewRequest(nethttp.MethodPost, "/", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *ServiceProviderClient) DecryptResource(resource NotifyEncryptResource) ([]byte, error) {
	if resource.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("不支持的加密算法: %s", resource.Algorithm)
	}
	key := []byte(c.apiv3Key)
	if len(key) != 32 {
		h := sha256.Sum256([]byte(c.apiv3Key))
		key = h[:]
	}
	nonce := []byte(resource.Nonce)
	additionalData := []byte(resource.AssociatedData)

	ciphertext, err := base64.StdEncoding.DecodeString(resource.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("密文解码失败: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES密钥解析失败: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("初始化GCM失败: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, fmt.Errorf("AEAD解密失败: %w", err)
	}
	return plaintext, nil
}

func (c *ServiceProviderClient) DecryptPayNotify(nr *NotifyRequest) (*PayTransactionResource, error) {
	raw, err := c.DecryptResource(nr.Resource)
	if err != nil {
		return nil, err
	}
	var intermediate struct {
		PayTransactionResource
		Payer  refundNotifyRawPayer  `json:"payer"`
		Amount refundNotifyRawAmount `json:"amount"`
	}
	if err := json.Unmarshal(raw, &intermediate); err != nil {
		return nil, fmt.Errorf("支付通知明文解析失败: %w", err)
	}
	intermediate.PayTransactionResource.PayerSpOpenid = intermediate.Payer.SpOpenid
	intermediate.PayTransactionResource.PayerSubOpenid = intermediate.Payer.SubOpenid
	intermediate.PayTransactionResource.AmountTotal = intermediate.Amount.Total
	intermediate.PayTransactionResource.AmountPayerTotal = intermediate.Amount.PayerTotal
	return &intermediate.PayTransactionResource, nil
}

func (c *ServiceProviderClient) DecryptRefundNotify(nr *NotifyRequest) (*RefundNotifyResource, error) {
	raw, err := c.DecryptResource(nr.Resource)
	if err != nil {
		return nil, err
	}
	var res RefundNotifyResource
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("退款通知明文解析失败: %w", err)
	}
	return &res, nil
}

// ============================================
// 工具
// ============================================

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func refundStatusString(s *refunddomestic.Status) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		// 降级使用伪随机
		for i := range b {
			b[i] = letters[int(uint32(i*131))%len(letters)]
		}
		return string(b)
	}
	for i, v := range buf {
		b[i] = letters[int(v)%len(letters)]
	}
	return string(b)
}

func signRSA(message string, key *rsa.PrivateKey) (string, error) {
	h := crypto.Hash.New(crypto.SHA256)
	if _, err := io.WriteString(h, message); err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}
