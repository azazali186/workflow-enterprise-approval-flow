package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLooksLikePlaceholder(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"CHANGE_ME_with_a_random_secret", true},
		{"your-secret-key-change-in-production", true},
		{"change-me", true},
		{"cHANGE_ME_uppercase_variant", true},
		{"a-real-9f2c4e1b7a-secret-that-is-long", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Equal(t, tt.expected, looksLikePlaceholder(tt.value))
		})
	}
}

func TestJWTExpiryDuration(t *testing.T) {
	tests := []struct {
		name     string
		expiry   string
		expected time.Duration
	}{
		{"default when unset", "", 24 * time.Hour},
		{"default on invalid input", "not-a-duration", 24 * time.Hour},
		{"default on zero", "0s", 24 * time.Hour},
		{"parsed value", "1h", time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{JWTExpiry: tt.expiry}
			assert.Equal(t, tt.expected, cfg.JWTExpiryDuration())
		})
	}
}
