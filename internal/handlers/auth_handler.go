package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"taxi-backend/internal/models"
	"taxi-backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Phone     string `json:"phone" binding:"required,e164"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Token   string              `json:"token,omitempty"`
	User    models.UserResponse `json:"user,omitempty"`
}

func AuthRegister(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Логируем тело запроса
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		log.Printf("Тело запроса: %s", string(body))

		// Логируем заголовки
		log.Printf("Заголовки запроса: %v", c.Request.Header)

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			log.Printf("Полученные данные: %+v", req)
			log.Printf("Тип ошибки: %T", err)

			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: fmt.Sprintf("Неверный формат данных: %v", err),
			})
			return
		}

		// Логируем успешно распарсенные данные
		log.Printf("Успешно распарсенные данные: %+v", req)

		// Проверяем, существует ли пользователь с таким email
		var existingUser models.User
		if result := db.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Пользователь с таким email уже существует",
			})
			return
		}

		// Проверяем, существует ли пользователь с таким телефоном
		if result := db.Where("phone = ?", req.Phone).First(&existingUser); result.Error == nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Пользователь с таким номером телефона уже существует",
			})
			return
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при создании пользователя",
			})
			return
		}

		// Создаем нового пользователя
		user := models.User{
			Email:     req.Email,
			Password:  string(hashedPassword),
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
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Phone:     user.Phone,
				Role:      user.Role,
				CreatedAt: user.CreatedAt,
			},
		})
	}
}

func AuthLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "Неверный формат данных",
			})
			return
		}

		// Ищем пользователя
		var user models.User
		if result := db.Where("email = ?", req.Email).First(&user); result.Error != nil {
			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Неверный email или пароль",
			})
			return
		}

		// Проверяем пароль
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Неверный email или пароль",
			})
			return
		}

		// Получаем актуальные данные пользователя с предзагрузкой всех связанных данных
		if err := db.Preload("DriverDocuments").First(&user, user.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Ошибка при получении данных пользователя",
			})
			return
		}

		log.Printf("Данные пользователя при входе: %+v", user)
		log.Printf("PhotoUrl пользователя при входе: '%s'", user.PhotoUrl)

		// Если PhotoUrl пустой, пробуем получить его из базы данных
		if user.PhotoUrl == "" {
			var photoUrl string
			err := db.Raw("SELECT photo_url FROM users WHERE id = ? AND photo_url IS NOT NULL AND photo_url != ''", user.ID).Scan(&photoUrl).Error
			if err == nil && photoUrl != "" {
				user.PhotoUrl = photoUrl
				// Обновляем photo_url в базе данных
				if err := db.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoUrl, user.ID).Error; err != nil {
					log.Printf("Ошибка при обновлении photo_url: %v", err)
				} else {
					log.Printf("Успешно обновлен photo_url: %s", photoUrl)
				}
			}
		}

		// Формируем ответ с данными пользователя
		userResponse := models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
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

		log.Printf("Отправляем ответ при входе: %+v", userResponse)

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
		role, _ := c.Get("role")

		// Если это админ, возвращаем специальный ответ
		if role == "admin" {
			c.JSON(http.StatusOK, models.UserResponse{
				ID:        0,
				Email:     "admin@admin.com",
				FirstName: "Admin",
				LastName:  "",
				Phone:     "",
				Role:      "admin",
				CreatedAt: time.Now(),
			})
			return
		}

		var user models.User

		// Получаем пользователя со всеми связанными данными
		if err := db.Preload("DriverDocuments").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		log.Printf("GetCurrentUser: Данные пользователя: %+v", user)
		log.Printf("GetCurrentUser: PhotoUrl пользователя: '%s'", user.PhotoUrl)

		// Если PhotoUrl пустой, пробуем получить его из базы данных
		if user.PhotoUrl == "" {
			var photoUrl string
			err := db.Raw("SELECT photo_url FROM users WHERE id = ? AND photo_url IS NOT NULL AND photo_url != ''", userID).Scan(&photoUrl).Error
			if err == nil && photoUrl != "" {
				user.PhotoUrl = photoUrl
				// Обновляем photo_url в базе данных
				if err := db.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoUrl, userID).Error; err != nil {
					log.Printf("GetCurrentUser: Ошибка при обновлении photo_url: %v", err)
				} else {
					log.Printf("GetCurrentUser: Успешно обновлен photo_url: %s", photoUrl)
				}
			}
		}

		response := models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			PhotoUrl:  user.PhotoUrl,
		}

		// Добавляем документы водителя, если они есть
		if user.DriverDocuments != nil {
			response.DriverDocuments = &models.DriverDocumentsResponse{
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

		log.Printf("GetCurrentUser: Отправляем ответ: %+v", response)
		c.JSON(http.StatusOK, response)
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
