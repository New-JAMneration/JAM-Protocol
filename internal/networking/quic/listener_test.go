package quic

import (
	"context"
	"testing"
	"time"
)

func TestNewListener(t *testing.T) {
	listenAddr := "localhost:0"
	quicCfg := NewQuicConfig()

	listener, err := NewListener(listenAddr, NewTLSConfig, quicCfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	addr := listener.ListenAddress()
	if err != nil {
		t.Fatalf("Failed to get listen address: %v", err)
	}
	if addr == "" {
		t.Error("Listen address is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = listener.Accept(ctx)
	if err == nil {
		t.Error("Expected Accept to timeout with no incoming connection")
	}
}
