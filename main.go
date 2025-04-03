package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"taxi-backend/internal/middleware"
	"taxi-backend/internal/models"
	"taxi-backend/internal/routes"
	"taxi-backend/internal/services"
	"taxi-backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func connectWithRetry(dsn string, maxAttempts int, delay time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxAttempts; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err == nil {
			// Настройка пула соединений с БД
			sqlDB, err := db.DB()
			if err != nil {
				return nil, fmt.Errorf("не удалось получить доступ к sql.DB: %w", err)
			}

			// Получаем параметры из конфигурации или используем значения по умолчанию
			maxOpenConns := 100
			maxIdleConns := 25
			connMaxLifetime := 60

			if val, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS")); err == nil && val > 0 {
				maxOpenConns = val
			}

			if val, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS")); err == nil && val > 0 {
				maxIdleConns = val
			}

			if val, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME_MINUTES")); err == nil && val > 0 {
				connMaxLifetime = val
			}

			sqlDB.SetMaxOpenConns(maxOpenConns)
			sqlDB.SetMaxIdleConns(maxIdleConns)
			sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)

			return db, nil
		}
		log.Printf("Попытка подключения к БД %d из %d не удалась: %v\n", i+1, maxAttempts, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("не удалось подключиться к базе данных после %d попыток: %v", maxAttempts, err)
}

// connectToRedis устанавливает соединение с Redis
func connectToRedis() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisHost == "" {
		redisHost = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	// Инициализируем Redis клиент
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:     redisPassword,
		DB:           0,
		PoolSize:     50,              // Максимальное количество соединений в пуле
		MinIdleConns: 10,              // Минимальное количество простаивающих соединений
		MaxRetries:   3,               // Максимальное количество повторных попыток
		DialTimeout:  5 * time.Second, // Тайм-аут при установке соединения
		ReadTimeout:  3 * time.Second, // Тайм-аут при чтении
		WriteTimeout: 3 * time.Second, // Тайм-аут при записи
	})

	// Проверяем подключение к Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	return rdb, nil
}

func initializeRedisForWhatsApp(redisClient *redis.Client) error {
	// Проверяем подключение к Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ошибка проверки подключения к Redis: %w", err)
	}

	// Устанавливаем глобальный Redis клиент для WhatsApp сервиса
	services.SetRedisClient(redisClient)
	return nil
}

func main() {
	// Устанавливаем режим релиза для продакшена
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения")
	}

	// Настраиваем логирование
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		log.SetFlags(0)
		// TODO: Заменить на структурированное JSON логирование
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
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	// Подключение к Redis
	redisClient, err := connectToRedis()
	if err != nil {
		log.Println("Предупреждение: Redis недоступен, продолжаем без кэширования:", err)
	} else {
		log.Println("Успешное подключение к Redis")
		defer redisClient.Close()

		// Инициализируем глобальный Redis клиент для WhatsApp сервиса
		if err := initializeRedisForWhatsApp(redisClient); err != nil {
			log.Printf("Ошибка инициализации Redis для WhatsApp: %v\n", err)
		} else {
			log.Println("Redis успешно инициализирован для WhatsApp сервиса")
		}
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
		log.Fatal("Ошибка миграции базы данных:", err)
	}

	// Запускаем WebSocket менеджер
	websocket.StartManager()

	// Создаем Gin роутер
	r := gin.New()

	// Используем наш собственный логгер и middleware для восстановления после паники
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Добавляем middleware для сбора метрик
	r.Use(middleware.PrometheusMiddleware())

	// Настройка доверенных прокси
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Статическая директория для загруженных файлов
	r.Static("/uploads", "./uploads")

	// Добавляем эндпоинт для метрик Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Проверка работоспособности системы
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API группа
	api := r.Group("/api")

	// Настраиваем маршруты
	routes.SetupRoutes(api, db)

	// Добавляем WebSocket маршрут вне группы /api для совместимости с клиентом
	r.GET("/ws", websocket.Handler(db))

	// Получаем порт из переменных окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Создаем HTTP сервер с настроенными таймаутами
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Сервер запущен на порту %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %s", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал завершения, закрываем соединения...")

	// Даем 30 секунд на завершение текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при graceful shutdown: %s", err)
	}

	log.Println("Сервер корректно завершил работу")
}
