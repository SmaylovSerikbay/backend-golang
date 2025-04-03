package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Константы для типов сообщений WebSocket
const (
	RideStatusUpdateType     = "RIDE_STATUS_UPDATE"
	BookingStatusUpdateType  = "BOOKING_STATUS_UPDATE"
	DriverLocationUpdateType = "DRIVER_LOCATION_UPDATE"
	DocumentStatusUpdateType = "DOCUMENT_STATUS_UPDATE"
)

// WebSocketMessage представляет формат сообщения WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebsocketManager управляет всеми подключениями WebSocket
type WebSocketManager struct {
	clients       map[string]map[*websocket.Conn]bool
	clientsByUser map[uint]map[*websocket.Conn]bool
	register      chan *WebSocketClient
	unregister    chan *WebSocketClient
	broadcast     chan *WebSocketMessage
	mutex         sync.RWMutex
}

// WebSocketClient представляет клиентское соединение WebSocket
type WebSocketClient struct {
	conn     *websocket.Conn
	userID   uint
	clientID string
}

// Глобальный менеджер WebSocket
var wsManager = NewWebSocketManager()

// Настройка для обновления WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключения с любых источников
	},
}

// NewWebSocketManager создает новый менеджер WebSocket
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:       make(map[string]map[*websocket.Conn]bool),
		clientsByUser: make(map[uint]map[*websocket.Conn]bool),
		register:      make(chan *WebSocketClient),
		unregister:    make(chan *WebSocketClient),
		broadcast:     make(chan *WebSocketMessage),
		mutex:         sync.RWMutex{},
	}
}

// Start запускает обработку сообщений WebSocket
func (manager *WebSocketManager) Start() {
	log.Printf("Запуск WebSocket Manager")
	go func() {
		for {
			select {
			case client := <-manager.register:
				log.Printf("Регистрация нового клиента: ID=%s, userID=%v", client.clientID, client.userID)
				manager.mutex.Lock()
				// Регистрация по clientID
				if _, ok := manager.clients[client.clientID]; !ok {
					manager.clients[client.clientID] = make(map[*websocket.Conn]bool)
				}
				manager.clients[client.clientID][client.conn] = true
				log.Printf("Клиент %s добавлен в map clients", client.clientID)

				// Регистрация по userID если авторизован
				if client.userID > 0 {
					if _, ok := manager.clientsByUser[client.userID]; !ok {
						manager.clientsByUser[client.userID] = make(map[*websocket.Conn]bool)
					}
					manager.clientsByUser[client.userID][client.conn] = true
					log.Printf("Клиент с userID=%d добавлен в map clientsByUser", client.userID)
				}
				manager.mutex.Unlock()

			case client := <-manager.unregister:
				log.Printf("Отмена регистрации клиента: ID=%s, userID=%v", client.clientID, client.userID)
				manager.mutex.Lock()
				// Удаление по clientID
				if _, ok := manager.clients[client.clientID]; ok {
					if _, exists := manager.clients[client.clientID][client.conn]; exists {
						delete(manager.clients[client.clientID], client.conn)
						client.conn.Close()
						log.Printf("Соединение клиента %s удалено из map clients", client.clientID)
					}
					if len(manager.clients[client.clientID]) == 0 {
						delete(manager.clients, client.clientID)
						log.Printf("Клиент %s полностью удален из map clients", client.clientID)
					}
				}

				// Удаление по userID
				if client.userID > 0 {
					if _, ok := manager.clientsByUser[client.userID]; ok {
						if _, exists := manager.clientsByUser[client.userID][client.conn]; exists {
							delete(manager.clientsByUser[client.userID], client.conn)
							log.Printf("Соединение для userID=%d удалено из map clientsByUser", client.userID)
						}
						if len(manager.clientsByUser[client.userID]) == 0 {
							delete(manager.clientsByUser, client.userID)
							log.Printf("Пользователь %d полностью удален из map clientsByUser", client.userID)
						}
					}
				}
				manager.mutex.Unlock()

			case message := <-manager.broadcast:
				// Широковещательная рассылка не реализована, т.к. обычно отправляем конкретным пользователям
				log.Printf("Получено сообщение для широковещательной рассылки: %v", message)
			}
		}
	}()
	log.Printf("WebSocket Manager успешно запущен")
}

