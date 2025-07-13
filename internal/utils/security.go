package utils

import (
	"crypto/rand"
	"math/big"
	"time"
)

// SecurityConfig holds configuration for security-related utilities
type SecurityConfig struct {
	MinDelayMs int // Minimum delay in milliseconds
	MaxDelayMs int // Maximum delay in milliseconds
}

// DefaultSecurityConfig provides sensible defaults for security utilities
var DefaultSecurityConfig = SecurityConfig{
	MinDelayMs: 50,
	MaxDelayMs: 150,
}

// AddTimingRandomization adds a small random delay to prevent timing attacks.
// This helps normalize response times across different authentication scenarios
// to prevent attackers from using timing differences to enumerate valid accounts.
//
// The delay is configurable but defaults to 50-150ms which provides good security
// without significantly impacting user experience.
func AddTimingRandomization() {
	AddTimingRandomizationWithConfig(DefaultSecurityConfig)
}

// AddTimingRandomizationWithConfig adds timing randomization with custom configuration
func AddTimingRandomizationWithConfig(config SecurityConfig) {
	delayRange := config.MaxDelayMs - config.MinDelayMs
	if delayRange <= 0 {
		// Invalid config, use fixed delay
		time.Sleep(time.Duration(config.MinDelayMs) * time.Millisecond)
		return
	}

	// Generate random number in range [0, delayRange)
	randomOffset, err := rand.Int(rand.Reader, big.NewInt(int64(delayRange)))
	if err != nil {
		// Fallback to middle of range if randomization fails
		midDelay := config.MinDelayMs + delayRange/2
		time.Sleep(time.Duration(midDelay) * time.Millisecond)
		return
	}

	delay := time.Duration(config.MinDelayMs+int(randomOffset.Int64())) * time.Millisecond
	time.Sleep(delay)
}
