package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("mangahub_secret_key") // Trong thực tế nên dùng biến môi trường

// HashPassword mã hóa mật khẩu người dùng
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash kiểm tra mật khẩu khớp với bản mã hóa không
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT tạo token cho người dùng sau khi đăng nhập thành công
func GenerateJWT(username string) (string, error) {
	// Tạo các "claims" (thông tin đính kèm trong token)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token hết hạn sau 24 giờ
	})

	// Ký tên vào token bằng chìa khóa bí mật
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, err
}