// BroadcastToUser отправляет сообщение всем подключениям конкретного пользователя
func (manager *WebSocketManager) BroadcastToUser(userID uint, message *WebSocketMessage) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	log.Printf("BroadcastToUser: Поиск подключений для пользователя ID %d", userID)

	// Получаем соединения для указанного пользователя
	connections, exists := manager.clientsByUser[userID]
	if !exists || len(connections) == 0 {
		log.Printf("BroadcastToUser: Нет активных подключений для пользователя ID %d", userID)
		return
	}

	log.Printf("BroadcastToUser: Найдено %d подключений для пользователя ID %d", len(connections), userID)

	// Кодируем сообщение в JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("BroadcastToUser: Ошибка при кодировании сообщения: %v", err)
		return
	}

	// Отправляем сообщение по каждому соединению
	for conn := range connections {
		go func(c *websocket.Conn) {
			if err := c.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
				log.Printf("BroadcastToUser: Ошибка при отправке сообщения: %v", err)
				// Отключаем клиента при ошибке отправки
				manager.unregister <- &WebSocketClient{
					conn:   c,
					userID: userID,
				}
			} else {
				log.Printf("BroadcastToUser: Сообщение успешно отправлено в одно из соединений пользователя ID %d", userID)
			}
		}(conn)
	}
}

