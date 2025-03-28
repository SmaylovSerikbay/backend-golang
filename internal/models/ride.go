package models

import (
	"time"
)

type RideStatus string

const (
	RideStatusActive    RideStatus = "active"    // Активная поездка
	RideStatusStarted   RideStatus = "started"   // Начатая поездка
	RideStatusCompleted RideStatus = "completed" // Завершенная поездка
	RideStatusCancelled RideStatus = "cancelled" // Отмененная поездка
)

type Ride struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	PassengerID        *uint      `json:"passenger_id,omitempty" gorm:"column:passenger_id;type:integer;default:null;->:false;<-:create"`
	DriverID           uint       `json:"driver_id" gorm:"not null"`
	FromAddress        string     `json:"from_address" gorm:"not null"`
	ToAddress          string     `json:"to_address" gorm:"not null"`
	FromLocation       string     `json:"from_location" gorm:"type:point;not null"`
	ToLocation         string     `json:"to_location" gorm:"type:point;not null"`
	Status             RideStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Price              float64    `json:"price" gorm:"not null"`
	SeatsCount         int        `json:"seats_count" gorm:"not null"`
	BookedSeats        int        `json:"booked_seats" gorm:"default:0"`
	DepartureDate      time.Time  `json:"departure_date" gorm:"not null"`
	Comment            string     `json:"comment" gorm:"default:''"`
	FrontSeatPrice     *float64   `json:"front_seat_price,omitempty" gorm:"default:null"`
	BackSeatPrice      *float64   `json:"back_seat_price,omitempty" gorm:"default:null"`
	CancellationReason string     `json:"cancellation_reason,omitempty" gorm:"default:''"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	Passenger          User       `json:"-" gorm:"foreignKey:PassengerID"`
	Driver             User       `json:"-" gorm:"foreignKey:DriverID"`
	Bookings           []Booking  `json:"-" gorm:"foreignKey:RideID"`
}

type RideResponse struct {
	ID                 uint       `json:"id"`
	PassengerID        *uint      `json:"passenger_id,omitempty"`
	DriverID           uint       `json:"driver_id"`
	FromAddress        string     `json:"from_address"`
	ToAddress          string     `json:"to_address"`
	FromLocation       string     `json:"from_location"`
	ToLocation         string     `json:"to_location"`
	Status             RideStatus `json:"status"`
	Price              float64    `json:"price"`
	SeatsCount         int        `json:"seats_count"`
	BookedSeats        int        `json:"booked_seats"`
	DepartureDate      time.Time  `json:"departure_date"`
	Comment            string     `json:"comment,omitempty"`
	FrontSeatPrice     *float64   `json:"front_seat_price,omitempty"`
	BackSeatPrice      *float64   `json:"back_seat_price,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	PassengerName      string     `json:"passenger_name,omitempty"`
	DriverName         string     `json:"driver_name"`
}

// RideCreate используется только для создания новой поездки
type RideCreate struct {
	DriverID       uint       `gorm:"not null"`
	FromAddress    string     `gorm:"not null"`
	ToAddress      string     `gorm:"not null"`
	FromLocation   string     `gorm:"type:point;not null"`
	ToLocation     string     `gorm:"type:point;not null"`
	Status         RideStatus `gorm:"type:varchar(20);default:'active'"`
	Price          float64    `gorm:"not null"`
	SeatsCount     int        `gorm:"not null"`
	DepartureDate  time.Time  `gorm:"not null"`
	Comment        string     `gorm:"default:''"`
	FrontSeatPrice *float64   `gorm:"default:null"`
	BackSeatPrice  *float64   `gorm:"default:null"`
}

func (rc *RideCreate) TableName() string {
	return "rides"
}
