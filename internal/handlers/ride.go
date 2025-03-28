package handlers

import "github.com/gin-gonic/gin"

func RequestRide(c *gin.Context) {
	c.JSON(200, gin.H{"message": "request ride endpoint"})
}

func GetActiveRide(c *gin.Context) {
	c.JSON(200, gin.H{"message": "get active ride endpoint"})
}

func CancelRide(c *gin.Context) {
	c.JSON(200, gin.H{"message": "cancel ride endpoint"})
}

func GetRideHistory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "get ride history endpoint"})
}
