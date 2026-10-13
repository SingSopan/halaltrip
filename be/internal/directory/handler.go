package directory

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/middleware"
)

const maxRequestBytes = 1 << 20

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

type Handler struct{ service *Service }

func NewHandler(store Store) Handler { return Handler{service: NewService(store)} }

func errorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func (h Handler) List(kind Kind) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, err := queryInt(c, "page", 1, 1, 1<<30)
		if err != nil {
			errorResponse(c, 400, "invalid_request", "page harus integer positif.")
			return
		}
		pageSize, err := queryInt(c, "page_size", 20, 1, 100)
		if err != nil {
			errorResponse(c, 400, "invalid_request", "page_size harus antara 1 dan 100.")
			return
		}
		filters := Filters{
			Query: strings.TrimSpace(c.Query("q")), City: strings.TrimSpace(c.Query("city")), Country: strings.TrimSpace(c.Query("country")),
			Cuisine: strings.TrimSpace(c.Query("cuisine")), HalalStatus: strings.TrimSpace(c.Query("halal_status")), PriceRange: strings.TrimSpace(c.Query("price_range")),
			Facility: strings.TrimSpace(c.Query("facility")), Feature: strings.TrimSpace(c.Query("feature")),
			Sort: c.DefaultQuery("sort", "name"), Order: strings.ToLower(c.DefaultQuery("order", "asc")),
			Limit: pageSize, Offset: (page - 1) * pageSize,
		}
		if (filters.Sort != "name" && filters.Sort != "city" && filters.Sort != "created_at" && filters.Sort != "updated_at") || (filters.Order != "asc" && filters.Order != "desc") {
			errorResponse(c, 400, "invalid_request", "sort atau order tidak didukung.")
			return
		}
		if err := validateFilters(kind, filters); err != nil {
			errorResponse(c, 400, "invalid_request", err.Error())
			return
		}
		items, total, err := h.service.List(c.Request.Context(), kind, filters)
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		c.JSON(200, gin.H{"data": items, "pagination": gin.H{"page": page, "page_size": pageSize, "total": total, "total_pages": (total + pageSize - 1) / pageSize}})
	}
}

func (h Handler) Get(kind Kind) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !validID(id) {
			errorResponse(c, 400, "invalid_request", "ID tempat tidak valid.")
			return
		}
		p, err := h.service.Get(c.Request.Context(), kind, id)
		if errors.Is(err, ErrNotFound) {
			errorResponse(c, 404, "not_found", "Tempat tidak ditemukan.")
			return
		}
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		c.JSON(200, gin.H{"data": p})
	}
}

func (h Handler) Create(kind Kind) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p Place
		if !bindPlace(c, &p) {
			return
		}
		created, err := h.service.Create(c.Request.Context(), kind, p)
		if invalidInput(err) {
			errorResponse(c, 400, "invalid_request", err.Error())
			return
		}
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created})
	}
}

func (h Handler) Update(kind Kind) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !validID(id) {
			errorResponse(c, 400, "invalid_request", "ID tempat tidak valid.")
			return
		}
		var p Place
		if !bindPlace(c, &p) {
			return
		}
		updated, err := h.service.Update(c.Request.Context(), kind, id, p)
		if invalidInput(err) {
			errorResponse(c, 400, "invalid_request", err.Error())
			return
		}
		if errors.Is(err, ErrNotFound) {
			errorResponse(c, 404, "not_found", "Tempat tidak ditemukan.")
			return
		}
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		c.JSON(200, gin.H{"data": updated})
	}
}

func (h Handler) Delete(kind Kind) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if !validID(id) {
			errorResponse(c, 400, "invalid_request", "ID tempat tidak valid.")
			return
		}
		err := h.service.Delete(c.Request.Context(), kind, id)
		if errors.Is(err, ErrNotFound) {
			errorResponse(c, 404, "not_found", "Tempat tidak ditemukan.")
			return
		}
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type SavedPlace struct {
	Type    Kind      `json:"place_type"`
	Place   Place     `json:"place"`
	SavedAt time.Time `json:"saved_at"`
}

func (h Handler) ListSaved(c *gin.Context) {
	refs, err := h.service.Saved(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	saved := make([]SavedPlace, 0, len(refs))
	for _, ref := range refs {
		p, err := h.service.Get(c.Request.Context(), ref.Kind, ref.ID)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
			return
		}
		saved = append(saved, SavedPlace{Type: ref.Kind, Place: p, SavedAt: ref.CreatedAt})
	}
	c.JSON(200, gin.H{"data": saved})
}

func (h Handler) Save(c *gin.Context) {
	var req struct {
		Type Kind   `json:"place_type"`
		ID   string `json:"place_id"`
	}
	if !bindPlace(c, &req) {
		return
	}
	if !validKind(req.Type) || !validID(req.ID) {
		errorResponse(c, 400, "invalid_request", "place_type atau place_id tidak valid.")
		return
	}
	err := h.service.Save(c.Request.Context(), middleware.UserID(c), req.Type, req.ID)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, 404, "not_found", "Tempat tidak ditemukan.")
		return
	}
	if err != nil {
		errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"place_type": req.Type, "place_id": req.ID})
}

