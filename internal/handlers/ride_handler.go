package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"taxi-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Получение активных поездок пользователя
func RideGetActive(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var rides []models.Ride
		if err := db.Where("(passenger_id = ? OR driver_id = ?) AND status IN (?)",
			userID, userID, []string{string(models.RideStatusActive), string(models.RideStatusStarted)}).
			Preload("Passenger").
			Preload("Driver").
			Order("created_at DESC").
			Find(&rides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении поездок"})
			return
		}

		// Формируем ответ
		var response []models.RideResponse
		for _, ride := range rides {
			// Подсчитываем количество забронированных мест
			var bookedSeats int64
			db.Model(&models.Booking{}).
				Where("ride_id = ? AND status = ?", ride.ID, "approved").
				Count(&bookedSeats)

			response = append(response, models.RideResponse{
				ID:             ride.ID,
				PassengerID:    ride.PassengerID,
				DriverID:       ride.DriverID,
				FromAddress:    ride.FromAddress,
				ToAddress:      ride.ToAddress,
				FromLocation:   ride.FromLocation,
				ToLocation:     ride.ToLocation,
				Status:         ride.Status,
				Price:          ride.Price,
				SeatsCount:     ride.SeatsCount,
				BookedSeats:    int(bookedSeats),
				DepartureDate:  ride.DepartureDate,
				Comment:        ride.Comment,
				FrontSeatPrice: ride.FrontSeatPrice,
				BackSeatPrice:  ride.BackSeatPrice,
				CreatedAt:      ride.CreatedAt,
				UpdatedAt:      ride.UpdatedAt,
				PassengerName:  ride.Passenger.FirstName + " " + ride.Passenger.LastName,
				DriverName:     ride.Driver.FirstName + " " + ride.Driver.LastName,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// Получение завершенных поездок пользователя
func RideGetCompleted(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var rides []models.Ride

		if err := db.Where("(passenger_id = ? OR driver_id = ?) AND status = ?",
			userID, userID, models.RideStatusCompleted).
			Preload("Passenger").
			Preload("Driver").
			Order("created_at DESC").
			Find(&rides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении поездок"})
			return
		}

		var response []models.RideResponse
		for _, ride := range rides {
			response = append(response, models.RideResponse{
				ID:             ride.ID,
				PassengerID:    ride.PassengerID,
				DriverID:       ride.DriverID,
				FromAddress:    ride.FromAddress,
				ToAddress:      ride.ToAddress,
				FromLocation:   ride.FromLocation,
				ToLocation:     ride.ToLocation,
				Status:         ride.Status,
				Price:          ride.Price,
				SeatsCount:     ride.SeatsCount,
				DepartureDate:  ride.DepartureDate,
				Comment:        ride.Comment,
				FrontSeatPrice: ride.FrontSeatPrice,
				BackSeatPrice:  ride.BackSeatPrice,
				CreatedAt:      ride.CreatedAt,
				UpdatedAt:      ride.UpdatedAt,
				PassengerName:  ride.Passenger.FirstName + " " + ride.Passenger.LastName,
				DriverName:     ride.Driver.FirstName + " " + ride.Driver.LastName,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// Создание новой поездки
func RideCreate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Сначала выведем информацию о входящем запросе
		body, err := c.GetRawData()
		if err != nil {
			fmt.Printf("DEBUG: Ошибка при чтении тела запроса: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при чтении запроса"})
			return
		}

		// Выведем сырые данные
		fmt.Printf("DEBUG: Полученное сырое тело запроса: %s\n", string(body))

		// Парсим JSON для детального вывода
		var jsonData map[string]interface{}
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			fmt.Printf("DEBUG: Ошибка при разборе JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат JSON"})
			return
		}

		// Проверим тип поля departureDate
		if departureDate, ok := jsonData["departureDate"]; ok {
			fmt.Printf("DEBUG: Тип поля departureDate: %T, значение: %v\n",
				departureDate, departureDate)
		} else {
			fmt.Printf("DEBUG: Поле departureDate отсутствует в запросе\n")
		}

		// Теперь создаем новый Reader из сохраненного тела для c.ShouldBindJSON
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var req struct {
			FromAddress    string          `json:"fromAddress" binding:"required"`
			ToAddress      string          `json:"toAddress" binding:"required"`
			FromLocation   models.Location `json:"fromLocation" binding:"required"`
			ToLocation     models.Location `json:"toLocation" binding:"required"`
			Price          float64         `json:"price" binding:"required"`
			SeatsCount     int             `json:"seatsCount" binding:"required"`
			DepartureDate  time.Time       `json:"departureDate" binding:"required"`
			Comment        string          `json:"comment"`
			FrontSeatPrice float64         `json:"frontSeatPrice"`
			BackSeatPrice  float64         `json:"backSeatPrice"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("DEBUG: Ошибка при разборе JSON в структуру: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		userID, _ := c.Get("user_id")

		// Вывод отладочной информации о получаемой дате
		fmt.Printf("DEBUG: Получена дата отправления: %v, в UTC: %v\n",
			req.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"),
			req.DepartureDate.UTC().Format("2006-01-02 15:04:05 -0700 MST"))

		var (
			frontSeatPrice *float64
			backSeatPrice  *float64
		)

		if req.FrontSeatPrice > 0 {
			fp := req.FrontSeatPrice
			frontSeatPrice = &fp
		}
		if req.BackSeatPrice > 0 {
			bp := req.BackSeatPrice
			backSeatPrice = &bp
		}

		// Создаем поездку с правильным форматом координат для PostgreSQL point
		ride := &models.Ride{
			DriverID:       userID.(uint),
			FromAddress:    req.FromAddress,
			ToAddress:      req.ToAddress,
			FromLocation:   fmt.Sprintf("(%f,%f)", req.FromLocation.Latitude, req.FromLocation.Longitude),
			ToLocation:     fmt.Sprintf("(%f,%f)", req.ToLocation.Latitude, req.ToLocation.Longitude),
			Status:         models.RideStatusActive,
			Price:          req.Price,
			SeatsCount:     req.SeatsCount,
			DepartureDate:  req.DepartureDate, // Используем дату точно как она пришла, без преобразования
			Comment:        req.Comment,
			FrontSeatPrice: frontSeatPrice,
			BackSeatPrice:  backSeatPrice,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		// Дополнительный вывод отладочной информации перед сохранением
		fmt.Printf("DEBUG: Сохраняем поездку с датой отправления: %v\n",
			ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))

		// Создаем запись в базе данных
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(ride).Error; err != nil {
				return fmt.Errorf("ошибка создания поездки: %v", err)
			}
			return nil
		})

		if err != nil {
			fmt.Printf("DEBUG: Ошибка при создании поездки: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании поездки"})
			return
		}

		// Получаем созданную поездку с данными водителя
		if err := db.Preload("Driver").First(ride, ride.ID).Error; err != nil {
			fmt.Printf("DEBUG: Ошибка при получении данных поездки: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении данных поездки"})
			return
		}

		// Дополнительный вывод отладочной информации после сохранения
		fmt.Printf("DEBUG: После сохранения, дата отправления: %v\n",
			ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))

		c.JSON(http.StatusOK, models.RideResponse{
			ID:             ride.ID,
			PassengerID:    ride.PassengerID,
			DriverID:       ride.DriverID,
			FromAddress:    ride.FromAddress,
			ToAddress:      ride.ToAddress,
			FromLocation:   ride.FromLocation,
			ToLocation:     ride.ToLocation,
			Status:         ride.Status,
			Price:          ride.Price,
			SeatsCount:     ride.SeatsCount,
			BookedSeats:    ride.BookedSeats,
			DepartureDate:  ride.DepartureDate,
			Comment:        ride.Comment,
			FrontSeatPrice: ride.FrontSeatPrice,
			BackSeatPrice:  ride.BackSeatPrice,
			CreatedAt:      ride.CreatedAt,
			UpdatedAt:      ride.UpdatedAt,
			DriverName:     ride.Driver.FirstName + " " + ride.Driver.LastName,
		})
	}
}

// Отмена поездки
func RideCancel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Reason string `json:"reason" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		rideID := c.Param("id")
		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, rideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является участником поездки
		currentUserID := userID.(uint)
		var isParticipant bool
		if ride.PassengerID != nil {
			isParticipant = *ride.PassengerID == currentUserID || ride.DriverID == currentUserID
		} else {
			isParticipant = ride.DriverID == currentUserID
		}

		if !isParticipant {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой поездке"})
			return
		}

		// Обновляем статус поездки
		ride.Status = models.RideStatusCancelled
		ride.CancellationReason = req.Reason

		if err := tx.Save(&ride).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при отмене поездки"})
			return
		}

		// Обновляем статусы всех активных бронирований
		if err := tx.Model(&models.Booking{}).
			Where("ride_id = ? AND status IN (?)", rideID, []string{"pending", "approved"}).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"updated_at": time.Now(),
			}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении бронирований"})
			return
		}

		// Удаляем маршрут
		if err := tx.Where("ride_id = ?", rideID).Delete(&models.RoutePoint{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении маршрута"})
			return
		}
		if err := tx.Where("ride_id = ?", rideID).Delete(&models.OptimizedRoute{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении маршрута"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Отправляем WebSocket уведомление водителю
		SendRideStatusUpdate(ride.DriverID, ride.ID, string(ride.Status))

		// Если есть пассажир, отправляем уведомление и ему
		if ride.PassengerID != nil {
			SendRideStatusUpdate(*ride.PassengerID, ride.ID, string(ride.Status))
		}

		// Отправляем уведомления всем, у кого есть активные бронирования
		var bookings []models.Booking
		db.Where("ride_id = ?", rideID).Find(&bookings)
		for _, booking := range bookings {
			SendBookingStatusUpdate(booking.PassengerID, booking.ID, "cancelled")
		}

		c.JSON(http.StatusOK, gin.H{"message": "Поездка успешно отменена"})
	}
}

// Завершение поездки
func RideComplete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rideID := c.Param("id")
		userID, _ := c.Get("user_id")
		var ride models.Ride

		if err := db.First(&ride, rideID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Только водитель может завершить поездку
		if ride.DriverID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Только водитель может завершить поездку"})
			return
		}

		ride.Status = models.RideStatusCompleted

		if err := db.Save(&ride).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при завершении поездки"})
			return
		}

		// Отправляем WebSocket уведомление водителю
		SendRideStatusUpdate(ride.DriverID, ride.ID, string(ride.Status))

		// Если есть пассажир, отправляем уведомление и ему
		if ride.PassengerID != nil {
			SendRideStatusUpdate(*ride.PassengerID, ride.ID, string(ride.Status))
		}

		// Отправляем уведомления всем, у кого есть бронирования для этой поездки
		var bookings []models.Booking
		db.Where("ride_id = ?", rideID).Find(&bookings)
		for _, booking := range bookings {
			SendBookingStatusUpdate(booking.PassengerID, booking.ID, "completed")
		}

		c.JSON(http.StatusOK, gin.H{"message": "Поездка успешно завершена"})
	}
}

// Обновление поездки
func RideUpdate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FromAddress    string          `json:"fromAddress" binding:"required"`
			ToAddress      string          `json:"toAddress" binding:"required"`
			FromLocation   models.Location `json:"fromLocation" binding:"required"`
			ToLocation     models.Location `json:"toLocation" binding:"required"`
			Price          float64         `json:"price" binding:"required"`
			SeatsCount     int             `json:"seatsCount" binding:"required"`
			DepartureDate  time.Time       `json:"departureDate" binding:"required"`
			Comment        string          `json:"comment"`
			FrontSeatPrice float64         `json:"frontSeatPrice"`
			BackSeatPrice  float64         `json:"backSeatPrice"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("DEBUG: Ошибка при разборе JSON в RideUpdate: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		// Вывод отладочной информации о получаемой дате
		fmt.Printf("DEBUG: RideUpdate - Получена дата отправления: %v, в UTC: %v\n",
			req.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"),
			req.DepartureDate.UTC().Format("2006-01-02 15:04:05 -0700 MST"))

		rideID := c.Param("id")
		userID, _ := c.Get("user_id")
		var ride models.Ride

		// Получаем поездку из базы
		if err := db.First(&ride, rideID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является водителем этой поездки
		if ride.DriverID != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Только водитель может редактировать поездку"})
			return
		}

		// Проверяем, что поездка активна
		if ride.Status != models.RideStatusActive {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Можно редактировать только активные поездки"})
			return
		}

		// Проверяем, что дата отправления в будущем
		if req.DepartureDate.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Дата отправления должна быть в будущем"})
			return
		}

		// Создаем поездку с правильной обработкой nullable полей
		var (
			frontSeatPrice *float64
			backSeatPrice  *float64
		)

		// Конвертируем цены только если они указаны
		if req.FrontSeatPrice > 0 {
			fp := req.FrontSeatPrice
			frontSeatPrice = &fp
		}
		if req.BackSeatPrice > 0 {
			bp := req.BackSeatPrice
			backSeatPrice = &bp
		}

		// Отладочная информация о текущей дате поездки перед обновлением
		fmt.Printf("DEBUG: RideUpdate - Текущая дата отправления: %v\n",
			ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))

		// Обновляем поля поездки
		ride.FromAddress = req.FromAddress
		ride.ToAddress = req.ToAddress
		ride.FromLocation = fmt.Sprintf("(%f,%f)", req.FromLocation.Latitude, req.FromLocation.Longitude)
		ride.ToLocation = fmt.Sprintf("(%f,%f)", req.ToLocation.Latitude, req.ToLocation.Longitude)
		ride.Price = req.Price
		ride.SeatsCount = req.SeatsCount
		ride.DepartureDate = req.DepartureDate // Используем дату точно как она пришла, без преобразования
		ride.Comment = req.Comment
		ride.FrontSeatPrice = frontSeatPrice
		ride.BackSeatPrice = backSeatPrice

		// Логируем значения перед сохранением
		fmt.Printf("DEBUG: Updating ride with ID: %d\n", ride.ID)
		fmt.Printf("DEBUG: FrontSeatPrice: %v\n", ride.FrontSeatPrice)
		fmt.Printf("DEBUG: BackSeatPrice: %v\n", ride.BackSeatPrice)
		fmt.Printf("DEBUG: DepartureDate: %v\n", ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))

		// Сохраняем изменения
		if err := db.Save(&ride).Error; err != nil {
			fmt.Printf("DEBUG: RideUpdate - Ошибка при сохранении: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении поездки"})
			return
		}

		// Получаем обновленную поездку с данными водителя
		if err := db.Preload("Driver").First(&ride, ride.ID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении данных поездки"})
			return
		}

		// Логируем значения после получения из базы
		fmt.Printf("DEBUG: Retrieved updated ride with ID: %d\n", ride.ID)
		fmt.Printf("DEBUG: FrontSeatPrice after retrieval: %v\n", ride.FrontSeatPrice)
		fmt.Printf("DEBUG: BackSeatPrice after retrieval: %v\n", ride.BackSeatPrice)
		fmt.Printf("DEBUG: DepartureDate after retrieval: %v\n", ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))

		c.JSON(http.StatusOK, models.RideResponse{
			ID:             ride.ID,
			PassengerID:    ride.PassengerID,
			DriverID:       ride.DriverID,
			FromAddress:    ride.FromAddress,
			ToAddress:      ride.ToAddress,
			FromLocation:   ride.FromLocation,
			ToLocation:     ride.ToLocation,
			Status:         ride.Status,
			Price:          ride.Price,
			SeatsCount:     ride.SeatsCount,
			BookedSeats:    ride.BookedSeats,
			DepartureDate:  ride.DepartureDate,
			Comment:        ride.Comment,
			FrontSeatPrice: ride.FrontSeatPrice,
			BackSeatPrice:  ride.BackSeatPrice,
			CreatedAt:      ride.CreatedAt,
			UpdatedAt:      ride.UpdatedAt,
			DriverName:     ride.Driver.FirstName + " " + ride.Driver.LastName,
		})
	}
}

// Поиск поездок для пассажиров
func RideSearch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FromCity      string    `json:"fromCity" binding:"required"`
			ToCity        string    `json:"toCity" binding:"required"`
			DepartureDate time.Time `json:"departureDate"`
			SeatsCount    int       `json:"seatsCount"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("DEBUG: RideSearch - Ошибка при разборе JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		// Для отладки выводим дату в том виде, в котором она пришла от клиента
		if !req.DepartureDate.IsZero() {
			fmt.Printf("DEBUG: RideSearch - Получена дата отправления: %v\n",
				req.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))
		} else {
			fmt.Printf("DEBUG: RideSearch - Дата отправления не указана\n")
		}

		fmt.Printf("DEBUG: Searching rides from %s to %s on %s for %d seats\n",
			req.FromCity, req.ToCity, req.DepartureDate.Format("2006-01-02"), req.SeatsCount)

		// Используем обычный запрос без транзакции для поиска
		query := db.Where("status IN (?)", []string{string(models.RideStatusActive), string(models.RideStatusStarted)})

		// Поиск по городам отправления и назначения - используем более гибкий поиск
		// Ищем поездки, где адрес содержит название города, а не начинается с него
		query = query.Where("from_address LIKE ? AND to_address LIKE ?",
			"%"+req.FromCity+"%", "%"+req.ToCity+"%")

		// Если указана дата отправления, ищем поездки на эту дату или позже
		if !req.DepartureDate.IsZero() {
			// Устанавливаем начало дня в том же часовом поясе, что и полученная дата
			startOfDay := time.Date(
				req.DepartureDate.Year(),
				req.DepartureDate.Month(),
				req.DepartureDate.Day(),
				0, 0, 0, 0,
				req.DepartureDate.Location(),
			)

			// Выводим для отладки время начала дня
			fmt.Printf("DEBUG: RideSearch - Начало дня для поиска: %v\n",
				startOfDay.Format("2006-01-02 15:04:05 -0700 MST"))

			// Ищем поездки на указанную дату или позже (в течение недели)
			endOfWeek := startOfDay.Add(7 * 24 * time.Hour)

			// Выводим для отладки конец периода поиска
			fmt.Printf("DEBUG: RideSearch - Конец периода поиска: %v\n",
				endOfWeek.Format("2006-01-02 15:04:05 -0700 MST"))

			query = query.Where("departure_date BETWEEN ? AND ?", startOfDay, endOfWeek)
		}

		// Загружаем данные о водителе и его документах
		query = query.Preload("Driver").Preload("Driver.DriverDocuments")

		// Получаем результаты
		var rides []models.Ride
		if err := query.Order("departure_date ASC").Find(&rides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске поездок"})
			return
		}

		fmt.Printf("DEBUG: Found %d rides\n", len(rides))

		// Для отладки выводим информацию о найденных поездках
		for i, ride := range rides {
			fmt.Printf("DEBUG: RideSearch - Найдена поездка #%d, ID: %d, Дата: %v\n",
				i+1, ride.ID, ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))
		}

		// Если поездок не найдено, попробуем более широкий поиск
		if len(rides) == 0 {
			// Сбрасываем запрос и ищем только по городам, без учета даты и количества мест
			query = db.Where("status IN (?)", []string{string(models.RideStatusActive), string(models.RideStatusStarted)})
			query = query.Where("from_address LIKE ? AND to_address LIKE ?",
				"%"+req.FromCity+"%", "%"+req.ToCity+"%")
			query = query.Preload("Driver").Preload("Driver.DriverDocuments")

			if err := query.Order("departure_date ASC").Find(&rides).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске поездок"})
				return
			}

			fmt.Printf("DEBUG: Found %d rides after broader search\n", len(rides))

			// Для отладки выводим информацию о найденных поездках при широком поиске
			for i, ride := range rides {
				fmt.Printf("DEBUG: RideSearch (широкий поиск) - Найдена поездка #%d, ID: %d, Дата: %v\n",
					i+1, ride.ID, ride.DepartureDate.Format("2006-01-02 15:04:05 -0700 MST"))
			}
		}

		// Получаем и обрабатываем информацию о забронированных местах поодиночке,
		// без обновления базы данных внутри этого запроса
		for i := range rides {
			var totalBookedSeats int
			if err := db.Model(&models.Booking{}).
				Where("ride_id = ? AND status = ?", rides[i].ID, "approved").
				Select("COALESCE(SUM(seats_count), 0)").
				Scan(&totalBookedSeats).Error; err != nil {
				// Логируем ошибку, но продолжаем выполнение
				fmt.Printf("ERROR: Failed to count booked seats for ride %d: %v\n", rides[i].ID, err)
				// Установим значение по умолчанию
				totalBookedSeats = 0
			}

			// Просто обновляем поле в памяти, без записи в базу данных
			rides[i].BookedSeats = totalBookedSeats
		}

		// Если указано количество мест, фильтруем поездки с достаточным количеством свободных мест
		var filteredRides []models.Ride
		if req.SeatsCount > 0 {
			for _, ride := range rides {
				if ride.SeatsCount-ride.BookedSeats >= req.SeatsCount {
					filteredRides = append(filteredRides, ride)
				}
			}
		} else {
			filteredRides = rides
		}

		// Формируем ответ с расширенной информацией
		var response []map[string]interface{}
		for _, ride := range filteredRides {
			// Базовая информация о поездке
			rideData := map[string]interface{}{
				"id":               ride.ID,
				"from_address":     ride.FromAddress,
				"to_address":       ride.ToAddress,
				"from_location":    ride.FromLocation,
				"to_location":      ride.ToLocation,
				"price":            ride.Price,
				"seats_count":      ride.SeatsCount,
				"booked_seats":     ride.BookedSeats,
				"departure_date":   ride.DepartureDate,
				"comment":          ride.Comment,
				"front_seat_price": ride.FrontSeatPrice,
				"back_seat_price":  ride.BackSeatPrice,
				"status":           string(ride.Status),
			}

			// Информация о водителе
			driverData := map[string]interface{}{
				"id":        ride.Driver.ID,
				"full_name": ride.Driver.FirstName + " " + ride.Driver.LastName,
				"phone":     ride.Driver.Phone,
			}

			// Информация об автомобиле, если есть документы водителя
			carData := map[string]interface{}{}
			if ride.Driver.DriverDocuments != nil {
				carData = map[string]interface{}{
					"car_brand":          ride.Driver.DriverDocuments.CarBrand,
					"car_model":          ride.Driver.DriverDocuments.CarModel,
					"car_year":           ride.Driver.DriverDocuments.CarYear,
					"car_color":          ride.Driver.DriverDocuments.CarColor,
					"car_number":         ride.Driver.DriverDocuments.CarNumber,
					"car_photo_front":    ride.Driver.DriverDocuments.CarPhotoFront,
					"car_photo_side":     ride.Driver.DriverDocuments.CarPhotoSide,
					"car_photo_interior": ride.Driver.DriverDocuments.CarPhotoInterior,
				}
			}

			// Объединяем все данные
			rideData["driver"] = driverData
			rideData["car"] = carData

			response = append(response, rideData)
		}

		c.JSON(http.StatusOK, response)
	}
}

