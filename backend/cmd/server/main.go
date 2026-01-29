package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"rag-backend/internal/auth"
	"rag-backend/internal/config"
	"rag-backend/internal/handlers"
	"rag-backend/internal/store"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.SupabaseJWKSURL == "" {
		log.Fatal("SUPABASE_JWKS_URL or SUPABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	store := store.New(pool)
	api := handlers.New(store)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware(cfg.AllowedOrigins))

	api.RegisterPublic(router)

	jwksCache := auth.NewJWKSCache(cfg.SupabaseJWKSURL, 10*time.Minute)
	authMiddleware := auth.NewMiddleware(jwksCache, cfg.SupabaseIssuer)
	api.RegisterProtected(router, authMiddleware)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowed["*"]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func init() {
	if os.Getenv("CORS_ALLOWED_ORIGINS") == "" {
		os.Setenv("CORS_ALLOWED_ORIGINS", strings.Join([]string{
			"http://localhost:3000",
		}, ","))
	}
}
