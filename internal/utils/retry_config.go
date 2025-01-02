package utils

import (
	"context"
	"errors"
	"time"
)

// RetryConfig holds configuration for the retry mechanism.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     func(attempt int) time.Duration // Optional backoff strategy
}

// RetryOption is a function that modifies the RetryConfig.
type RetryOption func(*RetryConfig)

// WithMaxAttempts sets the maximum number of retry attempts.
func WithMaxAttempts(attempts int) RetryOption {
	return func(cfg *RetryConfig) {
		cfg.MaxAttempts = attempts
	}
}

// WithDelay sets the delay between retry attempts.
func WithDelay(delay time.Duration) RetryOption {
	return func(cfg *RetryConfig) {
		cfg.Delay = delay
	}
}

// WithBackoff sets a custom backoff strategy.
func WithBackoff(backoff func(attempt int) time.Duration) RetryOption {
	return func(cfg *RetryConfig) {
		cfg.Backoff = backoff
	}
}

// Retry retries the given function according to the provided options.
func Retry(ctx context.Context, operation func() error, options ...RetryOption) error {
	// Default configuration
	cfg := &RetryConfig{
		MaxAttempts: 3,
		Delay:       time.Second,
		Backoff:     nil, // No backoff by default
	}

	// Apply options
	for _, opt := range options {
		opt(cfg)
	}

	var err error
	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		if err = operation(); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cfg.getDelay(attempt)):
			// Wait for the configured delay
		}
	}

	return errors.New("operation failed after retries: " + err.Error())
}

// getDelay calculates the delay for a given attempt using the backoff strategy.
func (cfg *RetryConfig) getDelay(attempt int) time.Duration {
	if cfg.Backoff != nil {
		return cfg.Backoff(attempt)
	}
	return cfg.Delay
}
