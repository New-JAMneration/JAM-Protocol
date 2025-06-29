package quic

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func TestStreamReadWrite(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer store.CloseMiniRedis()

	// context
	ctx := context.Background()
	// Start a QUIC listener (server) on "localhost:0" to let the OS assign a free port.
	listenAddr := "localhost:0"
	quicCfg := NewQuicConfig()
	listener, err := NewListener(listenAddr, false, NewTLSConfig, quicCfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	// Get the actual listening address.
	addr := listener.ListenAddress()
	if addr == "" {
		t.Fatal("Listener address is empty")
	}

	// Start a server goroutine that accepts a connection and stream.
	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			serverDone <- err
			return
		}
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			serverDone <- err
			return
		}
		// Read data sent by the client.
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			serverDone <- err
			return
		}
		// Echo the received data back to the client.
		response := "Echo: " + string(buf[:n])
		_, err = stream.Write([]byte(response))
		serverDone <- err
	}()

	// Create a client connection and open a stream.
	tlsCfg, err := NewTLSConfig(false, false)
	tlsCfg.InsecureSkipVerify = true // For testing, skip TLS verification.
	if err != nil {
		t.Fatalf("Failed to create client TLS config: %v", err)
	}
	clientConn, err := Dial(ctx, addr, tlsCfg, quicCfg, Validator)
	if err != nil {
		t.Fatalf("Client dial failed: %v", err)
	}
	streamQ, err := clientConn.OpenStream(context.Background())
	if err != nil {
		t.Fatalf("Client open stream failed: %v", err)
	}

	// Wrap the underlying quic.Stream with our Stream module.
	stream := &Stream{Stream: streamQ}

	// Write a test message to the stream.
	msg := "Hello Stream"
	_, err = stream.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Client write failed: %v", err)
	}

	// Set a read deadline and read the echo response.
	err = stream.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		t.Fatalf("SetReadDeadline failed: %v", err)
	}
	buf := make([]byte, 1024)
	n, err := stream.Read(buf)
	if err != nil {
		t.Fatalf("Client read failed: %v", err)
	}
	expected := "Echo: " + msg
	if string(buf[:n]) != expected {
		t.Errorf("Unexpected stream response: got %q, expected %q", string(buf[:n]), expected)
	}

	// Wait for the server goroutine to complete.
	if err := <-serverDone; err != nil {
		t.Fatalf("Server encountered error: %v", err)
	}

	// Close the stream and client connection.
	if err := stream.Close(); err != nil {
		t.Logf("Error closing stream: %v", err)
	}
	if err := clientConn.Close(); err != nil {
		t.Logf("Error closing client connection: %v", err)
	}
}
