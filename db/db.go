package db

import (
    "fmt"
    "os"

    "github.com/go-redis/redis/v8"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB
var RedisClient *redis.Client

func InitDB() error {
    dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbUser, dbPassword, dbName)

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %v", err)
    }

    // Test the connection
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get sql.DB instance: %v", err)
    }

    err = sqlDB.Ping()
    if err != nil {
        return fmt.Errorf("failed to ping database: %v", err)
    }

    return nil
}


func InitRedis() error {
    redisAddr := os.Getenv("REDIS_ADDR")
    fmt.Println("Redis address: ", redisAddr) 
    RedisClient = redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    _, err := RedisClient.Ping(RedisClient.Context()).Result()
    if err != nil {
        return fmt.Errorf("failed to connect to Redis: %v", err)
    }

    return nil
}

func CloseDB() {
    if DB != nil {
        sqlDB, _ := DB.DB()
        sqlDB.Close()
    }
}

func CloseRedis() {
    if RedisClient != nil {
        err := RedisClient.Close()
        if err != nil {
            fmt.Printf("Failed to close Redis connection: %v\n", err)
        }
    }
}