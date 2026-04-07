package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketJSONRPC(t *testing.T) {
	server := NewRPCServer(":19801")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Fatalf("Failed to start RPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:19801/", nil)
	if err != nil {
		t.Fatalf("WebSocket dial error: %v", err)
	}
	defer conn.Close()

	// Test ping
	pingReq := `{"jsonrpc":"2.0","id":1,"method":"ping","params":[]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(pingReq))
	if err != nil {
		t.Fatalf("Failed to send ping request: %v", err)
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read ping response: %v", err)
	}

	var pingResp JSONRPCResponse
	if err := json.Unmarshal(message, &pingResp); err != nil {
		t.Fatalf("Failed to unmarshal ping response: %v", err)
	}

	if pingResp.Result != "pong" {
		t.Fatalf("Unexpected ping response: %v", pingResp.Result)
	}

	// Test invalid method
	unknownReq := `{"jsonrpc":"2.0","id":2,"method":"unknownMethod","params":[]}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(unknownReq))
	if err != nil {
		t.Fatalf("Failed to send unknown method request: %v", err)
	}

	_, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read unknown method response: %v", err)
	}

	var unknownResp JSONRPCResponse
	if err := json.Unmarshal(message, &unknownResp); err != nil {
		t.Fatalf("Failed to unmarshal unknown method response: %v", err)
	}
	if unknownResp.Error == nil {
		t.Fatalf("Expected error, got nil")
	}
	if unknownResp.Error.Code != ErrCodeMethodNotFound {
		t.Fatalf("Unexpected error response for unknown method: %v", unknownResp.Error)
	}
}

func TestMultipleConnections(t *testing.T) {
	server := NewRPCServer(":19802")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Fatalf("Failed to start RPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	numClients := 5
	done := make(chan bool, numClients)
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:19802/", nil)
			if err != nil {
				t.Errorf("Client %d: WebSocket dial error: %v", clientID, err)
				done <- false
				return
			}
			defer conn.Close()

			pingReq := `{"jsonrpc":"2.0","id":` + string(rune('0'+clientID)) + `,"method":"ping"}`
			err = conn.WriteMessage(websocket.TextMessage, []byte(pingReq))
			if err != nil {
				t.Errorf("Client %d: Failed to send ping request: %v", clientID, err)
				done <- false
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				t.Errorf("Client %d: Failed to read ping response: %v", clientID, err)
				done <- false
				return
			}

			var resp JSONRPCResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				t.Errorf("Client %d: Failed to unmarshal ping response: %v", clientID, err)
				done <- false
				return
			}

			if resp.Result != "pong" {
				t.Errorf("Client %d: Unexpected ping response: %v", clientID, resp.Result)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	for i := 0; i < numClients; i++ {
		success := <-done
		if !success {
			t.Fatalf("One or more clients failed")
		}
	}
}
