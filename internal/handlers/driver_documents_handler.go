package handlers

import (
	"net/http"
	"taxi-backend/internal/models"
	"time"

	"log"

	"bytes"
	"io"

	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DriverDocumentsRequest struct {
	CarBrand             string `json:"carBrand" binding:"required"`
	CarModel             string `json:"carModel" binding:"required"`
	CarYear              string `json:"carYear" binding:"required"`
	CarColor             string `json:"carColor" binding:"required"`
	CarNumber            string `json:"carNumber" binding:"required"`
	DriverLicenseFront   string `json:"driverLicenseFront" binding:"required"`
	DriverLicenseBack    string `json:"driverLicenseBack" binding:"required"`
	CarRegistrationFront string `json:"carRegistrationFront" binding:"required"`
	CarRegistrationBack  string `json:"carRegistrationBack" binding:"required"`
	CarPhotoFront        string `json:"carPhotoFront"`
	CarPhotoSide         string `json:"carPhotoSide"`
	CarPhotoInterior     string `json:"carPhotoInterior"`
}

// Создание/обновление документов водителя
func DriverDocumentsSubmit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var req DriverDocumentsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		log.Printf("Получены данные для создания документов: %+v", req)

		// Проверяем существующие документы
		var docs models.DriverDocuments
		result := db.Where("user_id = ?", userID).First(&docs)

		if result.Error == nil {
			// Обновляем существующие документы
			log.Printf("Обновляем существующие документы для пользователя ID: %v", userID)
			docs.CarBrand = req.CarBrand
			docs.CarModel = req.CarModel
			docs.CarYear = req.CarYear
			docs.CarColor = req.CarColor
			docs.CarNumber = req.CarNumber
			docs.DriverLicenseFront = req.DriverLicenseFront
			docs.DriverLicenseBack = req.DriverLicenseBack
			docs.CarRegistrationFront = req.CarRegistrationFront
			docs.CarRegistrationBack = req.CarRegistrationBack
			docs.CarPhotoFront = req.CarPhotoFront
			docs.CarPhotoSide = req.CarPhotoSide
			docs.CarPhotoInterior = req.CarPhotoInterior
			docs.Status = models.DocumentStatusPending
			docs.UpdatedAt = time.Now()

			if err := db.Save(&docs).Error; err != nil {
				log.Printf("Ошибка при обновлении документов: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении документов"})
				return
			}
		} else {
			// Создаем новые документы
			log.Printf("Создаем новые документы для пользователя ID: %v", userID)
			docs = models.DriverDocuments{
				UserID:               userID.(uint),
				CarBrand:             req.CarBrand,
				CarModel:             req.CarModel,
				CarYear:              req.CarYear,
				CarColor:             req.CarColor,
				CarNumber:            req.CarNumber,
				DriverLicenseFront:   req.DriverLicenseFront,
				DriverLicenseBack:    req.DriverLicenseBack,
				CarRegistrationFront: req.CarRegistrationFront,
				CarRegistrationBack:  req.CarRegistrationBack,
				CarPhotoFront:        req.CarPhotoFront,
				CarPhotoSide:         req.CarPhotoSide,
				CarPhotoInterior:     req.CarPhotoInterior,
				Status:               models.DocumentStatusPending,
			}

			if err := db.Create(&docs).Error; err != nil {
				log.Printf("Ошибка при создании документов: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании документов"})
				return
			}
		}

		log.Printf("Документы успешно сохранены: %+v", docs)
		c.JSON(http.StatusOK, models.DriverDocumentsResponse{
			ID:                   docs.ID,
			CarBrand:             docs.CarBrand,
			CarModel:             docs.CarModel,
			CarYear:              docs.CarYear,
			CarColor:             docs.CarColor,
			CarNumber:            docs.CarNumber,
			DriverLicenseFront:   docs.DriverLicenseFront,
			DriverLicenseBack:    docs.DriverLicenseBack,
			CarRegistrationFront: docs.CarRegistrationFront,
			CarRegistrationBack:  docs.CarRegistrationBack,
			CarPhotoFront:        docs.CarPhotoFront,
			CarPhotoSide:         docs.CarPhotoSide,
			CarPhotoInterior:     docs.CarPhotoInterior,
			Status:               docs.Status,
			CreatedAt:            docs.CreatedAt,
			UpdatedAt:            docs.UpdatedAt,
		})
	}
}

