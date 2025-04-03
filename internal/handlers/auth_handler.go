package handlers

import (
	"fmt"
	"log"
	"net/http"
	"taxi-backend/internal/models"
	"taxi-backend/internal/services"
	"taxi-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Phone     string `json:"phone" binding:"required,e164"`
}

type VerifyCodeRequest struct {
	Phone string `json:"phone" binding:"required,e164"`
	Code  string `json:"code" binding:"required"`
}

type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required,e164"`
}

type AuthResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Token   string              `json:"token,omitempty"`
	User    models.UserResponse `json:"user,omitempty"`
	Error   string              `json:"error,omitempty"`
}

func AuthRegister(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: fmt.Sprintf("Неверный формат данных: %v", err),
			})
			return
		}

		// Проверяем, существует ли пользователь с таким телефоном
		var existingUser models.User
		if result := db.Where("phone = ?", req.Phone).First(&existingUser); result.Error == nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Пользователь с таким номером телефона уже существует",
			})
			return
		}

		// Создаем нового пользователя
		user := models.User{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Phone:     req.Phone,
			Role:      "user",
		}

		if result := db.Create(&user); result.Error != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при создании пользователя",
			})
			return
		}

		// Генерируем JWT токен
		token, err := utils.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при создании токена",
			})
			return
		}

		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Token:   token,
			User: models.UserResponse{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Phone:     user.Phone,
				Role:      user.Role,
				CreatedAt: user.CreatedAt,
			},
		})
	}
}

func SendVerificationCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка привязки JSON: %v", err)
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Неверный формат данных",
				Error:   err.Error(),
			})
			return
		}

		// Проверяем формат номера телефона
		if req.Phone == "" {
			log.Printf("Пустой номер телефона")
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Номер телефона не может быть пустым",
			})
			return
		}

		log.Printf("Получен запрос на отправку кода для номера: %s", req.Phone)

		whatsappService := services.NewWhatsAppService()
		code := whatsappService.GenerateVerificationCode()
		log.Printf("Сгенерирован код подтверждения: %s для номера: %s", code, req.Phone)

		// Отправляем код через WhatsApp
		err := whatsappService.SendVerificationCode(req.Phone, code)
		if err != nil {
			log.Printf("Ошибка при отправке кода через WhatsApp: %v", err)
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при отправке кода подтверждения",
				Error:   err.Error(),
			})
			return
		}

		log.Printf("Код успешно отправлен на номер: %s", req.Phone)

		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Message: "Код подтверждения отправлен",
		})
	}
}

func VerifyCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Неверный формат данных",
				Error:   err.Error(),
			})
			return
		}

		// Проверяем код через WhatsApp сервис
		whatsappService := services.NewWhatsAppService()
		isValid, err := whatsappService.VerifyCode(req.Phone, req.Code)
		if err != nil {
			log.Printf("Ошибка при проверке кода: %v", err)
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Ошибка при проверке кода",
				Error:   err.Error(),
			})
			return
		}

		if !isValid {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Неверный код подтверждения",
			})
			return
		}

		// Ищем пользователя по телефону
		var user models.User
		if result := db.Where("phone = ?", req.Phone).First(&user); result.Error != nil {
			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Пользователь не найден",
			})
			return
		}

		// Получаем актуальные данные пользователя
		if err := db.Preload("DriverDocuments").First(&user, user.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при получении данных пользователя",
			})
			return
		}

		// Формируем ответ с данными пользователя
		userResponse := models.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			PhotoUrl:  user.PhotoUrl,
		}

		// Добавляем документы водителя, если они есть
		if user.DriverDocuments != nil {
			userResponse.DriverDocuments = &models.DriverDocumentsResponse{
				ID:                   user.DriverDocuments.ID,
				CarBrand:             user.DriverDocuments.CarBrand,
				CarModel:             user.DriverDocuments.CarModel,
				CarYear:              user.DriverDocuments.CarYear,
				CarColor:             user.DriverDocuments.CarColor,
				CarNumber:            user.DriverDocuments.CarNumber,
				DriverLicenseFront:   user.DriverDocuments.DriverLicenseFront,
				DriverLicenseBack:    user.DriverDocuments.DriverLicenseBack,
				CarRegistrationFront: user.DriverDocuments.CarRegistrationFront,
				CarRegistrationBack:  user.DriverDocuments.CarRegistrationBack,
				Status:               user.DriverDocuments.Status,
				CreatedAt:            user.DriverDocuments.CreatedAt,
				UpdatedAt:            user.DriverDocuments.UpdatedAt,
			}
		}

		// Генерируем JWT токен
		token, err := utils.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при создании токена",
			})
			return
		}

		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Token:   token,
			User:    userResponse,
		})
	}
}

// Получение информации о текущем пользователе
func GetCurrentUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var user models.User
		if err := db.Preload("DriverDocuments").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, AuthResponse{
				Success: false,
				Message: "Пользователь не найден",
			})
			return
		}

		userResponse := models.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			PhotoUrl:  user.PhotoUrl,
		}

		if user.DriverDocuments != nil {
			userResponse.DriverDocuments = &models.DriverDocumentsResponse{
				ID:                   user.DriverDocuments.ID,
				CarBrand:             user.DriverDocuments.CarBrand,
				CarModel:             user.DriverDocuments.CarModel,
				CarYear:              user.DriverDocuments.CarYear,
				CarColor:             user.DriverDocuments.CarColor,
				CarNumber:            user.DriverDocuments.CarNumber,
				DriverLicenseFront:   user.DriverDocuments.DriverLicenseFront,
				DriverLicenseBack:    user.DriverDocuments.DriverLicenseBack,
				CarRegistrationFront: user.DriverDocuments.CarRegistrationFront,
				CarRegistrationBack:  user.DriverDocuments.CarRegistrationBack,
				Status:               user.DriverDocuments.Status,
				CreatedAt:            user.DriverDocuments.CreatedAt,
				UpdatedAt:            user.DriverDocuments.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			User:    userResponse,
		})
	}
}

// UpdateFCMToken обновляет FCM токен пользователя
func UpdateFCMToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FCMToken string `json:"fcmToken" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		userID, _ := c.Get("user_id")

		// Обновляем FCM токен пользователя
		if err := db.Model(&models.User{}).Where("id = ?", userID).Update("fcm_token", req.FCMToken).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении FCM токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "FCM токен успешно обновлен"})
	}
}
