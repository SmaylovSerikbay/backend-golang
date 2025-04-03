package services

import (
	"log"
	"taxi-backend/internal/websocket"
)

// SendBookingStatusUpdate отправляет уведомление об изменении статуса бронирования
// всем заинтересованным сторонам (пассажиру и водителю)
func SendBookingStatusUpdate(data map[string]interface{}) {
	// Получаем данные из карты
	var bookingID uint
	if id, ok := data["booking_id"].(uint); ok {
		bookingID = id
	} else if id, ok := data["booking_id"].(int); ok {
		bookingID = uint(id)
	} else {
		log.Printf("SendBookingStatusUpdate: некорректный booking_id: %v", data["booking_id"])
		return
	}

	status, ok := data["status"].(string)
	if !ok {
		log.Printf("SendBookingStatusUpdate: некорректный status: %v", data["status"])
		return
	}

	// Получаем ID пассажира
	var passengerID uint
	if id, ok := data["passenger_id"].(uint); ok {
		passengerID = id
	} else if id, ok := data["passenger_id"].(int); ok {
		passengerID = uint(id)
	} else {
		log.Printf("SendBookingStatusUpdate: не указан passenger_id, бронирование %d не будет отправлено пассажиру", bookingID)
	}

	// Получаем ID водителя
	var driverID uint
	if id, ok := data["driver_id"].(uint); ok {
		driverID = id
	} else if id, ok := data["driver_id"].(int); ok {
		driverID = uint(id)
	}

	log.Printf("SendBookingStatusUpdate: отправка уведомления о бронировании %d, статус %s", bookingID, status)

	// Отправляем уведомления всем участникам
	if passengerID > 0 {
		log.Printf("SendBookingStatusUpdate: отправка пассажиру %d", passengerID)
		websocket.SendBookingStatusUpdate(passengerID, bookingID, status)
	}

	if driverID > 0 {
		log.Printf("SendBookingStatusUpdate: отправка водителю %d", driverID)
		websocket.SendBookingStatusUpdate(driverID, bookingID, status)
	}
}
