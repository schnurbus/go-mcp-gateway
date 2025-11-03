package jsonrpc

import "encoding/json"

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type JSONRPCErrorResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Error   *JSONRPCError `json:"error"`
}

type AuthErrorData struct {
	Type           string `json:"type,omitempty"`
	Reason         string `json:"reason,omitempty"`
	ReauthURL      string `json:"reauthUrl,omitempty"`
	RequiresReauth bool   `json:"requiresReauth,omitempty"`
}

func NewErrorResponse(id any, message string, code int, data any) *JSONRPCErrorResponse {
	return &JSONRPCErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Message: message,
			Code:    code,
			Data:    data,
		},
	}
}
