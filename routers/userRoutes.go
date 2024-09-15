package routers

import (
    "github.com/gin-gonic/gin"
    "21BCE2661_Backend/handlers"
)

func SetupUserRoutes(r *gin.Engine) {
    userGroup := r.Group("/user")
    {
        userGroup.POST("/register", handlers.Register)
        userGroup.POST("/login", handlers.Login)
    }
}