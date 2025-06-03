package api

import (
	"context"
	"net/http"
	"strings"
)

// Context keys for audit information
type contextKey string

const (
	AuditIPAddressKey contextKey = "audit_ip_address"
	AuditUserAgentKey contextKey = "audit_user_agent"
)

// AuditContextMiddleware captures IP address and User-Agent for audit logging
func AuditContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP address with support for proxies
		ipAddress := getClientIPAddress(r)
		
		// Extract User-Agent
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "Unknown"
		}

		// Add to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, AuditIPAddressKey, ipAddress)
		ctx = context.WithValue(ctx, AuditUserAgentKey, userAgent)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClientIPAddress extracts the real client IP address, handling proxies
func getClientIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for load balancers/proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (nginx proxy)
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Check X-Forwarded header
	xForwarded := r.Header.Get("X-Forwarded")
	if xForwarded != "" {
		return xForwarded
	}

	// Check Forwarded header (RFC 7239)
	forwarded := r.Header.Get("Forwarded")
	if forwarded != "" {
		// Parse "for=" part of Forwarded header
		if strings.Contains(forwarded, "for=") {
			parts := strings.Split(forwarded, "for=")
			if len(parts) > 1 {
				ip := strings.Split(parts[1], ";")[0]
				ip = strings.Trim(ip, `"`)
				return ip
			}
		}
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// GetAuditInfoFromContext extracts audit information from request context
func GetAuditInfoFromContext(ctx context.Context) (ipAddress, userAgent string) {
	if ip, ok := ctx.Value(AuditIPAddressKey).(string); ok {
		ipAddress = ip
	}
	if ua, ok := ctx.Value(AuditUserAgentKey).(string); ok {
		userAgent = ua
	}
	return ipAddress, userAgent
} 