package http

import (
	"bufio"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TCPBridgeHandler(c *gin.Context) {
	// 1. Nâng cấp HTTP → WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("[Bridge] Lỗi Upgrade WS:", err)
		return
	}
	defer ws.Close()

	// 2. Kết nối tới TCP Server nội bộ (:9090)
	tcpConn, err := net.Dial("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Println("[Bridge] Không thể kết nối tới TCP Server:", err)
		ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"ERROR","message":"Cannot connect to TCP server"}`))
		return
	}
	defer tcpConn.Close()

	done := make(chan struct{})

	// ── Luồng A: TCP → WebSocket ──────────────────────────────────────────
	go func() {
		defer close(done)
		defer ws.Close()

		scanner := bufio.NewScanner(tcpConn)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			if err := ws.WriteMessage(websocket.TextMessage, line); err != nil {
				log.Printf("[Bridge] Lỗi ghi WebSocket: %v", err)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("[Bridge] TCP Scanner lỗi: %v", err)
		}
	}()

	// ── Luồng B: WebSocket → TCP ──────────────────────────────────────────
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {

			break
		}
		// Đảm bảo message kết thúc bằng \n
		if len(msg) == 0 || msg[len(msg)-1] != '\n' {
			msg = append(msg, '\n')
		}
		if _, err := tcpConn.Write(msg); err != nil {
			log.Printf("[Bridge] Lỗi ghi TCP: %v", err)
			break
		}
	}

	tcpConn.Close()

	// Chờ Luồng A kết thúc hoàn toàn
	<-done
}
