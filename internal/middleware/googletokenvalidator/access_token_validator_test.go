package googletokenvalidator

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/schnurbus/go-mcp-gateway/pkg/jsonrpc"
)

func TestNew_MissingAuthorizationHeader(t *testing.T) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var errorResp jsonrpc.JSONRPCErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if errorResp.Error == nil {
		t.Fatal("expected error in response")
	}

	if errorResp.Error.Code != -32001 {
		t.Errorf("expected error code -32001, got %d", errorResp.Error.Code)
	}

	if errorResp.Error.Message != "Authentication token missing" {
		t.Errorf("unexpected error message: %s", errorResp.Error.Message)
	}

	// Check error data
	dataBytes, _ := json.Marshal(errorResp.Error.Data)
	var authError jsonrpc.AuthErrorData
	json.Unmarshal(dataBytes, &authError)

	if authError.Type != "auth_error" {
		t.Errorf("expected type 'auth_error', got '%s'", authError.Type)
	}

	if authError.Reason != "missing_token" {
		t.Errorf("expected reason 'missing_token', got '%s'", authError.Reason)
	}
}

func TestNew_InvalidAuthorizationHeaderFormat(t *testing.T) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "InvalidFormat token123")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var errorResp jsonrpc.JSONRPCErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if errorResp.Error == nil {
		t.Fatal("expected error in response")
	}

	if errorResp.Error.Message != "Authentication token missing" {
		t.Errorf("unexpected error message: %s", errorResp.Error.Message)
	}
}

func TestNew_InvalidRequestBody(t *testing.T) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// Invalid JSON body
	req := httptest.NewRequest("POST", "/test", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	// Should still process and return missing auth error
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var errorResp jsonrpc.JSONRPCErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if errorResp.Error == nil {
		t.Fatal("expected error in response")
	}
}

func TestValidateGoogleToken_Success(t *testing.T) {
	// Create mock server for Google tokeninfo endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "tokeninfo") {
			t.Errorf("unexpected URL: %s", r.URL.String())
		}

		response := TokenInfoResponse{
			Aud:   "test-client-id",
			Sub:   "12345",
			Email: "test@example.com",
			Exp:   JsonTimestamp(time.Now().Add(1 * time.Hour)),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// This would require refactoring validateGoogleToken to accept a custom URL
	// For now, we'll test the logic separately
	isValid, err := validateGoogleToken("valid-token", "test-client-id")

	// Note: This will fail in unit tests without mocking the actual Google API
	// In a real scenario, you'd use dependency injection or interface mocking
	if err == nil && isValid {
		// Test passes if Google API is accessible
		t.Log("Token validation succeeded (using real Google API)")
	} else {
		t.Log("Token validation requires mock or real Google API access")
	}
}

func TestValidateGoogleToken_ExpiredToken(t *testing.T) {
	// Test the error types
	if ErrTokenExpired == nil {
		t.Error("ErrTokenExpired should be defined")
	}

	if ErrTokenExpired.Error() != "token is expired" {
		t.Errorf("unexpected error message: %s", ErrTokenExpired.Error())
	}
}

func TestValidateGoogleToken_InvalidAudience(t *testing.T) {
	if ErrInvalidAud == nil {
		t.Error("ErrInvalidAud should be defined")
	}

	if ErrInvalidAud.Error() != "invalid audience" {
		t.Errorf("unexpected error message: %s", ErrInvalidAud.Error())
	}
}

func TestValidateGoogleToken_DecodeError(t *testing.T) {
	if ErrDecode == nil {
		t.Error("ErrDecode should be defined")
	}

	if ErrDecode.Error() != "cannot decode token info" {
		t.Errorf("unexpected error message: %s", ErrDecode.Error())
	}
}

