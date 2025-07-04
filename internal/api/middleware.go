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

// SecurityHeadersMiddleware adds comprehensive security headers to all responses
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy - comprehensive policy for Svelte app
			csp := strings.Join([]string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline'", // Svelte needs unsafe-inline for now
				"style-src 'self' 'unsafe-inline'",  // Svelte/Tailwind needs unsafe-inline
				"img-src 'self' data: https:",       // Allow images from self, data URLs, and HTTPS
				"font-src 'self' https:",            // Allow fonts from self and HTTPS
				"connect-src 'self'",                // API calls only to same origin
				"frame-ancestors 'none'",            // Prevent embedding in frames
				"base-uri 'self'",                   // Restrict base tag to same origin
				"form-action 'self'",                // Forms only to same origin
				"object-src 'none'",                 // Disable plugins
				"upgrade-insecure-requests",         // Upgrade HTTP to HTTPS
			}, "; ")
			w.Header().Set("Content-Security-Policy", csp)
			
			// Prevent clickjacking attacks
			w.Header().Set("X-Frame-Options", "DENY")
			
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")
			
			// Enable XSS protection in browsers
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			
			// Strict Transport Security (HSTS) - 1 year with subdomains
			// Only set in production or when HTTPS is detected
			if isHTTPS(r) {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}
			
			// Referrer Policy - only send referrer to same origin
			w.Header().Set("Referrer-Policy", "same-origin")
			
			// Prevent Adobe Flash and PDF from loading
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
			
			// Additional security headers
			w.Header().Set("X-Download-Options", "noopen")           // IE download protection
			w.Header().Set("X-DNS-Prefetch-Control", "off")         // Disable DNS prefetching
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()") // Disable sensitive permissions
			
			next.ServeHTTP(w, r)
		})
	}
}

// isHTTPS checks if the request was made over HTTPS
func isHTTPS(r *http.Request) bool {
	// Check the request scheme
	if r.TLS != nil {
		return true
	}
	
	// Check X-Forwarded-Proto header (for proxies/load balancers)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}
	
	// Check X-Forwarded-SSL header
	if ssl := r.Header.Get("X-Forwarded-SSL"); ssl == "on" {
		return true
	}
	
	return false
}
