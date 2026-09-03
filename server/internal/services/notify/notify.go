package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"fz_yyc_api/internal/models"
	"fz_yyc_api/internal/utils"
	"fz_yyc_api/pkg/database"
)

// OrderAccepted 服务人员接单后，尽力而为地通知下单用户（当前仅微信小程序订阅消息通道）。
// 不阻塞接单主流程：任何失败仅写日志并返回 error，由调用方吞掉。
func OrderAccepted(ctx context.Context, order *models.Order, staffName, staffPhone string) error {
	if order == nil {
		return nil
	}

	// 幂等：该订单已发送过 wechat 通知则直接返回，不重复发送
	var existed models.OrderNotifyLog
	if err := database.DB.
		Where("order_id = ? AND channel = ?", order.ID, "wechat").
		First(&existed).Error; err == nil {
		return nil
	}

	logRecord := &models.OrderNotifyLog{
		OrderID: order.ID,
		UserID:  order.UserID,
		Channel: "wechat",
	}

	templateID := os.Getenv("WECHAT_STAFF_ACCEPT_TEMPLATE_ID")
	if templateID == "" {
		// 未配置订阅消息模板 ID：记一条未配置日志并返回，不真正调接口
		logRecord.Message = "未配置订阅消息模板ID"
		_ = database.DB.Create(logRecord).Error
		return nil
	}

	// 获取下单用户 openid
	var openID string
	if order.User != nil {
		openID = order.User.OpenID
	} else {
		var user models.User
		if err := database.DB.First(&user, order.UserID).Error; err == nil {
			openID = user.OpenID
		}
	}
	if openID == "" {
		logRecord.Message = "用户openid为空"
		_ = database.DB.Create(logRecord).Error
		return fmt.Errorf("用户openid为空")
	}

	if err := sendWechatSubscribe(ctx, templateID, openID, order, staffName, staffPhone); err != nil {
		logRecord.Message = err.Error()
		_ = database.DB.Create(logRecord).Error
		return err
	}

	logRecord.Success = 1
	logRecord.Message = "发送成功"
	_ = database.DB.Create(logRecord).Error
	return nil
}

// sendWechatSubscribe 通过小程序订阅消息接口 subscribeMessage.send 发送接单通知
func sendWechatSubscribe(ctx context.Context, templateID, openID string, order *models.Order, staffName, staffPhone string) error {
	accessToken, err := utils.GetAccessToken()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"thing1": map[string]string{"value": truncateThing(staffName)},  // 服务人员姓名
		"thing2": map[string]string{"value": truncateThing(staffPhone)}, // 联系电话
	}
	if order.ScheduledAt != nil {
		data["time3"] = map[string]string{"value": order.ScheduledAt.Format("2006-01-02 15:04")} // 预约时间
	}

	request := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"page":        "pages/order/detail?id=" + strconv.FormatUint(order.ID, 10),
		"data":        data,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化订阅消息失败: %v", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("构建订阅消息请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送订阅消息失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取订阅消息响应失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析订阅消息响应失败: %v", err)
	}
	if errcode, ok := result["errcode"].(float64); ok && errcode != 0 {
		errmsg, _ := result["errmsg"].(string)
		return fmt.Errorf("发送订阅消息失败: %d - %s", int(errcode), errmsg)
	}
	return nil
}

// truncateThing 微信订阅消息 thing 类型字段最长 20 个字符（按字符数截断）
func truncateThing(s string) string {
	runes := []rune(s)
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return s
}
