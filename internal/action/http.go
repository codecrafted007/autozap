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
	"github.com/codecrafted007/autozap/internal/metrics"
	"github.com/codecrafted007/autozap/internal/workflow"
)

// ExecuteHTTPAction executes an HTTP request defined in a workflow.Action.
// It handles method, URL, headers, body, timeout, and response validation.
func ExecuteHttpAction(action *workflow.Action, workflowName ...string) error {
	// Track action execution time
	startTime := time.Now()
	var executionError error

	// Defer metrics recording
	defer func() {
		if len(workflowName) > 0 && workflowName[0] != "" {
			status := "success"
			if executionError != nil {
				status = "failed"
			}
			metrics.RecordActionExecution(workflowName[0], action.Name, string(workflow.ActionTypeHTTP), status, time.Since(startTime))
		}
	}()

	if action.Type != workflow.ActionTypeHTTP {
		executionError = fmt.Errorf("invalid action type expected '%s' got '%s' ", workflow.ActionTypeHTTP.String(), action.Type.String())
		return executionError
	}
	if action.URL == "" {
		executionError = fmt.Errorf("http action '%s' has empty URL", action.Name)
		return executionError
	}
	if action.Method == "" {
		executionError = fmt.Errorf("http action '%s' has empty method", action.Name)
		return executionError
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
		executionError = fmt.Errorf("failed to create HTTP request: %w", err)
		return executionError
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
			executionError = fmt.Errorf("invalid timeout duration: %w", parseError)
			return executionError
		}

		ctx, cancel = context.WithTimeout(context.Background(), duration)
		defer cancel()
	}

	req = req.WithContext(ctx)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {

		if ctx.Err() == context.DeadlineExceeded {
			executionError = fmt.Errorf("HTTP action '%s' timed out after %s: %v", action.Name, action.Timeout, err)
			return executionError
		}
		executionError = fmt.Errorf("HTTP request failed for action '%s': %v", action.Name, err)
		return executionError
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.L().Errorw("Failed to close response body", "error", closeErr, "action_name", action.Name)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.L().Errorw("Failed to read HTTP response body", "error", err, "action_name", action.Name)
		executionError = fmt.Errorf("failed to read HTTP response body: %w", err)
		return executionError
	}
	responseBody := string(respBodyBytes)

	bodyOverview := responseBody
	if len(responseBody) > 200 {
		bodyOverview = responseBody[:200]
	}

	logFields := []interface{}{
		"action_name", action.Name,
		"method", action.Method,
		"url", action.URL,
		"status_code", resp.StatusCode,
		"respone_body_overview", bodyOverview, // print only first few charcters
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
					executionError = fmt.Errorf("HTTP action '%s': invalid type in expect_status list. Expected integer, got %T", action.Name, s)
					logger.L().Errorw("Invalid type in expect_status list", "error", executionError, "action_name", action.Name)
					return executionError
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
			executionError = fmt.Errorf("HTTP action '%s' failed: unexpected status code %d. Expected one of: %v", action.Name, resp.StatusCode, expectedStatuses)
			logger.L().Errorw("Unexpected status code", "error", executionError, "action_name", action.Name, "status_code", resp.StatusCode, "expected_statuses", expectedStatuses)
			return executionError
		}
	}

	// Validate response if body has the expected string

	if action.ExpectBodyContains != "" {
		if !strings.Contains(responseBody, action.ExpectBodyContains) {
			executionError = fmt.Errorf("HTTP action '%s' failed: response body does not contain expected string '%s'", action.Name, action.ExpectBodyContains)
			logger.L().Errorw("Response body does not contain expected string", "error", executionError, "action_name", action.Name)
			return executionError
		}
	}

	logger.L().Infow("Http action completed succesfully", "action_name", action.Name, "status_code", resp.Status)

	return nil
}
