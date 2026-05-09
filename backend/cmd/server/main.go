package main

import (
	"context"
	"mangahub/internal/database"
	"mangahub/internal/protocols/grpc"
	"mangahub/internal/protocols/http"
	"mangahub/internal/protocols/tcp"
	"mangahub/internal/protocols/udp"
	"mangahub/internal/protocols/websocket"
	"mangahub/internal/protocols/bridge"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"
	net_http "net/http"

	"github.com/gin-gonic/gin"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// 0. Load Configuration from Environment
	host := getEnv("HOST", "0.0.0.0")
	httpPort := getEnv("HTTP_PORT", "8080")
	tcpPort := getEnv("TCP_PORT", "9090")
	udpPort := getEnv("UDP_PORT", "9999")
	bridgePort := getEnv("BRIDGE_PORT", "8888")
	grpcPort := getEnv("GRPC_PORT", "50051")
	dbPath := getEnv("DB_PATH", "./data/mangahub.db")

	log.Printf("Starting MangaHub Backend on %s", host)

	// 1. Khởi tạo Database
	database.InitDB(dbPath)

	// 2. Setup Context cho Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Khởi tạo Gin
	r := gin.Default()
	r.Use(CORSMiddleware())

	// 4. Khởi tạo Hub cho WebSocket
	chatHub := websocket.NewHub()
	go chatHub.Run() // Chạy Hub ở một Goroutine riêng

	// 5. Gọi Router đã tách
	http.SetupRouter(r, chatHub)

	// 6. Khởi tạo TCP Server (AC1, AC4)
	go func() {
		if err := tcp.StartTCPServer(ctx, fmt.Sprintf("%s:%s", host, tcpPort)); err != nil {
			log.Printf("[TCP Error] %v", err)
		}
	}()

	// 7. Khởi tạo gRPC Server (AC1, AC4)
	go func() {
		if err := grpc.StartGRPCServer(ctx, fmt.Sprintf("%s:%s", host, grpcPort)); err != nil {
			log.Printf("[gRPC Error] %v", err)
		}
	}()

	// 8. Khởi tạo UDP Notifier Server (AC1)
	udpPortInt := 9999
	fmt.Sscanf(udpPort, "%d", &udpPortInt)
	if err := udp.InitUDPServer(udpPortInt); err != nil {
		log.Printf("[UDP Warning] %v", err)
	}

	// 9. Khởi tạo Bridge Service
	bridgePortInt := 8888
	fmt.Sscanf(bridgePort, "%d", &bridgePortInt)
	udpBridge := bridge.NewUDPBridge(bridgePortInt, chatHub)
	if err := udpBridge.Start(); err != nil {
		log.Printf("[Bridge Warning] %v", err)
	} else {
		// Đăng ký Bridge với UDP Server
		bridgeAddr := fmt.Sprintf("127.0.0.1:%d", bridgePortInt)
		udp.AddBridge(bridgeAddr)
	}

	// 10. Chạy server HTTP với cơ chế Graceful Shutdown (AC4)
	srv := &net_http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, httpPort),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != net_http.ErrServerClosed {
			log.Fatalf("[HTTP Error] listen: %s\n", err)
		}
	}()

	log.Printf("[HTTP] Server đang chạy tại %s:%s", host, httpPort)

	// Chờ tín hiệu dừng
	<-ctx.Done()

	// Đóng các resource
	log.Println("Shutting down MangaHub Backend...")

	// HTTP Shutdown với timeout 5s
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("MangaHub Backend exited cleanly.")
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
