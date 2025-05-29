package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	otpLength = 6 // Keeping OTP length as a fixed const for now
	// otpValidityMinutes = 5 // Moved to config
)

// OTPStore defines the interface for OTP storage and validation
type OTPStore interface {
	StoreOTP(identifier string, otp string, validityDuration time.Duration)
	ValidateOTP(identifier string, otpToValidate string) bool
}

// OTPStoreEntry holds the OTP and its expiry time.
type OTPStoreEntry struct {
	OTP       string
	ExpiresAt time.Time
}

// InMemoryOTPStore is a simple in-memory store for OTPs.
// For production, a more robust solution like Redis would be used.
type InMemoryOTPStore struct {
	store map[string]OTPStoreEntry
	mu    sync.RWMutex
}

// Ensure InMemoryOTPStore implements OTPStore interface
var _ OTPStore = (*InMemoryOTPStore)(nil)

// NewInMemoryOTPStore now takes OTPValidityMinutes from outside (e.g. config)
// However, the cleanup goroutine is internal. For simplicity, we might keep its ticker fixed
// or make InMemoryOTPStore take the duration as a parameter for StoreOTP's Expiry calculation.
// Let's adjust StoreOTP to take validity duration.
func NewInMemoryOTPStore() *InMemoryOTPStore {
	s := &InMemoryOTPStore{
		store: make(map[string]OTPStoreEntry),
	}
	go s.cleanupExpiredOTPs() // Cleanup interval is still internal, every 1 minute
	return s
}

// GenerateOTP generates a random numeric OTP of a defined length.
func GenerateOTP() (string, error) {
	var otpChars = make([]byte, otpLength)
	max := big.NewInt(10) // Max value for each digit (0-9)
	for i := 0; i < otpLength; i++ {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit for OTP: %w", err)
		}
		otpChars[i] = byte(num.Int64() + '0') // Convert int64 to byte char '0'-'9'
	}
	return string(otpChars), nil
}

// StoreOTP stores an OTP for a given identifier (e.g., phone number) with a specific validity duration.
func (s *InMemoryOTPStore) StoreOTP(identifier string, otp string, validityDuration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[identifier] = OTPStoreEntry{
		OTP:       otp,
		ExpiresAt: time.Now().Add(validityDuration),
	}
}

// ValidateOTP checks if the provided OTP is valid for the identifier and not expired.
func (s *InMemoryOTPStore) ValidateOTP(identifier string, otpToValidate string) bool {
	s.mu.RLock()
	entry, exists := s.store[identifier]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(entry.ExpiresAt) {
		s.mu.Lock()
		delete(s.store, identifier)
		s.mu.Unlock()
		return false
	}

	valid := entry.OTP == otpToValidate
	if valid {
		s.mu.Lock()
		delete(s.store, identifier)
		s.mu.Unlock()
	}
	return valid
}

// cleanupExpiredOTPs periodically removes expired OTPs from the store.
func (s *InMemoryOTPStore) cleanupExpiredOTPs() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for id, entry := range s.store {
			if time.Now().After(entry.ExpiresAt) {
				delete(s.store, id)
			}
		}
		s.mu.Unlock()
	}
}
