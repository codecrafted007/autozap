package action

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

// ExecuteHTTPAction executes an HTTP request defined in a workflow.Action.
// It handles method, URL, headers, body, timeout, and response validation.
func ExecuteHttpAction(action *workflow.Action) error {

	if action.Type != workflow.ActionTypeHTTP {
		return fmt.Errorf("invalid action type expected '%s' got '%s' ", workflow.ActionTypeHTTP.String(), action.Type.String())
	}
	if action.URL == "" {
		return fmt.Errorf("http action '%s' has empty URL", action.Name)
	}
	if action.Method == "" {
		return fmt.Errorf("http action '%s' has empty method", action.Name)
	}

	logger.L().Infow("Executing http action",
		"action_name", action.Name,
		"method", action.Method,
		"url", action.URL)

	var requestBody io.Reader
	if action.Body != "" {
		requestBody = bytes.NewBufferString(action.Body)
	}

	req, err := http.NewRequest(action.Method, action.URL, requestBody)
	if err != nil {
		logger.L().Errorw("Failed to create HTTP request", "error", err, "action_name", action.Name)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range action.Headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // This ensures context is cancelled when function exits

	if action.Timeout != "" {
		duration, parseError := time.ParseDuration(action.Timeout)
		if parseError != nil {
			logger.L().Errorw("Invalid timeout duration", "error", parseError, "timeout", action.Timeout, "action_name", action.Name)
			return fmt.Errorf("invalid timeout duration: %w", parseError)
		}

		ctx, cancel = context.WithTimeout(context.Background(), duration)
		defer cancel()
	}

	req = req.WithContext(ctx)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {

		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("HTTP action '%s' timed out after %s: %v", action.Name, action.Timeout, err)
		}
		return fmt.Errorf("HTTP request failed for action '%s': %v", action.Name, err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.L().Errorw("Failed to read HTTP response body", "error", err, "action_name", action.Name)
		return fmt.Errorf("failed to read HTTP response body: %w", err)
	}
	responseBody := string(respBodyBytes)

	logFields := []interface{}{
		"action_name", action.Name,
		"method", action.Method,
		"url", action.URL,
		"status_code", resp.StatusCode,
		"respone_body_overview", responseBody[:200], // print only first few charcters
	}
	logger.L().Infow("HTTP action response received", logFields...)

	if action.ExpectStatus != nil {
		expectedStatuses := []int{}

		if singleStatus, ok := action.ExpectStatus.(int); ok {
			expectedStatuses = append(expectedStatuses, singleStatus)
		} else if statusList, ok := action.ExpectStatus.([]interface{}); ok {
			for _, s := range statusList {
				if val, isInt := s.(int); isInt {
					expectedStatuses = append(expectedStatuses, val)
				} else {
					// Status cannot have other data type other than Int
					err := fmt.Errorf("HTTP action '%s': invalid type in expect_status list. Expected integer, got %T", action.Name, s)
					logger.L().Errorw("Invalid type in expect_status list", "error", err, "action_name", action.Name)
					return err
				}
			}
		}
		statusMatch := false

		for _, es := range expectedStatuses {
			if resp.StatusCode == es {
				statusMatch = true
			}
		}

		if !statusMatch {
			err := fmt.Errorf("HTTP action '%s' failed: unexpected status code %d. Expected one of: %v", action.Name, resp.StatusCode, expectedStatuses)
			logger.L().Errorw("Unexpected status code", "error", err, "action_name", action.Name, "status_code", resp.StatusCode, "expected_statuses", expectedStatuses)
			return err
		}
	}

	// Validate response if body has the expected string

	if action.ExpectBodyContains != "" {
		if !strings.Contains(responseBody, action.ExpectBodyContains) {
			err := fmt.Errorf("HTTP action '%s' failed: response body does not contain expected string '%s'", action.Name, action.ExpectBodyContains)
			logger.L().Errorw("Response body does not contain expected string", "error", err, "action_name", action.Name)
			return err
		}
	}

	logger.L().Infow("Http action completed succesfully", "action_name", action.Name, "status_code", resp.Status)

	return nil
}
