package auth

import (
    "fmt"
    "strings"
	"os"
    "net/http"
    "time"
    "github.com/go-redis/redis/v8"


    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
    "context"
)

var secretKey = []byte(os.Getenv("JWT_SECRET")) 
var ctx = context.Background()

// Claims struct to define JWT claims
type Claims struct {
    Email string `json:"email"`
    jwt.StandardClaims
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // The Authorization header should be in the format: "Bearer <token>"
        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid Authorization header format"})
            c.Abort()
            return
        }

        token := bearerToken[1]
        claims, err := validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token: " + err.Error()})
            c.Abort()
            return
        }

        // Set the user ID in the context for use in subsequent handlers
        c.Set("email", claims.Email)
        c.Next()
    }
}

func validateToken(tokenString string) (*Claims, error) {
    // Parse the token
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate the signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secretKey, nil
    })

    if err != nil {
        return nil, err
    }

    // Validate the token and return the claims
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}

// RateLimitMiddleware limits requests to 100 requests per user per minute
func RateLimitMiddleware(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        userKey := c.ClientIP() // or use a unique identifier like user's email from context

        // Define a unique key for the user in Redis
        redisKey := fmt.Sprintf("rate_limit:%s", userKey)

        // Increment the counter
        count, err := redisClient.Incr(ctx, redisKey).Result()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            c.Abort()
            return
        }

        // If this is the first request, set an expiration of 1 minute
        if count == 1 {
            redisClient.Expire(ctx, redisKey, 1*time.Minute)
        }

        // If the user has exceeded the limit, return 429
        if count > 100 {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }

        // Continue to the next handler
        c.Next()
    }
}