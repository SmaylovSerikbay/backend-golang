package models

import (
	"time"
)

type User struct {
	ID              uint             `json:"id" gorm:"primaryKey;column:id;autoIncrement"`
	FirstName       string           `json:"firstName" gorm:"column:first_name;not null;type:varchar(255)"`
	LastName        string           `json:"lastName" gorm:"column:last_name;not null;type:varchar(255)"`
	Phone           string           `json:"phone" gorm:"column:phone;unique;not null;type:varchar(20)"`
	PhotoUrl        string           `json:"photoUrl" gorm:"column:photo_url;type:text"`
	Role            string           `json:"role" gorm:"column:role;default:'user';type:varchar(20)"`
	FCMToken        string           `json:"fcmToken" gorm:"column:fcm_token;type:text"`
	CreatedAt       time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime;type:timestamp with time zone"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;type:timestamp with time zone"`
	DriverDocuments *DriverDocuments `json:"driver_documents,omitempty" gorm:"foreignKey:UserID"`
}

type UserResponse struct {
	ID              uint                     `json:"id"`
	FirstName       string                   `json:"firstName"`
	LastName        string                   `json:"lastName"`
	Phone           string                   `json:"phone"`
	PhotoUrl        string                   `json:"photoUrl"`
	Role            string                   `json:"role"`
	FCMToken        string                   `json:"fcmToken"`
	CreatedAt       time.Time                `json:"created_at"`
	DriverDocuments *DriverDocumentsResponse `json:"driver_documents,omitempty"`
}

// AfterFind вызывается после загрузки модели из базы данных
func (u *User) AfterFind() error {
	if u.PhotoUrl == "" {
		return nil
	}

	if u.PhotoUrl != "" && u.PhotoUrl[0] != '/' {
		u.PhotoUrl = "/" + u.PhotoUrl
	}

	return nil
}
