package server

import (
	"net"

	"github.com/aeroxe/approval-flow/internal/config"
	"go.uber.org/zap"
)

// parseTrustedCIDRs converts the configured TRUSTED_PROXIES list into CIDR
// nets for the Hertz client-IP resolver.
//
// An empty list returns nil, which makes Hertz trust *no* proxy: ClientIP()
// then always returns the socket peer address, so a remote client cannot
// spoof X-Forwarded-For / X-Real-IP to bypass the per-IP rate limiter, the
// login IP lockout, or poison audit/login logs.
//
// In production behind a reverse proxy or ingress controller, set
// TRUSTED_PROXIES to the proxy's CIDRs (e.g. "10.0.0.0/8,172.16.0.0/12") so
// the real client IP from the forwarded headers is used.
func parseTrustedCIDRs(cfg *config.Config) []*net.IPNet {
	if len(cfg.TrustedProxies) == 0 {
		return nil
	}

	cidrs := make([]*net.IPNet, 0, len(cfg.TrustedProxies))
	for _, raw := range cfg.TrustedProxies {
		_, ipnet, err := net.ParseCIDR(raw)
		if err != nil {
			cfg.Fatal("invalid TRUSTED_PROXIES CIDR",
				zap.String("cidr", raw),
				zap.Error(err),
			)
		}
		cidrs = append(cidrs, ipnet)
	}
	return cidrs
}
