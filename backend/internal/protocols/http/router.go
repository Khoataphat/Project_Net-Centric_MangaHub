package http

import (
	"mangahub/internal/protocols/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, chatHub *websocket.Hub) {
	// 1. Khi vào trang chủ "/" -> Trả về auth.html
	r.GET("/", func(c *gin.Context) {
		c.File("./web/auth.html")
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

	// 2. Với tất cả các đường dẫn khác (như /dashboard.html, /notification.js)
	// Tự động tìm trong thư mục "./web"
	r.NoRoute(func(c *gin.Context) {
		c.File("./web" + c.Request.URL.Path)
	})
}
