package server

import (
	"net"
	"testing"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestParseTrustedCIDRs(t *testing.T) {
	t.Run("empty list trusts no proxy", func(t *testing.T) {
		cfg := &config.Config{}
		got := parseTrustedCIDRs(cfg)
		assert.Nil(t, got, "no configured proxies must yield nil so ClientIP() uses the peer address")
	})

	t.Run("single CIDR", func(t *testing.T) {
		cfg := &config.Config{TrustedProxies: []string{"10.0.0.0/8"}}
		got := parseTrustedCIDRs(cfg)
		requireLen(t, got, 1)
		_, expected, _ := net.ParseCIDR("10.0.0.0/8")
		assert.Equal(t, expected.String(), got[0].String())
	})

	t.Run("multiple CIDRs preserved in order", func(t *testing.T) {
		cfg := &config.Config{TrustedProxies: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}}
		got := parseTrustedCIDRs(cfg)
		requireLen(t, got, 3)
		for i, raw := range cfg.TrustedProxies {
			_, expected, _ := net.ParseCIDR(raw)
			assert.Equal(t, expected.String(), got[i].String(), "CIDR %d must parse correctly", i)
		}
	})
}

func requireLen(t *testing.T, got []*net.IPNet, want int) {
	t.Helper()
	if len(got) != want {
		t.Fatalf("expected %d CIDRs, got %d", want, len(got))
	}
}