// Получение поездки по ID
func RideGetByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ride models.Ride
		rideID := c.Param("id")

		// Начинаем транзакцию
		tx := db.Begin()

		// Получаем поездку с информацией о водителе
		if err := tx.Preload("Driver").First(&ride, rideID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Подсчитываем количество забронированных мест
		var bookedSeats int64
		if err := tx.Model(&models.Booking{}).Where("ride_id = ? AND status IN ?", rideID, []string{"approved", "pending"}).Count(&bookedSeats).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Завершаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Формируем ответ
		response := gin.H{
			"id":               ride.ID,
			"from_address":     ride.FromAddress,
			"to_address":       ride.ToAddress,
			"from_location":    ride.FromLocation,
			"to_location":      ride.ToLocation,
			"price":            ride.Price,
			"seats_count":      ride.SeatsCount,
			"booked_seats":     bookedSeats,
			"departure_date":   ride.DepartureDate,
			"comment":          ride.Comment,
			"front_seat_price": ride.FrontSeatPrice,
			"back_seat_price":  ride.BackSeatPrice,
			"status":           string(ride.Status),
			"driver": gin.H{
				"id":        ride.Driver.ID,
				"full_name": ride.Driver.FirstName + " " + ride.Driver.LastName,
				"phone":     ride.Driver.Phone,
			},
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetCancelledRides возвращает список отмененных поездок
func GetCancelledRides(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var rides []models.Ride

		query := db.Where("status = ?", models.RideStatusCancelled)

		// Если пользователь не админ, показываем только его поездки
		if c.GetString("user_role") != "admin" {
			query = query.Where("driver_id = ?", userID)
		}

		if err := query.Order("created_at DESC").Find(&rides).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rides)
	}
}

