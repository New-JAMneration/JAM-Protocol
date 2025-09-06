package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func printUsage() {
	usage := `Usage: go run cmd/fuzz/fuzz.go COMMAND [ARGS...]

Valid commands are:
  serve FILE
  handshake FILE
  import_block SOCKET JSON_FILE
  set_state SOCKET JSON_FILE
  get_state SOCKET JSON_FILE
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
	log.Printf("  Name: %s\n", resp.Name)
	log.Printf("  App Version: %v\n", resp.AppVersion)
	log.Printf("  JAM Version: %v\n", resp.JamVersion)
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
	stateRoot, err := client.ImportBlock(block)
	if err != nil {
		log.Fatalf("error sending import_block request: %v\n", err)
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
	default:
		printUsage()
	}
}
