package models

import (
	"time"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"   // Ожидает подтверждения
	BookingStatusApproved  BookingStatus = "approved"  // Подтверждено
	BookingStatusStarted   BookingStatus = "started"   // Поездка началась
	BookingStatusRejected  BookingStatus = "rejected"  // Отклонено
	BookingStatusCancelled BookingStatus = "cancelled" // Отменено
	BookingStatusCompleted BookingStatus = "completed" // Завершено
)

type BookingType string

const (
	BookingTypeRegular   BookingType = "regular"    // Обычное бронирование
	BookingTypeFrontSeat BookingType = "front_seat" // Бронирование всего салона
	BookingTypeBackSeat  BookingType = "back_seat"  // Бронирование заднего ряда
)

// Booking представляет бронирование поездки
type Booking struct {
	ID              uint          `json:"id" gorm:"primaryKey"`
	RideID          uint          `json:"ride_id" gorm:"not null"`
	PassengerID     uint          `json:"passenger_id" gorm:"not null"`
	PickupAddress   string        `json:"pickup_address" gorm:"not null"`
	DropoffAddress  string        `json:"dropoff_address" gorm:"not null"`
	PickupLocation  string        `json:"pickup_location" gorm:"type:point;not null"`
	DropoffLocation string        `json:"dropoff_location" gorm:"type:point;not null"`
	SeatsCount      int           `json:"seats_count" gorm:"not null"`
	Status          BookingStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	BookingType     BookingType   `json:"booking_type" gorm:"type:varchar(20);default:'regular'"`
	Price           float64       `json:"price" gorm:"not null"`
	Comment         string        `json:"comment" gorm:"default:''"`
	RejectReason    string        `json:"reject_reason,omitempty" gorm:"default:''"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Ride            Ride          `json:"-" gorm:"foreignKey:RideID"`
	Passenger       User          `json:"-" gorm:"foreignKey:PassengerID"`
}

// BookingResponse представляет ответ API с информацией о бронировании
type BookingResponse struct {
	ID              uint          `json:"id"`
	RideID          uint          `json:"ride_id"`
	PassengerID     uint          `json:"passenger_id"`
	PickupAddress   string        `json:"pickup_address"`
	DropoffAddress  string        `json:"dropoff_address"`
	PickupLocation  string        `json:"pickup_location"`
	DropoffLocation string        `json:"dropoff_location"`
	SeatsCount      int           `json:"seats_count"`
	Status          BookingStatus `json:"status"`
	BookingType     BookingType   `json:"booking_type"`
	Price           float64       `json:"price"`
	Comment         string        `json:"comment,omitempty"`
	RejectReason    string        `json:"reject_reason,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	PassengerName   string        `json:"passenger_name"`
	PassengerPhone  string        `json:"passenger_phone"`
	RideInfo        RideResponse  `json:"ride_info,omitempty"`
}

// BookingCreate используется только для создания нового бронирования
type BookingCreate struct {
	RideID          uint        `json:"ride_id" binding:"required"`
	PassengerID     uint        `json:"passenger_id" binding:"required"`
	PickupAddress   string      `json:"pickup_address" binding:"required"`
	DropoffAddress  string      `json:"dropoff_address" binding:"required"`
	PickupLocation  string      `json:"pickup_location" binding:"required"`
	DropoffLocation string      `json:"dropoff_location" binding:"required"`
	SeatsCount      int         `json:"seats_count" binding:"required"`
	BookingType     BookingType `json:"booking_type" binding:"required"`
	Price           float64     `json:"price" binding:"required"`
	Comment         string      `json:"comment"`
}

func (bc *BookingCreate) TableName() string {
	return "bookings"
}
