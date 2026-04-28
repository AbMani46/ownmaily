package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AbMani46/ownmaily/internal/auth"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type contextKey string

const (
	ClaimsKey    contextKey = "claims"
	APIKeyAuthed contextKey = "apikeyAuthed"
)

func RequireJWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearer(r)
			if tokenStr == "" {
				if c, err := r.Cookie("jwt"); err == nil {
					tokenStr = c.Value
				}
			}
			if tokenStr == "" {
				writeUnauthorized(w, "missing token")
				return
			}
			claims, err := auth.ValidateToken(tokenStr, secret)
			if err != nil {
				writeUnauthorized(w, "invalid token")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAPIKey(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractBearer(r)
			if key == "" || !strings.HasPrefix(key, "om_") {
				writeUnauthorized(w, "invalid api key")
				return
			}
			apiKey, err := queries.GetLatestAPIKey(r.Context())
			if err != nil {
				writeUnauthorized(w, "unauthorized")
				return
			}
			if !auth.CheckKey(key, apiKey.KeyHash) {
				writeUnauthorized(w, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), APIKeyAuthed, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(secret string, queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := extractBearer(r)
			if strings.HasPrefix(bearer, "om_") {
				RequireAPIKey(queries)(next).ServeHTTP(w, r)
				return
			}
			RequireJWT(secret)(next).ServeHTTP(w, r)
		})
	}
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","message":"` + message + `"}`))
}
