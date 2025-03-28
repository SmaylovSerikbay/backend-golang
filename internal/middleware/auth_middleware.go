package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"taxi-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("\n=== JWT Auth Middleware Start ===\n")
		fmt.Printf("Request URL: %s\n", c.Request.URL.String())
		fmt.Printf("Request Method: %s\n", c.Request.Method)
		fmt.Printf("Request Path: %s\n", c.Request.URL.Path)
		fmt.Printf("Request Headers: %v\n", c.Request.Header)
		fmt.Printf("Request Remote Addr: %s\n", c.Request.RemoteAddr)
		fmt.Printf("Using JWT_SECRET: %s\n", os.Getenv("JWT_SECRET"))

		authHeader := c.GetHeader("Authorization")
		fmt.Printf("Authorization header: %s\n", authHeader)

		if authHeader == "" {
			fmt.Printf("Error: Authorization header is missing\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует токен авторизации"})
			c.Abort()
			fmt.Printf("=== JWT Auth Middleware End (Unauthorized - No Token) ===\n\n")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			fmt.Printf("Error: Invalid token format - parts: %v\n", parts)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			fmt.Printf("=== JWT Auth Middleware End (Unauthorized - Invalid Format) ===\n\n")
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			fmt.Printf("Error validating token: %v\n", err)
			fmt.Printf("Token parts: %v\n", strings.Split(parts[1], "."))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			fmt.Printf("=== JWT Auth Middleware End (Unauthorized - Invalid Token) ===\n\n")
			return
		}

		// Проверяем роль пользователя
		if claims.Role == "admin" {
			fmt.Printf("Admin access granted\n")
			c.Set("user_id", uint(0)) // Для админа устанавливаем user_id = 0
			c.Set("role", "admin")
			c.Next()
			fmt.Printf("=== JWT Auth Middleware End (Success - Admin) ===\n\n")
			return
		}

		if claims.UserID == 0 {
			fmt.Printf("Error: Invalid user ID\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный ID пользователя"})
			c.Abort()
			fmt.Printf("=== JWT Auth Middleware End (Unauthorized - Invalid User ID) ===\n\n")
			return
		}

		fmt.Printf("Token validated successfully for user ID: %v\n", claims.UserID)
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
		fmt.Printf("=== JWT Auth Middleware End (Success) ===\n\n")
	}
}
