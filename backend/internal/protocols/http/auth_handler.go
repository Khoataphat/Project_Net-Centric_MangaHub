package http

import (
	"mangahub/internal/auth"
	"mangahub/internal/database"
	"mangahub/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHandler xử lý yêu cầu đăng ký tài khoản
func RegisterHandler(c *gin.Context) {
	var user models.User

	// 1. Nhận dữ liệu từ Frontend
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 2. Mã hóa mật khẩu bằng hàm đã viết ở thư mục auth/
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi mã hóa"})
		return
	}

	// 3. Lưu vào SQLite
	_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, hashedPassword)
	if err != nil {
		// Thường lỗi ở bước này là do username đã bị trùng (ràng buộc UNIQUE trong SQL)
		c.JSON(http.StatusConflict, gin.H{"error": "Tên đăng nhập đã tồn tại"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký thành công"})
}

// LoginHandler xử lý yêu cầu đăng nhập và cấp phát JWT
func LoginHandler(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập đủ tài khoản và mật khẩu"})
		return
	}

	// 1. Tìm user trong database
	var hashedPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", req.Username).Scan(&hashedPassword)

	// Nếu không tìm thấy username hoặc quét dữ liệu lỗi
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	// 2. So sánh mật khẩu người dùng nhập với bản mã hóa trong DB
	if !auth.CheckPasswordHash(req.Password, hashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	// 3. Mật khẩu đúng -> Tạo Token thông hành
	token, err := auth.GenerateJWT(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo phiên đăng nhập"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
