package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  json.RawMessage(`{"limit":10}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded JSONRPCRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got '%s'", decoded.JSONRPC)
	}

	if decoded.Method != "tools/list" {
		t.Errorf("expected method 'tools/list', got '%s'", decoded.Method)
	}

	if id, ok := decoded.ID.(float64); !ok || int(id) != 1 {
		t.Errorf("expected ID 1, got %v", decoded.ID)
	}
}

func TestJSONRPCRequest_UnmarshalVariousIDTypes(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		expected interface{}
	}{
		{
			name:     "numeric ID",
			json:     `{"jsonrpc":"2.0","id":42,"method":"test"}`,
			expected: float64(42),
		},
		{
			name:     "string ID",
			json:     `{"jsonrpc":"2.0","id":"test-123","method":"test"}`,
			expected: "test-123",
		},
		{
			name:     "null ID",
			json:     `{"jsonrpc":"2.0","id":null,"method":"test"}`,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req JSONRPCRequest
			if err := json.Unmarshal([]byte(tc.json), &req); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			switch expected := tc.expected.(type) {
			case float64:
				if id, ok := req.ID.(float64); !ok || id != expected {
					t.Errorf("expected ID %v, got %v", expected, req.ID)
				}
			case string:
				if id, ok := req.ID.(string); !ok || id != expected {
					t.Errorf("expected ID %v, got %v", expected, req.ID)
				}
			case nil:
				if req.ID != nil {
					t.Errorf("expected ID nil, got %v", req.ID)
				}
			}
		})
	}
}

func TestJSONRPCError_Marshal(t *testing.T) {
	err := JSONRPCError{
		Code:    -32001,
		Message: "Authentication failed",
		Data: map[string]interface{}{
			"type":   "auth_error",
			"reason": "invalid_token",
		},
	}

	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("failed to marshal error: %v", marshalErr)
	}

	var decoded JSONRPCError
	if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal error: %v", unmarshalErr)
	}

	if decoded.Code != -32001 {
		t.Errorf("expected code -32001, got %d", decoded.Code)
	}

	if decoded.Message != "Authentication failed" {
		t.Errorf("expected message 'Authentication failed', got '%s'", decoded.Message)
	}

	if decoded.Data == nil {
		t.Fatal("expected data to be present")
	}
}

func TestJSONRPCErrorResponse_Marshal(t *testing.T) {
	resp := JSONRPCErrorResponse{
		JSONRPC: "2.0",
		ID:      "test-id",
		Error: &JSONRPCError{
			Code:    -32001,
			Message: "Test error",
			Data: AuthErrorData{
				Type:           "auth_error",
				Reason:         "test_reason",
				RequiresReauth: true,
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded JSONRPCErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got '%s'", decoded.JSONRPC)
	}

	if decoded.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%v'", decoded.ID)
	}

	if decoded.Error == nil {
		t.Fatal("expected error to be present")
	}

	if decoded.Error.Code != -32001 {
		t.Errorf("expected error code -32001, got %d", decoded.Error.Code)
	}
}

func TestAuthErrorData_Marshal(t *testing.T) {
	authErr := AuthErrorData{
		Type:           "auth_error",
		Reason:         "token_expired",
		ReauthURL:      "https://example.com/oauth/authorize",
		RequiresReauth: true,
	}

	data, err := json.Marshal(authErr)
	if err != nil {
		t.Fatalf("failed to marshal auth error: %v", err)
	}

	var decoded AuthErrorData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal auth error: %v", err)
	}

	if decoded.Type != "auth_error" {
		t.Errorf("expected type 'auth_error', got '%s'", decoded.Type)
	}

	if decoded.Reason != "token_expired" {
		t.Errorf("expected reason 'token_expired', got '%s'", decoded.Reason)
	}

	if decoded.ReauthURL != "https://example.com/oauth/authorize" {
		t.Errorf("expected reauthUrl, got '%s'", decoded.ReauthURL)
	}

	if !decoded.RequiresReauth {
		t.Error("expected requiresReauth to be true")
	}
}

func TestAuthErrorData_OmitEmpty(t *testing.T) {
	// Test that empty fields are omitted in JSON
	authErr := AuthErrorData{
		Type:   "auth_error",
		Reason: "missing_token",
		// ReauthURL and RequiresReauth are omitted
	}

	data, err := json.Marshal(authErr)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)

	// Check that empty fields are omitted
	if containsKey(jsonStr, "reauthUrl") {
		t.Error("expected reauthUrl to be omitted when empty")
	}

	// RequiresReauth is false (zero value), should be omitted due to omitempty
	if containsKey(jsonStr, "requiresReauth") {
		t.Error("expected requiresReauth to be omitted when false")
	}

	// Non-empty fields should be present
	if !containsKey(jsonStr, "type") {
		t.Error("expected type to be present")
	}

	if !containsKey(jsonStr, "reason") {
		t.Error("expected reason to be present")
	}
}

func TestNewErrorResponse(t *testing.T) {
	testCases := []struct {
		name    string
		id      interface{}
		message string
		code    int
		data    interface{}
	}{
		{
			name:    "with numeric ID",
			id:      42,
			message: "Test error",
			code:    -32001,
			data: AuthErrorData{
				Type:   "auth_error",
				Reason: "test",
			},
		},
		{
			name:    "with string ID",
			id:      "test-id-123",
			message: "Another error",
			code:    -32002,
			data: map[string]string{
				"detail": "error detail",
			},
		},
		{
			name:    "with nil ID",
			id:      nil,
			message: "Error with nil ID",
			code:    -32003,
			data:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := NewErrorResponse(tc.id, tc.message, tc.code, tc.data)

			if resp.JSONRPC != "2.0" {
				t.Errorf("expected jsonrpc '2.0', got '%s'", resp.JSONRPC)
			}

			if resp.ID != tc.id {
				t.Errorf("expected ID %v, got %v", tc.id, resp.ID)
			}

			if resp.Error == nil {
				t.Fatal("expected error to be present")
			}

			if resp.Error.Code != tc.code {
				t.Errorf("expected code %d, got %d", tc.code, resp.Error.Code)
			}

			if resp.Error.Message != tc.message {
				t.Errorf("expected message '%s', got '%s'", tc.message, resp.Error.Message)
			}

			// Marshal to ensure it's valid JSON
			_, err := json.Marshal(resp)
			if err != nil {
				t.Errorf("failed to marshal response: %v", err)
			}
		})
	}
}

func TestNewErrorResponse_JSONCompliance(t *testing.T) {
	// Test that the response structure complies with JSON-RPC 2.0 spec
	resp := NewErrorResponse(1, "Test error", -32600, nil)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Parse as generic map to check structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Check required fields according to JSON-RPC 2.0
	requiredFields := []string{"jsonrpc", "id", "error"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("required field '%s' is missing", field)
		}
	}

	// Check error object structure
	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("error field should be an object")
	}

	errorFields := []string{"code", "message"}
	for _, field := range errorFields {
		if _, exists := errorObj[field]; !exists {
			t.Errorf("required error field '%s' is missing", field)
		}
	}
}

func TestJSONRPCRequest_RawParams(t *testing.T) {
	// Test that params can be parsed as raw JSON
	jsonStr := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "test",
		"params": {
			"nested": {
				"value": 123
			},
			"array": [1, 2, 3]
		}
	}`

	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(req.Params) == 0 {
		t.Error("expected params to be present")
	}

	// Parse params into a map
	var params map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		t.Fatalf("failed to parse params: %v", err)
	}

	if _, exists := params["nested"]; !exists {
		t.Error("expected 'nested' field in params")
	}

	if _, exists := params["array"]; !exists {
		t.Error("expected 'array' field in params")
	}
}

// Helper function to check if a JSON string contains a key
func containsKey(jsonStr, key string) bool {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return false
	}
	_, exists := data[key]
	return exists
}

func BenchmarkNewErrorResponse(b *testing.B) {
	authData := AuthErrorData{
		Type:           "auth_error",
		Reason:         "invalid_token",
		RequiresReauth: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewErrorResponse(i, "Test error", -32001, authData)
	}
}

func BenchmarkJSONRPCErrorResponse_Marshal(b *testing.B) {
	resp := NewErrorResponse(1, "Test error", -32001, AuthErrorData{
		Type:           "auth_error",
		Reason:         "invalid_token",
		RequiresReauth: true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}
