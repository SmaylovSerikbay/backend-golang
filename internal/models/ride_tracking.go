package models

import (
	"time"
)

// RideTracking модель для отслеживания текущего состояния поездки
type RideTracking struct {
	ID               uint     `gorm:"primaryKey"`
	RideID           uint     `gorm:"not null"`
	CurrentLocation  Location `gorm:"embedded"`
	Status           string   // started, picking_up, dropping_off, completed
	CurrentBookingID uint     // текущее бронирование, которое обрабатывается
	EstimatedTime    int      // расчетное время в минутах до следующей точки
	UpdatedAt        time.Time

	// Связи
	Ride    Ride    `gorm:"foreignKey:RideID"`
	Booking Booking `gorm:"foreignKey:CurrentBookingID"`
}

// RideTrackingUpdate структура для обновления статуса поездки
type RideTrackingUpdate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    string  `json:"status,omitempty"`
	BookingID uint    `json:"booking_id,omitempty"`
}
