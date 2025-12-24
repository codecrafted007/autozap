package retry

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

// ExecuteWithRetry executes a function with retry logic based on the retry configuration
func ExecuteWithRetry(
	actionName string,
	retryConfig *workflow.RetryConfig,
	fn func() error,
) error {
	// If no retry config, execute once
	if retryConfig == nil || retryConfig.MaxAttempts <= 0 {
		return fn()
	}

	// Set defaults
	maxAttempts := retryConfig.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	initialDelay := parseDuration(retryConfig.InitialDelay, 1*time.Second)
	maxDelay := parseDuration(retryConfig.MaxDelay, 60*time.Second)
	multiplier := retryConfig.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			// Success
			if attempt > 1 {
				logger.L().Infow("Action succeeded after retry",
					"action_name", actionName,
					"attempt", attempt,
					"total_attempts", maxAttempts,
				)
			}
			return nil
		}

		lastErr = err

		// Check if we should retry
		if attempt >= maxAttempts {
			logger.L().Errorw("Action failed after all retry attempts",
				"action_name", actionName,
				"total_attempts", maxAttempts,
				"error", err,
			)
			break
		}

		// Check if error is retryable
		if !shouldRetry(err, retryConfig.RetryOn) {
			logger.L().Warnw("Action failed with non-retryable error",
				"action_name", actionName,
				"attempt", attempt,
				"error", err,
			)
			return err
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt-1, initialDelay, maxDelay, multiplier)

		logger.L().Infow("Action failed, retrying...",
			"action_name", actionName,
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"next_retry_in", delay,
			"error", err,
		)

		// Wait before retrying
		time.Sleep(delay)
	}

	return lastErr
}

// calculateDelay calculates the delay for exponential backoff with jitter
func calculateDelay(attempt int, initialDelay, maxDelay time.Duration, multiplier float64) time.Duration {
	// Exponential backoff: delay = initialDelay * (multiplier ^ attempt)
	delay := float64(initialDelay) * math.Pow(multiplier, float64(attempt))

	// Cap at maxDelay
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	// Add jitter (±10%) to avoid thundering herd
	jitter := delay * 0.1 * (rand.Float64()*2 - 1)
	finalDelay := time.Duration(delay + jitter)

	// Ensure delay is at least initialDelay
	if finalDelay < initialDelay {
		finalDelay = initialDelay
	}

	return finalDelay
}

// shouldRetry determines if an error should trigger a retry
func shouldRetry(err error, retryOn []string) bool {
	if err == nil {
		return false
	}

	// If no retry conditions specified, retry on all errors
	if len(retryOn) == 0 {
		return true
	}

	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	for _, condition := range retryOn {
		conditionLower := strings.ToLower(condition)

		// Check for specific conditions
		switch {
		case conditionLower == "timeout":
			if strings.Contains(errMsgLower, "timeout") ||
				strings.Contains(errMsgLower, "deadline exceeded") {
				return true
			}

		case conditionLower == "error":
			// Retry on any error
			return true

		case strings.HasPrefix(conditionLower, "status:"):
			// HTTP status code check (e.g., "status:500")
			statusCode := strings.TrimPrefix(conditionLower, "status:")
			if strings.Contains(errMsgLower, "status code "+statusCode) ||
				strings.Contains(errMsgLower, "status "+statusCode) {
				return true
			}

		case conditionLower == "network":
			if strings.Contains(errMsgLower, "network") ||
				strings.Contains(errMsgLower, "connection") ||
				strings.Contains(errMsgLower, "dns") {
				return true
			}

		default:
			// Check if error message contains the condition
			if strings.Contains(errMsgLower, conditionLower) {
				return true
			}
		}
	}

	return false
}

// parseDuration parses a duration string, returning defaultValue if parsing fails
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if s == "" {
		return defaultValue
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		logger.L().Warnw("Failed to parse duration, using default",
			"duration_string", s,
			"default", defaultValue,
			"error", err,
		)
		return defaultValue
	}

	return d
}

// IsRetryableHTTPStatus checks if an HTTP status code should trigger a retry
func IsRetryableHTTPStatus(statusCode int) bool {
	// Retry on server errors (5xx) and some client errors
	switch statusCode {
	case 408, // Request Timeout
		429, // Too Many Requests
		500, // Internal Server Error
		502, // Bad Gateway
		503, // Service Unavailable
		504: // Gateway Timeout
		return true
	default:
		return false
	}
}

// CreateRetryableError wraps an error with retry information
func CreateRetryableError(err error, retryable bool) error {
	return &RetryableError{
		Err:       err,
		Retryable: retryable,
	}
}

// WrapHTTPError wraps an HTTP error with retry information based on status code
func WrapHTTPError(err error, statusCode int) error {
	return &RetryableError{
		Err:       err,
		Retryable: IsRetryableHTTPStatus(statusCode),
	}
}

// GetDefaultRetryConfig returns a default retry configuration
func GetDefaultRetryConfig() *workflow.RetryConfig {
	return &workflow.RetryConfig{
		MaxAttempts:  3,
		InitialDelay: "1s",
		MaxDelay:     "60s",
		Multiplier:   2.0,
		RetryOn:      []string{"timeout", "network", "status:500", "status:502", "status:503", "status:504"},
	}
}
