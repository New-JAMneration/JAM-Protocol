package quic

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func setupTestGenesis(t *testing.T) func() {
	t.Helper()
	blockchain.ResetInstance()
	cs := blockchain.GetInstance()
	genesis := types.Block{Header: types.Header{Slot: 0}, Extrinsic: types.Extrinsic{}}
	if err := cs.GenerateGenesisBlock(genesis); err != nil {
		t.Fatalf("Failed to setup genesis: %v", err)
	}
	return func() { blockchain.ResetInstance() }
}

func TestNewListener(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer blockchain.CloseMiniRedis()
	cleanup := setupTestGenesis(t)
	defer cleanup()

	listenAddr := "localhost:0"
	quicCfg := NewQuicConfig()

	listener, err := NewListener(listenAddr, false, NewTLSConfig, quicCfg)
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
