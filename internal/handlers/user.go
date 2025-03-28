package handlers

import "github.com/gin-gonic/gin"

func GetProfile(c *gin.Context) {
	c.JSON(200, gin.H{"message": "get profile endpoint"})
}

func UpdateProfile(c *gin.Context) {
	c.JSON(200, gin.H{"message": "update profile endpoint"})
}
