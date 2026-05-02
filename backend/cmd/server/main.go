package main

import (
	"mangahub/internal/database"
	"mangahub/internal/protocols/grpc"
	"mangahub/internal/protocols/http"
	"mangahub/internal/protocols/tcp"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Khởi tạo Database
	database.InitDB("./data/mangahub.db")

	// 2. Khởi tạo Gin
	r := gin.Default()

	// 3. Sử dụng Middleware (CORS...)
	r.Use(CORSMiddleware())

	// 4. Gọi Router đã tách
	http.SetupRouter(r)

	// 5. Khởi tạo TCP Server (Chạy song song)
	go tcp.StartTCPServer(":9090")

	// 6. Khởi tạo gRPC Server
	go grpc.StartGRPCServer(":50051")

	// 7. Chạy server HTTP
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
