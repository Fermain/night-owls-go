package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"night-owls-go/internal/auth"
	"night-owls-go/internal/config" // For JWT secret

	"github.com/gorilla/sessions"
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
	// UserNameKey is the key used to store the user name in the request context.
	UserNameKey ContextKey = "userName"
)

// AuthMiddleware validates JWT tokens from headers or sessions
func AuthMiddleware(cfg *config.Config, logger *slog.Logger, sessionStore sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			var userID int64
			var phone, role, userName string
			
			// First try to get user info from session (preferred)
			if session, err := sessionStore.Get(r, "night-owls-session"); err == nil {
				if sessionUserID, ok := session.Values["user_id"].(int64); ok {
					userID = sessionUserID
				}
				if sessionPhone, ok := session.Values["phone"].(string); ok {
					phone = sessionPhone
				}
				if sessionRole, ok := session.Values["role"].(string); ok {
					role = sessionRole
				}
				if sessionName, ok := session.Values["name"].(string); ok {
					userName = sessionName
				}
				if sessionToken, ok := session.Values["token"].(string); ok {
					token = sessionToken
				}
			}
			
			// If no session info, fall back to JWT header (backward compatibility)
			if userID == 0 || phone == "" || role == "" {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Check if it's a Bearer token
				if !strings.HasPrefix(authHeader, "Bearer ") {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				token = strings.TrimPrefix(authHeader, "Bearer ")
				
				// Parse and validate JWT token
				claims, err := auth.ValidateJWT(token, cfg.JWTSecret)
				if err != nil {
					logger.WarnContext(r.Context(), "Invalid JWT token", "error", err.Error())
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Extract user information from JWT claims
				userID = claims.UserID
				phone = claims.Phone
				role = claims.Role
				userName = claims.Name
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserPhoneKey, phone)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			ctx = context.WithValue(ctx, UserNameKey, userName)

			// For debugging
			logger.DebugContext(r.Context(), "User authenticated", 
				"user_id", userID, 
				"phone", phone, 
				"role", role,
				"auth_method", map[bool]string{true: "session", false: "jwt_header"}[userID != 0])

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