// Получение документов водителя
func DriverDocumentsGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		log.Printf("UPDATED CODE: Получение документов для пользователя ID: %v", userID)

		// Если админ, возвращаем все документы
		if role == "admin" {
			var allDocs []models.DriverDocuments
			if err := db.Find(&allDocs).Error; err != nil {
				log.Printf("Ошибка при получении всех документов: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении документов"})
				return
			}

			var response []models.DriverDocumentsResponse
			for _, doc := range allDocs {
				// Получаем информацию о пользователе
				var user models.User
				if err := db.First(&user, doc.UserID).Error; err != nil {
					log.Printf("Ошибка при получении пользователя для документа ID %v: %v", doc.ID, err)
				}

				response = append(response, models.DriverDocumentsResponse{
					ID:     doc.ID,
					UserID: doc.UserID,
					User: &models.UserResponse{
						ID:        user.ID,
						FirstName: user.FirstName,
						LastName:  user.LastName,
						Phone:     user.Phone,
						PhotoUrl:  user.PhotoUrl,
						Role:      user.Role,
					},
					CarBrand:             doc.CarBrand,
					CarModel:             doc.CarModel,
					CarYear:              doc.CarYear,
					CarColor:             doc.CarColor,
					CarNumber:            doc.CarNumber,
					DriverLicenseFront:   doc.DriverLicenseFront,
					DriverLicenseBack:    doc.DriverLicenseBack,
					CarRegistrationFront: doc.CarRegistrationFront,
					CarRegistrationBack:  doc.CarRegistrationBack,
					CarPhotoFront:        doc.CarPhotoFront,
					CarPhotoSide:         doc.CarPhotoSide,
					CarPhotoInterior:     doc.CarPhotoInterior,
					Status:               doc.Status,
					RejectionReason:      doc.RejectionReason,
					CreatedAt:            doc.CreatedAt,
					UpdatedAt:            doc.UpdatedAt,
				})
			}

			c.JSON(http.StatusOK, response)
			return
		}

		// Для обычного пользователя возвращаем только его документы
		var docs models.DriverDocuments
		result := db.Where("user_id = ?", userID).First(&docs)
		if result.Error != nil {
			log.Printf("Ошибка при получении документов: %v", result.Error)
			if result.Error == gorm.ErrRecordNotFound {
				// Возвращаем пустой объект с пометкой, что документы не найдены
				log.Printf("ИСПРАВЛЕНО: Возвращаем объект с пометкой 'not_found' для пользователя ID: %v", userID)
				response := gin.H{
					"id":                     0,
					"user_id":                userID,
					"car_brand":              "",
					"car_model":              "",
					"car_year":               "",
					"car_color":              "",
					"car_number":             "",
					"driver_license_front":   "",
					"driver_license_back":    "",
					"car_registration_front": "",
					"car_registration_back":  "",
					"status":                 "not_found",
					"message":                "Документы для этого пользователя не найдены",
					"created_at":             time.Time{},
					"updated_at":             time.Time{},
				}
				c.JSON(http.StatusOK, response)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении документов"})
			return
		}

		// Получаем информацию о пользователе
		var user models.User
		if err := db.First(&user, docs.UserID).Error; err != nil {
			log.Printf("Ошибка при получении пользователя для документа ID %v: %v", docs.ID, err)
		}

		log.Printf("Документы найдены: %+v", docs)
		c.JSON(http.StatusOK, models.DriverDocumentsResponse{
			ID:     docs.ID,
			UserID: docs.UserID,
			User: &models.UserResponse{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Phone:     user.Phone,
				PhotoUrl:  user.PhotoUrl,
				Role:      user.Role,
			},
			CarBrand:             docs.CarBrand,
			CarModel:             docs.CarModel,
			CarYear:              docs.CarYear,
			CarColor:             docs.CarColor,
			CarNumber:            docs.CarNumber,
			DriverLicenseFront:   docs.DriverLicenseFront,
			DriverLicenseBack:    docs.DriverLicenseBack,
			CarRegistrationFront: docs.CarRegistrationFront,
			CarRegistrationBack:  docs.CarRegistrationBack,
			CarPhotoFront:        docs.CarPhotoFront,
			CarPhotoSide:         docs.CarPhotoSide,
			CarPhotoInterior:     docs.CarPhotoInterior,
			Status:               docs.Status,
			RejectionReason:      docs.RejectionReason,
			CreatedAt:            docs.CreatedAt,
			UpdatedAt:            docs.UpdatedAt,
		})
	}
}