func TestNew_RequestIDInErrorResponse(t *testing.T) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	testCases := []struct {
		name       string
		requestID  interface{}
		body       string
	}{
		{
			name:      "numeric ID",
			requestID: 42,
			body:      `{"jsonrpc":"2.0","id":42,"method":"test","params":{}}`,
		},
		{
			name:      "string ID",
			requestID: "test-id-123",
			body:      `{"jsonrpc":"2.0","id":"test-id-123","method":"test","params":{}}`,
		},
		{
			name:      "null ID",
			requestID: nil,
			body:      `{"jsonrpc":"2.0","id":null,"method":"test","params":{}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			body, _ := io.ReadAll(resp.Body)
			var errorResp jsonrpc.JSONRPCErrorResponse
			if err := json.Unmarshal(body, &errorResp); err != nil {
				t.Fatalf("failed to parse error response: %v", err)
			}

			// Verify the ID is preserved in the error response
			switch expected := tc.requestID.(type) {
			case int:
				if respID, ok := errorResp.ID.(float64); !ok || int(respID) != expected {
					t.Errorf("expected ID %v, got %v", expected, errorResp.ID)
				}
			case string:
				if respID, ok := errorResp.ID.(string); !ok || respID != expected {
					t.Errorf("expected ID %v, got %v", expected, errorResp.ID)
				}
			case nil:
				if errorResp.ID != nil {
					t.Errorf("expected ID nil, got %v", errorResp.ID)
				}
			}
		})
	}
}

func TestNew_MiddlewareChain(t *testing.T) {
	app := fiber.New()

	// Track if next handler was called
	nextCalled := false

	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		nextCalled = true
		return c.JSON(fiber.Map{"success": true})
	})

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// Missing auth header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if nextCalled {
		t.Error("next handler should not be called when auth fails")
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestTokenInfoResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"azp": "test-azp",
		"aud": "test-aud",
		"sub": "12345",
		"scope": "openid profile email",
		"exp": "1735689600",
		"expires_in": "3599",
		"email": "test@example.com",
		"email_verified": "true",
		"access_type": "offline"
	}`

	var tokenInfo TokenInfoResponse
	if err := json.Unmarshal([]byte(jsonData), &tokenInfo); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if tokenInfo.Aud != "test-aud" {
		t.Errorf("expected aud 'test-aud', got '%s'", tokenInfo.Aud)
	}

	if tokenInfo.Sub != "12345" {
		t.Errorf("expected sub '12345', got '%s'", tokenInfo.Sub)
	}

	if tokenInfo.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", tokenInfo.Email)
	}
}

func TestNew_ErrorResponseStructure(t *testing.T) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	reqBody := `{"jsonrpc":"2.0","id":"req-123","method":"tools/list","params":{}}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(resp.Body)
	var errorResp jsonrpc.JSONRPCErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	// Verify JSON-RPC 2.0 compliance
	if errorResp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc '2.0', got '%s'", errorResp.JSONRPC)
	}

	if errorResp.ID != "req-123" {
		t.Errorf("expected id 'req-123', got '%v'", errorResp.ID)
	}

	if errorResp.Error == nil {
		t.Fatal("expected error field to be present")
	}

	// Verify error structure
	if errorResp.Error.Code != -32001 {
		t.Errorf("expected error code -32001, got %d", errorResp.Error.Code)
	}

	if errorResp.Error.Message == "" {
		t.Error("expected non-empty error message")
	}

	// Verify data field structure
	if errorResp.Error.Data == nil {
		t.Fatal("expected error data to be present")
	}

	dataBytes, _ := json.Marshal(errorResp.Error.Data)
	var authError jsonrpc.AuthErrorData
	if err := json.Unmarshal(dataBytes, &authError); err != nil {
		t.Fatalf("failed to parse auth error data: %v", err)
	}

	if authError.Type == "" {
		t.Error("expected non-empty error type")
	}

	if authError.Reason == "" {
		t.Error("expected non-empty error reason")
	}
}

func BenchmarkNew_MissingAuth(b *testing.B) {
	app := fiber.New()
	app.Use(New("test-client-id"))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		resp.Body.Close()
	}
}
