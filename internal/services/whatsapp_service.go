package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type WhatsAppService struct {
	idInstance       string
	apiTokenInstance string
	baseURL          string
	mediaURL         string
	redisClient      *redis.Client
}

var (
	globalRedisClient *redis.Client
)

// SetRedisClient устанавливает глобальный Redis клиент для WhatsApp сервиса
func SetRedisClient(client *redis.Client) {
	globalRedisClient = client
}

func NewWhatsAppService() *WhatsAppService {
	// Используем глобальный Redis клиент, если он установлен
	var redisClient *redis.Client
	if globalRedisClient != nil {
		redisClient = globalRedisClient
	} else {
		// Инициализируем Redis клиент
		redisHost := os.Getenv("REDIS_HOST")
		redisPort := os.Getenv("REDIS_PORT")
		redisPassword := os.Getenv("REDIS_PASSWORD")

		if redisHost == "" {
			redisHost = "redis" // значение по умолчанию для Docker
		}
		if redisPort == "" {
			redisPort = "6379" // стандартный порт Redis
		}

		redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
		fmt.Printf("Подключение к Redis по адресу: %s\n", redisAddr)

		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       0,
		})

		// Проверяем подключение к Redis
		ctx := context.Background()
		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			fmt.Printf("Ошибка подключения к Redis: %v\n", err)
		} else {
			fmt.Println("Успешное подключение к Redis")
		}
	}

	return &WhatsAppService{
		idInstance:       os.Getenv("GREEN_API_INSTANCE_ID"),
		apiTokenInstance: os.Getenv("GREEN_API_TOKEN"),
		baseURL:          os.Getenv("GREEN_API_BASE_URL"),
		mediaURL:         os.Getenv("GREEN_API_MEDIA_URL"),
		redisClient:      redisClient,
	}
}

// CheckWhatsAppNumber проверяет, зарегистрирован ли номер в WhatsApp
func (w *WhatsAppService) CheckWhatsAppNumber(phone string) error {
	// Форматируем номер телефона - оставляем только цифры
	chatId := strings.TrimSpace(phone)
	chatId = strings.TrimPrefix(chatId, "+")

	// Проверяем, что номер содержит только цифры
	for _, r := range chatId {
		if r < '0' || r > '9' {
			return fmt.Errorf("номер телефона должен содержать только цифры: %s", phone)
		}
	}

	url := fmt.Sprintf("%s/waInstance%s/checkWhatsapp/%s", w.baseURL, w.idInstance, w.apiTokenInstance)

	payload := map[string]interface{}{
		"phoneNumber": chatId,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при проверке номера: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при чтении ответа: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("ошибка при разборе ответа: %w", err)
	}

	// Проверяем статус
	if exists, ok := response["existsWhatsapp"].(bool); ok && !exists {
		return fmt.Errorf("номер не зарегистрирован в WhatsApp")
	}

	return nil
}

// GenerateVerificationCode генерирует случайный код подтверждения
func (w *WhatsAppService) GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	return code
}

// SaveVerificationCode сохраняет код подтверждения в Redis
func (w *WhatsAppService) SaveVerificationCode(phone, code string) error {
	ctx := context.Background()
	key := fmt.Sprintf("verification_code:%s", phone)

	// Сохраняем код с временем жизни 5 минут
	err := w.redisClient.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("ошибка при сохранении кода в Redis: %w", err)
	}

	return nil
}

// VerifyCode проверяет код подтверждения
func (w *WhatsAppService) VerifyCode(phone, code string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verification_code:%s", phone)

	// Получаем сохраненный код
	savedCode, err := w.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, fmt.Errorf("код подтверждения истек или не существует")
	} else if err != nil {
		return false, fmt.Errorf("ошибка при получении кода из Redis: %w", err)
	}

	// Сравниваем коды
	isValid := savedCode == code

	// Если код верный, удаляем его из Redis
	if isValid {
		w.redisClient.Del(ctx, key)
	}

	return isValid, nil
}

