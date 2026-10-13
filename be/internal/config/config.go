package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                     string
	AppPort                    string
	AppBaseURL                 string
	DatabaseURL                string
	CORSAllowedOrigins         []string
	HTTPClientTimeout          time.Duration
	TravelRecommendationAPIURL string
	JWTSecret                  string
	JWTAccessTokenTTL          time.Duration
	JWTIssuer                  string
	JWTAudience                string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                     env("APP_ENV", "development"),
		AppPort:                    env("APP_PORT", "8080"),
		AppBaseURL:                 env("APP_BASE_URL", "http://localhost:8080"),
		DatabaseURL:                strings.TrimSpace(os.Getenv("DATABASE_URL")),
		HTTPClientTimeout:          15 * time.Second,
		TravelRecommendationAPIURL: strings.TrimRight(env("TRAVEL_RECOMMENDATION_API_URL", "https://travel-recommendation-api-production-75f5.up.railway.app"), "/"),
		JWTSecret:                  strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTIssuer:                  strings.TrimSpace(os.Getenv("JWT_ISSUER")),
		JWTAudience:                strings.TrimSpace(os.Getenv("JWT_AUDIENCE")),
		JWTAccessTokenTTL:          time.Hour,
	}
	port, err := strconv.Atoi(cfg.AppPort)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("APP_PORT must be an integer between 1 and 65535")
	}
	baseURL, err := url.ParseRequestURI(cfg.AppBaseURL)
	if err != nil || baseURL.Host == "" || (baseURL.Scheme != "http" && baseURL.Scheme != "https") {
		return Config{}, fmt.Errorf("APP_BASE_URL must be a valid http or https URL")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if raw := strings.TrimSpace(os.Getenv("JWT_ACCESS_TOKEN_TTL")); raw != "" {
		cfg.JWTAccessTokenTTL, err = time.ParseDuration(raw)
		if err != nil || cfg.JWTAccessTokenTTL < time.Second {
			return Config{}, fmt.Errorf("JWT_ACCESS_TOKEN_TTL must be at least 1s")
		}
	}
	databaseURL, err := url.Parse(cfg.DatabaseURL)
	if err != nil || (databaseURL.Scheme != "postgres" && databaseURL.Scheme != "postgresql") || databaseURL.Host == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must be a valid PostgreSQL URL")
	}
	if raw := strings.TrimSpace(os.Getenv("HTTP_CLIENT_TIMEOUT")); raw != "" {
		cfg.HTTPClientTimeout, err = time.ParseDuration(raw)
		if err != nil || cfg.HTTPClientTimeout <= 0 {
			return Config{}, fmt.Errorf("HTTP_CLIENT_TIMEOUT must be a positive duration")
		}
	}
	providerURL, err := url.ParseRequestURI(cfg.TravelRecommendationAPIURL)
	if err != nil || providerURL.Host == "" || (providerURL.Scheme != "http" && providerURL.Scheme != "https") {
		return Config{}, fmt.Errorf("TRAVEL_RECOMMENDATION_API_URL must be a valid http or https URL")
	}
	cfg.CORSAllowedOrigins = split(env("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))
	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
