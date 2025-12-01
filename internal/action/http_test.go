package action

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codecrafted007/autozap/internal/workflow"
)

func TestExecuteHttpAction(t *testing.T) {
	t.Run("Successful GET Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "test-get",
			URL:    server.URL,
			Method: "GET",
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Successful POST Request With Body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("Created"))
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "test-post",
			URL:    server.URL,
			Method: "POST",
			Body:   `{"key": "value"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid Action Type", func(t *testing.T) {
		action := &workflow.Action{
			Type:   workflow.ActionTypeBash,
			Name:   "wrong-type",
			URL:    "https://example.com",
			Method: "GET",
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for invalid action type, got nil")
		}
	})

	t.Run("Empty URL", func(t *testing.T) {
		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "empty-url",
			URL:    "",
			Method: "GET",
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for empty URL, got nil")
		}
	})

	t.Run("Empty Method", func(t *testing.T) {
		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "empty-method",
			URL:    "https://example.com",
			Method: "",
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for empty method, got nil")
		}
	})

	t.Run("ExpectStatus Single Value Match", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:         workflow.ActionTypeHTTP,
			Name:         "expect-200",
			URL:          server.URL,
			Method:       "GET",
			ExpectStatus: 200,
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error for matching status, got: %v", err)
		}
	})

	t.Run("ExpectStatus Single Value Mismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:         workflow.ActionTypeHTTP,
			Name:         "expect-200-get-404",
			URL:          server.URL,
			Method:       "GET",
			ExpectStatus: 200,
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for status mismatch, got nil")
		}
	})

	t.Run("ExpectStatus Multiple Values Match", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:         workflow.ActionTypeHTTP,
			Name:         "expect-multiple",
			URL:          server.URL,
			Method:       "POST",
			ExpectStatus: []interface{}{200, 201, 202},
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error for matching status in list, got: %v", err)
		}
	})

	t.Run("ExpectBodyContains Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello World from server"))
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:               workflow.ActionTypeHTTP,
			Name:               "expect-body",
			URL:                server.URL,
			Method:             "GET",
			ExpectBodyContains: "Hello World",
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error for body containing expected string, got: %v", err)
		}
	})

	t.Run("ExpectBodyContains Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello World"))
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:               workflow.ActionTypeHTTP,
			Name:               "expect-missing-body",
			URL:                server.URL,
			Method:             "GET",
			ExpectBodyContains: "NonExistentString",
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for body not containing expected string, got nil")
		}
	})

	t.Run("Timeout Specified", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:    workflow.ActionTypeHTTP,
			Name:    "with-timeout",
			URL:     server.URL,
			Method:  "GET",
			Timeout: "5s",
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid Timeout Format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:    workflow.ActionTypeHTTP,
			Name:    "invalid-timeout",
			URL:     server.URL,
			Method:  "GET",
			Timeout: "invalid",
		}

		err := ExecuteHttpAction(action)
		if err == nil {
			t.Fatal("Expected error for invalid timeout format, got nil")
		}
	})

	t.Run("Custom Headers", func(t *testing.T) {
		receivedHeaders := make(map[string]string)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders["Authorization"] = r.Header.Get("Authorization")
			receivedHeaders["X-Custom-Header"] = r.Header.Get("X-Custom-Header")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "custom-headers",
			URL:    server.URL,
			Method: "GET",
			Headers: map[string]string{
				"Authorization":   "Bearer token123",
				"X-Custom-Header": "custom-value",
			},
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if receivedHeaders["Authorization"] != "Bearer token123" {
			t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", receivedHeaders["Authorization"])
		}

		if receivedHeaders["X-Custom-Header"] != "custom-value" {
			t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", receivedHeaders["X-Custom-Header"])
		}
	})

	t.Run("PUT Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "test-put",
			URL:    server.URL,
			Method: "PUT",
			Body:   `{"update": "data"}`,
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("DELETE Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		action := &workflow.Action{
			Type:   workflow.ActionTypeHTTP,
			Name:   "test-delete",
			URL:    server.URL,
			Method: "DELETE",
		}

		err := ExecuteHttpAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
}
