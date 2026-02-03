package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		response := `{"data": {"viewer": {"id": "user-123", "name": "Test User"}}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		authHeader: "Bearer test-token",
	}

	resp, err := client.ExecuteRaw(context.Background(), "{ viewer { id name } }", nil)
	if err != nil {
		t.Fatalf("ExecuteRaw failed: %v", err)
	}

	if resp.Data == nil {
		t.Fatal("Expected Data to be non-nil")
	}

	// Unmarshal the raw JSON data
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal Data: %v", err)
	}

	viewer, ok := data["viewer"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected viewer in data")
	}

	if viewer["id"] != "user-123" {
		t.Errorf("Expected id 'user-123', got '%v'", viewer["id"])
	}
}

func TestExecuteRaw_WithGraphQLErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": null,
			"errors": [
				{
					"message": "Field 'invalid' doesn't exist on type 'Viewer'",
					"locations": [{"line": 1, "column": 10}],
					"path": ["viewer", "invalid"],
					"extensions": {"code": "FIELD_NOT_FOUND"}
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		authHeader: "Bearer test-token",
	}

	resp, err := client.ExecuteRaw(context.Background(), "{ viewer { invalid } }", nil)
	if err != nil {
		t.Fatalf("ExecuteRaw should not return error for GraphQL errors: %v", err)
	}

	if len(resp.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(resp.Errors))
	}

	if resp.Errors[0].Message != "Field 'invalid' doesn't exist on type 'Viewer'" {
		t.Errorf("Unexpected error message: %s", resp.Errors[0].Message)
	}

	if resp.Errors[0].Extensions["code"] != "FIELD_NOT_FOUND" {
		t.Errorf("Expected extension code 'FIELD_NOT_FOUND', got '%v'", resp.Errors[0].Extensions["code"])
	}
}

func TestExecuteRaw_WithVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req.Variables["id"] != "issue-123" {
			t.Errorf("Expected variable id 'issue-123', got '%v'", req.Variables["id"])
		}

		response := `{"data": {"issue": {"id": "issue-123"}}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		authHeader: "Bearer test-token",
	}

	variables := map[string]interface{}{"id": "issue-123"}
	resp, err := client.ExecuteRaw(context.Background(), "query($id: String!) { issue(id: $id) { id } }", variables)
	if err != nil {
		t.Fatalf("ExecuteRaw failed: %v", err)
	}

	if resp.Data == nil {
		t.Fatal("Expected Data to be non-nil")
	}
}

func TestExecuteRaw_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		authHeader: "Bearer invalid-token",
	}

	_, err := client.ExecuteRaw(context.Background(), "{ viewer { id } }", nil)
	if err == nil {
		t.Fatal("Expected error for HTTP 401")
	}
}

func TestGraphQLErrorUnmarshal(t *testing.T) {
	jsonData := `{
		"message": "Test error",
		"locations": [{"line": 1, "column": 5}],
		"path": ["viewer", "name"],
		"extensions": {"code": "TEST_ERROR", "field": "name"}
	}`

	var gqlErr GraphQLError
	err := json.Unmarshal([]byte(jsonData), &gqlErr)
	if err != nil {
		t.Fatalf("Failed to unmarshal GraphQLError: %v", err)
	}

	if gqlErr.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got '%s'", gqlErr.Message)
	}

	if len(gqlErr.Locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(gqlErr.Locations))
	}

	if gqlErr.Locations[0].Line != 1 || gqlErr.Locations[0].Column != 5 {
		t.Errorf("Unexpected location: line=%d, column=%d", gqlErr.Locations[0].Line, gqlErr.Locations[0].Column)
	}

	if gqlErr.Extensions["code"] != "TEST_ERROR" {
		t.Errorf("Expected extension code 'TEST_ERROR', got '%v'", gqlErr.Extensions["code"])
	}
}
