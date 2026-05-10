package ws

import "encoding/json"

type merchantEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func broadcastMerchantEvent(merchantID uint64, eventType string, payload interface{}) int {
	message, err := json.Marshal(merchantEvent{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		return 0
	}

	return BroadcastToMerchant(merchantID, message)
}

func BroadcastOrderNotify(merchantID uint64, orderNo string) int {
	return broadcastMerchantEvent(merchantID, "order_notify", map[string]interface{}{
		"merchant_id": merchantID,
		"order_no":    orderNo,
	})
}

func BroadcastStoreVisitNotify(merchantID uint64, visitorOpenID string, source string) int {
	return broadcastMerchantEvent(merchantID, "store_visit_notify", map[string]interface{}{
		"merchant_id":     merchantID,
		"visitor_openid":  visitorOpenID,
		"source":          source,
	})
}
