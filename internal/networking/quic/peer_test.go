package quic

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type mockUPHandler struct {
	encodeFunc func(kind string, message interface{}) ([]byte, error)
}

func (m *mockUPHandler) EncodeMessage(kind string, message interface{}) ([]byte, error) {
	if m.encodeFunc != nil {
		return m.encodeFunc(kind, message)
	}
	return []byte("test message"), nil
}

type mockCEHandler struct {
	handleFunc func(stream *Stream) error
}

func (m *mockCEHandler) HandleStream(stream *Stream) error {
	if m.handleFunc != nil {
		return m.handleFunc(stream)
	}
	return nil
}

func createTestPeerConfig(role PeerRole) PeerConfig {
	publicKey, _, _ := ed25519.GenerateKey(rand.Reader)

	hash := "5c743dbc514284b2ea57798787c5a155ef9d7ac1e9499ec65910a7a3d65897b7"
	byteArray, _ := hex.DecodeString(hash)
	genesisHeader := types.HeaderHash(byteArray)

	return PeerConfig{
		Role:          role,
		Addr:          &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		GenesisHeader: genesisHeader,
		PublicKey:     publicKey,
		UPHandler:     &mockUPHandler{},
		CEHandler:     &mockCEHandler{},
	}
}

func TestNewPeer(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	tests := []struct {
		name    string
		config  PeerConfig
		wantErr bool
	}{
		{
			name:    "Create validator peer",
			config:  createTestPeerConfig(Validator),
			wantErr: false,
		},
		{
			name:    "Create builder peer",
			config:  createTestPeerConfig(Builder),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peer, err := NewPeer(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPeer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && peer == nil {
				t.Error("NewPeer() returned nil peer when no error expected")
				return
			}
			if peer != nil {
				if peer.publicKey == nil {
					t.Error("Peer public key is nil")
				}
				if peer.Listener == nil {
					t.Error("Peer listener is nil")
				}
				if peer.tlsConfig == nil {
					t.Error("Peer TLS config is nil")
				}
				if peer.quicConfig == nil {
					t.Error("Peer QUIC config is nil")
				}
				if peer.connManager == nil {
					t.Error("Peer connection manager is nil")
				}
				if peer.CEHandler == nil {
					t.Error("Peer CE handler is nil")
				}
				if peer.UPHandler == nil {
					t.Error("Peer UP handler is nil")
				}
			}
		})
	}
}

func TestPeerConnect(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	config := createTestPeerConfig(Validator)
	peer, err := NewPeer(config)
	if err != nil {
		t.Fatalf("Failed to create peer: %v", err)
	}

	peer.tlsConfig.InsecureSkipVerify = true

	listener, err := NewListener("localhost:0", false, NewTLSConfig, NewQuicConfig())
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	addr := listener.ListenAddress()
	t.Logf("Test listener started on: %s", addr)

	t.Run("Connect to non-existent address", func(t *testing.T) {
		nonExistentAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999}
		_, err := peer.Connect(nonExistentAddr, *peer)
		if err == nil {
			t.Error("Expected error when connecting to non-existent address")
		}
	})

	t.Run("Connect to valid address", func(t *testing.T) {
		serverDone := make(chan error, 1)
		go func() {
			_, err := listener.Accept(context.Background())
			if err != nil {
				serverDone <- err
				return
			}
			defer listener.Close()
			serverDone <- nil
		}()

		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to resolve address: %v", err)
		}

		conn, err := peer.Connect(tcpAddr, *peer)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		if conn == nil {
			t.Error("Expected non-nil connection")
		}

		select {
		case err := <-serverDone:
			if err != nil {
				t.Errorf("Server error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Server timeout")
		}

		conn2, err := peer.Connect(tcpAddr, *peer)
		if err != nil {
			t.Fatalf("Failed to reconnect: %v", err)
		}

		if conn2 == nil {
			t.Error("Expected non-nil connection on reconnect")
		}

		if conn != conn2 {
			t.Error("Expected same connection to be reused")
		}
	})
}

func TestPeerBroadcast(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	encodeCalled := false
	encodeFunc := func(kind string, message interface{}) ([]byte, error) {
		encodeCalled = true
		return []byte("broadcast message"), nil
	}

	config := createTestPeerConfig(Validator)
	config.UPHandler = &mockUPHandler{
		encodeFunc: encodeFunc,
	}

	peer, err := NewPeer(config)
	if err != nil {
		t.Fatalf("Failed to create peer: %v", err)
	}

	peer.tlsConfig.InsecureSkipVerify = true

	t.Run("Broadcast with no connections", func(t *testing.T) {
		peer.Broadcast("test", "test message")
		if !encodeCalled {
			t.Error("Expected EncodeMessage to be called")
		}
	})

	t.Run("Broadcast with connections", func(t *testing.T) {
		// Create a listener and connection for testing
		listener, err := NewListener("localhost:0", false, NewTLSConfig, NewQuicConfig())
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		addr := listener.ListenAddress()
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to resolve address: %v", err)
		}

		serverDone := make(chan error, 1)
		connectionReady := make(chan struct{})
		go func() {
			conn, err := listener.Accept(context.Background())
			if err != nil {
				serverDone <- err
				return
			}
			defer listener.Close()

			close(connectionReady)

			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				serverDone <- err
				return
			}

			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil {
				serverDone <- err
				return
			}

			expected := "broadcast message"
			if string(buf[:n]) != expected {
				serverDone <- fmt.Errorf("expected %s, got %s", expected, string(buf[:n]))
				return
			}

			stream.Close()
			serverDone <- nil
		}()

		_, err = peer.Connect(tcpAddr, *peer)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		select {
		case <-connectionReady:
			peer.Broadcast("test", "test message")
		case <-time.After(2 * time.Second):
			t.Fatal("Connection not ready in time")
		}

		select {
		case err := <-serverDone:
			if err != nil {
				t.Errorf("Server error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Server timeout")
		}
	})
}