// SendVerificationCode отправляет код подтверждения через WhatsApp
func (w *WhatsAppService) SendVerificationCode(phone string, code string) error {
	fmt.Printf("Начало отправки кода подтверждения для номера %s\n", phone)

	// Проверяем номер телефона
	if phone == "" {
		return fmt.Errorf("номер телефона не может быть пустым")
	}

	// Форматируем номер телефона
	chatId := strings.TrimSpace(phone)
	chatId = strings.TrimPrefix(chatId, "+")

	fmt.Printf("Отформатированный номер телефона: %s\n", chatId)

	// Проверяем, что номер содержит только цифры
	for _, r := range chatId {
		if r < '0' || r > '9' {
			return fmt.Errorf("номер телефона должен содержать только цифры: %s", phone)
		}
	}

	// Проверяем длину номера (должно быть 11-12 цифр)
	if len(chatId) < 11 || len(chatId) > 12 {
		return fmt.Errorf("неверная длина номера телефона: %s", phone)
	}

	fmt.Printf("Сохранение кода %s в Redis для номера %s\n", code, phone)

	// Сохраняем код в Redis
	if err := w.SaveVerificationCode(phone, code); err != nil {
		fmt.Printf("Ошибка при сохранении кода в Redis: %v\n", err)
		return fmt.Errorf("ошибка при сохранении кода: %w", err)
	}

	fmt.Printf("Код успешно сохранен в Redis\n")

	// Проверяем наличие всех необходимых параметров
	if w.idInstance == "" || w.apiTokenInstance == "" || w.baseURL == "" {
		return fmt.Errorf("отсутствуют необходимые параметры Green API: idInstance=%s, apiTokenInstance=%s, baseURL=%s",
			w.idInstance, w.apiTokenInstance, w.baseURL)
	}

	chatId = fmt.Sprintf("%s@c.us", chatId)

	url := fmt.Sprintf("%s/waInstance%s/sendMessage/%s", w.baseURL, w.idInstance, w.apiTokenInstance)
	fmt.Printf("Отправка запроса к Green API:\nURL: %s\nID Instance: %s\nToken: %s\n", url, w.idInstance, w.apiTokenInstance)

	// Форматируем сообщение
	message := fmt.Sprintf("Ваш код подтверждения для ASTRA: %s\n\nНикому не сообщайте этот код.", code)

	fmt.Printf("Подготовленные данные:\nChatID: %s\nMessage: %s\n", chatId, message)

	payload := map[string]interface{}{
		"chatId":  chatId,
		"message": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Ошибка при маршалинге данных: %v\n", err)
		return fmt.Errorf("ошибка при маршалинге данных: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка при создании запроса: %v\n", err)
		return fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Добавляем отладочную информацию
	fmt.Printf("Отправляемые данные:\n%s\n", string(jsonData))

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка при отправке запроса: %v\n", err)
		return fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка при чтении ответа: %v\n", err)
		return fmt.Errorf("ошибка при чтении ответа: %w", err)
	}
	fmt.Printf("Ответ от Green API:\nStatus: %d\nBody: %s\n", resp.StatusCode, string(bodyBytes))

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		fmt.Printf("Ошибка при разборе ответа: %v\n", err)
		return fmt.Errorf("ошибка при разборе ответа: %w, тело: %s", err, string(bodyBytes))
	}

	// Проверяем статус ответа и наличие ошибок
	if resp.StatusCode != http.StatusOK {
		if errMsg, ok := response["error"]; ok {
			fmt.Printf("Ошибка от Green API: %v\n", errMsg)
			return fmt.Errorf("ошибка от Green API: %v", errMsg)
		}
		return fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}

	// Проверяем наличие idMessage в ответе
	if _, ok := response["idMessage"]; !ok {
		return fmt.Errorf("отсутствует idMessage в ответе")
	}

	fmt.Printf("Код успешно отправлен через WhatsApp\n")
	return nil
}