func (h Handler) Unsave(c *gin.Context) {
	kind, id := Kind(c.Param("place_type")), c.Param("place_id")
	if !validKind(kind) || !validID(id) {
		errorResponse(c, 400, "invalid_request", "place_type atau place_id tidak valid.")
		return
	}
	err := h.service.Unsave(c.Request.Context(), middleware.UserID(c), kind, id)
	if errors.Is(err, ErrNotFound) {
		errorResponse(c, 404, "not_found", "Tempat tersimpan tidak ditemukan.")
		return
	}
	if err != nil {
		errorResponse(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.Status(http.StatusNoContent)
}

func bindPlace(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	if err := c.ShouldBindJSON(dst); err != nil {
		errorResponse(c, 400, "invalid_request", "Request JSON tidak valid.")
		return false
	}
	return true
}

func queryInt(c *gin.Context, key string, fallback, min, max int) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < min || v > max {
		return 0, fmt.Errorf("invalid")
	}
	return v, nil
}

func validID(id string) bool { return uuidPattern.MatchString(id) }

func validateFilters(kind Kind, f Filters) error {
	for _, v := range []string{f.Query, f.City, f.Country, f.Cuisine, f.HalalStatus, f.PriceRange, f.Facility, f.Feature} {
		if utf8.RuneCountInString(v) > 200 {
			return errors.New("filter terlalu panjang.")
		}
	}
	if kind == Restaurant && f.HalalStatus != "" && !validHalalStatus(f.HalalStatus) {
		return errors.New("halal_status tidak valid.")
	}
	if kind != Restaurant && f.HalalStatus != "" {
		return errors.New("halal_status hanya berlaku untuk restoran.")
	}
	if kind != Restaurant && f.Cuisine != "" {
		return errors.New("cuisine hanya berlaku untuk restoran.")
	}
	if kind != Hotel && f.Feature != "" {
		return errors.New("feature hanya berlaku untuk hotel.")
	}
	if kind != Mosque && f.Facility != "" {
		return errors.New("facility hanya berlaku untuk masjid.")
	}
	return nil
}

func validHalalStatus(v string) bool {
	switch v {
	case "certified", "muslim_owned", "halal_friendly", "no_pork", "unknown":
		return true
	}
	return false
}

func ValidatePlace(kind Kind, p Place) error { _, err := NormalizePlace(kind, p); return err }

func NormalizePlace(kind Kind, p Place) (Place, error) {
	if !validKind(kind) {
		return p, errors.New("jenis tempat tidak valid.")
	}
	p.Name, p.City, p.Country = strings.TrimSpace(p.Name), strings.TrimSpace(p.City), strings.TrimSpace(p.Country)
	if p.Name == "" || utf8.RuneCountInString(p.Name) > 200 || p.City == "" || utf8.RuneCountInString(p.City) > 120 || p.Country == "" || utf8.RuneCountInString(p.Country) > 120 {
		return p, errors.New("name, city, dan country wajib diisi dan harus valid.")
	}
	for _, v := range []string{p.Description, p.Address, p.Cuisine, p.PriceRange, p.HalalVerificationNote} {
		if utf8.RuneCountInString(v) > 2000 {
			return p, errors.New("field teks terlalu panjang.")
		}
	}
	if (p.Latitude == nil) != (p.Longitude == nil) {
		return p, errors.New("latitude dan longitude harus diisi bersamaan.")
	}
	if p.Latitude != nil && (*p.Latitude < -90 || *p.Latitude > 90 || *p.Longitude < -180 || *p.Longitude > 180) {
		return p, errors.New("koordinat di luar rentang yang valid.")
	}
	if p.SourceURL != "" {
		u, err := url.ParseRequestURI(p.SourceURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return p, errors.New("source_url harus berupa URL http atau https.")
		}
	}
	if p.ImageURL != "" {
		u, err := url.ParseRequestURI(p.ImageURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return p, errors.New("image_url harus berupa URL http atau https.")
		}
	}
	switch kind {
	case Restaurant:
		if p.HalalStatus == "" {
			p.HalalStatus = "unknown"
		}
		if !validHalalStatus(p.HalalStatus) {
			return p, errors.New("halal_status tidak valid.")
		}
		if p.HalalStatus == "certified" && (strings.TrimSpace(p.HalalVerificationNote) == "" || p.SourceURL == "") {
			return p, errors.New("status certified memerlukan catatan verifikasi dan source_url bukti.")
		}
	case Mosque:
		if len(p.Facilities) > 50 {
			return p, errors.New("maksimal 50 fasilitas.")
		}
		for _, v := range p.Facilities {
			if strings.TrimSpace(v) == "" || utf8.RuneCountInString(v) > 100 {
				return p, errors.New("nama fasilitas tidak valid.")
			}
		}
	case Hotel:
		if len(p.MuslimFriendlyFeatures) > 50 {
			return p, errors.New("maksimal 50 fitur.")
		}
		for _, f := range p.MuslimFriendlyFeatures {
			if strings.TrimSpace(f.Name) == "" || utf8.RuneCountInString(f.Name) > 120 {
				return p, errors.New("nama fitur tidak valid.")
			}
			if f.Verified && (strings.TrimSpace(f.VerificationNote) == "" || p.SourceURL == "") {
				return p, errors.New("fitur terverifikasi memerlukan verification_note dan source_url.")
			}
		}
	}
	return p, nil
}
