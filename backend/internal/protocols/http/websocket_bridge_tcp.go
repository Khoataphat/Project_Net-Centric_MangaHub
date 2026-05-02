package http

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Bỏ qua CORS cho WS
}

func TCPBridgeHandler(c *gin.Context) {
	// 1. Nâng cấp HTTP lên WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Lỗi Upgrade WS:", err)
		return
	}
	defer ws.Close()

	// 2. Kết nối tới TCP Server nội bộ (cổng 9090)
	tcpConn, err := net.Dial("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Println("Không thể kết nối tới TCP Server:", err)
		return
	}
	defer tcpConn.Close()

	// 3. Chạy 2 luồng song song để bơm dữ liệu qua lại
	// Luồng A: Đọc từ TCP -> Gửi về WebSocket (Frontend)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := tcpConn.Read(buf)
			if err != nil {
				return
			}
			ws.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	// Luồng B: Đọc từ WebSocket (Frontend) -> Gửi xuống TCP
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		tcpConn.Write(msg)
	}
}
