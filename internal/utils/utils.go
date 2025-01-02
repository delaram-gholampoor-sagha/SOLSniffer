package utils

import "time"

func DefaultRetryConfig(maxAttempts int, delay time.Duration, backoff func(attempt int) time.Duration) *RetryConfig {
	return &RetryConfig{
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Backoff:     backoff,
	}
}

func DatabaseRetryConfig() *RetryConfig {
	return DefaultRetryConfig(3, 2*time.Second, func(attempt int) time.Duration {
		return time.Duration(attempt) * 2 * time.Second // Exponential backoff
	})
}

func WebSocketRetryConfig() *RetryConfig {
	return DefaultRetryConfig(10, 1*time.Second, func(attempt int) time.Duration {
		return time.Duration(attempt) * 500 * time.Millisecond // Linear backoff
	})
}
