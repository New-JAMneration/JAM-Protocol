package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

func printUsage() {
	usage := `Usage: go run cmd/fuzz/fuzz.go COMMAND [ARGS...]

Valid commands are:
  serve FILE
  handshake FILE
  import_block SOCKET JSON_FILE
  set_state SOCKET JSON_FILE
  get_state SOCKET JSON_FILE
  test_folder SOCKET FOLDER_PATH
  help [COMMAND]`

	log.Fatalln(usage)
}

type Handler func([]string)

var handlers = map[string]Handler{
	"serve":        serve,
	"handshake":    handshake,
	"import_block": importBlock,
	"set_state":    setState,
	"get_state":    getState,
	"test_folder":  testFolder,
	"help":         help,
}

func main() {
	if len(os.Args) == 1 {
		printUsage()
	}

	handler, valid := handlers[os.Args[1]]
	if !valid {
		printUsage()
	}

	config.InitConfig()

	handler(os.Args[2:])
}

func serve(args []string) {
	if len(args) == 0 {
		helpImpl("serve")
	}

	server, err := fuzz.NewFuzzServer("unix", args[0])
	if err != nil {
		log.Fatalf("error creating server: %v\n", err)
	}

	server.ListenAndServe(context.Background())
}

func handshake(args []string) {
	if len(args) == 0 {
		helpImpl("handshake")
	}

	client, err := fuzz.NewFuzzClient("unix", args[0])
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	var info fuzz.PeerInfo

	if err := info.FromConfig(); err != nil {
		log.Fatalf("error reading config: %v\n", err)
	}

	resp, err := client.Handshake(info)
	if err != nil {
		log.Fatalf("error sending request: %v\n", err)
	}

	log.Println("received handshake response:")
	log.Printf("  fuzz-version: %d\n", resp.FuzzVersion)
	log.Printf("  fuzz-features: %d\n", resp.FuzzFeatures)
	log.Printf("  jam-version: %v\n", resp.JamVersion)
	log.Printf("  app-version: %v\n", resp.AppVersion)
	log.Printf("  app-name: %s\n", resp.AppName)
}

func importBlock(args []string) {
	if len(args) == 0 {
		helpImpl("import_block")
	}

	client, err := fuzz.NewFuzzClient("unix", args[0])
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing block data
	if len(args) < 2 {
		log.Fatalln("import_block requires a JSON file path as second argument")
	}

	jsonFile := args[1]
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into Block structure
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		log.Fatalf("error parsing JSON: %v\n", err)
	}

	// Send import_block request
	stateRoot, errorMessage, err := client.ImportBlock(block)
	if err != nil {
		log.Fatalf("error sending import_block request: %v\n", err)
	} else if errorMessage != nil {
		log.Fatalf("error sending import_block request: %v\n", errorMessage.Error)
	}

	log.Printf("import_block successful, state root: %x\n", stateRoot)
}

func setState(args []string) {
	if len(args) == 0 {
		helpImpl("set_state")
	}

	client, err := fuzz.NewFuzzClient("unix", args[0])
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing header and state data
	if len(args) < 2 {
		log.Fatalln("set_state requires a JSON file path as second argument")
	}

	jsonFile := args[1]
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into header and state structures
	var requestData struct {
		Header types.Header       `json:"header"`
		State  types.StateKeyVals `json:"state"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		log.Fatalf("error parsing JSON: %v\n", err)
	}

	// Send set_state request
	stateRoot, err := client.SetState(requestData.Header, requestData.State)
	if err != nil {
		log.Fatalf("error sending set_state request: %v\n", err)
	}

	log.Printf("set_state successful, state root: %x\n", stateRoot)
}

func getState(args []string) {
	if len(args) == 0 {
		helpImpl("get_state")
	}

	client, err := fuzz.NewFuzzClient("unix", args[0])
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing header hash
	if len(args) < 2 {
		log.Fatalln("get_state requires a JSON file path as second argument")
	}

	jsonFile := args[1]
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("error reading JSON file: %v\n", err)
	}

	// Parse JSON data into header hash
	var requestData struct {
		HeaderHash string `json:"header_hash"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		log.Fatalf("error parsing JSON: %v\n", err)
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
		log.Fatalf("error parsing header hash: %v\n", err)
	}

	if len(hashBytes) != 32 {
		log.Fatalf("header hash must be 32 bytes, got %d bytes\n", len(hashBytes))
	}

	copy(headerHash[:], hashBytes)

	// Send get_state request
	state, err := client.GetState(headerHash)
	if err != nil {
		log.Fatalf("error sending get_state request: %v\n", err)
	}

	log.Printf("get_state successful, retrieved %d key-value pairs\n", len(state))
}

// TestData represents the structure of test JSON files
type TestData struct {
	PreState struct {
		StateRoot string             `json:"state_root"`
		KeyVals   types.StateKeyVals `json:"keyvals"`
	} `json:"pre_state"`
	PostState struct {
		StateRoot string             `json:"state_root"`
		KeyVals   types.StateKeyVals `json:"keyvals"`
	} `json:"post_state"`
	Block types.Block `json:"block"`
}

