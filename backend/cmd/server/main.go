package main

import (
	"mangahub/internal/database"
	"mangahub/internal/protocols/grpc"
	"mangahub/internal/protocols/http"
	"mangahub/internal/protocols/tcp"
	"mangahub/internal/protocols/udp"
	"mangahub/internal/protocols/websocket"
	"mangahub/internal/protocols/bridge"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Khởi tạo Database
	database.InitDB("./data/mangahub.db")

	// 2. Khởi tạo Gin
	r := gin.Default()

	// 3. Sử dụng Middleware (CORS...)
	r.Use(CORSMiddleware())

	// 4. Khởi tạo Hub cho WebSocket
	chatHub := websocket.NewHub()
	go chatHub.Run() // Chạy Hub ở một Goroutine riêng

	// 5. Gọi Router đã tách
	http.SetupRouter(r, chatHub)

	// 6. Khởi tạo TCP Server (Chạy song song)
	go tcp.StartTCPServer(":9090")

	// 7. Khởi tạo gRPC Server
	go grpc.StartGRPCServer(":50051")

	// 8. Khởi tạo UDP Notifier Server (AC1)
	if err := udp.InitUDPServer(9999); err != nil {
		log.Printf("[UDP] Cảnh báo: %v", err)
	}

	// 9. Khởi tạo Bridge Service (Kết nối UDP Server -> WebSocket Hub)
	udpBridge := bridge.NewUDPBridge(8888, chatHub)
	if err := udpBridge.Start(); err != nil {
		log.Printf("[Bridge] Cảnh báo: %v", err)
	} else {
		// Đăng ký Bridge với UDP Server
		udp.AddBridge("127.0.0.1:8888")
	}

	// 10. Chạy server HTTP
	r.Run(":8080")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
