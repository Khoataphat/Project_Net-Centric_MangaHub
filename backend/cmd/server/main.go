package main

import (
	"mangahub/internal/auth"
	"mangahub/internal/database"
	"mangahub/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Khởi tạo Database
	database.InitDB("./data/mangahub.db")

	// 2. Khởi tạo Router Gin
	r := gin.Default()

	// Middleware xử lý CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Cho phép tất cả các nguồn
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Trả về 204 No Content cho các yêu cầu OPTIONS (Preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 3. Định nghĩa các Route API cho Người A
	api := r.Group("/api")
	{
		// API Đăng ký
		api.POST("/register", func(c *gin.Context) {
			var user models.User
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// Hash mật khẩu và lưu vào DB
			hashed, _ := auth.HashPassword(user.Password)
			_, err := database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashed)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User already exists"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Đăng ký thành công"})
		})

		// API Đăng nhập
		api.POST("/login", func(c *gin.Context) {
			var req models.LoginRequest
			c.ShouldBindJSON(&req)

			var hashedPassword string
			err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", req.Username).Scan(&hashedPassword)

			if err != nil || !auth.CheckPasswordHash(req.Password, hashedPassword) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
				return
			}

			token, _ := auth.GenerateJWT(req.Username)
			c.JSON(http.StatusOK, gin.H{"token": token})
		})
	}

	r.Run(":8080")
}
