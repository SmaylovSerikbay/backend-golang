package handlers

import (
	"log"
	"net/http"
	"taxi-backend/internal/models"
	"taxi-backend/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetOptimizedRoute возвращает оптимизированный маршрут для поездки
func GetOptimizedRoute(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rideID := c.Param("id")
		var ride models.Ride

		log.Printf("Начало обработки маршрута для поездки %s", rideID)

		// Получаем поездку
		if err := db.First(&ride, rideID).Error; err != nil {
			log.Printf("Ошибка при получении поездки: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}
		log.Printf("Поездка найдена: %+v", ride)

		// Получаем все бронирования для этой поездки
		var bookings []models.Booking
		if err := db.Where("ride_id = ?", rideID).Find(&bookings).Error; err != nil {
			log.Printf("Ошибка при получении бронирований: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении бронирований"})
			return
		}
		log.Printf("Найдено бронирований: %d", len(bookings))

		// Создаем оптимизатор маршрута
		optimizer := services.NewRouteOptimizer("")
		log.Printf("Оптимизатор маршрута создан")

		// Получаем оптимизированный маршрут
		route, err := optimizer.OptimizeRoute(&ride, bookings)
		if err != nil {
			log.Printf("Ошибка при оптимизации маршрута: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при оптимизации маршрута"})
			return
		}
		log.Printf("Маршрут оптимизирован, количество точек: %d", len(route.Points))

		// Сохраняем маршрут в базе данных
		if err := db.Transaction(func(tx *gorm.DB) error {
			log.Printf("Начало транзакции")

			// Проверяем существование старого маршрута
			var existingRoute models.OptimizedRoute
			if err := tx.Where("ride_id = ?", ride.ID).First(&existingRoute).Error; err == nil {
				log.Printf("Найден существующий маршрут ID: %d", existingRoute.ID)
				// Если маршрут существует, обновляем его
				existingRoute.Distance = route.Distance
				existingRoute.Duration = route.Duration
				if err := tx.Save(&existingRoute).Error; err != nil {
					log.Printf("Ошибка при обновлении существующего маршрута: %v", err)
					return err
				}
				route.ID = existingRoute.ID
				log.Printf("Существующий маршрут обновлен")
			} else {
				log.Printf("Создание нового маршрута")
				// Если маршрута нет, создаем новый
				if err := tx.Create(route).Error; err != nil {
					log.Printf("Ошибка при создании нового маршрута: %v", err)
					return err
				}
				log.Printf("Новый маршрут создан, ID: %d", route.ID)
			}

			// Удаляем все старые точки маршрута
			if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.RoutePoint{}).Error; err != nil {
				log.Printf("Ошибка при удалении старых точек: %v", err)
				return err
			}
			log.Printf("Старые точки маршрута удалены")

			// Создаем новый массив точек, начиная с точки водителя
			allPoints := make([]models.RoutePoint, len(route.Points)+1)

			// Получаем текущее местоположение водителя
			var tracking models.RideTracking
			if err := tx.Where("ride_id = ?", ride.ID).First(&tracking).Error; err == nil {
				log.Printf("Найдено текущее местоположение водителя: %+v", tracking.CurrentLocation)
				allPoints[0] = models.RoutePoint{
					OrderNum:  1,
					RideID:    ride.ID,
					RouteID:   route.ID,
					Type:      "start",
					Address:   "Текущее местоположение",
					Latitude:  tracking.CurrentLocation.Latitude,
					Longitude: tracking.CurrentLocation.Longitude,
					Time:      time.Now(),
				}
			} else {
				log.Printf("Местоположение водителя не найдено, используем начальную точку маршрута")
				fromLoc := services.ParseLocation(ride.FromLocation)
				allPoints[0] = models.RoutePoint{
					OrderNum:  1,
					RideID:    ride.ID,
					RouteID:   route.ID,
					Type:      "start",
					Address:   "Текущее местоположение",
					Latitude:  fromLoc.Latitude,
					Longitude: fromLoc.Longitude,
					Time:      time.Now(),
				}
			}

			// Копируем остальные точки с обновленными номерами
			for i := range route.Points {
				route.Points[i].RouteID = route.ID
				route.Points[i].ID = 0
				route.Points[i].OrderNum = i + 2
				allPoints[i+1] = route.Points[i]
			}
			log.Printf("Подготовлено точек для сохранения: %d", len(allPoints))

			// Сохраняем все точки одним запросом
			if err := tx.Create(&allPoints).Error; err != nil {
				log.Printf("Ошибка при сохранении точек маршрута: %v", err)
				return err
			}
			log.Printf("Точки маршрута сохранены успешно")

			return nil
		}); err != nil {
			log.Printf("Ошибка в транзакции: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении маршрута"})
			return
		}

		// Загружаем сохраненный маршрут с точками
		var savedRoute models.OptimizedRoute
		if err := db.Preload("Points", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_num ASC")
		}).First(&savedRoute, route.ID).Error; err != nil {
			log.Printf("Ошибка при загрузке сохраненного маршрута: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении сохраненного маршрута"})
			return
		}
		log.Printf("Маршрут успешно загружен и готов к отправке")

		c.JSON(http.StatusOK, savedRoute)
	}
}
