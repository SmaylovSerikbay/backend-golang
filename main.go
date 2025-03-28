package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"taxi-backend/internal/handlers"
	"taxi-backend/internal/middleware"
	"taxi-backend/internal/models"
	"taxi-backend/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectWithRetry(dsn string, maxAttempts int, delay time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxAttempts; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		log.Printf("Попытка подключения к БД %d из %d не удалась: %v\n", i+1, maxAttempts, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток: %v", maxAttempts, err)
}

func main() {
	// Устанавливаем режим релиза для продакшена
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := connectWithRetry(dsn, 5, 5*time.Second)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автоматическая миграция моделей
	if err := db.AutoMigrate(
		&models.User{},
		&models.DriverDocuments{},
		&models.Booking{},
		&models.Ride{},
		&models.RideTracking{},
		&models.OptimizedRoute{},
		&models.RoutePoint{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := gin.Default()

	// Настройка доверенных прокси
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Статическая директория для загруженных файлов
	r.Static("/uploads", "./uploads")

	// Публичные роуты
	r.POST("/api/register", handlers.AuthRegister(db))
	r.POST("/api/login", handlers.AuthLogin(db))

	// Защищенные роуты
	api := r.Group("/api")
	api.Use(middleware.JWTAuth())
	{
		routes.SetupRoutes(api, db)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
