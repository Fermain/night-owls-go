package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ( // Define common errors for JWT validation.
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenMissing = errors.New("authorization token missing")
)

// Claims defines the JWT claims.
type Claims struct {
	UserID int64  `json:"user_id"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT for a given user ID, phone number, and role.
func GenerateJWT(userID int64, phone string, role string, jwtSecret string, expirationHours int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	claims := &Claims{
		UserID: userID,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "night-owls-go", // Optional: an identifier for the issuer
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateJWT validates a JWT string and returns the claims if valid.
func ValidateJWT(tokenString string, jwtSecret string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrTokenMissing
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) || errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("could not parse token: %w", err) // More generic parsing error
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
