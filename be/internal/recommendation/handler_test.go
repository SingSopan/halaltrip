package recommendation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"halaltrip/be/internal/auth"
	"halaltrip/be/internal/middleware"
	"halaltrip/be/internal/travelprovider"
	"halaltrip/be/internal/users"
)

func testRouter(providerURL string, timeout time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewHandler(travelprovider.NewClient(providerURL, timeout))
	r := gin.New()
	r.GET("/api/v1/health/recommendation-provider", h.ProviderHealth)
	r.POST("/api/v1/recommendations/cities", h.Cities)
	r.POST("/api/v1/recommendations/destinations", h.DestinationsRecommendation)
	r.GET("/api/v1/destinations", h.ListDestinations)
	r.GET("/api/v1/destinations/:id", h.GetDestination)
	r.POST("/api/v1/destinations", h.CreateDestination)
	r.PUT("/api/v1/destinations/:id", h.UpdateDestination)
	r.DELETE("/api/v1/destinations/:id", h.DeleteDestination)
	return r
}

type savedPreferenceStore struct{}

func (savedPreferenceStore) GetPreferences(context.Context, string) (users.Preferences, error) {
	return users.Preferences{Interests: []string{"culture", "history"}, Activity: "Exploration", TravelType: "Solo"}, nil
}

func TestAuthenticatedRecommendationUsesSavedPreferencesWhenOmitted(t *testing.T) {
	secret := []byte("recommendation-test-secret-at-least-32-bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["activity"] != "Exploration" || body["travel_type"] != "Solo" {
			t.Errorf("request body = %#v", body)
		}
		w.Write([]byte(`{"recommendations":[]}`))
	}))
	defer server.Close()
	h := NewHandler(travelprovider.NewClient(server.URL, time.Second), savedPreferenceStore{})
	r := gin.New()
	r.Use(middleware.JWT(secret, "", "", true))
	r.POST("/cities", h.Cities)
	unauth := perform(r, http.MethodPost, "/cities", `{}`)
	if unauth.Code != http.StatusBadRequest {
		t.Fatalf("without token status = %d", unauth.Code)
	}
	token, err := auth.SignToken(secret, "user-1", "", "", time.Hour, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/cities", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("with saved preferences status = %d body %s", res.Code, res.Body.String())
	}
}

func perform(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	r.ServeHTTP(res, req)
	return res
}

