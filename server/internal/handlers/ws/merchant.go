package ws

import (
	"net/http"

	"fz_yyc_api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MerchantWS(c *gin.Context) {
	merchantID := uint64(middleware.GetMerchantID(c))
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	AddMerchantConn(merchantID, conn)
	defer func() {
		RemoveMerchantConn(merchantID, conn)
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
