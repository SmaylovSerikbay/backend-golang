package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	"taxi-backend/internal/models"
	"taxi-backend/internal/services"
)

type VerificationData struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code,omitempty"`
}

var (
	whatsappService = services.NewWhatsAppService()
)

// RequestVerificationCode отправляет код подтверждения через WhatsApp
func RequestVerificationCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data VerificationData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Генерируем код подтверждения
		code := whatsappService.GenerateVerificationCode()

		// Отправляем код через WhatsApp
		if err := whatsappService.SendVerificationCode(data.Phone, code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки кода"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Код подтверждения отправлен"})
	}
}

// VerifyAndRegister проверяет код и регистрирует пользователя
func VerifyAndRegister(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			Phone     string `json:"phone" binding:"required"`
			Code      string `json:"code" binding:"required"`
			FirstName string `json:"firstName" binding:"required"`
			LastName  string `json:"lastName" binding:"required"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем код
		isValid, err := whatsappService.VerifyCode(data.Phone, data.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код подтверждения"})
			return
		}

		// Проверяем, существует ли пользователь с таким номером телефона
		var existingUser models.User
		if err := db.Where("phone = ?", data.Phone).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь с таким номером телефона уже существует"})
			return
		}

		// Создаем пользователя
		user := &models.User{
			Phone:     data.Phone,
			FirstName: data.FirstName,
			LastName:  data.LastName,
		}

		if err := db.Create(user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя"})
			return
		}

		// Генерируем JWT токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		jwtSecret := os.Getenv("JWT_SECRET")
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user":  user,
		})
	}
}

// VerifyAndLogin проверяет код и выполняет вход
func VerifyAndLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data VerificationData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем код
		isValid, err := whatsappService.VerifyCode(data.Phone, data.Code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код подтверждения"})
			return
		}

		// Ищем пользователя
		var user models.User
		if err := db.Where("phone = ?", data.Phone).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
			return
		}

		// Генерируем JWT токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		jwtSecret := os.Getenv("JWT_SECRET")
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user":  user,
		})
	}
}
