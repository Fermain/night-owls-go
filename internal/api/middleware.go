package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config" // For JWT secret
)

// ContextKey is a type used for context keys to avoid collisions.
type ContextKey string

const (
	// UserIDKey is the key used to store the user ID in the request context.
	UserIDKey    ContextKey = "userID"
	// UserPhoneKey is the key used to store the user phone in the request context.
	UserPhoneKey ContextKey = "userPhone"
)

// AuthMiddleware creates a middleware handler for JWT authentication.
func AuthMiddleware(cfg *config.Config, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondWithError(w, http.StatusUnauthorized, "Authorization header missing", logger)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				RespondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format (expected Bearer token)", logger)
				return
			}

			tokenString := parts[1]
			claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
			if err != nil {
				// All JWT validation errors should be treated as 401 Unauthorized
				// This includes malformed tokens, expired tokens, wrong signatures, etc.
				RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token", logger)
				return
			}

			// Store user ID and phone in context for downstream handlers
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserPhoneKey, claims.Phone)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
} 