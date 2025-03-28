package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"taxi-backend/internal/models"
	"taxi-backend/internal/services"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var firebaseService = services.NewFirebaseService()

// BookingCreate создает новое бронирование
func BookingCreate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RideID          uint            `json:"rideId" binding:"required"`
			PickupAddress   string          `json:"pickupAddress" binding:"required"`
			DropoffAddress  string          `json:"dropoffAddress" binding:"required"`
			PickupLocation  models.Location `json:"pickupLocation" binding:"required"`
			DropoffLocation models.Location `json:"dropoffLocation" binding:"required"`
			SeatsCount      int             `json:"seatsCount" binding:"required"`
			BookingType     string          `json:"bookingType" binding:"required"`
			Comment         string          `json:"comment"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		userID, _ := c.Get("user_id")
		passengerID := userID.(uint)

		// Проверяем существование поездки
		var ride models.Ride
		if err := db.First(&ride, req.RideID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь не бронирует свою же поездку
		if ride.DriverID == passengerID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Вы не можете забронировать свою собственную поездку"})
			return
		}

		// Проверяем, что поездка активна
		if ride.Status != models.RideStatusActive && ride.Status != models.RideStatusStarted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Поездка недоступна для бронирования"})
			return
		}

		// Проверяем, что у поездки достаточно свободных мест
		// Получаем сумму забронированных мест для этой поездки
		var totalBookedSeats int
		if err := db.Model(&models.Booking{}).
			Where("ride_id = ? AND status = ?",
				req.RideID,
				models.BookingStatusApproved).
			Select("COALESCE(SUM(seats_count), 0)").
			Scan(&totalBookedSeats).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при подсчете забронированных мест"})
			return
		}

		// Проверяем, что запрашиваемое количество мест доступно
		availableSeats := ride.SeatsCount - totalBookedSeats
		if req.SeatsCount > availableSeats {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно свободных мест"})
			return
		}

		// Определяем тип бронирования и цену
		var bookingType models.BookingType
		var price float64

		switch req.BookingType {
		case "regular":
			bookingType = models.BookingTypeRegular
			price = ride.Price * float64(req.SeatsCount)
		case "front_seat":
			bookingType = models.BookingTypeFrontSeat
			if ride.FrontSeatPrice == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование всего салона недоступно для этой поездки"})
				return
			}
			price = *ride.FrontSeatPrice
		case "back_seat":
			bookingType = models.BookingTypeBackSeat
			if ride.BackSeatPrice == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование заднего ряда недоступно для этой поездки"})
				return
			}
			price = *ride.BackSeatPrice
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный тип бронирования"})
			return
		}

		// Создаем бронирование
		booking := models.Booking{
			RideID:          req.RideID,
			PassengerID:     passengerID,
			PickupAddress:   req.PickupAddress,
			DropoffAddress:  req.DropoffAddress,
			PickupLocation:  fmt.Sprintf("(%f,%f)", req.PickupLocation.Longitude, req.PickupLocation.Latitude),
			DropoffLocation: fmt.Sprintf("(%f,%f)", req.DropoffLocation.Longitude, req.DropoffLocation.Latitude),
			SeatsCount:      req.SeatsCount,
			Status:          models.BookingStatusPending,
			BookingType:     bookingType,
			Price:           price,
			Comment:         req.Comment,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := db.Create(&booking).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании бронирования"})
			return
		}

		// Загружаем информацию о пассажире
		var passenger models.User
		db.First(&passenger, passengerID)

		// Формируем ответ
		response := models.BookingResponse{
			ID:              booking.ID,
			RideID:          booking.RideID,
			PassengerID:     booking.PassengerID,
			PickupAddress:   booking.PickupAddress,
			DropoffAddress:  booking.DropoffAddress,
			PickupLocation:  booking.PickupLocation,
			DropoffLocation: booking.DropoffLocation,
			SeatsCount:      booking.SeatsCount,
			Status:          booking.Status,
			BookingType:     booking.BookingType,
			Price:           booking.Price,
			Comment:         booking.Comment,
			CreatedAt:       booking.CreatedAt,
			UpdatedAt:       booking.UpdatedAt,
			PassengerName:   passenger.FirstName + " " + passenger.LastName,
			PassengerPhone:  passenger.Phone,
		}

		// После успешного создания бронирования отправляем уведомление водителю
		var driver models.User
		if err := db.First(&driver, ride.DriverID).Error; err == nil && driver.FCMToken != "" {
			notificationData := map[string]interface{}{
				"type":       "new_booking",
				"booking_id": booking.ID,
				"ride_id":    booking.RideID,
			}
			firebaseService.SendNotification(
				driver.FCMToken,
				"Новое бронирование",
				fmt.Sprintf("Пассажир %s хочет забронировать %d мест", passenger.FirstName+" "+passenger.LastName, req.SeatsCount),
				notificationData,
			)
		}

		c.JSON(http.StatusCreated, response)
	}
}

// BookingGetByUser получает бронирования пользователя
func BookingGetByUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var bookings []models.Booking

		// Получаем бронирования, где пользователь является пассажиром
		if err := db.Where("passenger_id = ?", userID).
			Order("created_at DESC").
			Find(&bookings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении бронирований"})
			return
		}

		// Если бронирований нет, возвращаем пустой массив
		if len(bookings) == 0 {
			c.JSON(http.StatusOK, []models.BookingResponse{})
			return
		}

		// Формируем ответ
		var response []models.BookingResponse
		for _, booking := range bookings {
			// Загружаем информацию о пассажире
			var passenger models.User
			db.First(&passenger, booking.PassengerID)

			// Загружаем информацию о поездке
			var ride models.Ride
			db.First(&ride, booking.RideID)

			// Загружаем информацию о водителе
			var driver models.User
			db.First(&driver, ride.DriverID)

			// Формируем ответ для поездки
			rideResponse := models.RideResponse{
				ID:             ride.ID,
				DriverID:       ride.DriverID,
				FromAddress:    ride.FromAddress,
				ToAddress:      ride.ToAddress,
				FromLocation:   ride.FromLocation,
				ToLocation:     ride.ToLocation,
				Status:         ride.Status,
				Price:          ride.Price,
				SeatsCount:     ride.SeatsCount,
				DepartureDate:  ride.DepartureDate,
				FrontSeatPrice: ride.FrontSeatPrice,
				BackSeatPrice:  ride.BackSeatPrice,
				DriverName:     driver.FirstName + " " + driver.LastName,
			}

			// Добавляем бронирование в ответ
			response = append(response, models.BookingResponse{
				ID:              booking.ID,
				RideID:          booking.RideID,
				PassengerID:     booking.PassengerID,
				PickupAddress:   booking.PickupAddress,
				DropoffAddress:  booking.DropoffAddress,
				PickupLocation:  booking.PickupLocation,
				DropoffLocation: booking.DropoffLocation,
				SeatsCount:      booking.SeatsCount,
				Status:          booking.Status,
				BookingType:     booking.BookingType,
				Price:           booking.Price,
				Comment:         booking.Comment,
				RejectReason:    booking.RejectReason,
				CreatedAt:       booking.CreatedAt,
				UpdatedAt:       booking.UpdatedAt,
				PassengerName:   passenger.FirstName + " " + passenger.LastName,
				PassengerPhone:  passenger.Phone,
				RideInfo:        rideResponse,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// BookingGetByRide получает бронирования для поездки
func BookingGetByRide(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rideID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID поездки"})
			return
		}

		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		// Проверяем, что пользователь является водителем этой поездки или имеет бронирование в этой поездке
		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, rideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, является ли пользователь водителем или имеет бронирование
		var hasBooking bool
		err = tx.Model(&models.Booking{}).
			Where("ride_id = ? AND passenger_id = ?", rideID, userID).
			Select("COUNT(*) > 0").
			Scan(&hasBooking).Error

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при проверке бронирования"})
			return
		}

		if ride.DriverID != userID.(uint) && !hasBooking {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет доступа к этой поездке"})
			return
		}

		// Получаем сумму всех подтвержденных бронирований для этой поездки
		var totalBookedSeats int
		if err := tx.Model(&models.Booking{}).
			Where("ride_id = ? AND status = ?", rideID, models.BookingStatusApproved).
			Select("COALESCE(SUM(seats_count), 0)").
			Scan(&totalBookedSeats).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при подсчете забронированных мест"})
			return
		}

		// Обновляем количество забронированных мест в поездке
		if err := tx.Model(&ride).Updates(map[string]interface{}{
			"booked_seats": totalBookedSeats,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении количества мест"})
			return
		}

		// Получаем бронирования для этой поездки
		var bookings []models.Booking
		if err := tx.Where("ride_id = ?", rideID).
			Order("created_at DESC").
			Find(&bookings).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении бронирований"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Если бронирований нет, возвращаем пустой массив
		if len(bookings) == 0 {
			c.JSON(http.StatusOK, []models.BookingResponse{})
			return
		}

		// Формируем ответ
		var response []models.BookingResponse
		for _, booking := range bookings {
			// Загружаем информацию о пассажире
			var passenger models.User
			db.First(&passenger, booking.PassengerID)

			// Добавляем бронирование в ответ
			response = append(response, models.BookingResponse{
				ID:              booking.ID,
				RideID:          booking.RideID,
				PassengerID:     booking.PassengerID,
				PickupAddress:   booking.PickupAddress,
				DropoffAddress:  booking.DropoffAddress,
				PickupLocation:  booking.PickupLocation,
				DropoffLocation: booking.DropoffLocation,
				SeatsCount:      booking.SeatsCount,
				Status:          booking.Status,
				BookingType:     booking.BookingType,
				Price:           booking.Price,
				Comment:         booking.Comment,
				RejectReason:    booking.RejectReason,
				CreatedAt:       booking.CreatedAt,
				UpdatedAt:       booking.UpdatedAt,
				PassengerName:   passenger.FirstName + " " + passenger.LastName,
				PassengerPhone:  passenger.Phone,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// BookingApprove подтверждает бронирование
func BookingApprove(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		// Получаем бронирование с блокировкой
		var booking models.Booking
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&booking, bookingID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Бронирование не найдено"})
			return
		}

		// Получаем поездку с блокировкой
		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, booking.RideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является водителем этой поездки
		if ride.DriverID != userID.(uint) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для подтверждения этого бронирования"})
			return
		}

		// Проверяем статус бронирования
		if booking.Status != "pending" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование не может быть подтверждено"})
			return
		}

		// Проверяем наличие свободных мест
		var totalBookedSeats int
		if err := tx.Model(&models.Booking{}).
			Where("ride_id = ? AND status = ? AND id != ?",
				booking.RideID,
				"approved",
				booking.ID).
			Select("COALESCE(SUM(seats_count), 0)").
			Scan(&totalBookedSeats).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при подсчете мест"})
			return
		}

		// Проверяем, достаточно ли мест
		if totalBookedSeats+booking.SeatsCount > ride.SeatsCount {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно свободных мест"})
			return
		}

		// Обновляем статус бронирования
		booking.Status = "approved"
		if err := tx.Save(&booking).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении бронирования"})
			return
		}

		// Обновляем количество забронированных мест
		if err := tx.Model(&ride).Updates(map[string]interface{}{
			"booked_seats": totalBookedSeats + booking.SeatsCount,
			"updated_at":   time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении поездки"})
			return
		}

		// Удаляем существующий маршрут для пересчета
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.RoutePoint{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.OptimizedRoute{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Загружаем информацию о пассажире
		var passenger models.User
		db.First(&passenger, booking.PassengerID)

		// Формируем ответ
		response := models.BookingResponse{
			ID:              booking.ID,
			RideID:          booking.RideID,
			PassengerID:     booking.PassengerID,
			PickupAddress:   booking.PickupAddress,
			DropoffAddress:  booking.DropoffAddress,
			PickupLocation:  booking.PickupLocation,
			DropoffLocation: booking.DropoffLocation,
			SeatsCount:      booking.SeatsCount,
			Status:          booking.Status,
			BookingType:     booking.BookingType,
			Price:           booking.Price,
			Comment:         booking.Comment,
			CreatedAt:       booking.CreatedAt,
			UpdatedAt:       booking.UpdatedAt,
			PassengerName:   passenger.FirstName + " " + passenger.LastName,
			PassengerPhone:  passenger.Phone,
		}

		// После успешного подтверждения бронирования отправляем уведомление пассажиру
		if err := db.First(&passenger, booking.PassengerID).Error; err == nil && passenger.FCMToken != "" {
			notificationData := map[string]interface{}{
				"type":       "booking_approved",
				"booking_id": booking.ID,
				"ride_id":    booking.RideID,
			}
			firebaseService.SendNotification(
				passenger.FCMToken,
				"Бронирование подтверждено",
				fmt.Sprintf("Ваше бронирование на %d мест подтверждено", booking.SeatsCount),
				notificationData,
			)
		}

		c.JSON(http.StatusOK, response)
	}
}

// BookingReject отклоняет бронирование
func BookingReject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		userID, _ := c.Get("user_id")

		var req struct {
			Reason string `json:"reason" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать причину отклонения"})
			return
		}

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		// Получаем бронирование с блокировкой
		var booking models.Booking
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&booking, bookingID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Бронирование не найдено"})
			return
		}

		// Получаем поездку с блокировкой
		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, booking.RideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является водителем этой поездки
		if ride.DriverID != userID.(uint) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для отклонения этого бронирования"})
			return
		}

		// Проверяем статус бронирования
		if booking.Status != models.BookingStatusPending && booking.Status != models.BookingStatusRejected {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование не может быть изменено"})
			return
		}

		// Определяем новый статус (если был rejected -> pending, если был pending -> rejected)
		var newStatus models.BookingStatus
		if booking.Status == models.BookingStatusRejected {
			newStatus = models.BookingStatusPending
		} else {
			newStatus = models.BookingStatusRejected
		}

		// Обновляем статус бронирования
		booking.Status = newStatus
		booking.RejectReason = req.Reason
		if err := tx.Save(&booking).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении бронирования"})
			return
		}

		// Удаляем существующий маршрут для пересчета
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.RoutePoint{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.OptimizedRoute{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Загружаем информацию о пассажире
		var passenger models.User
		db.First(&passenger, booking.PassengerID)

		// Формируем ответ
		response := models.BookingResponse{
			ID:              booking.ID,
			RideID:          booking.RideID,
			PassengerID:     booking.PassengerID,
			PickupAddress:   booking.PickupAddress,
			DropoffAddress:  booking.DropoffAddress,
			PickupLocation:  booking.PickupLocation,
			DropoffLocation: booking.DropoffLocation,
			SeatsCount:      booking.SeatsCount,
			Status:          booking.Status,
			BookingType:     booking.BookingType,
			Price:           booking.Price,
			Comment:         booking.Comment,
			RejectReason:    booking.RejectReason,
			CreatedAt:       booking.CreatedAt,
			UpdatedAt:       booking.UpdatedAt,
			PassengerName:   passenger.FirstName + " " + passenger.LastName,
			PassengerPhone:  passenger.Phone,
		}

		// После отклонения бронирования отправляем уведомление пассажиру
		if err := db.First(&passenger, booking.PassengerID).Error; err == nil && passenger.FCMToken != "" {
			notificationData := map[string]interface{}{
				"type":       "booking_rejected",
				"booking_id": booking.ID,
				"ride_id":    booking.RideID,
				"reason":     req.Reason,
			}
			firebaseService.SendNotification(
				passenger.FCMToken,
				"Бронирование отклонено",
				fmt.Sprintf("Ваше бронирование отклонено. Причина: %s", req.Reason),
				notificationData,
			)
		}

		c.JSON(http.StatusOK, response)
	}
}

