package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/urfave/cli/v3"
)

var (
	modeFlag = &cli.StringFlag{
		Name:  "mode",
		Usage: "Node mode: tiny or full or custom",
		Value: "tiny",
	}

	configPathFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "Path to configuration file",
		Value: "example.json",
	}

	socketAddrArg = &cli.StringArg{
		Name: "socket-addr",
	}

	jsonFileArg = &cli.StringArg{
		Name: "json-file",
	}

	folderPathArg = &cli.StringArg{
		Name: "folder-path",
	}

	folderWiseArg = &cli.BoolFlag{
		Name:  "folderwise",
		Usage: "SetState once(true) or SetState each block(false)",
		Value: false,
	}
)

var cmd = cli.Command{
	Name:        "fuzz",
	Usage:       "JAM Fuzz Testing Server",
	Description: `JAM Fuzz Testing Server`,
	Authors:     []any{"New JAMneration"},
	Action:      serve,
	ArgsUsage:   "<socket-addr>",
	Arguments: []cli.Argument{
		socketAddrArg,
	},
	Flags: []cli.Flag{
		configPathFlag,
		modeFlag,
	},
	Commands: []*cli.Command{
		handshakeCmd,
		importBlockCmd,
		setStateCmd,
		getStateCmd,
		testTraceFolderCmd,
		testStepFolderCmd,
	},
}

var (
	handshakeCmd = &cli.Command{
		Name:        "handshake",
		Usage:       "Fuzz peer handshake",
		Description: "Fuzz peer handshake",
		Action:      handshake,
		ArgsUsage:   "<socket-addr>",
		Arguments: []cli.Argument{
			socketAddrArg,
		},
	}

	importBlockCmd = &cli.Command{
		Name:        "import_block",
		Usage:       "Fuzz import block",
		Description: "Fuzz import block",
		Action:      importBlock,
		ArgsUsage:   "<socket-addr> <json-file>",
		Arguments: []cli.Argument{
			socketAddrArg,
			jsonFileArg,
		},
	}

	setStateCmd = &cli.Command{
		Name:        "set_state",
		Usage:       "Fuzz set state",
		Description: "Fuzz set state",
		Action:      setState,
		ArgsUsage:   "<socket-addr> <json-file>",
		Arguments: []cli.Argument{
			socketAddrArg,
			jsonFileArg,
		},
	}

	getStateCmd = &cli.Command{
		Name:        "get_state",
		Usage:       "Fuzz get state",
		Description: "Fuzz get state",
		Action:      getState,
		ArgsUsage:   "<socket-addr> <json-file>",
		Arguments: []cli.Argument{
			socketAddrArg,
			jsonFileArg,
		},
	}
)

func main() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("serve requires a socket path argument")
	}

	configPath := cmd.String(configPathFlag.Name)
	mode := cmd.String(modeFlag.Name)

	config.InitConfig(configPath, mode)

	server, err := fuzz.NewFuzzServer("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating server: %v", err)
	}

	err = server.ListenAndServe(ctx)
	if err != nil {
		return fmt.Errorf("error running server: %v", err)
	}

	return nil
}

func handshake(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("handshake requires a socket path argument")
	}

	configPath := cmd.String(configPathFlag.Name)
	mode := cmd.String(modeFlag.Name)

	config.InitConfig(configPath, mode)

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	var info fuzz.PeerInfo

	if err := info.FromConfig(); err != nil {
		return fmt.Errorf("error reading config: %v", err)
	}

	resp, err := client.Handshake(info)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}

	log.Println("received handshake response:")
	log.Printf("  fuzz-version: %d\n", resp.FuzzVersion)
	log.Printf("  fuzz-features: %d\n", resp.FuzzFeatures)
	log.Printf("  jam-version: %v\n", resp.JamVersion)
	log.Printf("  app-version: %v\n", resp.AppVersion)
	log.Printf("  app-name: %s\n", resp.AppName)

	return nil
}

func importBlock(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("import_block requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return fmt.Errorf("import_block requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		log.Fatalf("error creating client: %v\n", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Parse JSON data into Block structure
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// Send import_block request
	stateRoot, errorMessage, err := client.ImportBlock(block)
	if err != nil {
		return fmt.Errorf("error sending import_block request: %v", err)
	} else if errorMessage != nil {
		return fmt.Errorf("error sending import_block request: %v", errorMessage.Error)
	}

	log.Printf("import_block successful, state root: %x\n", stateRoot)

	return nil
}

func setState(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("set_state requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return fmt.Errorf("set_state requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}

	// Parse JSON data into header and state structures
	var requestData struct {
		Header types.Header       `json:"header"`
		State  types.StateKeyVals `json:"state"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// Send set_state request
	stateRoot, err := client.SetState(requestData.Header, requestData.State)
	if err != nil {
		return fmt.Errorf("error sending set_state request: %v", err)
	}

	log.Printf("set_state successful, state root: %x\n", stateRoot)

	return nil
}

func getState(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("get_state requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return fmt.Errorf("get_state requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}

	// Parse JSON data into header hash
	var requestData struct {
		HeaderHash string `json:"header_hash"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
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
		return fmt.Errorf("error parsing header hash: %v", err)
	}

	if len(hashBytes) != 32 {
		return fmt.Errorf("header hash must be 32 bytes, got %d bytes", len(hashBytes))
	}

	copy(headerHash[:], hashBytes)

	// Send get_state request
	state, err := client.GetState(headerHash)
	if err != nil {
		return fmt.Errorf("error sending get_state request: %v", err)
	}

	log.Printf("get_state successful, retrieved %d key-value pairs\n", len(state))

	return nil
}
