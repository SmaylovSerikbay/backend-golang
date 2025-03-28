package handlers

import (
	"net/http"
	"taxi-backend/internal/models"
	"time"

	"log"

	"bytes"
	"io"

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

		log.Printf("Получение документов для пользователя ID: %v", userID)

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
						Email:     user.Email,
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
				c.JSON(http.StatusNotFound, gin.H{"error": "Документы не найдены"})
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
				Email:     user.Email,
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
			RejectionReason string                      `json:"rejectionReason"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Ошибка валидации данных: %v", err)
			log.Printf("Тело запроса: %s", bodyString)

			// Попробуем другой формат данных (с snake_case)
			var altReq struct {
				Status          models.DriverDocumentStatus `json:"status" binding:"required"`
				RejectionReason string                      `json:"rejection_reason"`
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

		log.Printf("Получен запрос на обновление статуса документа: %+v", req)

		docID := c.Param("id")
		log.Printf("ID документа из параметров: %s", docID)

		var docs models.DriverDocuments

		if err := db.First(&docs, docID).Error; err != nil {
			log.Printf("Документы не найдены: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Документы не найдены"})
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

		// Получаем информацию о пользователе для ответа
		var user models.User
		if err := db.First(&user, docs.UserID).Error; err != nil {
			log.Printf("Ошибка при получении пользователя для документа ID %v: %v", docs.ID, err)
		}

		log.Printf("Статус документа успешно обновлен: %+v", docs)
		c.JSON(http.StatusOK, models.DriverDocumentsResponse{
			ID:     docs.ID,
			UserID: docs.UserID,
			User: &models.UserResponse{
				ID:        user.ID,
				Email:     user.Email,
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
