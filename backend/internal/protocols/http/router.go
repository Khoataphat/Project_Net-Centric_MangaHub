package http

import "github.com/gin-gonic/gin"

func SetupRouter(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/register", RegisterHandler)
		api.POST("/login", LoginHandler)
		api.GET("/mangas", GetMangas)
	}
}
