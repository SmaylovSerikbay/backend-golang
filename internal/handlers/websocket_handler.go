package handlers

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

					// Выводим все текущие подключения для этого пользователя
					conns := manager.clientsByUser[client.userID]
					log.Printf("Текущее количество подключений для userID=%d: %d", client.userID, len(conns))
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

// WSHandler обрабатывает подключения WebSocket
func WSHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("=== Начало обработки WebSocket подключения ===")
		log.Printf("URL запроса: %s", c.Request.URL.String())
		log.Printf("Метод запроса: %s", c.Request.Method)
		log.Printf("Параметры запроса: %v", c.Request.URL.Query())
		log.Printf("HTTP заголовки запроса: %v", c.Request.Header)

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
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
				log.Printf("Ошибка при обновлении WebSocket соединения: статус=%d, причина=%v", status, reason)
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

		// Устанавливаем userID из контекста или из clientID
		if exists {
			client.userID = userID.(uint)
			log.Printf("Установлен userID из контекста для WebSocket клиента: %d", client.userID)
		} else if parsedUserID > 0 {
			client.userID = parsedUserID
			log.Printf("Установлен userID из clientID для WebSocket клиента: %d", client.userID)
		}

		// Регистрируем клиента
		wsManager.register <- client
		log.Printf("Клиент отправлен на регистрацию: ID=%s, userID=%v", client.clientID, client.userID)

		// Отправляем приветственное сообщение
		welcomeMsg := map[string]interface{}{
			"type": "CONNECTION_ESTABLISHED",
			"payload": map[string]interface{}{
				"client_id": client.clientID,
				"user_id":   client.userID,
				"timestamp": time.Now().Unix(),
			},
		}

		if err := conn.WriteJSON(welcomeMsg); err != nil {
			log.Printf("Ошибка отправки приветственного сообщения: %v", err)
		} else {
			log.Printf("Приветственное сообщение отправлено клиенту %s", client.clientID)
		}

		// Обрабатываем входящие сообщения
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Перехвачена паника в обработчике сообщений: %v", r)
				}
				wsManager.unregister <- client
				conn.Close()
				log.Printf("WebSocket соединение закрыто для клиента: %s", client.clientID)
			}()

			for {
				// Устанавливаем таймаут чтения в 1 час
				conn.SetReadDeadline(time.Now().Add(1 * time.Hour))

				_, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("Ошибка чтения сообщения: %v", err)
					} else {
						log.Printf("Соединение закрыто: %v", err)
					}
					break // Выходим при ошибке чтения
				}
				log.Printf("Получено сообщение от клиента %s: %s", client.clientID, string(message))

				// Обрабатываем сообщение
				var msgData map[string]interface{}
				if err := json.Unmarshal(message, &msgData); err == nil {
					// Проверяем, является ли сообщение ping
					if msgType, ok := msgData["type"].(string); ok {
						if msgType == "PING" {
							// Отправляем pong в ответ
							pongMsg := map[string]interface{}{
								"type": "PONG",
								"payload": map[string]interface{}{
									"timestamp": time.Now().Unix(),
								},
							}
							if err := conn.WriteJSON(pongMsg); err != nil {
								log.Printf("Ошибка отправки PONG: %v", err)
							} else {
								log.Printf("PONG отправлен клиенту %s", client.clientID)
							}
						}
					}
				}
			}
		}()

		log.Printf("=== Конец обработки WebSocket подключения ===")
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
