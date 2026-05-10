package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu              sync.RWMutex
	merchantClients map[uint64]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{merchantClients: make(map[uint64]map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(merchantID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.merchantClients[merchantID] == nil {
		h.merchantClients[merchantID] = make(map[*websocket.Conn]struct{})
	}
	h.merchantClients[merchantID][conn] = struct{}{}
}

func (h *Hub) Remove(merchantID uint64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.merchantClients[merchantID]
	if clients == nil {
		return
	}
	delete(clients, conn)
	if len(clients) == 0 {
		delete(h.merchantClients, merchantID)
	}
}

func (h *Hub) Broadcast(merchantID uint64, msg []byte) int {
	h.mu.RLock()
	clients := h.merchantClients[merchantID]
	conns := make([]*websocket.Conn, 0, len(clients))
	for conn := range clients {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			h.Remove(merchantID, conn)
			_ = conn.Close()
			continue
		}
		delivered++
	}
	return delivered
}

var defaultHub = NewHub()

func AddMerchantConn(merchantID uint64, conn *websocket.Conn) {
	defaultHub.Add(merchantID, conn)
}

func RemoveMerchantConn(merchantID uint64, conn *websocket.Conn) {
	defaultHub.Remove(merchantID, conn)
}

func BroadcastToMerchant(merchantID uint64, msg []byte) int {
	return defaultHub.Broadcast(merchantID, msg)
}
