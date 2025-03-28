package routes

import (
	"taxi-backend/internal/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// Защищенные маршруты (требуют аутентификации)
	{
		// Получение информации о текущем пользователе
		api.GET("/user", handlers.GetCurrentUser(db))

		// Роуты для пользователей
		api.GET("/profile", handlers.UserGetProfile(db))
		api.PUT("/profile", handlers.UserUpdateProfile(db))
		api.PUT("/fcm-token", handlers.UpdateFCMToken(db))

		// Роуты для документов водителя
		api.POST("/driver/documents", handlers.DriverDocumentsSubmit(db))
		api.GET("/driver/documents", handlers.DriverDocumentsGet(db))
		api.PUT("/driver/documents/:id/status", handlers.DriverDocumentsUpdateStatus(db))
		api.DELETE("/driver/documents", handlers.DriverDocumentsDelete(db))

		// Роуты для поездок
		api.POST("/rides", handlers.RideCreate(db))
		api.GET("/rides/active", handlers.RideGetActive(db))
		api.GET("/rides/completed", handlers.RideGetCompleted(db))
		api.GET("/rides/cancelled", handlers.GetCancelledRides(db))
		api.GET("/rides/:id", handlers.RideGetByID(db))
		api.GET("/rides/:id/route", handlers.GetOptimizedRoute(db))
		api.PUT("/rides/:id", handlers.RideUpdate(db))
		api.PUT("/rides/:id/start", handlers.RideStart(db))
		api.PUT("/rides/:id/cancel", handlers.RideCancel(db))
		api.PUT("/rides/:id/complete", handlers.RideComplete(db))
		api.POST("/rides/search", handlers.RideSearch(db))

		// Роуты для бронирований
		api.POST("/bookings", handlers.BookingCreate(db))
		api.GET("/bookings", handlers.BookingGetByUser(db))
		api.GET("/rides/:id/bookings", handlers.BookingGetByRide(db))
		api.PUT("/bookings/:id/approve", handlers.BookingApprove(db))
		api.PUT("/bookings/:id/reject", handlers.BookingReject(db))
		api.PUT("/bookings/:id/cancel", handlers.BookingCancel(db))
		api.PUT("/bookings/:id/pickup", handlers.BookingPickup(db))

		// Роуты для водителей
		api.GET("/drivers/nearby", handlers.DriverGetNearby(db))
		api.POST("/drivers/location", handlers.DriverUpdateLocation(db))

		// Роуты для адресов
		api.POST("/addresses/search", handlers.SearchAddress)

		// Загрузка файлов
		api.POST("/upload", handlers.UploadFile)

		// Роуты для отслеживания поездки
		api.PUT("/rides/:id/tracking", handlers.RideTrackingUpdate(db))
		api.GET("/rides/:id/tracking", handlers.RideTrackingGet(db))
	}
}
