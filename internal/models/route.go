package models

import (
	"time"

	"gorm.io/gorm"
)

// RoutePoint представляет точку маршрута
type RoutePoint struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OrderNum  int       `json:"order_num"` // Порядковый номер в маршруте
	RideID    uint      `json:"ride_id" gorm:"index"`
	RouteID   uint      `json:"route_id" gorm:"index"`
	BookingID *uint     `json:"booking_id" gorm:"index"` // Может быть nil для точек маршрута водителя
	Type      string    `json:"type"`                    // "pickup", "dropoff" или "start"
	Address   string    `json:"address"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Time      time.Time `json:"time"` // Расчетное время прибытия
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OptimizedRoute представляет оптимизированный маршрут
type OptimizedRoute struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	RideID    uint         `json:"ride_id" gorm:"uniqueIndex;not null"`
	Points    []RoutePoint `json:"points" gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE"`
	Distance  float64      `json:"distance"` // Общее расстояние в метрах
	Duration  int          `json:"duration"` // Общее время в минутах
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

func (or *OptimizedRoute) TableName() string {
	return "optimized_routes"
}

// BeforeCreate хук для установки RideID перед созданием записи
func (or *OptimizedRoute) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("ride_id", or.RideID)
	return nil
}

// AfterFind хук для заполнения RideID после получения записи
func (or *OptimizedRoute) AfterFind(tx *gorm.DB) error {
	var rideID uint
	if err := tx.Raw("SELECT ride_id FROM optimized_routes WHERE id = ?", or.ID).Scan(&rideID).Error; err != nil {
		return err
	}
	or.RideID = rideID
	return nil
}
