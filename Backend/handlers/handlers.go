package handlers

import (
	"net/http"

	"github.com/AlexTsandas/Trialog/Backend/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetTest(c *gin.Context) {

	message := "Hello, World!"
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

type signupRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SignUp(c *gin.Context) {
	var input signupRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signup input"})
		return
	}

	if len(input.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password exceeds 72 bytes"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	_, err = config.DB.Exec(
		c.Request.Context(),
		`INSERT INTO users (username, email, password_hash)
     VALUES ($1, $2, $3)`,
		input.Username,
		input.Email,
		string(passwordHash),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

type signInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SignIn(c *gin.Context) {
	var input signInRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sign-in input"})
		return
	}

	// to be continued...
}
