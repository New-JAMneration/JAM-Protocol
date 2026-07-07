//go:build integration

package node_test

import (
	"os"
	"os/exec"
	"testing"
)

// TestSyncSmoke_dualNode is a placeholder for #567 dual-node QUIC E2E.
// Run manually when two nodes and chainspec are available:
//
//	go test -tags=integration ./internal/node/... -run TestSyncSmoke_dualNode -count=1
//
// Full harness lands in #567; this test documents the expected smoke command.
func TestSyncSmoke_dualNode(t *testing.T) {
	if os.Getenv("JAM_SYNC_SMOKE") == "" {
		t.Skip("set JAM_SYNC_SMOKE=1 to run dual-node sync smoke (see #567)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not in PATH: %v", err)
	}
	t.Log(`smoke: go run ./cmd/node --chain cmd/node/test_data/dev.chainspec.json --listen-addr 127.0.0.1:0 --role full`)
}
