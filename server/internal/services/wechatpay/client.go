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
	"os"
	"strconv"
	"strings"
	"time"

	"fz_yyc_api/internal/config"

	wxpay "github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/partnerpayments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/profitsharing"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
)

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
	pubKeyID       string
	callbackURL    string
	client         *wxpay.Client
	notifyHandler  *notify.Handler
}

// SPMchID 返回服务商商户号（跨包访问）
func (c *ServiceProviderClient) SPMchID() string {
	return c.spMchID
}

// resolveKeyFileOrRaw 兼容两种配置方式：
// - 配置值为已存在的文件路径 → 读取文件内容作为密钥材料
// - 否则视为内联 PEM / Base64 原文
func resolveKeyFileOrRaw(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if info, err := os.Stat(trimmed); err == nil && !info.IsDir() {
		content, readErr := os.ReadFile(trimmed)
		if readErr != nil {
			return "", fmt.Errorf("读取密钥文件失败 %s: %w", trimmed, readErr)
		}
		return strings.TrimSpace(string(content)), nil
	}
	return trimmed, nil
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

func NewServiceProviderClient() (*ServiceProviderClient, error) {
	if config.Config == nil {
		return nil, fmt.Errorf("应用配置未初始化")
	}
	wc := config.Config.WechatPay

	spMchID := strings.TrimSpace(wc.SPMchID)
	apiv3Key := strings.TrimSpace(wc.APIV3Key)
	certSerialNo := strings.TrimSpace(wc.CertSerialNo)
	privateKeyStr, err := resolveKeyFileOrRaw(wc.PrivateKey)
	if err != nil {
		return nil, err
	}
	callbackURL := strings.TrimSpace(wc.CallbackURL)

	if spMchID == "" || apiv3Key == "" || certSerialNo == "" || privateKeyStr == "" {
		return nil, fmt.Errorf("微信支付服务商凭证未配置（mch_id/api_v3_key/cert/private_key）")
	}

	privKey, err := loadRSAPrivateKey(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("解析支付私钥失败: %w", err)
	}

	publicKeyStr, err := resolveKeyFileOrRaw(wc.PublicKey)
	if err != nil {
		return nil, err
	}
	// 解析微信支付公钥（用于响应验签 + 敏感字段加密）
	var pubKey *rsa.PublicKey
	if publicKeyStr != "" {
		if k, kErr := loadPublicKey(publicKeyStr); kErr == nil {
			pubKey = k
		}
	}
	pubKeyID := strings.TrimSpace(wc.PublicKeyID)

	httpClient := &nethttp.Client{Timeout: 15 * time.Second}

	var clientOpts []wxpay.ClientOption
	clientOpts = append(clientOpts, option.WithHTTPClient(httpClient))
	if pubKey != nil && pubKeyID != "" {
		// 微信支付公钥模式：负责签名、响应验签、敏感字段加密（自动携带 Wechatpay-Serial）
		clientOpts = append(clientOpts,
			option.WithWechatPayPublicKeyAuthCipher(spMchID, certSerialNo, privKey, pubKeyID, pubKey))
	} else {
		// 兜底：仅配置签名，不做响应验签
		clientOpts = append(clientOpts,
			option.WithMerchantCredential(spMchID, certSerialNo, privKey),
			option.WithoutValidator(),
		)
	}
	wxClient, err := wxpay.NewClient(context.Background(), clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("初始化微信支付客户端失败: %w", err)
	}

	client := &ServiceProviderClient{
		spMchID:      spMchID,
		apiv3Key:     apiv3Key,
		certSerialNo: certSerialNo,
		privateKey:   privKey,
		wechatPubKey: publicKeyStr,
		pubKeyID:     pubKeyID,
		callbackURL:  callbackURL,
		client:       wxClient,
	}

	if err := client.initNotifyHandler(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ServiceProviderClient) initNotifyHandler() error {
	// 本项目使用微信支付公钥模式验签回调：公钥与公钥ID必须成对配置，不再回退下载平台证书
	if c.wechatPubKey != "" && c.pubKeyID == "" {
		return fmt.Errorf("微信支付通知验签配置不完整：已配置公钥 WECHAT_PAY_SP_PUBLIC_KEY，但缺少公钥ID WECHAT_PAY_SP_PUBLIC_KEY_ID")
	}
	if c.pubKeyID != "" && c.wechatPubKey == "" {
		return fmt.Errorf("微信支付通知验签配置不完整：已配置公钥ID WECHAT_PAY_SP_PUBLIC_KEY_ID，但缺少公钥 WECHAT_PAY_SP_PUBLIC_KEY")
	}
	if c.wechatPubKey == "" {
		return fmt.Errorf("微信支付通知验签未配置：请在配置中提供 WECHAT_PAY_SP_PUBLIC_KEY(微信支付V3公钥) 与 WECHAT_PAY_SP_PUBLIC_KEY_ID(公钥ID)")
	}
	pubKey, err := loadPublicKey(c.wechatPubKey)
	if err != nil {
		return fmt.Errorf("微信支付V3公钥解析失败: %w", err)
	}
	handler, err := notify.NewRSANotifyHandler(
		c.apiv3Key,
		verifiers.NewSHA256WithRSAPubkeyVerifier(c.pubKeyID, *pubKey),
	)
	if err != nil {
		return fmt.Errorf("初始化通知处理器失败: %w", err)
	}
	c.notifyHandler = handler
	return nil
}

// loadPublicKey 解析微信支付公钥（PEM 或 Base64 原文，PKCS1/PKCS8）
func loadPublicKey(pemOrBase64 string) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(pemOrBase64)
	var raw []byte
	if strings.Contains(trimmed, "-----BEGIN") {
		block, _ := pem.Decode([]byte(trimmed))
		if block == nil {
			return nil, errors.New("公钥PEM解析失败")
		}
		raw = block.Bytes
	} else {
		decoded, err := base64.StdEncoding.DecodeString(trimmed)
		if err != nil {
			return nil, fmt.Errorf("公钥Base64解码失败: %w", err)
		}
		raw = decoded
	}
	if pkixKey, err := x509.ParsePKIXPublicKey(raw); err == nil {
		if rsaKey, ok := pkixKey.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("公钥不是RSA类型")
	}
	rsaKey, err2 := x509.ParsePKCS1PublicKey(raw)
	if err2 != nil {
		return nil, fmt.Errorf("解析RSA公钥失败: %w", err2)
	}
	return rsaKey, nil
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

	// 固定 sub_app 模式：sp_appid 取服务商 appid，sub_appid 取特约商户主体小程序 appid
	spAppid := strings.TrimSpace(config.Config.Wechat.AppID)
	if spAppid == "" {
		spAppid = req.AppID
	}
	subAppid := strings.TrimSpace(config.Config.Wechat.SubAppID)
	if subAppid == "" {
		subAppid = req.AppID
	}

	r := jsapi.PrepayRequest{
		SpAppid:     wxpay.String(spAppid),
		SpMchid:     wxpay.String(c.spMchID),
		SubAppid:    wxpay.String(subAppid),
		SubMchid:    wxpay.String(req.SubMchID),
		Description: wxpay.String(req.Description),
		OutTradeNo:  wxpay.String(req.OrderNo),
		NotifyUrl:   wxpay.String(req.NotifyURL),
		Amount: &jsapi.Amount{
			Total:    wxpay.Int64(req.TotalAmount),
			Currency: wxpay.String("CNY"),
		},
		Payer: &jsapi.Payer{
			SpOpenid:  nil,
			SubOpenid: wxpay.String(req.OpenID),
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
	if subAppid != "" {
		appID = subAppid
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
// 分账 (Profit Sharing)
// ============================================

// ProfitSharingReceiverItem 一次分账指令中单个接收方的信息
type ProfitSharingReceiverItem struct {
	Type        string // 类型: MERCHANT_ID 商户号 / PERSONAL_OPENID 个人openid
	Account     string // 商户号 或 个人openid
	Name        string // 个人类型时的真实姓名(选传, 传则校实名)
	Amount      int64  // 分账金额(单位:分)
	Description string // 分账原因描述
}

// ProfitSharingRequest 创建分账单请求
type ProfitSharingRequest struct {
	SubMchID      string // 特约商户号(分账出资方)
	AppID         string // 特约商户主体小程序appid
	TransactionID string // 微信支付交易单号
	OutOrderNo    string // 商户分账单号
	Receivers     []ProfitSharingReceiverItem
}

// ProfitSharingReceiverResult 分账单内单方结果
type ProfitSharingReceiverResult struct {
	Type        string     // MERCHANT_ID / PERSONAL_OPENID
	Account     string     // 接收方账号
	Amount      int64      // 分账金额(分)
	Result      string     // PENDING/SUCCESS/CLOSED
	DetailID    string     // 微信分账明细单号
	FailReason  string     // 失败原因
	CreateTime  time.Time  // 分账创建时间
	FinishTime  *time.Time // 分账完成时间
}

// ProfitSharingResult 分账单查询/创建结果
type ProfitSharingResult struct {
	OutOrderNo string
	OrderID    string // 微信分账单号
	Status     string // PROCESSING/FINISHED
	Receivers  []ProfitSharingReceiverResult
}

// AddProfitSharingReceiverRequest 添加分账接收方请求
type AddProfitSharingReceiverRequest struct {
	SubMchID       string // 特约商户号(分账出资方)
	AppID          string // 特约商户主体小程序appid
	Type           string // MERCHANT_ID / PERSONAL_OPENID
	Account        string // 商户号 或 个人openid
	Name           string // 商户全称或开户人姓名(MERCHANT_ID必传) / 个人姓名(PERSONAL_OPENID选传)
	RelationType   string // 与特约商户关系, 如 SERVICE_PROVIDER
}

// CreateProfitSharingOrder 创建分账单（将一笔已支付订单按各方金额分给多个接收方）
func (c *ServiceProviderClient) CreateProfitSharingOrder(
	ctx context.Context,
	req ProfitSharingRequest,
) (*ProfitSharingResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	if len(req.Receivers) == 0 {
		return nil, fmt.Errorf("分账接收方列表为空")
	}
	svc := profitsharing.OrdersApiService{Client: c.client}
	receivers := make([]profitsharing.CreateOrderReceiver, 0, len(req.Receivers))
	for _, r := range req.Receivers {
		item := profitsharing.CreateOrderReceiver{
			Type:        wxpay.String(r.Type),
			Account:     wxpay.String(r.Account),
			Amount:      wxpay.Int64(r.Amount),
			Description: wxpay.String(r.Description),
		}
		if r.Name != "" {
			item.Name = wxpay.String(r.Name)
		}
		receivers = append(receivers, item)
	}
	unfreezeUnsplit := false
	resp, _, err := svc.CreateOrder(ctx, profitsharing.CreateOrderRequest{
		Appid:            wxpay.String(req.AppID),
		OutOrderNo:       wxpay.String(req.OutOrderNo),
		SubMchid:         wxpay.String(req.SubMchID),
		TransactionId:    wxpay.String(req.TransactionID),
		Receivers:        receivers,
		UnfreezeUnsplit:  &unfreezeUnsplit,
	})
	if err != nil {
		return nil, fmt.Errorf("微信分账下单失败: %w", err)
	}
	return mapProfitSharingOrderEntity(resp), nil
}

// QueryProfitSharingOrder 查询分账结果
func (c *ServiceProviderClient) QueryProfitSharingOrder(
	ctx context.Context,
	subMchID, transactionID, outOrderNo string,
) (*ProfitSharingResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("微信支付客户端未初始化")
	}
	svc := profitsharing.OrdersApiService{Client: c.client}
	resp, _, err := svc.QueryOrder(ctx, profitsharing.QueryOrderRequest{
		SubMchid:      wxpay.String(subMchID),
		TransactionId: wxpay.String(transactionID),
		OutOrderNo:    wxpay.String(outOrderNo),
	})
	if err != nil {
		return nil, fmt.Errorf("微信分账结果查询失败: %w", err)
	}
	return mapProfitSharingOrderEntity(resp), nil
}

// QueryProfitSharingMerchantRatio 查询特约商户允许父商户分账的最大比例（单位: 万分比）
func (c *ServiceProviderClient) QueryProfitSharingMerchantRatio(ctx context.Context, subMchID string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("微信支付客户端未初始化")
	}
	if subMchID == "" {
		return 0, fmt.Errorf("缺少子商户号")
	}
	svc := profitsharing.MerchantsApiService{Client: c.client}
	resp, _, err := svc.QueryMerchantRatio(ctx, profitsharing.QueryMerchantRatioRequest{
		SubMchid: wxpay.String(subMchID),
	})
	if err != nil {
		return 0, fmt.Errorf("微信分账比例查询失败: %w", err)
	}
	if resp == nil || resp.MaxRatio == nil {
		return 0, fmt.Errorf("微信未返回允许分账比例")
	}
	return *resp.MaxRatio, nil
}

// AddProfitSharingReceiver 建立分账接收方关系
func (c *ServiceProviderClient) AddProfitSharingReceiver(
	ctx context.Context,
	req AddProfitSharingReceiverRequest,
) error {
	if c.client == nil {
		return fmt.Errorf("微信支付客户端未初始化")
	}
	if req.SubMchID == "" || req.AppID == "" || req.Type == "" || req.Account == "" || req.RelationType == "" {
		return fmt.Errorf("缺少分账接收方信息(sub_mchid/appid/type/account/relation_type)")
	}
	svc := profitsharing.ReceiversApiService{Client: c.client}
	recvType := profitsharing.ReceiverType(req.Type)
	recvRelation := profitsharing.ReceiverRelationType(req.RelationType)
	body := profitsharing.AddReceiverRequest{
		Appid:        wxpay.String(req.AppID),
		SubMchid:     wxpay.String(req.SubMchID),
		Type:         &recvType,
		Account:      wxpay.String(req.Account),
		RelationType: &recvRelation,
	}
	if req.Name != "" {
		body.Name = wxpay.String(req.Name)
	}
	_, _, err := svc.AddReceiver(ctx, body)
	if err != nil {
		// 已存在的接收方关系视为成功，避免重复建关系报错
		if IsProfitSharingReceiverAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("微信建立分账接收方关系失败: %w", err)
	}
	return nil
}

// DeleteProfitSharingReceiver 解除分账接收方关系
func (c *ServiceProviderClient) DeleteProfitSharingReceiver(
	ctx context.Context,
	req AddProfitSharingReceiverRequest,
) error {
	if c.client == nil {
		return fmt.Errorf("微信支付客户端未初始化")
	}
	if req.SubMchID == "" || req.AppID == "" || req.Type == "" || req.Account == "" {
		return fmt.Errorf("缺少分账接收方信息(sub_mchid/appid/type/account)")
	}
	svc := profitsharing.ReceiversApiService{Client: c.client}
	delType := profitsharing.ReceiverType(req.Type)
	_, _, err := svc.DeleteReceiver(ctx, profitsharing.DeleteReceiverRequest{
		Appid:    wxpay.String(req.AppID),
		SubMchid: wxpay.String(req.SubMchID),
		Type:     &delType,
		Account:  wxpay.String(req.Account),
	})
	if err != nil {
		return fmt.Errorf("微信解除分账接收方关系失败: %w", err)
	}
	return nil
}

// mapProfitSharingOrderEntity 将微信分账单实体映射为内部结果结构
func mapProfitSharingOrderEntity(resp *profitsharing.OrdersEntity) *ProfitSharingResult {
	if resp == nil {
		return &ProfitSharingResult{}
	}
	result := &ProfitSharingResult{
		OutOrderNo: derefString(resp.OutOrderNo),
		OrderID:    derefString(resp.OrderId),
		Status:     "",
		Receivers:  []ProfitSharingReceiverResult{},
	}
	if resp.State != nil {
		result.Status = string(*resp.State)
	}
	for _, r := range resp.Receivers {
		item := ProfitSharingReceiverResult{
			Type:       derefReceiverType(r.Type),
			Account:    derefString(r.Account),
			Amount:     derefInt64(r.Amount),
			Result:     derefDetailStatus(r.Result),
			DetailID:   derefString(r.DetailId),
			FailReason: derefDetailFailReason(r.FailReason),
		}
		if r.CreateTime != nil {
			item.CreateTime = *r.CreateTime
		}
		if r.FinishTime != nil {
			item.FinishTime = r.FinishTime
		}
		result.Receivers = append(result.Receivers, item)
	}
	return result
}

func derefReceiverType(t *profitsharing.ReceiverType) string {
	if t == nil {
		return ""
	}
	return string(*t)
}

func derefDetailStatus(s *profitsharing.DetailStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func derefDetailFailReason(r *profitsharing.DetailFailReason) string {
	if r == nil {
		return ""
	}
	return string(*r)
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// IsProfitSharingReceiverAlreadyExists 判断是否为「分账接收方已存在」错误
func IsProfitSharingReceiverAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *wxpay.APIError
	if errors.As(err, &apiErr) && strings.Contains(apiErr.Message, "已存在") {
		return true
	}
	return strings.Contains(err.Error(), "已存在")
}

// IsProfitSharingReceiverRelationNotExist 判断是否为「分账接收方关系不存在」错误
func IsProfitSharingReceiverRelationNotExist(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *wxpay.APIError
	if errors.As(err, &apiErr) && apiErr.Code == "PARAM_ERROR" && strings.Contains(apiErr.Message, "关系不存在") {
		return true
	}
	return strings.Contains(err.Error(), "关系不存在")
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
