package main_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const FUZZ_TEST_DATA_DIR = "./test_data"

func TestFuzzTestSuite(t *testing.T) {
	os.Args = []string{os.Args[0], "-config", "../../example.json"}
	config.InitConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverConn, clientConn := net.Pipe()

	server := fuzz.NewFuzzServer()
	go server.Serve(ctx, serverConn)

	client := fuzz.NewFuzzClientInternal(clientConn)
	defer client.Close()

	sendHandshake(t, client)
	sendImportBlock(t, client, "example_import_block.json")
	sendGetState(t, client, "example_get_state.json")
}

func sendHandshake(t *testing.T, client fuzz.FuzzService) {
	var info fuzz.PeerInfo

	if err := info.FromConfig(); err != nil {
		t.Errorf("error reading config: %v\n", err)
	}

	t.Log("Sending handshake")
	resp, err := client.Handshake(info)
	if err != nil {
		t.Errorf("error sending request: %v\n", err)
	}

	t.Log("Received handshake response:")
	t.Logf("  Name: %s\n", resp.Name)
	t.Logf("  App Version: %v\n", resp.AppVersion)
	t.Logf("  JAM Version: %v\n", resp.JamVersion)
}

func sendImportBlock(t *testing.T, client fuzz.FuzzService, filename string) {
	// Read JSON file containing block data
	jsonFile := filepath.Join(FUZZ_TEST_DATA_DIR, filename)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Errorf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into Block structure
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		t.Errorf("error parsing JSON: %v\n", err)
	}

	// Send ImportBlock request
	stateRoot, err := client.ImportBlock(block)
	if err != nil {
		t.Errorf("error sending import_block request: %v\n", err)
	}

	t.Logf("importBlock successful, state root: %x\n", stateRoot)
}

/*
func sendSetState(t *testing.T, client fuzz.FuzzService, filename string) {
	// Read JSON file containing header and state data
	jsonFile := filepath.Join(FUZZ_TEST_DATA_DIR, filename)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Errorf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into header and state structures
	var requestData struct {
		Header types.Header       `json:"header"`
		State  types.StateKeyVals `json:"state"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		t.Errorf("error parsing JSON: %v\n", err)
	}

	// Send SetState request
	stateRoot, err := client.SetState(requestData.Header, requestData.State)
	if err != nil {
		t.Errorf("error sending set_state request: %v\n", err)
	}

	t.Logf("set_state successful, state root: %x\n", stateRoot)
}
*/

func sendGetState(t *testing.T, client fuzz.FuzzService, filename string) {
	// Read JSON file containing header hash
	jsonFile := filepath.Join(FUZZ_TEST_DATA_DIR, filename)
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Errorf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into header hash
	var requestData struct {
		HeaderHash string `json:"header_hash"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		t.Errorf("error parsing JSON: %v\n", err)
	}

	// Parse header hash from hex string
	headerHashStr := requestData.HeaderHash

	// Remove 0x prefix if present
	if len(headerHashStr) > 2 && headerHashStr[:2] == "0x" {
		headerHashStr = headerHashStr[2:]
	}

	var headerHash types.HeaderHash
	hashBytes, err := hex.DecodeString(headerHashStr)
	if err != nil {
		t.Errorf("error parsing header hash: %v\n", err)
	}

	if len(hashBytes) != 32 {
		t.Errorf("header hash must be 32 bytes, got %d bytes\n", len(hashBytes))
	}

	copy(headerHash[:], hashBytes)

	// Send get_state request
	state, err := client.GetState(headerHash)
	if err != nil {
		t.Errorf("error sending get_state request: %v\n", err)
	}

	t.Logf("get_state successful, retrieved %d key-value pairs\n", len(state))
}
