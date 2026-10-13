package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(context.Context) error
}

type Handler struct{ database Pinger }

func NewHandler(database Pinger) Handler { return Handler{database: database} }

func (h Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h Handler) Ready(c *gin.Context) {
	if h.database == nil {
		c.JSON(http.StatusOK, gin.H{"status": "ready", "dependencies": gin.H{"database": "not_configured"}})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.database.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "dependencies": gin.H{"database": "unavailable"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready", "dependencies": gin.H{"database": "available"}})
}
