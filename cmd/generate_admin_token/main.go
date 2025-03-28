package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	fmt.Printf("Using JWT_SECRET: %s\n", jwtSecret)

	claims := Claims{
		UserID: 0, // Для админа устанавливаем user_id = 0
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(1, 0, 0)), // Токен действителен 1 год
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatalf("Error generating admin token: %v", err)
	}

	fmt.Printf("Generated admin token: %s\n", tokenString)
}
