package rpc

import (
	"encoding/json"
	"testing"
)

func TestHandlePing(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"2.0","id":1,"method":"ping", "params":[]}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Fatalf("Expected JSONRPC version '2.0', got: %s", resp.JSONRPC)
	}

	if resp.Error != nil {
		t.Fatalf("Expected no error, got: %v", resp.Error)
	}

	expectedResult := "pong"
	if resp.Result != expectedResult {
		t.Fatalf("Expected result %q, got %q", expectedResult, resp.Result)
	}
}

func TestMethodNotFound(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"2.0","id":1,"method":"unknownMethod", "params":[]}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatalf("Expected an error for unknown method")
	}

	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Fatalf("Expected error code %d, got %d", ErrCodeMethodNotFound, resp.Error.Code)
	}

	if resp.Error.Message != "Method not found" {
		t.Fatalf("Expected error message 'Method not found', got %q", resp.Error.Message)
	}
}

func TestParseError(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{invalid json}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatalf("Expected a parse error")
	}

	if resp.Error.Code != ErrCodeParseError {
		t.Fatalf("Expected error code %d, got %d", ErrCodeParseError, resp.Error.Code)
	}

	if resp.Error.Message != "Parse error" {
		t.Fatalf("Expected error message 'Parse error', got %q", resp.Error.Message)
	}
}

func TestInvalidRequest(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"1.0","id":3,"method":"ping"}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatalf("Expected an invalid request error")
	}

	if resp.Error.Code != ErrCodeInvalidRequest {
		t.Fatalf("Expected error code %d, got %d", ErrCodeInvalidRequest, resp.Error.Code)
	}

	if resp.Error.Message != "Invalid Request" {
		t.Fatalf("Expected error message 'Invalid Request', got %q", resp.Error.Message)
	}
}

func TestMissingMethod(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"2.0","id":4, "params":[]}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error == nil {
		t.Fatalf("Expected an invalid request error for missing method")
	}

	if resp.Error.Code != ErrCodeInvalidRequest {
		t.Fatalf("Expected error code %d, got %d", ErrCodeInvalidRequest, resp.Error.Code)
	}

	if resp.Error.Message != "Invalid Request" {
		t.Fatalf("Expected error message 'Invalid Request', got %q", resp.Error.Message)
	}
}

func TestPingWithStringID(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"2.0","id":"abc","method":"ping", "params":[]}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedResult := "pong"
	if resp.Result != expectedResult {
		t.Fatalf("Expected result %q, got %q", expectedResult, resp.Result)
	}

	var idStr string
	if err := json.Unmarshal(*resp.ID, &idStr); err != nil {
		t.Fatalf("Failed to unmarshal ID: %v", err)
	}

	if idStr != "abc" {
		t.Fatalf("Expected ID 'abc', got %q", idStr)
	}
}

func TestPingWithNullID(t *testing.T) {
	handler := NewHandler()

	reqJSON := `{"jsonrpc":"2.0","id":null,"method":"ping", "params":[]}`
	respBytes := handler.HandleMessage([]byte(reqJSON))

	var resp JSONRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedResult := "pong"
	if resp.Result != expectedResult {
		t.Fatalf("Expected result %q, got %q", expectedResult, resp.Result)
	}
}