// BookingCancel отменяет бронирование
func BookingCancel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		// Получаем бронирование с блокировкой
		var booking models.Booking
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&booking, bookingID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Бронирование не найдено"})
			return
		}

		// Получаем поездку с блокировкой
		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, booking.RideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является пассажиром этого бронирования
		if booking.PassengerID != userID.(uint) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав для отмены этого бронирования"})
			return
		}

		// Проверяем статус бронирования
		if booking.Status == "cancelled" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование уже отменено"})
			return
		}

		// Если бронирование было подтверждено, уменьшаем количество забронированных мест
		if booking.Status == "approved" {
			// Пересчитываем общее количество забронированных мест для поездки
			var totalBookedSeats int
			if err := tx.Model(&models.Booking{}).
				Where("ride_id = ? AND status = ? AND id != ?",
					booking.RideID,
					"approved",
					booking.ID).
				Select("COALESCE(SUM(seats_count), 0)").
				Scan(&totalBookedSeats).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при подсчете забронированных мест"})
				return
			}

			// Обновляем количество забронированных мест в поездке
			if err := tx.Model(&ride).Updates(map[string]interface{}{
				"booked_seats": totalBookedSeats,
				"updated_at":   time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении количества мест"})
				return
			}
		}

		// Обновляем статус бронирования
		booking.Status = "cancelled"
		if err := tx.Save(&booking).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении бронирования"})
			return
		}

		// Удаляем существующий маршрут для пересчета
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.RoutePoint{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}
		if err := tx.Where("ride_id = ?", ride.ID).Delete(&models.OptimizedRoute{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении маршрута"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Загружаем информацию о пассажире
		var passenger models.User
		db.First(&passenger, booking.PassengerID)

		// Формируем ответ
		response := models.BookingResponse{
			ID:              booking.ID,
			RideID:          booking.RideID,
			PassengerID:     booking.PassengerID,
			PickupAddress:   booking.PickupAddress,
			DropoffAddress:  booking.DropoffAddress,
			PickupLocation:  booking.PickupLocation,
			DropoffLocation: booking.DropoffLocation,
			SeatsCount:      booking.SeatsCount,
			Status:          booking.Status,
			BookingType:     booking.BookingType,
			Price:           booking.Price,
			Comment:         booking.Comment,
			CreatedAt:       booking.CreatedAt,
			UpdatedAt:       booking.UpdatedAt,
			PassengerName:   passenger.FirstName + " " + passenger.LastName,
			PassengerPhone:  passenger.Phone,
		}

		c.JSON(http.StatusOK, response)
	}
}