// Начало поездки
func RideStart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rideID := c.Param("id")
		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, rideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		fmt.Printf("DEBUG: Начало поездки %d, текущий статус: %s\n", ride.ID, ride.Status)

		// Только водитель может начать поездку
		if ride.DriverID != userID.(uint) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "Только водитель может начать поездку"})
			return
		}

		// Проверяем, что поездка в статусе active
		if ride.Status != models.RideStatusActive {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Поездка не может быть начата"})
			return
		}

		// Обновляем статус поездки на started
		if err := tx.Model(&ride).Update("status", models.RideStatusStarted).Error; err != nil {
			tx.Rollback()
			fmt.Printf("DEBUG: Ошибка при обновлении статуса: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале поездки"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			fmt.Printf("DEBUG: Ошибка при сохранении изменений: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Отправляем WebSocket уведомление водителю
		SendRideStatusUpdate(ride.DriverID, ride.ID, string(models.RideStatusStarted))

		// Если есть пассажир, отправляем уведомление и ему
		if ride.PassengerID != nil {
			SendRideStatusUpdate(*ride.PassengerID, ride.ID, string(models.RideStatusStarted))
		}

		// Отправляем уведомления всем, у кого есть бронирования для этой поездки
		var bookings []models.Booking
		db.Where("ride_id = ?", rideID).Find(&bookings)
		for _, booking := range bookings {
			SendBookingStatusUpdate(booking.PassengerID, booking.ID, "started")
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Поездка успешно начата",
			"status":  string(models.RideStatusStarted),
		})
	}
}

// UpdateRideBookedSeats обновляет количество забронированных мест для поездки
func UpdateRideBookedSeats(db *gorm.DB, rideID uint) (int, error) {
	var totalBookedSeats int

	// Получаем сумму забронированных мест только для подтвержденных бронирований
	if err := db.Model(&models.Booking{}).
		Where("ride_id = ? AND status = ?", rideID, models.BookingStatusApproved).
		Select("COALESCE(SUM(seats_count), 0)").
		Scan(&totalBookedSeats).Error; err != nil {
		return 0, err
	}

	// Обновляем поездку с новым количеством мест
	if err := db.Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(map[string]interface{}{
			"booked_seats": totalBookedSeats,
			"updated_at":   time.Now(),
		}).Error; err != nil {
		return 0, err
	}

	return totalBookedSeats, nil
}
