package quic

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func TestConnectionDialAndClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer store.CloseMiniRedis()

	listenAddr := "localhost:0"
	quicCfg := NewQuicConfig()
	listener, err := NewListener(listenAddr, false, NewTLSConfig, quicCfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	addr := listener.ListenAddress()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.Accept(ctx)
		if err != nil {
			t.Errorf("Server Accept error: %v", err)
			return
		}
		log.Printf("Server accepted connection from %s", conn.RemoteAddr())
		err = conn.CloseWithError(0, "test close")
		if err != nil {
			t.Errorf("Server CloseWithError error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond) // Increase wait time for server readiness

	clientTLS, err := NewTLSConfig(false, false)
	if err != nil {
		t.Fatalf("Failed to create client TLS config: %v", err)
	}

	// Set up client TLS config to skip verification for testing
	clientTLS.InsecureSkipVerify = true // Skip certificate verification for testing

	clientConn, err := Dial(ctx, addr, clientTLS, quicCfg, Validator)
	if err != nil {
		t.Fatalf("Client dial failed: %v", err)
	}
	log.Printf("Client connected to server at %s", addr)

	time.Sleep(3 * time.Second) // Allow some time for the connection to be established

	err = clientConn.Close()
	if err != nil {
		t.Errorf("Client Close error: %v", err)
	}

	select {
	case <-serverDone:
		t.Logf("Step 10: Server completed connection handling")
	case <-ctx.Done():
		t.Fatalf("Test timed out")
	}
}
