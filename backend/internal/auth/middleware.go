package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	value := ctx.Value(userIDKey)
	id, ok := value.(string)
	return id, ok
}

type Middleware struct {
	cache  *JWKSCache
	issuer string
}

func NewMiddleware(cache *JWKSCache, issuer string) *Middleware {
	return &Middleware{cache: cache, issuer: issuer}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerToken(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			kid, _ := t.Header["kid"].(string)
			if kid == "" {
				return nil, errors.New("missing kid")
			}
			return m.cache.GetKey(kid)
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithLeeway(30*time.Second))
		if err != nil || !parsed.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := parsed.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		if m.issuer != "" {
			if iss, _ := claims["iss"].(string); iss != m.issuer {
				http.Error(w, "invalid issuer", http.StatusUnauthorized)
				return
			}
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			http.Error(w, "missing subject", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid header")
	}
	return strings.TrimSpace(parts[1]), nil
}
