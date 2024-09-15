package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "21BCE2661_Backend/db"
    "21BCE2661_Backend/handlers"
    "21BCE2661_Backend/routers" 
    "21BCE2661_Backend/auth"
)

func main() {
    if err := db.InitDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.CloseDB()

    if err := db.InitRedis(); err != nil {
        log.Fatalf("Failed to initialize Redis: %v", err)
    }
    defer db.CloseRedis()

    r := gin.Default()
    r.Use(auth.RateLimitMiddleware(db.RedisClient))

	routers.SetupUserRoutes(r)
	routers.SetupFileRoutes(r)

    go handlers.StartWorker()

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}


