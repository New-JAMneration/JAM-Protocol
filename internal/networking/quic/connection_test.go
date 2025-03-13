package quic

import (
	"context"
	"testing"
	"time"
)

func TestConnectionDialAndClose(t *testing.T) {
	ctx := context.Background()
	listenAddr := "localhost:0"
	quicCfg := NewQuicConfig()
	listener, err := NewListener(listenAddr, NewTLSConfig, quicCfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	addr := listener.ListenAddress()
	if addr == "" {
		t.Fatalf("Failed to get listener address: %v", err)
	}

	serverDone := make(chan struct{})
	go func() {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			t.Errorf("Server Accept error: %v", err)
			close(serverDone)
			return
		}
		conn.CloseWithError(0, "test close")
		close(serverDone)
	}()

	tlsCfg, err := NewTLSConfig(false)
	if err != nil {
		t.Fatalf("Failed to create client TLS config: %v", err)
	}
	clientConn, err := Dial(ctx, addr, tlsCfg, quicCfg)
	if err != nil {
		t.Fatalf("Client dial failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	err = clientConn.Close()
	if err != nil {
		t.Errorf("Client Close error: %v", err)
	}
	<-serverDone
}