func TestRecommendationRequestsAreForwardedAndResponsePreserved(t *testing.T) {
	cases := []struct{ name, internal, external, request, response string }{
		{"cities", "/api/v1/recommendations/cities", "/recommend/cities", `{"interests":["culture","history"],"activity":"Exploration","travel_type":"Solo","top_n":5}`, `{"recommendations":[{"city":"Yogyakarta"}]}`},
		{"destinations", "/api/v1/recommendations/destinations", "/recommend/destinations", `{"city":"Yogyakarta","interests":["culture"],"activity":"Relaxing","travel_type":"Family"}`, `{"recommendations":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != tc.external {
					t.Errorf("upstream request = %s %s", r.Method, r.URL.Path)
				}
				body, _ := io.ReadAll(r.Body)
				var got, want any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(tc.request), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("request body = %s, want %s", body, tc.request)
				}
				w.Header().Set("Content-Type", "application/json")
				io.WriteString(w, tc.response)
			}))
			defer server.Close()
			res := perform(testRouter(server.URL, time.Second), http.MethodPost, tc.internal, tc.request)
			if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != tc.response {
				t.Fatalf("response = %d %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestDestinationCRUDProxy(t *testing.T) {
	requests := []struct {
		method, internal, external, body, response string
		status                                     int
	}{
		{http.MethodGet, "/api/v1/destinations", "/destinations", "", `[{"id":1}]`, 200},
		{http.MethodGet, "/api/v1/destinations/12", "/destinations/12", "", `{"id":12}`, 200},
		{http.MethodPost, "/api/v1/destinations", "/destinations", `{"name":"Yogyakarta","custom":"kept"}`, `{"id":12,"name":"Yogyakarta"}`, 201},
		{http.MethodPut, "/api/v1/destinations/12", "/destinations/12", `{"name":"Jogja","custom":"kept"}`, `{"id":12,"name":"Jogja"}`, 200},
		{http.MethodDelete, "/api/v1/destinations/12", "/destinations/12", "", "", 204},
	}
	for _, tc := range requests {
		t.Run(fmt.Sprintf("%s %s", tc.method, tc.internal), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method || r.URL.Path != tc.external {
					t.Errorf("upstream request = %s %s", r.Method, r.URL.Path)
				}
				if tc.body != "" {
					body, _ := io.ReadAll(r.Body)
					if string(body) != tc.body {
						t.Errorf("body = %s, want %s", body, tc.body)
					}
				}
				w.WriteHeader(tc.status)
				if tc.response != "" {
					io.WriteString(w, tc.response)
				}
			}))
			defer server.Close()
			res := perform(testRouter(server.URL, time.Second), tc.method, tc.internal, tc.body)
			if res.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", res.Code, tc.status, res.Body.String())
			}
			if tc.response != "" && strings.TrimSpace(res.Body.String()) != tc.response {
				t.Errorf("response = %s", res.Body.String())
			}
		})
	}
}

func TestInvalidRecommendationRequestDoesNotCallProvider(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; io.WriteString(w, `{}`) }))
	defer server.Close()
	r := testRouter(server.URL, time.Second)
	invalid := []struct{ path, body string }{
		{"/api/v1/recommendations/cities", `{`},
		{"/api/v1/recommendations/cities", `{"interests":[],"activity":"Relaxing","travel_type":"Solo"}`},
		{"/api/v1/recommendations/cities", `{"interests":["beach"],"activity":"Relaxing","travel_type":"Solo"}`},
		{"/api/v1/recommendations/cities", `{"interests":["nature"],"activity":"relaxing","travel_type":"Solo"}`},
		{"/api/v1/recommendations/cities", `{"interests":["nature"],"activity":"Relaxing","travel_type":"solo"}`},
		{"/api/v1/recommendations/cities", `{"interests":["nature"],"activity":"Relaxing","travel_type":"Solo","top_n":101}`},
		{"/api/v1/recommendations/destinations", `{"interests":["nature"],"activity":"Relaxing","travel_type":"Solo"}`},
	}
	for _, tc := range invalid {
		res := perform(r, http.MethodPost, tc.path, tc.body)
		if res.Code != http.StatusBadRequest {
			t.Errorf("%s response status = %d, want 400", tc.body, res.Code)
		}
	}
	if calls != 0 {
		t.Errorf("provider called %d times for invalid requests", calls)
	}
}

func TestProviderFailuresMapToGatewayErrors(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(80 * time.Millisecond)
			io.WriteString(w, `{}`)
		}))
		defer server.Close()
		res := perform(testRouter(server.URL, 10*time.Millisecond), http.MethodGet, "/api/v1/destinations", "")
		if res.Code != http.StatusGatewayTimeout {
			t.Fatalf("status = %d, want 504", res.Code)
		}
	})
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   int
	}{
		{"provider error", 503, `{"secret":"hidden"}`, http.StatusBadGateway},
		{"invalid json", 200, `not-json`, http.StatusBadGateway},
		{"not found", 404, `{"detail":"hidden"}`, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); io.WriteString(w, tc.body) }))
			defer server.Close()
			res := perform(testRouter(server.URL, time.Second), http.MethodGet, "/api/v1/destinations/9", "")
			if res.Code != tc.want {
				t.Fatalf("status = %d, want %d", res.Code, tc.want)
			}
			if strings.Contains(res.Body.String(), "hidden") || strings.Contains(res.Body.String(), "secret") {
				t.Errorf("provider detail leaked: %s", res.Body.String())
			}
		})
	}
}

func TestProviderHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/health" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"status":"ok"}`)
	}))
	defer server.Close()
	res := perform(testRouter(server.URL, time.Second), http.MethodGet, "/api/v1/health/recommendation-provider", "")
	if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}
