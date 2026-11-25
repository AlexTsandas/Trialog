package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTest(c *gin.Context) {

	message := "Hello, World!"
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}
