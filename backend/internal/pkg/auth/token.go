package auth

import (
	"fmt"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	pkguuid "github.com/aeroxe/approval-flow/internal/pkg/uuid"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenService handles JWT token operations
type TokenService struct {
	secret []byte
	expiry time.Duration
	logger *config.Config
}

// NewTokenService creates a new token service
func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		secret: []byte(cfg.JWTSecret),
		expiry: cfg.JWTExpiryDuration(),
		logger: cfg,
	}
}

// Expiry returns the configured access token lifetime.
func (t *TokenService) Expiry() time.Duration {
	return t.expiry
}

// Generate generates a new access token
func (t *TokenService) Generate(userID, email string, roles []string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "access-token",
			ID:        pkguuid.NewV7().String(), // UUID v7 for token ID
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

// Validate validates a token and returns claims
func (t *TokenService) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// Refresh refreshes an expired token
func (t *TokenService) Refresh(tokenString string) (string, error) {
	claims, err := t.Validate(tokenString)
	if err != nil {
		return "", err
	}
	return t.Generate(claims.UserID, claims.Email, claims.Roles)
}

// GenerateRefresh generates a refresh token with longer expiry
func (t *TokenService) GenerateRefresh(userID, email string, roles []string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   "refresh-token",
			ID:        pkguuid.NewV7().String(), // UUID v7 for token ID
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.secret)
}

// ExtractTokenID extracts the token ID from a token string without validation
// WARNING: This does not validate the token signature - use only for non-security purposes
func (t *TokenService) ExtractTokenID(tokenString string) string {
	// Parse without validation to extract claims
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return ""
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims.ID
	}
	return ""
}