// BookingPickup обрабатывает подбор пассажира водителем
func BookingPickup(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		userID, _ := c.Get("user_id")

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при начале транзакции"})
			return
		}

		// Получаем бронирование с блокировкой
		var booking models.Booking
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&booking, bookingID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Бронирование не найдено"})
			return
		}

		// Получаем поездку с блокировкой
		var ride models.Ride
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&ride, booking.RideID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Поездка не найдена"})
			return
		}

		// Проверяем, что пользователь является водителем поездки
		if ride.DriverID != userID.(uint) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{"error": "Только водитель может отметить подбор пассажира"})
			return
		}

		// Проверяем статус бронирования
		if booking.Status != "approved" && booking.Status != "started" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Бронирование не может быть обновлено"})
			return
		}

		// Обновляем статус бронирования на "started"
		booking.Status = "started"
		if err := tx.Save(&booking).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении бронирования"})
			return
		}

		// Получаем информацию о пассажире
		var passenger models.User
		if err := tx.First(&passenger, booking.PassengerID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении информации о пассажире"})
			return
		}

		// Подтверждаем транзакцию
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении изменений"})
			return
		}

		// Формируем ответ
		response := models.BookingResponse{
			ID:              booking.ID,
			RideID:          booking.RideID,
			PassengerID:     booking.PassengerID,
			PickupAddress:   booking.PickupAddress,
			DropoffAddress:  booking.DropoffAddress,
			PickupLocation:  booking.PickupLocation,
			DropoffLocation: booking.DropoffLocation,
			SeatsCount:      booking.SeatsCount,
			Status:          booking.Status,
			BookingType:     booking.BookingType,
			Price:           booking.Price,
			Comment:         booking.Comment,
			CreatedAt:       booking.CreatedAt,
			UpdatedAt:       booking.UpdatedAt,
			PassengerName:   passenger.FirstName + " " + passenger.LastName,
			PassengerPhone:  passenger.Phone,
		}

		// После отметки о забрании пассажира отправляем уведомление
		if err := db.First(&passenger, booking.PassengerID).Error; err == nil && passenger.FCMToken != "" {
			notificationData := map[string]interface{}{
				"type":       "passenger_picked_up",
				"booking_id": booking.ID,
				"ride_id":    booking.RideID,
			}
			firebaseService.SendNotification(
				passenger.FCMToken,
				"Вы забраны",
				"Водитель отметил вас как забранного",
				notificationData,
			)
		}

		c.JSON(http.StatusOK, response)
	}
}
