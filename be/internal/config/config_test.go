package config

import "testing"

func TestLoadDefaultsAndRequiredDatabase(t *testing.T) {
	t.Setenv("APP_PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/halaltrip?sslmode=disable")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	for _, key := range []string{"APP_ENV", "APP_BASE_URL", "CORS_ALLOWED_ORIGINS", "HTTP_CLIENT_TIMEOUT", "TRAVEL_RECOMMENDATION_API_URL", "JWT_ACCESS_TOKEN_TTL", "JWT_ISSUER", "JWT_AUDIENCE"} {
		t.Setenv(key, "")
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppEnv != "development" || cfg.AppPort != "8080" || cfg.HTTPClientTimeout.String() != "15s" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.JWTAccessTokenTTL.String() != "1h0m0s" {
		t.Fatalf("unexpected JWT token lifetime: %s", cfg.JWTAccessTokenTTL)
	}
	if cfg.TravelRecommendationAPIURL != "https://travel-recommendation-api-production-75f5.up.railway.app" {
		t.Fatalf("unexpected provider URL: %s", cfg.TravelRecommendationAPIURL)
	}
}

func TestLoadRejectsWeakJWTSecretAndInvalidTokenLifetime(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/halaltrip")
	t.Setenv("JWT_SECRET", "too-short")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a weak JWT_SECRET")
	}
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "500ms")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a token lifetime below one second")
	}
}

func TestLoadUsesTravelRecommendationAPIURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/halaltrip")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("TRAVEL_RECOMMENDATION_API_URL", "http://localhost:9000/")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TravelRecommendationAPIURL != "http://localhost:9000" {
		t.Fatalf("URL = %q", cfg.TravelRecommendationAPIURL)
	}
}

func TestLoadRejectsInvalidTravelRecommendationAPIURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/halaltrip")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("TRAVEL_RECOMMENDATION_API_URL", "file:///tmp/provider")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject a non-http provider URL")
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an empty DATABASE_URL")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("APP_PORT", "70000")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/halaltrip")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject an invalid APP_PORT")
	}
}
