package models

import (
	"time"
)

type DriverDocumentStatus string

const (
	DocumentStatusPending  DriverDocumentStatus = "pending"  // На модерации
	DocumentStatusApproved DriverDocumentStatus = "approved" // Принят
	DocumentStatusRejected DriverDocumentStatus = "rejected" // Отказ
	DocumentStatusRevision DriverDocumentStatus = "revision" // Доработка
)

type DriverDocuments struct {
	ID                   uint                 `json:"id" gorm:"primaryKey"`
	UserID               uint                 `json:"user_id" gorm:"not null"`
	CarBrand             string               `json:"car_brand" gorm:"not null"`
	CarModel             string               `json:"car_model" gorm:"not null"`
	CarYear              string               `json:"car_year" gorm:"not null"`
	CarColor             string               `json:"car_color" gorm:"not null"`
	CarNumber            string               `json:"car_number" gorm:"not null"`
	DriverLicenseFront   string               `json:"driver_license_front" gorm:"not null"`
	DriverLicenseBack    string               `json:"driver_license_back" gorm:"not null"`
	CarRegistrationFront string               `json:"car_registration_front" gorm:"not null"`
	CarRegistrationBack  string               `json:"car_registration_back" gorm:"not null"`
	CarPhotoFront        string               `json:"car_photo_front"`
	CarPhotoSide         string               `json:"car_photo_side"`
	CarPhotoInterior     string               `json:"car_photo_interior"`
	Status               DriverDocumentStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	RejectionReason      string               `json:"rejection_reason"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	User                 User                 `json:"-" gorm:"foreignKey:UserID"`
}

type DriverDocumentsResponse struct {
	ID                   uint                 `json:"id"`
	UserID               uint                 `json:"user_id,omitempty"`
	User                 *UserResponse        `json:"user,omitempty"`
	CarBrand             string               `json:"car_brand"`
	CarModel             string               `json:"car_model"`
	CarYear              string               `json:"car_year"`
	CarColor             string               `json:"car_color"`
	CarNumber            string               `json:"car_number"`
	DriverLicenseFront   string               `json:"driver_license_front"`
	DriverLicenseBack    string               `json:"driver_license_back"`
	CarRegistrationFront string               `json:"car_registration_front"`
	CarRegistrationBack  string               `json:"car_registration_back"`
	CarPhotoFront        string               `json:"car_photo_front,omitempty"`
	CarPhotoSide         string               `json:"car_photo_side,omitempty"`
	CarPhotoInterior     string               `json:"car_photo_interior,omitempty"`
	Status               DriverDocumentStatus `json:"status"`
	RejectionReason      string               `json:"rejection_reason,omitempty"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}
