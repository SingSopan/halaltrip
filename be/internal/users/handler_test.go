package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"halaltrip/be/internal/auth"
	"halaltrip/be/internal/middleware"
)

type memoryStore struct {
	user  User
	hash  string
	prefs Preferences
}

func (s *memoryStore) Create(_ context.Context, name, email, hash string) (User, error) {
	if s.user.Email == email {
		return User{}, ErrConflict
	}
	s.user = User{ID: "user-1", Name: name, Email: email}
	s.hash = hash
	return s.user, nil
}
func (s *memoryStore) ByEmail(_ context.Context, email string) (User, string, error) {
	if s.user.Email != email {
		return User{}, "", ErrNotFound
	}
	return s.user, s.hash, nil
}
func (s *memoryStore) ByID(_ context.Context, id string) (User, error) {
	if s.user.ID != id {
		return User{}, ErrNotFound
	}
	return s.user, nil
}
func (s *memoryStore) Update(_ context.Context, id string, name, email *string) (User, error) {
	if s.user.ID != id {
		return User{}, ErrNotFound
	}
	if name != nil {
		s.user.Name = *name
	}
	if email != nil {
		s.user.Email = *email
	}
	return s.user, nil
}
func (s *memoryStore) GetPreferences(_ context.Context, id string) (Preferences, error) {
	if s.user.ID != id || len(s.prefs.Interests) == 0 {
		return Preferences{}, ErrNotFound
	}
	return s.prefs, nil
}
func (s *memoryStore) SavePreferences(_ context.Context, id string, p Preferences) error {
	if s.user.ID != id {
		return ErrNotFound
	}
	s.prefs = p
	return nil
}

func userTestRouter(store Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	secret := []byte("unit-test-secret-that-is-at-least-32-bytes")
	h := NewHandler(store, secret, "test-issuer", "test-audience", 3600)
	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	required := middleware.JWT(secret, "test-issuer", "test-audience", false)
	g := r.Group("/users/me", required)
	g.GET("", h.Me)
	g.PATCH("", h.PatchMe)
	g.PATCH("/preferences", h.SavePreferences)
	g.GET("/preferences", h.GetPreferences)
	return r
}
func send(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	r.ServeHTTP(w, req)
	return w
}
func tokenFor(t *testing.T, id string, expires time.Duration) string {
	t.Helper()
	token, err := auth.SignToken([]byte("unit-test-secret-that-is-at-least-32-bytes"), id, "test-issuer", "test-audience", expires, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRegisterSuccessAndDuplicateEmail(t *testing.T) {
	store := &memoryStore{}
	r := userTestRouter(store)
	res := send(r, http.MethodPost, "/register", `{"name":"Alya","email":"ALYA@example.com","password":"password-kuat"}`, "")
	if res.Code != http.StatusCreated {
		t.Fatalf("register = %d %s", res.Code, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "password_hash") || strings.Contains(res.Body.String(), "password-kuat") {
		t.Fatalf("sensitive data leaked: %s", res.Body.String())
	}
	if bcrypt.CompareHashAndPassword([]byte(store.hash), []byte("password-kuat")) != nil {
		t.Fatal("password was not bcrypt hashed")
	}
	res = send(r, http.MethodPost, "/register", `{"name":"Alya","email":"alya@example.com","password":"password-kuat"}`, "")
	if res.Code != http.StatusConflict {
		t.Fatalf("duplicate register = %d", res.Code)
	}
}
func TestLoginSuccessAndFailureHaveSafeMessage(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password-kuat"), bcrypt.MinCost)
	store := &memoryStore{user: User{ID: "user-1", Name: "Alya", Email: "alya@example.com"}, hash: string(hash)}
	r := userTestRouter(store)
	res := send(r, http.MethodPost, "/login", `{"email":"ALYA@example.com","password":"password-kuat"}`, "")
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "access_token") {
		t.Fatalf("login = %d %s", res.Code, res.Body.String())
	}
	res = send(r, http.MethodPost, "/login", `{"email":"missing@example.com","password":"wrong"}`, "")
	if res.Code != http.StatusUnauthorized || !strings.Contains(res.Body.String(), "Email atau password salah") || strings.Contains(res.Body.String(), "missing@example.com") {
		t.Fatalf("failure response leaked account status: %s", res.Body.String())
	}
}
func TestProtectedProfileAndOwnershipUseJWTSubject(t *testing.T) {
	store := &memoryStore{user: User{ID: "user-1", Name: "Alya", Email: "alya@example.com"}}
	r := userTestRouter(store)
	if res := send(r, http.MethodGet, "/users/me", "", ""); res.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous profile status = %d", res.Code)
	}
	res := send(r, http.MethodPatch, "/users/me", `{"id":"user-2","name":"Updated"}`, tokenFor(t, "user-1", time.Hour))
	if res.Code != http.StatusOK || store.user.Name != "Updated" || store.user.ID != "user-1" {
		t.Fatalf("profile patch = %d %s", res.Code, res.Body.String())
	}
}
func TestPreferencesValidationAndUserOwnership(t *testing.T) {
	store := &memoryStore{user: User{ID: "user-1", Name: "Alya", Email: "alya@example.com"}}
	r := userTestRouter(store)
	own := tokenFor(t, "user-1", time.Hour)
	bad := send(r, http.MethodPatch, "/users/me/preferences", `{"interests":["beach"],"activity":"Relaxing","travel_type":"Solo"}`, own)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("invalid preferences status = %d", bad.Code)
	}
	good := send(r, http.MethodPatch, "/users/me/preferences", `{"interests":["culture","history"],"activity":"Exploration","travel_type":"Solo"}`, own)
	if good.Code != http.StatusOK {
		t.Fatalf("save preferences = %d %s", good.Code, good.Body.String())
	}
	foreign := send(r, http.MethodPatch, "/users/me/preferences", `{"interests":["nature"],"activity":"Relaxing","travel_type":"Solo"}`, tokenFor(t, "user-2", time.Hour))
	if foreign.Code != http.StatusNotFound || store.prefs.Interests[0] != "culture" {
		t.Fatalf("foreign token changed prefs: %d", foreign.Code)
	}
}
func TestExpiredAndInvalidTokensAreRejected(t *testing.T) {
	store := &memoryStore{user: User{ID: "user-1", Name: "Alya", Email: "alya@example.com"}}
	r := userTestRouter(store)
	for _, token := range []string{"not-a-jwt", tokenFor(t, "user-1", -time.Hour)} {
		res := send(r, http.MethodGet, "/users/me", "", token)
		if res.Code != http.StatusUnauthorized {
			t.Errorf("token %q status = %d", token, res.Code)
		}
	}
}
func TestValidatePreferencesRejectsDuplicatesAndUnsupportedValues(t *testing.T) {
	for _, p := range []Preferences{{Interests: []string{"culture", "culture"}, Activity: "Exploration", TravelType: "Solo"}, {Interests: []string{"nature"}, Activity: "relaxing", TravelType: "Solo"}, {Interests: []string{"nature"}, Activity: "Relaxing", TravelType: "solo"}} {
		if err := ValidatePreferences(p); err == nil {
			t.Errorf("accepted invalid preferences: %+v", p)
		}
	}
	if err := ValidatePreferences(Preferences{Interests: []string{"culture"}, Activity: "Exploration", TravelType: "Family"}); err != nil {
		t.Fatal(err)
	}
}
