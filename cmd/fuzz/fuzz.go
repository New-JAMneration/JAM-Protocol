package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func printUsage() {
	usage := `Usage: go run cmd/fuzz/fuzz.go COMMAND [ARGS...]

Valid commands are:
  serve ADDRESS
  handshake ADDRESS
  import_block ADDRESS JSON_FILE
  set_state ADDRESS JSON_FILE
  get_state ADDRESS JSON_FILE
  help [COMMAND]

Where:
  ADDRESS - can be a "host:port" address (must contains a ':' separator) 
    or a UNIX domain socket (must not contain a ':' character).

	examples:
	  - :8080
	  - localhost:8080
	  - 4.2.2.2:8080
	  - /tmp/socket
  JSON_FILE - path to a JSON file containing block, header, and/or state data.
`

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

	server := fuzz.NewFuzzServer()
	network, address := splitAddress(args[0])
	err := server.ListenAndServe(context.Background(), network, address)
	if err != nil {
		log.Fatalf("error while serving: %v", err)
	}
}

func handshake(args []string) {
	if len(args) == 0 {
		helpImpl("handshake")
	}

	network, address := splitAddress(args[0])
	client, err := fuzz.NewFuzzClient(network, address)
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
	if len(args) < 2 {
		helpImpl("import_block")
	}

	network, address := splitAddress(args[0])
	client, err := fuzz.NewFuzzClient(network, address)
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing block data
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
	if len(args) < 2 {
		helpImpl("set_state")
	}

	network, address := splitAddress(args[0])
	client, err := fuzz.NewFuzzClient(network, address)
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing header and state data
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
	if len(args) < 2 {
		helpImpl("get_state")
	}

	network, address := splitAddress(args[0])
	client, err := fuzz.NewFuzzClient(network, address)
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}

	defer client.Close()

	// Read JSON file containing header hash
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

	var message string

	switch args[0] {
	case "serve":
		message = `serve ADDRESS - starts a server listening on an address
  where ADDRESS can be a network address like ":port", "localhost:port", "host:port", or a UNIX domain socket`
	case "handshake":
		message = `handshake ADDRESS - connects to a server listening on ADDRESS and sends a handshake
  where ADDRESS can be a network address like ":port", "localhost:port", "host:port", or a UNIX domain socket`
	case "import_block":
		message = `import_block ADDRESS JSON_FILE - connects to a server listening on ADDRESS and sends an import_block request with block data from JSON_FILE
  where ADDRESS can be a network address like ":port", "localhost:port", "host:port", or a UNIX domain socket
    and JSON_FILE is the path to a JSON file containing block data`
	case "set_state":
		message = `set_state ADDRESS JSON_FILE - connects to a server listening on ADDRESS and sends a set_state request with header and state data from JSON_FILE
  where ADDRESS can be a network address like ":port", "localhost:port", "host:port", or a UNIX domain socket
    and JSON_FILE is the path to a JSON file containing header and state data`
	case "get_state":
		message = `get_state ADDRESS JSON_FILE - connects to a server listening on ADDRESS and sends a get_state request with header hash from JSON_FILE
  where ADDRESS can be a network address like ":port", "localhost:port", "host:port", or a UNIX domain socket
    AND JSON_FILE is the path to a JSON file containing a header hash`
	default:
		printUsage()
	}

	log.Fatalln(message)
}

func splitAddress(address string) (string, string) {
	if strings.ContainsRune(address, ':') {
		return "tcp", address
	}

	return "unix", address
}
