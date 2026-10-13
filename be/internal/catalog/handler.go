package catalog

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Destination struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	Country          string   `json:"country"`
	City             string   `json:"city"`
	Description      string   `json:"description"`
	ImageURL         string   `json:"image_url"`
	Latitude         *float64 `json:"latitude,omitempty"`
	Longitude        *float64 `json:"longitude,omitempty"`
	SourceURL        string   `json:"source_url,omitempty"`
	ImageAttribution string   `json:"image_attribution,omitempty"`
	ImageLicenseURL  string   `json:"image_license_url,omitempty"`
}

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) Handler { return Handler{pool: pool} }

func (h Handler) ListDestinations(c *gin.Context) {
	city := strings.TrimSpace(c.Query("city"))
	if utf8.RuneCountInString(city) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request", "message": "Filter kota terlalu panjang."}})
		return
	}
	rows, err := h.pool.Query(c.Request.Context(), `SELECT id,slug,name,country,city,description,image_url,latitude,longitude,source_url,image_attribution,image_license_url FROM destinations WHERE ($1 = '' OR city ILIKE $1) ORDER BY name,id`, city)
	if err != nil {
		writeInternalError(c)
		return
	}
	defer rows.Close()
	items := make([]Destination, 0)
	for rows.Next() {
		var item Destination
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name, &item.Country, &item.City, &item.Description, &item.ImageURL, &item.Latitude, &item.Longitude, &item.SourceURL, &item.ImageAttribution, &item.ImageLicenseURL); err != nil {
			writeInternalError(c)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		writeInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h Handler) GetDestination(c *gin.Context) {
	id := c.Param("id")
	if !uuidPattern.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request", "message": "ID destinasi tidak valid."}})
		return
	}
	var item Destination
	err := h.pool.QueryRow(c.Request.Context(), `SELECT id,slug,name,country,city,description,image_url,latitude,longitude,source_url,image_attribution,image_license_url FROM destinations WHERE id=$1`, id).Scan(&item.ID, &item.Slug, &item.Name, &item.Country, &item.City, &item.Description, &item.ImageURL, &item.Latitude, &item.Longitude, &item.SourceURL, &item.ImageAttribution, &item.ImageLicenseURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "not_found", "message": "Destinasi tidak ditemukan."}})
			return
		}
		writeInternalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func writeInternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "Terjadi kesalahan internal."}})
}
