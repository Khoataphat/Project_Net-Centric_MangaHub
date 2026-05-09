package main

import (
	"io"
	"log"
	"mangahub/internal/database"
	"mangahub/internal/protocols/bridge"
	"mangahub/internal/protocols/grpc"
	"mangahub/internal/protocols/http"
	"mangahub/internal/protocols/tcp"
	"mangahub/internal/protocols/udp"
	"mangahub/internal/protocols/websocket"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	multiWriter := io.MultiWriter(os.Stdout, websocket.WsLogWriter{})
	log.SetOutput(multiWriter)
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = multiWriter

	database.InitDB("./data/mangahub.db")

	go websocket.ServerLogs.Run()

	r := gin.Default()
	r.Use(CORSMiddleware())

	chatHub := websocket.NewHub()
	go chatHub.Run()

	http.SetupRouter(r, chatHub)

	go tcp.StartTCPServer(":9090")
	go grpc.StartGRPCServer(":50051")

	if err := udp.InitUDPServer(9999); err != nil {
		log.Printf("[UDP] Cảnh báo: %v", err)
	}

	udpBridge := bridge.NewUDPBridge(8888, chatHub)
	if err := udpBridge.Start(); err != nil {
		log.Printf("[Bridge] Cảnh báo: %v", err)
	} else {
		if err := udp.AddBridge("127.0.0.1:8888"); err != nil {
			log.Printf("[UDP] Cảnh báo đăng ký bridge: %v", err)
		}
	}

	log.Println("[HTTP] Server đang chạy tại port :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("[HTTP] Lỗi khi chạy server: %v", err)
	}
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
