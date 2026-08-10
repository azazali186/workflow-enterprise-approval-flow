package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyRefreshToken(t *testing.T) {
	tests := []struct {
		name      string
		current   string
		previous  string
		presented string
		want      refreshTokenStatus
	}{
		{
			name:      "current hash matches → valid rotation",
			current:   "hash-current",
			previous:  "hash-prev",
			presented: "hash-current",
			want:      refreshTokenValid,
		},
		{
			name:      "previous hash matches → benign race (second tab)",
			current:   "hash-current",
			previous:  "hash-prev",
			presented: "hash-prev",
			want:      refreshTokenRace,
		},
		{
			name:      "neither matches → replay/reuse, revoke",
			current:   "hash-current",
			previous:  "hash-prev",
			presented: "hash-stolen",
			want:      refreshTokenReused,
		},
		{
			name:      "empty current → session missing/revoked",
			current:   "",
			previous:  "",
			presented: "anything",
			want:      refreshTokenMissing,
		},
		{
			name:      "no previous stored + mismatch → reuse",
			current:   "hash-current",
			previous:  "",
			presented: "hash-old",
			want:      refreshTokenReused,
		},
		{
			name:      "empty presented is rejected",
			current:   "hash-current",
			previous:  "hash-prev",
			presented: "",
			want:      refreshTokenReused,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifyRefreshToken(tt.current, tt.previous, tt.presented))
		})
	}
}

func TestRefreshTokenStatusFromCode(t *testing.T) {
	tests := []struct {
		name string
		code int64
		want refreshTokenStatus
	}{
		{"valid rotation", 0, refreshTokenValid},
		{"benign race", 1, refreshTokenRace},
		{"reuse/forged", 2, refreshTokenReused},
		{"missing session", 3, refreshTokenMissing},
		{"unknown code defaults to reused", 99, refreshTokenReused},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, refreshTokenStatusFromCode(tt.code))
		})
	}
}

func TestRefreshTokenConstants(t *testing.T) {
	assert.Equal(t, "auth:refresh:", RefreshTokenCachePrefix)
	assert.Equal(t, "168h0m0s", RefreshTokenTTL.String(), "refresh token TTL must match the 7-day JWT lifetime")
}
