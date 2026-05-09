package http

import (
	"mangahub/internal/protocols/websocket"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, chatHub *websocket.Hub) {
	api := r.Group("/api")
	{
		api.POST("/register", RegisterHandler)
		api.POST("/login", LoginHandler)
		api.GET("/mangas", GetMangas)
		api.GET("/mangas/:id", GetMangaByID)
		api.GET("/ws-tcp-bridge", TCPBridgeHandler)
		api.GET("/admin/scan-manga", ScanMangaHandler)
		api.GET("/ws/chat", func(c *gin.Context) {
			websocket.ServeWS(chatHub, c)
		})
	}
}
