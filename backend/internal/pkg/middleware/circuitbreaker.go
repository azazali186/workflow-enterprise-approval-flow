package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aeroxe/approval-flow/internal/config"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"go.uber.org/zap"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Failing, reject requests
	CircuitHalfOpen                     // Testing if service recovered
)

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
	logger           *config.Config
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration, cfg *config.Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		logger:           cfg,
	}
}

// CanExecute checks if a request can be executed
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failureCount = 0
	case CircuitHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.failureCount = 0
			if cb.logger != nil {
				cb.logger.Info("circuit breaker closed (service recovered)")
			}
		}
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.lastFailureTime = time.Now()
			if cb.logger != nil {
				cb.logger.Warn("circuit breaker opened (service failing)")
			}
		}
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		cb.lastFailureTime = time.Now()
		if cb.logger != nil {
			cb.logger.Warn("circuit breaker re-opened (recovery failed)")
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// Global circuit breakers for external services
var (
	DBCircuitBreaker    *CircuitBreaker
	RedisCircuitBreaker *CircuitBreaker
	NATSCircuitBreaker  *CircuitBreaker
)

// InitCircuitBreakers initializes global circuit breakers for external dependencies
func InitCircuitBreakers(cfg *config.Config) {
	timeout := 30 * time.Second
	DBCircuitBreaker = NewCircuitBreaker(5, 3, timeout, cfg)
	RedisCircuitBreaker = NewCircuitBreaker(3, 2, timeout, cfg)
	NATSCircuitBreaker = NewCircuitBreaker(3, 2, timeout, cfg)
}

const (
	// circuitFailureThreshold is how many consecutive 5xx responses open a route's breaker.
	circuitFailureThreshold = 10
	// circuitSuccessThreshold is how many consecutive successes re-close a half-open breaker.
	circuitSuccessThreshold = 3
	// circuitOpenTimeout is how long a breaker stays open before probing again.
	circuitOpenTimeout = 30 * time.Second
)

// CircuitBreakerMiddleware wraps every request in a per-route circuit breaker.
// Routes that repeatedly return 5xx responses are short-circuited with a 503
// until they recover, preventing cascading failures from flaky dependencies.
func CircuitBreakerMiddleware(cfg *config.Config) app.HandlerFunc {
	var breakers sync.Map // routeKey -> *CircuitBreaker

	breakerFor := func(routeKey string) *CircuitBreaker {
		if cb, ok := breakers.Load(routeKey); ok {
			return cb.(*CircuitBreaker)
		}
		cb := NewCircuitBreaker(circuitFailureThreshold, circuitSuccessThreshold, circuitOpenTimeout, cfg)
		actual, _ := breakers.LoadOrStore(routeKey, cb)
		return actual.(*CircuitBreaker)
	}

	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())
		// Never short-circuit infrastructure endpoints — probes and scrapers
		// must always reach the real handlers so their signals stay accurate.
		if isCircuitExcludedPath(path) {
			c.Next(ctx)
			return
		}

		routeKey := fmt.Sprintf("%s %s", c.Request.Method(), path)
		cb := breakerFor(routeKey)

		if !cb.CanExecute() {
			cfg.Warn("circuit breaker open; short-circuiting request",
				zap.String("route", routeKey),
			)
			c.JSON(consts.StatusServiceUnavailable, map[string]string{
				"error": "service temporarily unavailable",
			})
			c.Abort()
			return
		}

		c.Next(ctx)

		if c.Response.StatusCode() >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}

// isCircuitExcludedPath reports whether a path should bypass the circuit breaker.
func isCircuitExcludedPath(path string) bool {
	for _, excluded := range []string{"/health", "/metrics", "/docs", "/version", "/ws"} {
		if strings.HasPrefix(path, excluded) {
			return true
		}
	}
	return false
}
