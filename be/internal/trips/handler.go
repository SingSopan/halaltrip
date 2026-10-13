package trips

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/middleware"
)

const maxRequestBytes = 1 << 20

var idPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Handler struct{ service *Service }

func NewHandler(store Store) Handler { return Handler{service: NewService(store)} }
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
func bind(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	if err := c.ShouldBindJSON(target); err != nil {
		writeError(c, 400, "invalid_request", "Request JSON tidak valid.")
		return false
	}
	return true
}
func validID(id string) bool { return idPattern.MatchString(id) }
func invalid(c *gin.Context) {
	writeError(c, 400, "invalid_request", "ID perjalanan atau item tidak valid.")
}
func serviceError(c *gin.Context, err error) {
	switch {
	case IsValidation(err):
		writeError(c, 400, "invalid_request", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(c, 404, "not_found", "Perjalanan atau item itinerary tidak ditemukan.")
	default:
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
	}
}

func (h Handler) CreateTrip(c *gin.Context) {
	var t Trip
	if !bind(c, &t) {
		return
	}
	created, err := h.service.CreateTrip(c.Request.Context(), middleware.UserID(c), t)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": created})
}
func (h Handler) ListTrips(c *gin.Context) {
	items, err := h.service.ListTrips(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": items})
}
func (h Handler) GetTrip(c *gin.Context) {
	id := c.Param("tripId")
	if !validID(id) {
		invalid(c)
		return
	}
	trip, err := h.service.GetTrip(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": trip})
}
func (h Handler) PatchTrip(c *gin.Context) {
	id := c.Param("tripId")
	if !validID(id) {
		invalid(c)
		return
	}
	var p TripPatch
	if !bind(c, &p) {
		return
	}
	if p.Title == nil && p.StartDate == nil && p.EndDate == nil && p.Origin == nil && p.Notes == nil && p.Status == nil {
		writeError(c, 400, "invalid_request", "Sertakan field perjalanan untuk diperbarui.")
		return
	}
	trip, err := h.service.UpdateTrip(c.Request.Context(), middleware.UserID(c), id, p)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": trip})
}
func (h Handler) DeleteTrip(c *gin.Context) {
	id := c.Param("tripId")
	if !validID(id) {
		invalid(c)
		return
	}
	if err := h.service.DeleteTrip(c.Request.Context(), middleware.UserID(c), id); err != nil {
		serviceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h Handler) CreateItem(c *gin.Context) {
	tripID := c.Param("tripId")
	if !validID(tripID) {
		invalid(c)
		return
	}
	var item Item
	if !bind(c, &item) {
		return
	}
	created, err := h.service.CreateItem(c.Request.Context(), middleware.UserID(c), tripID, item)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": created})
}
func (h Handler) ListItems(c *gin.Context) {
	tripID := c.Param("tripId")
	if !validID(tripID) {
		invalid(c)
		return
	}
	items, err := h.service.ListItems(c.Request.Context(), middleware.UserID(c), tripID)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": items})
}
func (h Handler) PatchItem(c *gin.Context) {
	tripID, itemID := c.Param("tripId"), c.Param("itemId")
	if !validID(tripID) || !validID(itemID) {
		invalid(c)
		return
	}
	var p ItemPatch
	if !bind(c, &p) {
		return
	}
	if p.Title == nil && p.Description == nil && p.PlaceType == nil && p.ExternalPlaceID == nil && p.ScheduledDate == nil && p.StartTime == nil && p.EndTime == nil && p.TransportMode == nil && p.Notes == nil {
		writeError(c, 400, "invalid_request", "Sertakan field itinerary untuk diperbarui.")
		return
	}
	item, err := h.service.UpdateItem(c.Request.Context(), middleware.UserID(c), tripID, itemID, p)
	if err != nil {
		serviceError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": item})
}
func (h Handler) DeleteItem(c *gin.Context) {
	tripID, itemID := c.Param("tripId"), c.Param("itemId")
	if !validID(tripID) || !validID(itemID) {
		invalid(c)
		return
	}
	if err := h.service.DeleteItem(c.Request.Context(), middleware.UserID(c), tripID, itemID); err != nil {
		serviceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h Handler) ReorderItems(c *gin.Context) {
	tripID := c.Param("tripId")
	if !validID(tripID) {
		invalid(c)
		return
	}
	var req struct {
		ItemIDs []string `json:"item_ids"`
	}
	if !bind(c, &req) {
		return
	}
	if req.ItemIDs == nil {
		writeError(c, 400, "invalid_request", "item_ids wajib berupa array.")
		return
	}
	if err := h.service.ReorderItems(c.Request.Context(), middleware.UserID(c), tripID, req.ItemIDs); err != nil {
		serviceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
