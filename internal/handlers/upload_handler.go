package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadFile(c *gin.Context) {
	// Получаем файл из запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}

	// Создаем уникальное имя файла
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Создаем директорию для загрузок, если она не существует
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании директории"})
		return
	}

	// Создаем поддиректорию по дате
	now := time.Now()
	dateDir := filepath.Join(uploadDir, now.Format("2006/01/02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании директории"})
		return
	}

	// Полный путь к файлу
	filePath := filepath.Join(dateDir, newFileName)

	// Сохраняем файл
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении файла"})
		return
	}

	// Возвращаем URL файла
	fileURL := fmt.Sprintf("/uploads/%s/%s", now.Format("2006/01/02"), newFileName)
	c.JSON(http.StatusOK, gin.H{
		"url": fileURL,
	})
}
