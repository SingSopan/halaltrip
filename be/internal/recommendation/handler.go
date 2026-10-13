package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/middleware"
	"halaltrip/be/internal/travelprovider"
	"halaltrip/be/internal/users"
)

const maxRequestBytes = 1 << 20

type Provider interface {
	Do(ctx context.Context, method, path string, body []byte) (int, []byte, error)
}

type PreferenceStore interface {
	GetPreferences(context.Context, string) (users.Preferences, error)
}
type Handler struct {
	provider    Provider
	preferences PreferenceStore
}

func NewHandler(provider Provider, stores ...PreferenceStore) Handler {
	h := Handler{provider: provider}
	if len(stores) > 0 {
		h.preferences = stores[0]
	}
	return h
}

type recommendRequest struct {
	Interests  []string `json:"interests"`
	Activity   string   `json:"activity"`
	TravelType string   `json:"travel_type"`
	TopN       *int     `json:"top_n,omitempty"`
	City       string   `json:"city,omitempty"`
}

var validInterests = map[string]bool{"culinary": true, "culture": true, "nature": true, "shopping": true, "adventure": true, "history": true}
var validActivities = map[string]bool{"Relaxing": true, "Exploration": true, "Adventure": true}
var validTravelTypes = map[string]bool{"Solo": true, "Couple": true, "Family": true, "Friends": true}

func (h Handler) ProviderHealth(c *gin.Context) { h.forward(c, http.MethodGet, "/health", nil, false) }

func (h Handler) Cities(c *gin.Context) {
	var req recommendRequest
	if !h.bindRecommendation(c, &req, false) {
		return
	}
	h.forwardJSON(c, http.MethodPost, "/recommend/cities", req)
}

func (h Handler) DestinationsRecommendation(c *gin.Context) {
	var req recommendRequest
	if !h.bindRecommendation(c, &req, true) {
		return
	}
	h.forwardJSON(c, http.MethodPost, "/recommend/destinations", req)
}

func (h Handler) bindRecommendation(c *gin.Context, req *recommendRequest, requireCity bool) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	if err := c.ShouldBindJSON(req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request JSON tidak valid.")
		return false
	}
	if len(req.Interests) == 0 && req.Activity == "" && req.TravelType == "" {
		if userID := middleware.UserID(c); userID != "" && h.preferences != nil {
			prefs, err := h.preferences.GetPreferences(c.Request.Context(), userID)
			if err != nil {
				writeError(c, http.StatusBadRequest, "preferences_required", "Kirim preferensi atau simpan preferensi pengguna terlebih dahulu.")
				return false
			}
			if err := users.ValidatePreferences(prefs); err != nil {
				writeError(c, http.StatusBadRequest, "preferences_invalid", "Preferensi tersimpan tidak valid; perbarui preferensi sebelum meminta rekomendasi.")
				return false
			}
			req.Interests, req.Activity, req.TravelType = prefs.Interests, prefs.Activity, prefs.TravelType
		} else {
			writeError(c, http.StatusBadRequest, "invalid_request", "Kirim interests, activity, dan travel_type.")
			return false
		}
	} else if len(req.Interests) == 0 || req.Activity == "" || req.TravelType == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "Kirim interests, activity, dan travel_type secara lengkap.")
		return false
	}
	if len(req.Interests) == 0 {
		writeError(c, http.StatusBadRequest, "invalid_request", "Pilih minimal satu interest yang valid.")
		return false
	}
	for _, interest := range req.Interests {
		if !validInterests[interest] {
			writeError(c, http.StatusBadRequest, "invalid_request", "Interest tidak valid.")
			return false
		}
	}
	if !validActivities[req.Activity] || !validTravelTypes[req.TravelType] {
		writeError(c, http.StatusBadRequest, "invalid_request", "Activity atau travel_type tidak valid.")
		return false
	}
	if req.TopN != nil && (*req.TopN < 1 || *req.TopN > 100) {
		writeError(c, http.StatusBadRequest, "invalid_request", "top_n harus berupa integer antara 1 dan 100.")
		return false
	}
	if requireCity && strings.TrimSpace(req.City) == "" {
		writeError(c, http.StatusBadRequest, "invalid_request", "Field city wajib diisi.")
		return false
	}
	return true
}

func (h Handler) ListDestinations(c *gin.Context) {
	h.forward(c, http.MethodGet, "/destinations", nil, false)
}
func (h Handler) GetDestination(c *gin.Context) {
	h.forward(c, http.MethodGet, "/destinations/"+c.Param("id"), nil, false)
}

func (h Handler) CreateDestination(c *gin.Context) {
	h.forwardRawJSON(c, http.MethodPost, "/destinations")
}
func (h Handler) UpdateDestination(c *gin.Context) {
	h.forwardRawJSON(c, http.MethodPut, "/destinations/"+c.Param("id"))
}
func (h Handler) DeleteDestination(c *gin.Context) {
	h.forward(c, http.MethodDelete, "/destinations/"+c.Param("id"), nil, true)
}

func (h Handler) forwardJSON(c *gin.Context, method, path string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	h.forward(c, method, path, body, false)
}

func (h Handler) forwardRawJSON(c *gin.Context, method, path string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(strings.TrimSpace(string(body))) == 0 || !json.Valid(body) {
		writeError(c, http.StatusBadRequest, "invalid_request", "Request JSON tidak valid.")
		return
	}
	h.forward(c, method, path, body, method == http.MethodDelete)
}

func (h Handler) forward(c *gin.Context, method, path string, body []byte, allowEmpty bool) {
	status, response, err := h.provider.Do(c.Request.Context(), method, path, body)
	if err != nil {
		var providerHTTPError *travelprovider.HTTPError
		switch {
		case errors.Is(err, travelprovider.ErrTimeout):
			writeError(c, http.StatusGatewayTimeout, "provider_timeout", "Layanan rekomendasi tidak merespons tepat waktu.")
		case errors.As(err, &providerHTTPError) && providerHTTPError.StatusCode == http.StatusNotFound:
			writeError(c, http.StatusNotFound, "not_found", "Data tidak ditemukan.")
		default:
			writeError(c, http.StatusBadGateway, "provider_error", "Layanan rekomendasi sedang tidak tersedia.")
		}
		return
	}
	if allowEmpty && status == http.StatusNoContent {
		c.Status(status)
		return
	}
	if !json.Valid(response) {
		writeError(c, http.StatusBadGateway, "provider_error", "Layanan rekomendasi mengembalikan respons tidak valid.")
		return
	}
	c.Data(status, "application/json; charset=utf-8", response)
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