func testFolder(args []string) {
	if len(args) < 2 {
		helpImpl("test_folder")
	}

	socketPath := args[0]
	folderPath := args[1]

	// Connect to server
	client, err := fuzz.NewFuzzClient("unix", socketPath)
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}
	defer client.Close()

	// Read all JSON files in the folder
	var jsonFiles []string
	err = filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			jsonFiles = append(jsonFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error walking directory: %v\n", err)
	}

	if len(jsonFiles) == 0 {
		log.Fatalln("no JSON files found in the specified folder")
	}

	log.Printf("Found %d JSON files to test\n", len(jsonFiles))

	successCount := 0
	failureCount := 0

	// Do the handshake
	info, err := client.Handshake(fuzz.PeerInfo{AppName: "Fuzz-Test"})
	if err != nil {
		log.Fatalf("error doing handshake: %v\n", err)
	}
	log.Printf("handshake successful, fuzz-version: %d, fuzz-features: %d, jam-version: %v, app-version: %v, app-name: %s\n", info.FuzzVersion, info.FuzzFeatures, info.JamVersion, info.AppVersion, info.AppName)

	// Process each JSON file
	for _, jsonFile := range jsonFiles {
		if err := testSingleFile(client, jsonFile); err != nil {
			logger.ColorRed("FAILED!!: %s - %v", jsonFile, err)
			failureCount++
		} else {
			logger.ColorGreen("PASSED: %s", jsonFile)
			successCount++
		}
	}
}

func testSingleFile(client *fuzz.FuzzClient, jsonFile string) error {
	mismatchCount := 0
	importBlockMismatch := false
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}

	var testData TestData
	if err := json.Unmarshal(data, &testData); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// Step 1: SetState with pre_state data
	expectedPreStateRoot, err := parseStateRoot(testData.PreState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing pre_state state_root: %v", err)
	}

	// Print Sending SetState
	logger.ColorGreen("[SetState][Request] block_header_hash=0x%v", hex.EncodeToString(testData.Block.Header.Parent[:]))
	actualPreStateRoot, err := client.SetState(testData.Block.Header, testData.PreState.KeyVals)
	logger.ColorYellow("[SetState][Response] state_root=0x%v", hex.EncodeToString(actualPreStateRoot[:]))
	if err != nil {
		return fmt.Errorf("SetState failed: %v", err)
	}

	if actualPreStateRoot != expectedPreStateRoot {
		logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x",
			expectedPreStateRoot, actualPreStateRoot)
		mismatchCount++
	}

	// Step 2: ImportBlock
	expectedPostStateRoot, err := parseStateRoot(testData.PostState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing post_state state_root: %v", err)
	}

	// Print Sending ImportBlock
	serializedHeader, err := utilities.HeaderSerialization(testData.Block.Header)
	if err != nil {
		return fmt.Errorf("error serializing header: %v", err)
	}
	blockHeaderHashHex := hash.Blake2bHash(serializedHeader)
	logger.ColorGreen("[ImportBlock][Request] block_header_hash=0x%x", blockHeaderHashHex)

	// Print ImportBlock Response
	actualPostStateRoot, errorMessage, err := client.ImportBlock(testData.Block)
	if err != nil {
		logger.ColorYellow("[ImportBlock][Response] error=%v", err)
	} else if errorMessage != nil {
		logger.ColorYellow("[ImportBlock][Response] error message=%v", errorMessage.Error)
	} else {
		logger.ColorYellow("[ImportBlock][Response] state_root=0x%v", hex.EncodeToString(actualPostStateRoot[:]))
	}

	if actualPostStateRoot != expectedPostStateRoot {
		logger.ColorBlue("[ImportBlock][Check] state_root mismatch: expected 0x%x, got 0x%x",
			expectedPostStateRoot, actualPostStateRoot)
		mismatchCount++
		importBlockMismatch = true
	}

	if importBlockMismatch {
		// request GetState
		logger.ColorGreen("[GetState][Request] header_hash=0x%v", blockHeaderHashHex)
		actualPostState, err := client.GetState(types.HeaderHash(blockHeaderHashHex))
		stateRoot := merklization.MerklizationSerializedState(actualPostState)
		if err != nil {
			return fmt.Errorf("error sending get_state request: %v", err)
		}
		logger.ColorYellow("[GetState][Response] state_root=0x%v", hex.EncodeToString(stateRoot[:]))
	}

	if mismatchCount > 0 {
		return fmt.Errorf("mismatch count: %d", mismatchCount)
	}

	return nil
}

func parseStateRoot(stateRootStr string) (types.StateRoot, error) {
	// Remove 0x prefix if present
	if len(stateRootStr) > 2 && stateRootStr[:2] == "0x" {
		stateRootStr = stateRootStr[2:]
	}

	var stateRoot types.StateRoot
	hashBytes, err := hex.DecodeString(stateRootStr)
	if err != nil {
		return types.StateRoot{}, err
	}

	if len(hashBytes) != 32 {
		return types.StateRoot{}, fmt.Errorf("state root must be 32 bytes, got %d bytes", len(hashBytes))
	}

	copy(stateRoot[:], hashBytes)
	return stateRoot, nil
}

func help(args []string) {
	helpImpl(args...)
}

func helpImpl(args ...string) {
	if len(args) == 0 {
		printUsage()
	}

	switch args[0] {
	case "serve":
		log.Fatalln("serve FILE - starts a server listening on FILE via named Unix socket")
	case "handshake":
		log.Fatalln("handshake FILE - connects to a server listening on FILE and sends a handshake")
	case "import_block":
		log.Fatalln("import_block SOCKET JSON_FILE - connects to a server listening on SOCKET and sends an import_block request with block data from JSON_FILE")
	case "set_state":
		log.Fatalln("set_state SOCKET JSON_FILE - connects to a server listening on SOCKET and sends a set_state request with header and state data from JSON_FILE")
	case "get_state":
		log.Fatalln("get_state SOCKET JSON_FILE - connects to a server listening on SOCKET and sends a get_state request with header hash from JSON_FILE")
	case "test_folder":
		log.Fatalln("test_folder SOCKET FOLDER_PATH - tests all JSON files in FOLDER_PATH by running SetState and ImportBlock operations")
	default:
		printUsage()
	}
}
