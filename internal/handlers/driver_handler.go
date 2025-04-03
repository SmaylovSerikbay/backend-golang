package handlers

import (
	"net/http"
	"taxi-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DriverGetNearby(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Ближайшие водители"})
	}
}

func DriverUpdateLocation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Latitude  float64 `json:"latitude" binding:"required"`
			Longitude float64 `json:"longitude" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		// Получаем ID пользователя (водителя) из контекста
		driverID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
			return
		}

		// Обновляем местоположение водителя в базе данных
		// (эта часть может быть реализована в зависимости от структуры БД)

		// Получаем активные поездки этого водителя
		var activeRides []models.Ride
		if err := db.Where("driver_id = ? AND status IN (?)", driverID, []string{string(models.RideStatusActive), string(models.RideStatusStarted)}).Find(&activeRides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении поездок"})
			return
		}

		// Отправляем WebSocket уведомления пассажирам активных поездок
		for _, ride := range activeRides {
			if ride.PassengerID != nil {
				SendDriverLocationUpdate(*ride.PassengerID, driverID.(uint), req.Latitude, req.Longitude)
			}

			// Также отправляем уведомления всем пассажирам с бронированиями
			var bookings []models.Booking
			if err := db.Where("ride_id = ? AND status IN (?)", ride.ID, []string{"approved", "started"}).Find(&bookings).Error; err != nil {
				continue
			}

			for _, booking := range bookings {
				SendDriverLocationUpdate(booking.PassengerID, driverID.(uint), req.Latitude, req.Longitude)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Локация обновлена"})
	}
}
