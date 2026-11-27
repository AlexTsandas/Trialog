package handlers

import (
	"errors"
	"net/http"

	"time"

	"github.com/AlexTsandas/Trialog/Backend/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
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

	var userID int
	var storedHash string

	err := config.DB.QueryRow(
		c.Request.Context(),
		`SELECT id, password_hash FROM users WHERE email = $1`,
		input.Email,
	).Scan(&userID, &storedHash)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not sign in"})
		return
	}

	if len(input.Password) > 72 ||
		bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	accessToken, err := createAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create access token"})
		return
	}

	refreshToken := rand.Text()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = config.DB.Exec(
		c.Request.Context(),
		`INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
     VALUES ($1, $2, $3)`,
		userID,
		hashRefreshToken(refreshToken),
		expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
		return
	}

	setRefreshCookie(c, refreshToken, expiresAt)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func Refresh(c *gin.Context) {
	oldToken, err := c.Cookie("refresh_token")
	if err != nil || oldToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Please sign in again"})
		return
	}

	newToken := rand.Text()
	ctx := c.Request.Context()

	tx, err := config.DB.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not refresh session"})
		return
	}
	defer tx.Rollback(ctx)

	var userID int
	var expiresAt time.Time

	err = tx.QueryRow(
		ctx,
		`UPDATE sessions
         SET refresh_token_hash = $1
         WHERE refresh_token_hash = $2
           AND expires_at > CURRENT_TIMESTAMP
           AND revoked_at IS NULL
         RETURNING user_id, expires_at`,
		hashRefreshToken(newToken),
		hashRefreshToken(oldToken),
	).Scan(&userID, &expiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Please sign in again"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not refresh session"})
		return
	}

	accessToken, err := createAccessToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create access token"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not refresh session"})
		return
	}

	setRefreshCookie(c, newToken, expiresAt)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