// Handler обрабатывает подключения WebSocket
func Handler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("=== Начало обработки WebSocket подключения ===")
		log.Printf("URL запроса: %s", c.Request.URL.String())

		// Проверяем, что это действительно запрос на установление WebSocket соединения
		if c.GetHeader("Upgrade") != "websocket" {
			log.Printf("Не найден заголовок Upgrade: websocket. Отклоняем соединение.")
			c.String(http.StatusBadRequest, "Требуется WebSocket соединение")
			return
		}

		// Получаем userID из контекста (если пользователь авторизован)
		userID, exists := c.Get("user_id")
		clientID := c.Query("client_id") // Уникальный ID клиента (можно использовать UUID)
		testMode := c.Query("test") == "true"

		log.Printf("userID из контекста: %v (exists: %v)", userID, exists)
		log.Printf("clientID из запроса: %s", clientID)
		log.Printf("Тестовый режим: %v", testMode)

		// Проверяем, есть ли userID в client_id (например, "user_123")
		var parsedUserID uint
		if clientID != "" {
			var tmp int
			_, err := fmt.Sscanf(clientID, "user_%d", &tmp)
			if err == nil && tmp > 0 {
				parsedUserID = uint(tmp)
				log.Printf("Распознан userID из clientID: %d", parsedUserID)
			}
		}

		// Если client_id не указан, используем userID в качестве clientID
		if clientID == "" && exists {
			clientID = fmt.Sprintf("user_%v", userID)
			log.Printf("clientID сгенерирован из userID: %s", clientID)
		} else if clientID == "" {
			clientID = fmt.Sprintf("anon_%d", time.Now().UnixNano())
			log.Printf("clientID сгенерирован как анонимный: %s", clientID)
		}

		// Настройка для обновления соединения
		wsUpgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Разрешаем подключения с любых источников
			},
		}

		// Обновляем соединение до WebSocket
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Ошибка обновления соединения до WebSocket: %v", err)
			c.String(http.StatusInternalServerError, "Не удалось установить WebSocket соединение")
			return
		}

		log.Printf("WebSocket соединение успешно установлено")

		// Если это тестовое соединение, сразу закрываем его
		if testMode {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"TEST_SUCCESS"}`))
			conn.Close()
			log.Printf("Тестовое WebSocket соединение закрыто")
			return
		}

		// Создаем нового клиента
		client := &WebSocketClient{
			conn:     conn,
			clientID: clientID,
		}

		// Устанавливаем userID из контекста или из client_id
		if exists {
			client.userID = userID.(uint)
		} else if parsedUserID > 0 {
			client.userID = parsedUserID
		}

		// Регистрируем клиента
		wsManager.register <- client

		// Запускаем обработку сообщений
		go handleMessages(client)
	}
}

// handleMessages обрабатывает сообщения от клиента
func handleMessages(client *WebSocketClient) {
	defer func() {
		// Когда функция завершается, отменяем регистрацию клиента
		wsManager.unregister <- client
	}()

	// Читаем сообщения от клиента
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка при чтении сообщения от клиента %s: %v", client.clientID, err)
			break
		}

		log.Printf("Получено сообщение от клиента %s: %s", client.clientID, string(message))

		// Пытаемся разобрать сообщение как JSON
		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			log.Printf("Ошибка при разборе JSON: %v", err)
			continue
		}

		// Обрабатываем ping-сообщения
		if msgType, ok := data["type"].(string); ok && msgType == "ping" {
			log.Printf("Получен ping от клиента %s, отправляем pong", client.clientID)
			pongMsg := map[string]interface{}{
				"type": "pong",
				"time": time.Now().Unix(),
			}
			pongJSON, _ := json.Marshal(pongMsg)
			if err := client.conn.WriteMessage(websocket.TextMessage, pongJSON); err != nil {
				log.Printf("Ошибка при отправке pong: %v", err)
			}
		}
	}
}

// SendRideStatusUpdate отправляет обновление статуса поездки
func SendRideStatusUpdate(userID uint, rideID uint, status string) {
	payload := map[string]interface{}{
		"ride_id": rideID,
		"status":  status,
	}
	message := &WebSocketMessage{
		Type:    RideStatusUpdateType,
		Payload: payload,
	}
	wsManager.BroadcastToUser(userID, message)
}

// SendBookingStatusUpdate отправляет обновление статуса бронирования
func SendBookingStatusUpdate(userID uint, bookingID uint, status string) {
	payload := map[string]interface{}{
		"booking_id": bookingID,
		"status":     status,
	}
	message := &WebSocketMessage{
		Type:    BookingStatusUpdateType,
		Payload: payload,
	}
	wsManager.BroadcastToUser(userID, message)
}

// SendDriverLocationUpdate отправляет обновление местоположения водителя
func SendDriverLocationUpdate(userID uint, driverID uint, lat, lng float64) {
	payload := map[string]interface{}{
		"driver_id": driverID,
		"lat":       lat,
		"lng":       lng,
	}
	message := &WebSocketMessage{
		Type:    DriverLocationUpdateType,
		Payload: payload,
	}
	wsManager.BroadcastToUser(userID, message)
}

// SendDocumentStatusUpdate отправляет обновление статуса документа водителя
func SendDocumentStatusUpdate(userID uint, documentID uint, status string) {
	log.Printf("SendDocumentStatusUpdate: Подготовка уведомления для пользователя %d, документ %d, статус: %s", userID, documentID, status)

	payload := map[string]interface{}{
		"document_id": documentID,
		"status":      status,
	}
	message := &WebSocketMessage{
		Type:    DocumentStatusUpdateType,
		Payload: payload,
	}

	log.Printf("SendDocumentStatusUpdate: Отправка уведомления через BroadcastToUser: %+v", message)
	wsManager.BroadcastToUser(userID, message)
	log.Printf("SendDocumentStatusUpdate: Уведомление отправлено")
}

// GetManager возвращает глобальный экземпляр менеджера WebSocket
func GetManager() *WebSocketManager {
	return wsManager
}

// StartManager запускает менеджер WebSocket
func StartManager() {
	wsManager.Start()
}
