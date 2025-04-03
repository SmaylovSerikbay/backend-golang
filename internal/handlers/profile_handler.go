package handlers

import (
	"net/http"
	"taxi-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UserGetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		role, _ := c.Get("role")

		// Если это админ, возвращаем специальный ответ
		if role == "admin" {
			c.JSON(http.StatusOK, models.UserResponse{
				ID:        0,
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
			return
		}

		// Формируем ответ
		response := models.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			PhotoUrl:  user.PhotoUrl,
			FCMToken:  user.FCMToken,
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

		c.JSON(http.StatusOK, response)
	}
}

func UserUpdateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var user models.User

		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}

		var req struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			PhotoUrl  string `json:"photoUrl"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		// Обновляем только разрешенные поля
		updates := map[string]interface{}{}

		if req.FirstName != "" {
			updates["first_name"] = req.FirstName
			user.FirstName = req.FirstName
		}
		if req.LastName != "" {
			updates["last_name"] = req.LastName
			user.LastName = req.LastName
		}
		if req.PhotoUrl != "" {
			// Убеждаемся, что URL начинается с /
			photoUrl := req.PhotoUrl
			if photoUrl != "" && photoUrl[0] != '/' {
				photoUrl = "/" + photoUrl
			}
			updates["photo_url"] = photoUrl
			user.PhotoUrl = photoUrl
		}

		// Обновляем поля
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении профиля"})
			return
		}

		// Получаем обновленные данные пользователя
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении обновленных данных"})
			return
		}

		// Формируем ответ
		response := models.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			PhotoUrl:  user.PhotoUrl,
		}

		c.JSON(http.StatusOK, response)
	}
}
