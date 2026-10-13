package router

import (
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"halaltrip/be/internal/config"
)

func TestRegistersExpectedRoutes(t *testing.T) {
	pool, err := pgxpool.New(t.Context(), "postgres://postgres:postgres@localhost:5432/halaltrip?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	r := New(config.Config{AppEnv: "test", JWTSecret: "01234567890123456789012345678901", JWTAccessTokenTTL: 3600_000_000_000}, pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	want := map[string]bool{
		"GET /health": false, "GET /health/ready": false,
		"GET /api/v1/health/recommendation-provider": false,
		"POST /api/v1/auth/register":                 false, "POST /api/v1/auth/login": false,
		"GET /api/v1/users/me": false, "PATCH /api/v1/users/me": false,
		"GET /api/v1/users/me/preferences": false, "PATCH /api/v1/users/me/preferences": false,
		"POST /api/v1/recommendations/cities": false, "POST /api/v1/recommendations/destinations": false,
		"GET /api/v1/destinations": false, "GET /api/v1/destinations/:id": false,
		"POST /api/v1/destinations": false, "PUT /api/v1/destinations/:id": false, "DELETE /api/v1/destinations/:id": false,
		"GET /api/v1/restaurants": false, "GET /api/v1/restaurants/:id": false, "POST /api/v1/restaurants": false, "PUT /api/v1/restaurants/:id": false, "DELETE /api/v1/restaurants/:id": false,
		"GET /api/v1/mosques": false, "GET /api/v1/mosques/:id": false, "POST /api/v1/mosques": false, "PUT /api/v1/mosques/:id": false, "DELETE /api/v1/mosques/:id": false,
		"GET /api/v1/hotels": false, "GET /api/v1/hotels/:id": false, "POST /api/v1/hotels": false, "PUT /api/v1/hotels/:id": false, "DELETE /api/v1/hotels/:id": false,
		"GET /api/v1/saved-places": false, "POST /api/v1/saved-places": false, "DELETE /api/v1/saved-places/:place_type/:place_id": false,
		"POST /api/v1/trips": false, "GET /api/v1/trips": false, "GET /api/v1/trips/:tripId": false, "PATCH /api/v1/trips/:tripId": false, "DELETE /api/v1/trips/:tripId": false,
		"POST /api/v1/trips/:tripId/itinerary-items": false, "GET /api/v1/trips/:tripId/itinerary-items": false,
		"PUT /api/v1/trips/:tripId/itinerary-items/order": false, "PATCH /api/v1/trips/:tripId/itinerary-items/:itemId": false, "DELETE /api/v1/trips/:tripId/itinerary-items/:itemId": false,
	}
	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
