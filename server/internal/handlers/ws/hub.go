package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 单商户模式下的连接集合，所有商家端连接共享同一广播通道
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[conn] = struct{}{}
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, conn)
}

func (h *Hub) Broadcast(msg []byte) int {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			h.Remove(conn)
			_ = conn.Close()
			continue
		}
		delivered++
	}
	return delivered
}

var defaultHub = NewHub()

func AddMerchantConn(conn *websocket.Conn) {
	defaultHub.Add(conn)
}

func RemoveMerchantConn(conn *websocket.Conn) {
	defaultHub.Remove(conn)
}

func BroadcastToMerchant(msg []byte) int {
	return defaultHub.Broadcast(msg)
}
