package services

import (
	"crypto/sha1"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"go-report-management/structs"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Login(c *gin.Context, db *gorm.DB) {
	type loginCreds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var creds loginCreds
	var user structs.SysUser

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	// Recuperar usuario de la base de datos
	result := db.Where("username = ?", creds.Username).First(&user)

	fmt.Println(result)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !checkPasswordHash(creds.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	tokenString, err := GenerateJWT(creds.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	refreshToken, err := GenerateRefreshJWT(creds.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "refresh_token": refreshToken})
}

func checkPasswordHash(password, hash string) bool {
	return fmt.Sprintf("%x", sha1.Sum([]byte(password))) == hash
}

func GenerateJWT(username string) (string, error) {
	loadEnv()
	jwtKey, jwtKeyExists := os.LookupEnv("jwtSecret")
	if !jwtKeyExists {
		log.Fatalf("JWT secret is missing")
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "your_app_name",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))

	return tokenString, err
}

func GenerateRefreshJWT(username string) (string, error) {
	loadEnv()
	jwtKey, jwtKeyExists := os.LookupEnv("jwtSecret")

	if !jwtKeyExists {
		log.Fatalf("JWT secret is missing")
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "your_app_name",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))

	return tokenString, err
}

func AuthenticateJWT() gin.HandlerFunc {
	loadEnv()
	jwtKey, jwtKeyExists := os.LookupEnv("jwtSecret")

	if !jwtKeyExists {
		log.Fatalf("JWT secret is missing")
	}

	return func(c *gin.Context) {
		const BearerSchema = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerSchema)
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			var message string
			if err != nil {
				message = err.Error()
			} else {
				message = "Invalid token"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": message})
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}

func RefreshToken(c *gin.Context) {
	loadEnv()
	jwtKey, jwtKeyExists := os.LookupEnv("jwtSecret")

	if !jwtKeyExists {
		log.Fatalf("JWT secret is missing")
	}

	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := jwt.Parse(requestBody.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has expired"})
			return
		}
		
		expirationTime := time.Now().Add(15 * time.Minute)
		claims["exp"] = expirationTime.Unix()

		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := newToken.SignedString([]byte(jwtKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
	}
}
