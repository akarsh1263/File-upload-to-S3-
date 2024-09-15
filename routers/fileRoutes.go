package routers

import (
    "github.com/gin-gonic/gin"
    "21BCE2661_Backend/handlers"
    "21BCE2661_Backend/auth"
)

func SetupFileRoutes(r *gin.Engine) {
    fileGroup := r.Group("/file")
	fileGroup.Use(auth.AuthMiddleware())
    {
        fileGroup.POST("/upload", handlers.UploadFile)
        fileGroup.GET("/", handlers.GetFiles) 
        fileGroup.GET("/:file_id", handlers.GetFileURL) 
        fileGroup.GET("/search", handlers.SearchFiles) 
    }
}