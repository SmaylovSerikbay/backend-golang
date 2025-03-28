package handlers

import "github.com/gin-gonic/gin"

func GetNearbyDrivers(c *gin.Context) {
	c.JSON(200, gin.H{"message": "get nearby drivers endpoint"})
}

func UpdateDriverLocation(c *gin.Context) {
	c.JSON(200, gin.H{"message": "update driver location endpoint"})
}
