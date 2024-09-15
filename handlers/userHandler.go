package handlers

import (
    "net/http"
	"21BCE2661_Backend/db"
    "21BCE2661_Backend/models"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/dgrijalva/jwt-go"
    "time"
	"os"
)

func Register(c *gin.Context) {
    var input struct {
        Email    string `json:"email" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.BindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user := models.User{
        Email:    input.Email,
        Password: string(hashedPassword),
    }

    result := db.DB.Create(&user)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// Login handler
func Login(c *gin.Context) {
    jwtSecret := os.Getenv("JWT_SECRET")
    var input struct {
        Email    string `json:"email" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.BindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
	result := db.DB.Where("email = ?", input.Email).Take(&user)
    if result.Error != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
        return
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "email": user.Email,
        "exp": time.Now().Add(time.Minute * 15).Unix(),
    })

    tokenString, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
