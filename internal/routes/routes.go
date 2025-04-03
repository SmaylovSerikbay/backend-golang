package routes

import (
	"taxi-backend/internal/handlers"
	"taxi-backend/internal/middleware"
	"taxi-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// Публичные маршруты для аутентификации
	auth := api.Group("/auth")
	{
		auth.POST("/request-code", handlers.RequestVerificationCode(db))
		auth.POST("/verify-register", handlers.VerifyAndRegister(db))
		auth.POST("/verify-login", handlers.VerifyAndLogin(db))
	}

	// Защищенные маршруты (требуют аутентификации)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth())
	{
		// Получение информации о текущем пользователе
		protected.GET("/user", handlers.GetCurrentUser(db))

		// Роуты для пользователей
		protected.GET("/profile", handlers.UserGetProfile(db))
		protected.PUT("/profile", handlers.UserUpdateProfile(db))
		protected.PUT("/fcm-token", handlers.UpdateFCMToken(db))

		// Роуты для документов водителя
		protected.POST("/driver/documents", handlers.DriverDocumentsSubmit(db))
		protected.GET("/driver/documents", handlers.DriverDocumentsGet(db))
		protected.PUT("/driver/documents/:id/status", handlers.DriverDocumentsUpdateStatus(db))
		protected.DELETE("/driver/documents", handlers.DriverDocumentsDelete(db))

		// Роуты для поездок
		protected.POST("/rides", handlers.RideCreate(db))
		protected.GET("/rides/active", handlers.RideGetActive(db))
		protected.GET("/rides/completed", handlers.RideGetCompleted(db))
		protected.GET("/rides/cancelled", handlers.GetCancelledRides(db))
		protected.GET("/rides/:id", handlers.RideGetByID(db))
		protected.GET("/rides/:id/route", handlers.GetOptimizedRoute(db))
		protected.PUT("/rides/:id", handlers.RideUpdate(db))
		protected.PUT("/rides/:id/start", handlers.RideStart(db))
		protected.PUT("/rides/:id/cancel", handlers.RideCancel(db))
		protected.PUT("/rides/:id/complete", handlers.RideComplete(db))
		protected.POST("/rides/search", handlers.RideSearch(db))

		// Роуты для бронирований
		protected.POST("/bookings", handlers.BookingCreate(db))
		protected.GET("/bookings", handlers.BookingGetByUser(db))
		protected.GET("/rides/:id/bookings", handlers.BookingGetByRide(db))
		protected.PUT("/bookings/:id/approve", handlers.BookingApprove(db))
		protected.PUT("/bookings/:id/reject", handlers.BookingReject(db))
		protected.PUT("/bookings/:id/cancel", handlers.BookingCancel(db))
		protected.PUT("/bookings/:id/pickup", handlers.BookingPickup(db))

		// Роуты для водителей
		protected.GET("/drivers/nearby", handlers.DriverGetNearby(db))
		protected.POST("/drivers/location", handlers.DriverUpdateLocation(db))

		// Роуты для адресов
		protected.POST("/addresses/search", handlers.SearchAddress)

		// Загрузка файлов
		protected.POST("/upload", handlers.UploadFile)

		// Роуты для отслеживания поездки
		protected.PUT("/rides/:id/tracking", handlers.RideTrackingUpdate(db))
		protected.GET("/rides/:id/tracking", handlers.RideTrackingGet(db))

		// WebSocket подключение для получения обновлений в реальном времени
		protected.GET("/ws", websocket.Handler(db))
	}
}
