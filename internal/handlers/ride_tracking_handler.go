package handlers

import (
	"errors"
	"net/http"
	"taxi-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RideTrackingUpdate обновляет статус и местоположение поездки
func RideTrackingUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		rideID := c.Param("id")

		// Проверяем, является ли пользователь водителем этой поездки
		var ride models.Ride
		if err := db.First(&ride, rideID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		if ride.DriverID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Только водитель может обновлять статус поездки"})
			return
		}

		var update models.RideTrackingUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Начинаем транзакцию
		tx := db.Begin()

		// Получаем или создаем запись отслеживания
		var tracking models.RideTracking
		result := tx.Where("ride_id = ?", rideID).First(&tracking)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			tracking = models.RideTracking{
				RideID: ride.ID,
				Status: "started",
			}
		}

		// Обновляем местоположение и статус
		tracking.CurrentLocation.Latitude = update.Latitude
		tracking.CurrentLocation.Longitude = update.Longitude
		tracking.UpdatedAt = time.Now()

		if update.Status != "" {
			tracking.Status = update.Status
		}
		if update.BookingID != 0 {
			tracking.CurrentBookingID = update.BookingID
		}

		// Рассчитываем примерное время до следующей точки
		// TODO: Добавить реальный расчет времени через сервис маршрутизации
		tracking.EstimatedTime = 15 // Пока используем фиксированное значение

		if result.Error != nil {
			if err := tx.Create(&tracking).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании записи отслеживания"})
				return
			}
		} else {
			if err := tx.Save(&tracking).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении записи отслеживания"})
				return
			}
		}

		// Фиксируем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		c.JSON(http.StatusOK, tracking)
	}
}

// RideTrackingGet получает текущий статус поездки
func RideTrackingGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rideID := c.Param("id")

		var tracking models.RideTracking
		if err := db.Where("ride_id = ?", rideID).
			Preload("Ride").
			Preload("Booking").
			First(&tracking).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Информация об отслеживании не найдена"})
			return
		}

		c.JSON(http.StatusOK, tracking)
	}
}
