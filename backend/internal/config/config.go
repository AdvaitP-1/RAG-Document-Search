package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	DatabaseURL      string
	SupabaseURL      string
	SupabaseJWKSURL  string
	SupabaseIssuer   string
	AllowedOrigins   []string
}

func Load() Config {
	port := getEnv("PORT", "8080")
	supabaseURL := os.Getenv("SUPABASE_URL")
	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	if jwksURL == "" && supabaseURL != "" {
		jwksURL = strings.TrimRight(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	}

	return Config{
		Port:            port,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		SupabaseURL:     supabaseURL,
		SupabaseJWKSURL: jwksURL,
		SupabaseIssuer:  os.Getenv("SUPABASE_JWT_ISSUER"),
		AllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
