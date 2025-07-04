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
	UserIDKey ContextKey = "userID"
	// UserPhoneKey is the key used to store the user phone in the request context.
	UserPhoneKey ContextKey = "userPhone"
	// UserRoleKey is the key used to store the user role in the request context.
	UserRoleKey ContextKey = "userRole"
)

// AuthMiddleware creates a middleware handler for JWT authentication.
// Now supports both Authorization header and secure HTTP-only cookies for JWT tokens.
func AuthMiddleware(cfg *config.Config, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string
			var tokenSource string

			// Try to get token from Authorization header first (backward compatibility)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
					tokenSource = "header"
				}
			}

			// If no token in header, try to get from secure cookie
			if tokenString == "" {
				if cookie, err := r.Cookie("auth_token"); err == nil {
					tokenString = cookie.Value
					tokenSource = "cookie"
				}
			}

			// If still no token found, return unauthorized
			if tokenString == "" {
				RespondWithError(w, http.StatusUnauthorized, "Authentication required", logger)
				return
			}

			// Validate the JWT token
			claims, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
			if err != nil {
				// All JWT validation errors should be treated as 401 Unauthorized
				// This includes malformed tokens, expired tokens, wrong signatures, etc.
				logger.DebugContext(r.Context(), "JWT validation failed", "error", err, "token_source", tokenSource)
				RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token", logger)
				return
			}

			// Log successful authentication for monitoring
			logger.DebugContext(r.Context(), "JWT authentication successful", 
				"user_id", claims.UserID, 
				"token_source", tokenSource,
				"role", claims.Role,
			)

			// Store user ID, phone, and role in context for downstream handlers
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserPhoneKey, claims.Phone)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware creates a middleware handler that requires admin role.
// This middleware should be used after AuthMiddleware.
func AdminMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				RespondWithError(w, http.StatusUnauthorized, "User role not found in context", logger)
				return
			}

			if role != "admin" {
				RespondWithError(w, http.StatusForbidden, "Admin access required", logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
