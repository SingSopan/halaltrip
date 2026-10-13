package users

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"halaltrip/be/internal/auth"
	"halaltrip/be/internal/middleware"
)

const maxRequestBytes = 1 << 20

var allowedInterests = map[string]bool{"culinary": true, "culture": true, "nature": true, "shopping": true, "adventure": true, "history": true}
var allowedActivities = map[string]bool{"Relaxing": true, "Exploration": true, "Adventure": true}
var allowedTravelTypes = map[string]bool{"Solo": true, "Couple": true, "Family": true, "Friends": true}

type Handler struct {
	store            Store
	secret           []byte
	issuer, audience string
	ttlSeconds       int64
}

func NewHandler(store Store, secret []byte, issuer, audience string, ttlSeconds int64) Handler {
	return Handler{store: store, secret: secret, issuer: issuer, audience: audience, ttlSeconds: ttlSeconds}
}
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
func bind(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	if err := c.ShouldBindJSON(dst); err != nil {
		writeError(c, 400, "invalid_request", "Request JSON tidak valid.")
		return false
	}
	return true
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Register(c *gin.Context) {
	var req registerRequest
	if !bind(c, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !validName(req.Name) || !validEmail(req.Email) || len(req.Password) < 8 || len(req.Password) > 72 {
		writeError(c, 400, "invalid_request", "Nama, email, atau password tidak memenuhi format yang diperlukan.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password during registration", "error", err)
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	u, err := h.store.Create(c.Request.Context(), req.Name, req.Email, string(hash))
	if errors.Is(err, ErrConflict) {
		writeError(c, 409, "email_unavailable", "Email tidak dapat digunakan.")
		return
	}
	if err != nil {
		slog.Error("create user during registration", "error", err)
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": u, "access_token": h.token(u.ID), "token_type": "Bearer", "expires_in": h.ttlSeconds})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Login(c *gin.Context) {
	var req loginRequest
	if !bind(c, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	u, hash, err := h.store.ByEmail(c.Request.Context(), email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeError(c, 401, "invalid_credentials", "Email atau password salah.")
		return
	}
	c.JSON(200, gin.H{"user": u, "access_token": h.token(u.ID), "token_type": "Bearer", "expires_in": h.ttlSeconds})
}
func (h Handler) token(id string) string {
	token, _ := auth.SignToken(h.secret, id, h.issuer, h.audience, time.Duration(h.ttlSeconds)*time.Second, time.Now())
	return token
}

type profilePatch struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func (h Handler) Me(c *gin.Context) {
	u, err := h.store.ByID(c.Request.Context(), middleware.UserID(c))
	if errors.Is(err, ErrNotFound) {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan atau token tidak valid.")
		return
	}
	if err != nil {
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(200, gin.H{"user": u})
}
func (h Handler) PatchMe(c *gin.Context) {
	var req profilePatch
	if !bind(c, &req) {
		return
	}
	if req.Name == nil && req.Email == nil {
		writeError(c, 400, "invalid_request", "Sertakan name atau email untuk diperbarui.")
		return
	}
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if !validName(v) {
			writeError(c, 400, "invalid_request", "Nama tidak valid.")
			return
		}
		req.Name = &v
	}
	if req.Email != nil {
		v := strings.ToLower(strings.TrimSpace(*req.Email))
		if !validEmail(v) {
			writeError(c, 400, "invalid_request", "Email tidak valid.")
			return
		}
		req.Email = &v
	}
	u, err := h.store.Update(c.Request.Context(), middleware.UserID(c), req.Name, req.Email)
	if errors.Is(err, ErrConflict) {
		writeError(c, 409, "email_unavailable", "Email tidak dapat digunakan.")
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Autentikasi diperlukan atau token tidak valid.")
		return
	}
	if err != nil {
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(200, gin.H{"user": u})
}
func (h Handler) GetPreferences(c *gin.Context) {
	p, err := h.store.GetPreferences(c.Request.Context(), middleware.UserID(c))
	if errors.Is(err, ErrNotFound) {
		writeError(c, 404, "preferences_not_found", "Preferensi perjalanan belum disimpan.")
		return
	}
	if err != nil {
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(200, gin.H{"preferences": p})
}
func (h Handler) SavePreferences(c *gin.Context) {
	var p Preferences
	if !bind(c, &p) {
		return
	}
	if err := ValidatePreferences(p); err != nil {
		writeError(c, 400, "invalid_request", err.Error())
		return
	}
	if err := h.store.SavePreferences(c.Request.Context(), middleware.UserID(c), p); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
			return
		}
		writeError(c, 500, "internal_error", "Terjadi kesalahan internal.")
		return
	}
	c.JSON(200, gin.H{"preferences": p})
}

func validName(v string) bool {
	return utf8.RuneCountInString(v) > 0 && utf8.RuneCountInString(v) <= 100
}
func validEmail(v string) bool {
	if len(v) > 254 {
		return false
	}
	a, err := mail.ParseAddress(v)
	return err == nil && a.Address == v && strings.Contains(v, "@")
}
func ValidatePreferences(p Preferences) error {
	if len(p.Interests) == 0 {
		return errors.New("Pilih minimal satu interest yang valid.")
	}
	seen := map[string]bool{}
	for _, v := range p.Interests {
		if !allowedInterests[v] || seen[v] {
			return errors.New("Interest tidak valid atau duplikat.")
		}
		seen[v] = true
	}
	if !allowedActivities[p.Activity] || !allowedTravelTypes[p.TravelType] {
		return errors.New("Activity atau travel_type tidak valid.")
	}
	return nil
}
