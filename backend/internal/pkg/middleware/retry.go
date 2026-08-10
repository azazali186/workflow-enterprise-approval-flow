package middleware

import (
	"math"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
	}
}

// RetryWithBackoff executes a function with retry and exponential backoff
func RetryWithBackoff(config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt < config.MaxRetries {
			delay := calculateDelay(config, attempt)
			time.Sleep(delay)
		}
	}

	return lastErr
}

// calculateDelay calculates the delay for a given attempt
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	return time.Duration(delay)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if re, ok := err.(*RetryableError); ok {
		return re.Retryable
	}
	return false
}
