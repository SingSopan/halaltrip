package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/auth"
)

const userIDKey = "authenticated_user_id"

func UserID(c *gin.Context) string { value, _ := c.Get(userIDKey); id, _ := value.(string); return id }

func JWT(secret []byte, issuer, audience string, optional bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" && optional {
			c.Next()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			unauthorized(c)
			return
		}
		claims, err := auth.ParseToken(strings.TrimSpace(parts[1]), secret, issuer, audience, time.Now())
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(userIDKey, claims.Subject)
		c.Next()
	}
}
func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized", "message": "Autentikasi diperlukan atau token tidak valid."}})
}
