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
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	baseURL         = "https://catalog.api.2gis.com/3.0"
	cacheExpiration = 24 * time.Hour // Кэш на 24 часа
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	redis      *redis.Client
}

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

func NewClient(apiKey string) *Client {
	// Получаем конфигурацию Redis из переменных окружения
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
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Проверяем подключение к Redis
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Ошибка подключения к Redis: %v", err)
		// Если Redis недоступен, все равно создаем клиент
		// Приложение будет работать без кэширования
	} else {
		log.Printf("Успешное подключение к Redis")
	}

	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		redis:      rdb,
	}
}

func (c *Client) SearchAddress(query string) (*SearchResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("address:%s", query)

	// Пробуем получить результат из кэша
	cachedResult, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Если данные есть в кэше, десериализуем и возвращаем
		var result SearchResponse
		if err := json.Unmarshal([]byte(cachedResult), &result); err == nil {
			log.Printf("Получены данные из кэша для запроса: %s", query)
			return &result, nil
		}
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

	log.Printf("Получен ответ от 2GIS (тело): %s", string(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("2GIS вернул статус %d", resp.StatusCode)
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		log.Printf("Ошибка при декодировании ответа от 2GIS: %v", err)
		return nil, fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	// Сохраняем результат в кэш
	if resultJSON, err := json.Marshal(result); err == nil {
		c.redis.Set(ctx, cacheKey, resultJSON, cacheExpiration)
		log.Printf("Результат сохранен в кэш для запроса: %s", query)
	}

	log.Printf("Получен ответ от 2GIS: %+v", result)
	return &result, nil
}
