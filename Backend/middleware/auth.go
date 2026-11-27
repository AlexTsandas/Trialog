package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access token required"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not verify access token"})
		return
	}

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(
		parts[1],
		&claims,
		func(_ *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid || claims.Subject == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired access token"})
		return
	}

	c.Set("user_id", claims.Subject)
	c.Next()
}
