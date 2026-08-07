package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	// Set JWT_SECRET for tests
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-12345")
	os.Exit(m.Run())
}

func TestTokenService_Generate(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: 24 * time.Hour,
	}

	token, err := svc.Generate("user-123", "test@example.com", []string{"admin", "user"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestTokenService_Validate(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: 24 * time.Hour,
	}

	// Generate a token
	token, err := svc.Generate("user-123", "test@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected user ID 'user-123', got: %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got: %s", claims.Email)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got: %v", claims.Roles)
	}
}

func TestTokenService_ValidateInvalidToken(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: 24 * time.Hour,
	}

	// Validate invalid token
	_, err := svc.Validate("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestTokenService_ValidateWrongSecret(t *testing.T) {
	// Create two services with different secrets
	svc1 := &TokenService{
		secret: []byte("secret-1"),
		expiry: 24 * time.Hour,
	}
	svc2 := &TokenService{
		secret: []byte("secret-2"),
		expiry: 24 * time.Hour,
	}

	// Generate token with svc1
	token, err := svc1.Generate("user-123", "test@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate with svc2 (wrong secret)
	_, err = svc2.Validate(token)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestTokenService_GenerateRefresh(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: 24 * time.Hour,
	}

	token, err := svc.GenerateRefresh("user-123", "test@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Validate the refresh token
	claims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims.Subject != "refresh-token" {
		t.Errorf("expected subject 'refresh-token', got: %s", claims.Subject)
	}
}

func TestTokenService_Refresh(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: 24 * time.Hour,
	}

	// Generate a token
	originalToken, err := svc.Generate("user-123", "test@example.com", []string{"admin"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Refresh the token
	newToken, err := svc.Refresh(originalToken)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if newToken == "" {
		t.Error("expected non-empty new token")
	}
	if newToken == originalToken {
		t.Error("expected different token after refresh")
	}

	// Validate the new token
	claims, err := svc.Validate(newToken)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected user ID 'user-123', got: %s", claims.UserID)
	}
}

func TestTokenService_TokenExpiry(t *testing.T) {
	svc := &TokenService{
		secret: []byte("test-secret-key-for-testing-12345"),
		expiry: -1 * time.Hour, // Already expired
	}

	// Create a token with past expiry
	claims := Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		Roles:  []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   "access-token",
			ID:        "test-id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(svc.secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Try to validate expired token
	_, err = svc.Validate(tokenString)
	if err == nil {
		t.Error("expected error for expired token")
	}
}
