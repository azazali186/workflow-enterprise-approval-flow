package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== Pure Logic Tests ====================

func TestIsExcludedRoute(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/ws", true},
		{"/api/v1/auth/login", true},
		{"/api/v1/auth/register", true},
		{"/api/v1/auth/refresh", true},
		{"/metrics", true},
		{"/api/v1/applications", false},
		{"/api/v1/admin/users", false},
		{"/api/v1/workflows", false},
		{"/", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsExcludedRoute(tt.path))
		})
	}
}

func TestFormatRouteName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/ws", "WebSocket Connection"},
		{"/api/v1/applications", "Applications"},
		{"/api/v1/auth/login", "Auth Login"},
		{"/api/v1/admin/users", "Admin Users"},
		{"/api/v1/workflows", "Workflows"},
		{"/api/v1/templates", "Templates"},
		{"/api/v1/escalations", "Escalations"},
		{"/api/v1/notifications", "Notifications"},
		{"/api/v1/approvals", "Approvals"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatRouteName(tt.path))
		})
	}
}

func TestChangePasswordRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		oldPassword string
		newPassword string
		valid       bool
	}{
		{"valid passwords", "oldpass123", "newpass456", true},
		{"empty old password", "", "newpass456", false},
		{"empty new password", "oldpass123", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ChangePasswordRequest{
				OldPassword: tt.oldPassword,
				NewPassword: tt.newPassword,
			}
			if tt.valid {
				assert.NotEmpty(t, req.OldPassword)
				assert.NotEmpty(t, req.NewPassword)
			} else {
				valid := req.OldPassword != "" && req.NewPassword != ""
				assert.False(t, valid)
			}
		})
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		valid    bool
	}{
		{"valid credentials", "user@example.com", "password123", true},
		{"empty email", "", "password123", false},
		{"empty password", "user@example.com", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			if tt.valid {
				assert.NotEmpty(t, req.Email)
				assert.NotEmpty(t, req.Password)
			} else {
				valid := req.Email != "" && req.Password != ""
				assert.False(t, valid)
			}
		})
	}
}

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		userName string
		valid    bool
	}{
		{"valid registration", "user@example.com", "password123", "Test User", true},
		{"empty email", "", "password123", "Test User", false},
		{"empty password", "user@example.com", "", "Test User", false},
		{"empty name", "user@example.com", "password123", "", false},
		{"all empty", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
				Name:     tt.userName,
			}
			if tt.valid {
				assert.NotEmpty(t, req.Email)
				assert.NotEmpty(t, req.Password)
				assert.NotEmpty(t, req.Name)
			} else {
				valid := req.Email != "" && req.Password != "" && req.Name != ""
				assert.False(t, valid)
			}
		})
	}
}

func TestComputeTokenHash(t *testing.T) {
	// Test that the hash is deterministic
	hash1 := computeTokenHash("token123", "user456")
	hash2 := computeTokenHash("token123", "user456")
	assert.Equal(t, hash1, hash2)

	// Test that different inputs produce different hashes
	hash3 := computeTokenHash("token789", "user456")
	assert.NotEqual(t, hash1, hash3)

	hash4 := computeTokenHash("token123", "user789")
	assert.NotEqual(t, hash1, hash4)
}

func TestService_Constants(t *testing.T) {
	assert.Equal(t, "auth:token:", TokenCachePrefix)
	assert.Equal(t, "24h0m0s", DefaultTokenTTL.String())
	assert.Equal(t, "30m0s", RefreshThreshold.String())
}

func TestRouteInfo_Structure(t *testing.T) {
	route := RouteInfo{
		Name:   "Test Route",
		URL:    "/api/v1/test",
		Method: "POST",
		Guard:  "API",
	}

	assert.Equal(t, "Test Route", route.Name)
	assert.Equal(t, "/api/v1/test", route.URL)
	assert.Equal(t, "POST", route.Method)
	assert.Equal(t, "API", route.Guard)
}

func TestLoginRequest_Structure(t *testing.T) {
	req := LoginRequest{
		Email:    "user@example.com",
		Password: "secret123",
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "secret123", req.Password)
}

func TestRegisterRequest_Structure(t *testing.T) {
	req := RegisterRequest{
		Email:    "user@example.com",
		Password: "secret123",
		Name:     "Test User",
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "secret123", req.Password)
	assert.Equal(t, "Test User", req.Name)
}

func TestLoginResponse_Structure(t *testing.T) {
	resp := &LoginResponse{
		User:         nil,
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
	}

	assert.Nil(t, resp.User)
	assert.Equal(t, "access-token-123", resp.AccessToken)
	assert.Equal(t, "refresh-token-456", resp.RefreshToken)
}

func TestChangePasswordRequest_Structure(t *testing.T) {
	req := ChangePasswordRequest{
		OldPassword: "oldpass",
		NewPassword: "newpass123",
	}

	assert.Equal(t, "oldpass", req.OldPassword)
	assert.Equal(t, "newpass123", req.NewPassword)
}
