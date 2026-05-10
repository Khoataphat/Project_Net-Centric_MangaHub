package http

import (
	"fmt"
	"mangahub/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RoleMiddleware kiểm tra xem user có quyền truy cập không
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy token từ Header "Authorization"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token xác thực"})
			c.Abort()
			return
		}

		// Định dạng: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Định dạng token không đúng"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 2. Giải mã token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("phương thức ký không hợp lệ: %v", token.Header["alg"])
			}
			return auth.JWTKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		// 3. Kiểm tra Role
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không thể đọc thông tin từ token"})
			c.Abort()
			return
		}

		userRole := claims["role"].(string)
		if userRole != requiredRole && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền truy cập tính năng này"})
			c.Abort()
			return
		}

		// Lưu thông tin vào context để dùng ở các handler sau
		c.Set("userID", int(claims["id"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Set("role", userRole)

		c.Next()
	}
}
