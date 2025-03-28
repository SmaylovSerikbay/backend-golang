package handlers

import (
	"net/http"
	"os"
	"taxi-backend/internal/models"
	"taxi-backend/internal/services/dgis"

	"log"

	"github.com/gin-gonic/gin"
)

type AddressSearchRequest struct {
	Query string `json:"query" binding:"required"`
}

type AddressSearchResponse struct {
	Addresses []AddressResult `json:"addresses"`
}

type AddressResult struct {
	Name        string          `json:"name"`
	FullAddress string          `json:"fullAddress"`
	Location    models.Location `json:"location"`
}

func SearchAddress(c *gin.Context) {
	var req AddressSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ошибка при разборе запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	log.Printf("Поиск адреса с запросом: %s", req.Query)

	client := dgis.NewClient(os.Getenv("DGIS_API_KEY"))
	result, err := client.SearchAddress(req.Query)
	if err != nil {
		log.Printf("Ошибка при поиске адреса в 2GIS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при поиске адреса"})
		return
	}

	log.Printf("Найдено %d результатов", len(result.Result.Items))

	addresses := make([]AddressResult, 0)
	for _, item := range result.Result.Items {
		fullAddress := item.FullName
		if fullAddress == "" {
			fullAddress = item.Address.FullName
			if fullAddress == "" {
				fullAddress = item.Address.Name
			}
		}

		log.Printf("Добавляем адрес: %s (тип: %s)", fullAddress, item.Type)

		addresses = append(addresses, AddressResult{
			Name:        item.Name,
			FullAddress: fullAddress,
			Location: models.Location{
				Latitude:  item.Point.Lat,
				Longitude: item.Point.Lon,
			},
		})
	}

	c.JSON(http.StatusOK, AddressSearchResponse{
		Addresses: addresses,
	})
}