// Обновление статуса документов (для админов)
func DriverDocumentsUpdateStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("=== Начало обработки запроса на обновление статуса документа ===")
		log.Printf("URL запроса: %s", c.Request.URL.String())
		log.Printf("Метод запроса: %s", c.Request.Method)
		log.Printf("Параметры запроса: %v", c.Params)
		log.Printf("Заголовки запроса: %v", c.Request.Header)

		// Читаем тело запроса для логирования
		bodyBytes, _ := c.GetRawData()
		bodyString := string(bodyBytes)
		log.Printf("Тело запроса: %s", bodyString)

		// Восстанавливаем тело запроса для дальнейшей обработки
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var req struct {
			Status          models.DriverDocumentStatus `json:"status" binding:"required"`
			RejectionReason string                      `json:"rejectionReason,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			log.Printf("Тело запроса: %s", bodyString)

			// Попробуем другой формат данных (с snake_case)
			var altReq struct {
				Status          models.DriverDocumentStatus `json:"status" binding:"required"`
				RejectionReason string                      `json:"rejection_reason,omitempty"`
			}

			if err := c.Request.Body.Close(); err != nil {
				log.Printf("Ошибка при закрытии тела запроса: %v", err)
			}

			// Восстанавливаем тело запроса для повторной обработки
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := c.ShouldBindJSON(&altReq); err != nil {
				log.Printf("Ошибка валидации данных (альтернативный формат): %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
				return
			}

			// Копируем данные из альтернативного формата
			req.Status = altReq.Status
			req.RejectionReason = altReq.RejectionReason
		}

		log.Printf("Получен запрос на обновление статуса документа: статус=%s, причина=%s", req.Status, req.RejectionReason)

		docID := c.Param("id")
		log.Printf("ID документа из параметров: %s", docID)

		var docs models.DriverDocuments

		if err := db.First(&docs, docID).Error; err != nil {
			log.Printf("Документы не найдены: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Документы не найдены"})
			return
		}

		log.Printf("Текущий статус документа: %s", docs.Status)

		// Получаем пользователя для отправки уведомления
		var user models.User
		if err := db.First(&user, docs.UserID).Error; err != nil {
			log.Printf("Ошибка при получении пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении пользователя"})
			return
		}

		// Обновляем статус и причину отклонения
		docs.Status = req.Status
		docs.RejectionReason = req.RejectionReason
		docs.UpdatedAt = time.Now()

		log.Printf("Обновляем документ ID %s: статус=%s, причина=%s", docID, docs.Status, docs.RejectionReason)

		if err := db.Save(&docs).Error; err != nil {
			log.Printf("Ошибка при обновлении статуса: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении статуса"})
			return
		}

		log.Printf("Документ успешно обновлен, новый статус: %s", docs.Status)

		// Получаем обновленный документ из БД для проверки
		var updatedDocs models.DriverDocuments
		if err := db.First(&updatedDocs, docID).Error; err != nil {
			log.Printf("Ошибка при получении обновленного документа: %v", err)
		} else {
			log.Printf("Проверка обновления: ID=%d, статус=%s", updatedDocs.ID, updatedDocs.Status)
		}

		// Отправляем WebSocket уведомление пользователю
		log.Printf("Подготовка к отправке WebSocket уведомления пользователю ID %d о статусе документа %d", user.ID, docs.ID)
		log.Printf("Статус документа для WebSocket: %s, тип: %T", string(docs.Status), docs.Status)
		SendDocumentStatusUpdate(user.ID, docs.ID, string(docs.Status))
		log.Printf("Отправлено WebSocket уведомление пользователю ID %d о статусе документа %d", user.ID, docs.ID)

		// Отправляем уведомление пользователю
		if user.FCMToken != "" {
			var title, body string
			switch docs.Status {
			case models.DocumentStatusApproved:
				title = "Документы одобрены"
				body = "Ваши документы были проверены и одобрены. Теперь вы можете создавать поездки."
			case models.DocumentStatusRejected:
				title = "Документы отклонены"
				body = "Ваши документы были отклонены. Причина: " + docs.RejectionReason
			case models.DocumentStatusRevision:
				title = "Требуется доработка документов"
				body = "Ваши документы требуют доработки. Причина: " + docs.RejectionReason
			}

			if title != "" && body != "" {
				notification := map[string]interface{}{
					"notification": map[string]interface{}{
						"title": title,
						"body":  body,
					},
					"data": map[string]interface{}{
						"document_id": docID,
						"status":      string(docs.Status),
					},
					"to": user.FCMToken,
				}

				// Отправляем уведомление через FCM
				fcmURL := "https://fcm.googleapis.com/fcm/send"
				fcmKey := os.Getenv("FCM_SERVER_KEY")
				if fcmKey == "" {
					log.Printf("FCM_SERVER_KEY не установлен")
				} else {
					client := &http.Client{}
					jsonData, _ := json.Marshal(notification)
					req, _ := http.NewRequest("POST", fcmURL, bytes.NewBuffer(jsonData))
					req.Header.Set("Authorization", "key="+fcmKey)
					req.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(req)
					if err != nil {
						log.Printf("Ошибка при отправке FCM уведомления: %v", err)
					} else {
						defer resp.Body.Close()
						body, _ := io.ReadAll(resp.Body)
						log.Printf("Ответ FCM: %s", string(body))
					}
				}
			}
		} else {
			log.Printf("FCM токен не найден для пользователя ID %d", user.ID)
		}

		// Получаем информацию о пользователе для ответа
		var userResponse models.User
		if err := db.First(&userResponse, docs.UserID).Error; err != nil {
			log.Printf("Ошибка при получении пользователя для документа ID %v: %v", docs.ID, err)
		}

		log.Printf("Статус документа успешно обновлен: %+v", docs)
		c.JSON(http.StatusOK, models.DriverDocumentsResponse{
			ID:     docs.ID,
			UserID: docs.UserID,
			User: &models.UserResponse{
				ID:        userResponse.ID,
				FirstName: userResponse.FirstName,
				LastName:  userResponse.LastName,
				Phone:     userResponse.Phone,
				PhotoUrl:  userResponse.PhotoUrl,
				Role:      userResponse.Role,
			},
			CarBrand:             docs.CarBrand,
			CarModel:             docs.CarModel,
			CarYear:              docs.CarYear,
			CarColor:             docs.CarColor,
			CarNumber:            docs.CarNumber,
			DriverLicenseFront:   docs.DriverLicenseFront,
			DriverLicenseBack:    docs.DriverLicenseBack,
			CarRegistrationFront: docs.CarRegistrationFront,
			CarRegistrationBack:  docs.CarRegistrationBack,
			CarPhotoFront:        docs.CarPhotoFront,
			CarPhotoSide:         docs.CarPhotoSide,
			CarPhotoInterior:     docs.CarPhotoInterior,
			Status:               docs.Status,
			RejectionReason:      docs.RejectionReason,
			CreatedAt:            docs.CreatedAt,
			UpdatedAt:            docs.UpdatedAt,
		})
		log.Printf("=== Конец обработки запроса на обновление статуса документа ===")
	}
}

// Удаление документов водителя
func DriverDocumentsDelete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		log.Printf("Удаление документов для пользователя ID: %v", userID)

		result := db.Where("user_id = ?", userID).Delete(&models.DriverDocuments{})
		if result.Error != nil {
			log.Printf("Ошибка при удалении документов: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении документов"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Документы не найдены"})
			return
		}

		log.Printf("Документы успешно удалены для пользователя ID: %v", userID)
		c.JSON(http.StatusOK, gin.H{"message": "Документы успешно удалены"})
	}
}
