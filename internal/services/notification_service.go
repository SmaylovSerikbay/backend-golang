package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type NotificationService struct {
	serverKey string
}

type FCMPayload struct {
	To           string            `json:"to"`
	Data         map[string]string `json:"data,omitempty"`
	Notification struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	} `json:"notification"`
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		serverKey: os.Getenv("FIREBASE_SERVER_KEY"),
	}
}

func (s *NotificationService) SendPushNotification(ctx context.Context, token, title, body string, data map[string]string) error {
	payload := FCMPayload{
		To:   token,
		Data: data,
	}
	payload.Notification.Title = title
	payload.Notification.Body = body

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "key="+s.serverKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM returned error: %v", resp.Status)
	}

	return nil
}
