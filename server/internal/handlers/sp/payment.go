package sp

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/pkg/database"

	"github.com/gin-gonic/gin"
)

type WechatPayNotify struct {
	ReturnCode     string `xml:"return_code"`
	ReturnMsg      string `xml:"return_msg"`
	ResultCode     string `xml:"result_code"`
	TransactionID  string `xml:"transaction_id"`
	OrderID        string `xml:"out_trade_no"`
	TimeEnd        string `xml:"time_end"`
}

type WechatPayNotifyResponse struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

type DecryptedNotifyData struct {
	TransactionID string `json:"transaction_id"`
	Amount         struct {
		Total       int `json:"total"`
		PayerTotal  int `json:"payer_total"`
		Currency    string `json:"currency"`
		PayerCurrency string `json:"payer_currency"`
	} `json:"amount"`
	OutTradeNo     string `json:"out_trade_no"`
	PayerOpenID    string `json:"payer.openid"`
	TradeState     string `json:"trade_state"`
	TradeType      string `json:"trade_type"`
	Attach         string `json:"attach"`
	SuccessTime    string `json:"success_time"`
}

func PaymentNotify(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "读取请求失败"})
		return
	}

	fmt.Printf("微信支付回调原始数据: %s\n", string(body))

	contentType := c.GetHeader("Content-Type")
	if contentType == "application/json" {
		handleV3Notify(c, body)
	} else {
		handleV2Notify(c, body)
	}
}

func handleV3Notify(c *gin.Context, body []byte) {
	var notifyReq struct {
		EventType       string `json:"event_type"`
		ResourceType    string `json:"resource_type"`
		Resource        struct {
			Algorithm      string `json:"algorithm"`
			ciphertext     string `json:"ciphertext"`
			OriginalBytes  string `json:"original_bytes"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(body, &notifyReq); err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "JSON解析失败"})
		return
	}

	apiV3Key := []byte("C4856B4B9E5A4E5E9F5A4B5C6D7E8F9A")

	ciphertext, err := base64.StdEncoding.DecodeString(notifyReq.Resource.ciphertext)
	if err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "ciphertext解码失败"})
		return
	}

	block, err := aes.NewCipher(apiV3Key)
	if err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "创建cipher失败"})
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "创建GCM失败"})
		return
	}

	nonce := []byte(notifyReq.Resource.Nonce)
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(notifyReq.Resource.AssociatedData))
	if err != nil {
		fmt.Printf("GCM解密失败，尝试AES-CBC: %v\n", err)
		plaintext = tryAESCBC(apiV3Key, nonce, ciphertext)
		if plaintext == nil {
			c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "解密失败"})
			return
		}
	}

	var notifyData DecryptedNotifyData
	if err := json.Unmarshal(plaintext, &notifyData); err != nil {
		fmt.Printf("解析解密数据失败: %v\n", err)
		fmt.Printf("解密数据: %s\n", string(plaintext))
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "解析解密数据失败"})
		return
	}

	fmt.Printf("V3支付回调解密数据: %+v\n", notifyData)

	if err := processPaymentSuccess(notifyData.OutTradeNo, notifyData.TransactionID); err != nil {
		c.XML(http.StatusInternalServerError, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "处理失败"})
		return
	}

	c.XML(http.StatusOK, WechatPayNotifyResponse{ReturnCode: "SUCCESS", ReturnMsg: "OK"})
}

func handleV2Notify(c *gin.Context, body []byte) {
	var notify WechatPayNotify
	if err := xml.Unmarshal(body, &notify); err != nil {
		c.XML(http.StatusBadRequest, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "XML解析失败"})
		return
	}

	fmt.Printf("V2支付回调数据: %+v\n", notify)

	if notify.ReturnCode != "SUCCESS" || notify.ResultCode != "SUCCESS" {
		c.XML(http.StatusOK, WechatPayNotifyResponse{ReturnCode: "SUCCESS", ReturnMsg: "OK"})
		return
	}

	if err := processPaymentSuccess(notify.OrderID, notify.TransactionID); err != nil {
		c.XML(http.StatusInternalServerError, WechatPayNotifyResponse{ReturnCode: "FAIL", ReturnMsg: "处理失败"})
		return
	}

	c.XML(http.StatusOK, WechatPayNotifyResponse{ReturnCode: "SUCCESS", ReturnMsg: "OK"})
}

func tryAESCBC(key, nonce, ciphertext []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	plaintext := make([]byte, len(ciphertext))
	dec := cipher.NewCBCDecrypter(block, nonce)
	dec.CryptBlocks(plaintext, ciphertext)

	padding := int(plaintext[len(plaintext)-1])
	if padding > aes.BlockSize || padding < 1 {
		return nil
	}

	end := len(plaintext) - padding
	plaintext = plaintext[:end]

	for i := end; i < len(plaintext); i++ {
		if plaintext[i] != byte(padding) {
			return nil
		}
	}

	return plaintext
}

func processPaymentSuccess(orderID, transactionID string) error {
	var order models.Order
	if err := database.DB.Where("order_no = ?", orderID).First(&order).Error; err != nil {
		fmt.Printf("订单不存在: %s\n", orderID)
		return nil
	}

	if order.Status == 2 {
		fmt.Printf("订单已支付: %s\n", orderID)
		return nil
	}

	updates := map[string]interface{}{
		"status":         "paid",
		"pay_time":        database.DB.NowFunc(),
		"transaction_id": transactionID,
	}

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	fmt.Printf("订单支付成功: %s, 微信交易号: %s\n", orderID, transactionID)
	return nil
}
