package dgis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const baseURL = "https://catalog.api.2gis.com/3.0"

// Client представляет клиент для работы с API 2ГИС
type Client struct {
	apiKey        string
	httpClient    *http.Client
	cacheService  *CacheService
	rateLimiter   *time.Ticker
	requestsMutex sync.Mutex
	requestsCount int
	requestsLimit int
	resetTime     time.Time
}

// SearchResponse представляет ответ от API поиска адресов 2ГИС
type SearchResponse struct {
	Meta struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"meta"`
	Result struct {
		Total int `json:"total"`
		Items []struct {
			Type     string `json:"type"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			Address  struct {
				Name       string `json:"name"`
				FullName   string `json:"full_name"`
				PostalCode string `json:"postal_code"`
			} `json:"address"`
			Point struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"point"`
		} `json:"items"`
	} `json:"result"`
}

// RouteResponse представляет ответ от API построения маршрутов 2ГИС
type RouteResponse struct {
	Meta struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"meta"`
	Result struct {
		Routes []struct {
			Distance int    `json:"distance"`
			Duration int    `json:"duration"`
			Type     string `json:"type"`
			Points   []struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"points"`
		} `json:"routes"`
	} `json:"result"`
}

// NewClient создает новый клиент для работы с API 2ГИС
func NewClient(apiKey string) *Client {
	// Создаем сервис кэширования
	cacheService := NewCacheService()

	// Получаем ограничение на количество запросов в день из конфигурации
	// По умолчанию 5000 запросов в день
	requestsLimit := 5000
	if limitStr := os.Getenv("DGIS_DAILY_LIMIT"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			requestsLimit = val
		}
	}

	return &Client{
		apiKey:        apiKey,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		cacheService:  cacheService,
		rateLimiter:   time.NewTicker(200 * time.Millisecond), // Максимум 5 запросов в секунду
		requestsLimit: requestsLimit,
		resetTime:     time.Now().Add(24 * time.Hour),
	}
}

// checkRateLimit проверяет лимит запросов и ожидает, если необходимо
func (c *Client) checkRateLimit() error {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()

	// Если прошли сутки, сбрасываем счетчик
	if time.Now().After(c.resetTime) {
		c.requestsCount = 0
		c.resetTime = time.Now().Add(24 * time.Hour)
	}

	// Проверяем дневной лимит запросов
	if c.requestsCount >= c.requestsLimit {
		return fmt.Errorf("превышен дневной лимит запросов к API 2ГИС (%d)", c.requestsLimit)
	}

	// Ожидаем разрешения от rate limiter (не чаще 5 запросов в секунду)
	<-c.rateLimiter.C

	c.requestsCount++
	return nil
}

// SearchAddress выполняет поиск адреса
func (c *Client) SearchAddress(query string) (*SearchResponse, error) {
	ctx := context.Background()

	// Генерируем ключ для кэша
	cacheKey := c.cacheService.GenerateGeocodingKey(query)

	// Пробуем получить результат из кэша
	var result SearchResponse
	found, err := c.cacheService.Get(ctx, cacheKey, &result)
	if err != nil {
		log.Printf("Ошибка при получении данных из кэша: %v", err)
	} else if found {
		log.Printf("Получены данные из кэша для запроса: %s", query)
		return &result, nil
	}

	// Проверяем ограничение запросов
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}

	// Если в кэше нет, делаем запрос к API
	params := url.Values{}
	params.Add("q", query)
	params.Add("key", c.apiKey)
	params.Add("fields", "items.point,items.address,items.type,items.full_name")
	params.Add("type", "building,street,adm_div.city,adm_div.district")
	params.Add("locale", "ru_KZ")

	// Координаты для поиска в Астане:
	// viewpoint1 - левый верхний угол (северо-западный)
	// viewpoint2 - правый нижний угол (юго-восточный)
	params.Add("viewpoint1", "71.0,51.5") // Северо-западный угол
	params.Add("viewpoint2", "72.0,51.0") // Юго-восточный угол

	url := fmt.Sprintf("%s/items?%s", baseURL, params.Encode())
	log.Printf("Отправляем запрос к 2GIS: %s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса к 2GIS: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении тела ответа: %v", err)
		return nil, fmt.Errorf("ошибка при чтении ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("2GIS вернул статус %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		log.Printf("Ошибка при декодировании ответа от 2GIS: %v", err)
		return nil, fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	// Сохраняем результат в кэш
	if err := c.cacheService.Set(ctx, cacheKey, result); err != nil {
		log.Printf("Ошибка при сохранении данных в кэш: %v", err)
	} else {
		log.Printf("Результат сохранен в кэш для запроса: %s", query)
	}

	return &result, nil
}

// GetRoute получает маршрут между двумя точками
func (c *Client) GetRoute(startLat, startLng, endLat, endLng float64) (*RouteResponse, error) {
	ctx := context.Background()

	// Генерируем ключ для кэша
	cacheKey := c.cacheService.GenerateRouteKey(startLat, startLng, endLat, endLng)

	// Пробуем получить результат из кэша
	var result RouteResponse
	found, err := c.cacheService.Get(ctx, cacheKey, &result)
	if err != nil {
		log.Printf("Ошибка при получении маршрута из кэша: %v", err)
	} else if found {
		log.Printf("Получен маршрут из кэша: %f,%f -> %f,%f", startLat, startLng, endLat, endLng)
		return &result, nil
	}

	// Проверяем ограничение запросов
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}

	// Если в кэше нет, делаем запрос к API
	params := url.Values{}
	params.Add("key", c.apiKey)
	params.Add("locale", "ru_KZ")
	params.Add("point1", fmt.Sprintf("%f,%f", startLat, startLng))
	params.Add("point2", fmt.Sprintf("%f,%f", endLat, endLng))
	params.Add("type", "car") // Тип транспорта

	url := fmt.Sprintf("%s/directions?%s", baseURL, params.Encode())
	log.Printf("Отправляем запрос маршрута к 2GIS: %s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса маршрута к 2GIS: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса маршрута: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка при чтении тела ответа маршрута: %v", err)
		return nil, fmt.Errorf("ошибка при чтении ответа маршрута: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("2GIS вернул статус %d для маршрута: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("неверный статус ответа для маршрута: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		log.Printf("Ошибка при декодировании ответа маршрута от 2GIS: %v", err)
		return nil, fmt.Errorf("ошибка при декодировании ответа маршрута: %w", err)
	}

	// Сохраняем результат в кэш
	if err := c.cacheService.Set(ctx, cacheKey, result); err != nil {
		log.Printf("Ошибка при сохранении маршрута в кэш: %v", err)
	} else {
		log.Printf("Маршрут сохранен в кэш: %f,%f -> %f,%f", startLat, startLng, endLat, endLng)
	}

	return &result, nil
}

// Close закрывает ресурсы клиента
func (c *Client) Close() {
	c.rateLimiter.Stop()
	if c.cacheService != nil {
		c.cacheService.Close()
	}
}
