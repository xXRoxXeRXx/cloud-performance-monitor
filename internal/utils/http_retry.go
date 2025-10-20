package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPRetryConfig configures HTTP request retry behavior
type HTTPRetryConfig struct {
	*RetryConfig
	ClientLogger
}

// NewHTTPRetryConfig creates a new HTTP retry configuration with default settings
func NewHTTPRetryConfig() *HTTPRetryConfig {
	return &HTTPRetryConfig{
		RetryConfig:  DefaultRetryConfig(),
		ClientLogger: &DefaultClientLogger{},
	}
}

// DoWithRetry executes an HTTP request with retry logic
func (hrc *HTTPRetryConfig) DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, operation string) (*http.Response, error) {
	var resp *http.Response
	var err error
	
	retryErr := hrc.RetryConfig.WithRetry(ctx, operation, func(retryCtx context.Context) error {
		// Clone the request for each retry attempt
		reqClone := req.Clone(retryCtx)
		
		resp, err = client.Do(reqClone)
		if err != nil {
			return err
		}
		
		// Check if HTTP status code indicates a retryable error
		if hrc.isRetryableHTTPStatus(resp.StatusCode) {
			resp.Body.Close() // Close the body before retrying
			return fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
		}
		
		return nil
	})
	
	if retryErr != nil {
		return nil, retryErr
	}
	
	return resp, nil
}

// isRetryableHTTPStatus checks if an HTTP status code should trigger a retry
func (hrc *HTTPRetryConfig) isRetryableHTTPStatus(statusCode int) bool {
	switch statusCode {
	// 5xx server errors are generally retryable
	case 500, 502, 503, 504, 507, 508, 509, 510, 511:
		return true
	// 408 Request Timeout is retryable
	case 408:
		return true
	// 429 Too Many Requests is retryable (rate limiting)
	case 429:
		return true
	// Some 4xx errors that might be temporary
	case 409: // Conflict - might be resolved on retry
		return true
	default:
		return false
	}
}

// DoWithRetryAndLog executes an HTTP request with retry logic and detailed logging
func (hrc *HTTPRetryConfig) DoWithRetryAndLog(ctx context.Context, client *http.Client, req *http.Request, operation, service, instance string) (*http.Response, error) {
	startTime := time.Now()
	
	hrc.ClientLogger.LogOperation(DEBUG, service, instance, operation, "start",
		fmt.Sprintf("Starting HTTP %s request to %s", req.Method, req.URL.String()),
		map[string]interface{}{
			"method":    req.Method,
			"url":       req.URL.String(),
			"operation": operation,
		})
	
	resp, err := hrc.DoWithRetry(ctx, client, req, operation)
	duration := time.Since(startTime)
	
	if err != nil {
		hrc.ClientLogger.LogOperation(ERROR, service, instance, operation, "error",
			fmt.Sprintf("HTTP request failed after retries: %v", err),
			map[string]interface{}{
				"method":    req.Method,
				"url":       req.URL.String(),
				"operation": operation,
				"duration":  duration.String(),
				"error":     err.Error(),
			})
		return nil, err
	}
	
	hrc.ClientLogger.LogOperation(DEBUG, service, instance, operation, "success",
		fmt.Sprintf("HTTP %s request completed with status %d", req.Method, resp.StatusCode),
		map[string]interface{}{
			"method":      req.Method,
			"url":         req.URL.String(),
			"operation":   operation,
			"duration":    duration.String(),
			"status_code": resp.StatusCode,
		})
	
	return resp, nil
}