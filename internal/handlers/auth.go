package handlers

import (
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	c.JSON(200, gin.H{"message": "register endpoint"})
}

func Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "login endpoint"})
}
