package ws

import "encoding/json"

type merchantEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func broadcastMerchantEvent(eventType string, payload interface{}) int {
	message, err := json.Marshal(merchantEvent{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		return 0
	}

	return BroadcastToMerchant(message)
}

func BroadcastOrderNotify(orderNo string) int {
	return broadcastMerchantEvent("order_notify", map[string]interface{}{
		"order_no": orderNo,
	})
}

func BroadcastStoreVisitNotify(visitorOpenID string, source string) int {
	return broadcastMerchantEvent("store_visit_notify", map[string]interface{}{
		"visitor_openid": visitorOpenID,
		"source":         source,
	})
}