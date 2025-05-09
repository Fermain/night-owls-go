package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

const (
	otpLength        = 6
	otpValidityMinutes = 5 // OTPs are valid for 5 minutes
)

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

// NewInMemoryOTPStore creates a new in-memory OTP store and starts a cleanup goroutine.
func NewInMemoryOTPStore() *InMemoryOTPStore {
	s := &InMemoryOTPStore{
		store: make(map[string]OTPStoreEntry),
	}
	go s.cleanupExpiredOTPs()
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

// StoreOTP stores an OTP for a given identifier (e.g., phone number).
func (s *InMemoryOTPStore) StoreOTP(identifier string, otp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[identifier] = OTPStoreEntry{
		OTP:       otp,
		ExpiresAt: time.Now().Add(otpValidityMinutes * time.Minute),
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
		// OTP has expired, remove it proactively (though cleanup goroutine also handles this)
		s.mu.Lock()
		delete(s.store, identifier)
		s.mu.Unlock()
		return false
	}

	valid := entry.OTP == otpToValidate
	if valid {
		// OTP is valid and used, remove it to prevent reuse
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
 