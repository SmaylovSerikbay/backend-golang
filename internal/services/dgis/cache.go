package dgis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheService представляет сервис кэширования для запросов к 2ГИС API
type CacheService struct {
	redisClient *redis.Client
	ttl         time.Duration
	enabled     bool
}

// NewCacheService создает новый сервис кэширования
func NewCacheService() *CacheService {
	// Проверяем, включено ли кэширование
	cacheEnabled := os.Getenv("CACHE_ENABLED") == "true"

	if !cacheEnabled {
		return &CacheService{
			enabled: false,
		}
	}

	// Получаем настройки подключения к Redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Получаем TTL для кэша
	cacheDuration := os.Getenv("DGIS_CACHE_DURATION")
	ttl := 86400 // 1 день по умолчанию

	if cacheDuration != "" {
		if val, err := strconv.Atoi(cacheDuration); err == nil {
			ttl = val
		}
	}

	// Создаем клиент Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB:       0,
	})

	return &CacheService{
		redisClient: client,
		ttl:         time.Duration(ttl) * time.Second,
		enabled:     true,
	}
}

// Get получает данные из кэша
func (c *CacheService) Get(ctx context.Context, key string, result interface{}) (bool, error) {
	if !c.enabled {
		return false, nil
	}

	val, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Ключ не найден в кэше
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка при получении данных из кэша: %w", err)
	}

	// Десериализуем данные из JSON
	if err := json.Unmarshal([]byte(val), result); err != nil {
		return false, fmt.Errorf("ошибка при десериализации данных из кэша: %w", err)
	}

	return true, nil
}

// Set сохраняет данные в кэш
func (c *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	if !c.enabled {
		return nil
	}

	// Сериализуем данные в JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации данных для кэша: %w", err)
	}

	// Сохраняем данные в Redis
	if err := c.redisClient.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("ошибка при сохранении данных в кэш: %w", err)
	}

	return nil
}

// GenerateRouteKey генерирует ключ для кэша маршрутов
func (c *CacheService) GenerateRouteKey(startLat, startLng, endLat, endLng float64) string {
	return fmt.Sprintf("route:%f:%f:%f:%f", startLat, startLng, endLat, endLng)
}

// GenerateGeocodingKey генерирует ключ для кэша геокодирования
func (c *CacheService) GenerateGeocodingKey(query string) string {
	return fmt.Sprintf("geocoding:%s", query)
}

// Close закрывает соединение с Redis
func (c *CacheService) Close() error {
	if c.enabled {
		return c.redisClient.Close()
	}
	return nil
}
