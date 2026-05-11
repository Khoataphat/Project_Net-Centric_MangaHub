package http

import (
	"mangahub/internal/protocols/websocket"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, chatHub *websocket.Hub) {
	// 1. Khi vào trang chủ "/" -> Tìm auth.html
	r.GET("/", func(c *gin.Context) {
		folders := []string{"./frontend", "../frontend", "./web", "../web"}
		for _, folder := range folders {
			target := filepath.Join(folder, "auth.html")
			if _, err := os.Stat(target); err == nil {
				c.File(target)
				return
			}
		}
	})

	r.GET("/health", HealthHandler)
	api := r.Group("/api")
	{
		api.POST("/register", RegisterHandler)
		api.POST("/login", LoginHandler)
		api.GET("/mangas", GetMangas)
		api.GET("/mangas/:id", GetMangaByID)
		api.GET("/mangas/:id/chapters", GetMangaChapters)
		api.GET("/chapters/:chapter_id/pages", GetChapterPages)
		api.GET("/ws-tcp-bridge", TCPBridgeHandler)
		api.GET("/admin/scan-manga", RoleMiddleware("admin"), ScanMangaHandler)
		api.GET("/ws-logs", websocket.ServeLogWS)
		api.GET("/ws/chat", func(c *gin.Context) {
			websocket.ServeWS(chatHub, c)
		})
	}

	// 2. Tự động tìm file trong thư mục "frontend"
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Danh sách các folder có thể chứa frontend
		folders := []string{"./frontend", "../frontend", "./web", "../web"}

		for _, folder := range folders {
			target := filepath.Join(folder, path)
			if _, err := os.Stat(target); err == nil {
				c.File(target)
				return
			}
		}
	})
}
