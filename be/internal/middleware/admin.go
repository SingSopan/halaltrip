package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Admin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var role string
		err := pool.QueryRow(c.Request.Context(), `SELECT role FROM users WHERE id=$1`, UserID(c)).Scan(&role)
		if err != nil || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "forbidden", "message": "Akses administrator diperlukan."}})
			return
		}
		c.Next()
	}
}
