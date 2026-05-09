package websocket

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWebSocketConcurrency(goTest *testing.T) {
	hub := NewHub()
	go hub.Run()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		ServeWS(hub, c)
	})

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	const numClients = 20
	var wg sync.WaitGroup
	wg.Add(numClients)

	clients := make([]*websocket.Conn, numClients)

	for i := 0; i < numClients; i++ {
		go func(idx int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				goTest.Errorf("Lỗi kết nối client %d: %v", idx, err)
				return
			}
			clients[idx] = conn
		}(i)
	}
	wg.Wait()

	// Broadcast test
	message := []byte(`{"type":"NEW_CHAPTER", "manga":"Naruto"}`)
	hub.Broadcast(message)

	// Kiểm tra xem tất cả các client có nhận được không
	for i, conn := range clients {
		if conn == nil {
			continue
		}
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			goTest.Errorf("Client %d không nhận được tin nhắn: %v", i, err)
			continue
		}

		var data map[string]interface{}
		json.Unmarshal(msg, &data)
		if data["manga"] != "Naruto" {
			goTest.Errorf("Client %d nhận sai dữ liệu: %s", i, string(msg))
		}
		conn.Close()
	}
}
