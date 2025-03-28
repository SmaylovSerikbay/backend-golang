package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type FirebaseService struct {
	serverKey string
}

type NotificationPayload struct {
	To           string                 `json:"to"`
	Notification NotificationContent    `json:"notification"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

type NotificationContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewFirebaseService() *FirebaseService {
	return &FirebaseService{
		serverKey: os.Getenv("FIREBASE_SERVER_KEY"),
	}
}

func (s *FirebaseService) SendNotification(token string, title, body string, data map[string]interface{}) error {
	payload := NotificationPayload{
		To: token,
		Notification: NotificationContent{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("key=%s", s.serverKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неуспешный статус ответа: %d", resp.StatusCode)
	}

	return nil
}

// Отправка уведомления всем подписчикам определенной темы
func (s *FirebaseService) SendTopicNotification(topic, title, body string, data map[string]interface{}) error {
	payload := NotificationPayload{
		To:           fmt.Sprintf("/topics/%s", topic),
		Notification: NotificationContent{Title: title, Body: body},
		Data:         data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге данных: %v", err)
	}

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("key=%s", s.serverKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неуспешный статус ответа: %d", resp.StatusCode)
	}

	return nil
}
