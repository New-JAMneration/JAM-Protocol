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
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/New-JAMneration/JAM-Protocol/logger"
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
		testFolderCmd,
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

	testFolderCmd = &cli.Command{
		Name:      "test_folder",
		Action:    testFolder,
		ArgsUsage: "<socket-addr> <folder-path>",
		Arguments: []cli.Argument{
			socketAddrArg,
			folderPathArg,
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

func testFolder(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("test_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return fmt.Errorf("test_folder requires a json file path argument")
	}

	// Connect to server
	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
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
		return fmt.Errorf("error walking directory: %v", err)
	}

	if len(jsonFiles) == 0 {
		return fmt.Errorf("no JSON files found in the specified folder")
	}

	log.Printf("Found %d JSON files to test\n", len(jsonFiles))

	successCount := 0
	failureCount := 0

	// Do the handshake
	info, err := client.Handshake(fuzz.PeerInfo{AppName: "Fuzz-Test"})
	if err != nil {
		return fmt.Errorf("error doing handshake: %v", err)
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

	return nil
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
	fmt.Println("File: ", jsonFile)
	// Step 1: SetState with pre_state data
	expectedPreStateRoot, err := parseStateRoot(testData.PreState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing pre_state state_root: %v", err)
	}

	// Print Sending SetState
	logger.ColorGreen("[SetState][Request] block_header_hash= 0x%v", hex.EncodeToString(testData.Block.Header.Parent[:]))
	actualPreStateRoot, err := client.SetState(testData.Block.Header, testData.PreState.KeyVals)
	logger.ColorYellow("[SetState][Response] state_root= 0x%v", hex.EncodeToString(actualPreStateRoot[:]))
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
	logger.ColorGreen("[ImportBlock][Request] block_header_hash= 0x%x", blockHeaderHashHex)

	// Print ImportBlock Response
	actualPostStateRoot, errorMessage, err := client.ImportBlock(testData.Block)
	if err != nil {
		logger.ColorYellow("[ImportBlock][Response] error= %v", err)
		return err
	} else if errorMessage != nil {
		logger.ColorYellow("[ImportBlock][Response] error message= %v", errorMessage.Error)
		return err
	} else {
		logger.ColorYellow("[ImportBlock][Response] state_root= 0x%v", hex.EncodeToString(actualPostStateRoot[:]))
	}

	if actualPostStateRoot != expectedPostStateRoot {
		logger.ColorBlue("[ImportBlock][Check] state_root mismatch: expected 0x%x, got 0x%x",
			expectedPostStateRoot, actualPostStateRoot)
		mismatchCount++
		importBlockMismatch = true
	}

	if importBlockMismatch {
		// request GetState
		logger.ColorGreen("[GetState][Request] header_hash= 0x%v", hex.EncodeToString(blockHeaderHashHex[:]))
		actualStateKeyVal, err := client.GetState(types.HeaderHash(blockHeaderHashHex))
		if err != nil {
			return fmt.Errorf("error sending get_state request: %v", err)
		}

		// actual state struct, key-val to state
		diffs, err := merklization.GetStateKeyValsDiff(testData.PostState.KeyVals, actualStateKeyVal)
		if err != nil {
			return fmt.Errorf("fuzzer GetStateKeyValsDiff error: %v", err)
		}
		// stateRoot := merklization.MerklizationSerializedState(actualPostState)
		logger.ColorYellow("[GetState][Response] %d different key-val: ", len(diffs))
		for _, v := range diffs {
			if state, keyExists := jamtests.KeyValMap[v.Key]; keyExists {
				logger.ColorYellow("state: %s, key: %v", state, v.Key)
				if len(v.ActualValue) > 600 {
					logger.ColorDebug("value too big, only check diff")
					/*
						actualStateStruct, err := m.SingleKeyValToState(v.Key, v.ActualValue)
						if err != nil {
							return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
						}
						expectStateStruct, err := m.SingleKeyValToState(v.Key, v.ExpectedValue)
						if err != nil {
							return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
						}

						diff := cmp.Diff(actualStateStruct, expectStateStruct)
						logger.ColorDebug("diffs: %+v", diff)
					*/
				} else {
					logger.ColorDebug("actualVal: %+v", v.ActualValue)
					logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
					/*
						// detailed state struct, except service-related
						actualStateStruct, err := m.SingleKeyValToState(v.Key, v.ActualValue)
						if err != nil {
							return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
						}
						expectStateStruct, err := m.SingleKeyValToState(v.Key, v.ExpectedValue)
						if err != nil {
							return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
						}

						logger.ColorDebug("actualState: %+v", actualStateStruct)
						logger.ColorDebug("actualState: %+v", expectStateStruct)

						diff := cmp.Diff(actualStateStruct, expectStateStruct)
						logger.ColorDebug("diffs: %+v", diff)
					*/
				}
			} else if v.Key[0] == byte(255) {
				serviceId, err := merklization.DecodeServiceIdFromType2(v.Key)
				if err != nil {
					return fmt.Errorf("fuzzer DecodeServiceIdFromType2 error: %v", err)
				}
				logger.ColorYellow("service: %d", serviceId)
				logger.ColorDebug("actualVal: %+v", v.ActualValue)
				logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
				/*
					actualStateStruct, err := m.SingleKeyValToState(v.Key, v.ActualValue)
					if err != nil {
						return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
					}
					expectStateStruct, err := m.SingleKeyValToState(v.Key, v.ExpectedValue)
					if err != nil {
						return fmt.Errorf("GetState SingleKeyValToState failed: %v", err)
					}

					logger.ColorDebug("actualStruct: %+v", actualStateStruct)
					logger.ColorDebug("expectStruct: %+v", expectStateStruct)
					diff := cmp.Diff(actualStateStruct, expectStateStruct)
					logger.ColorDebug("diffs: %+v", diff)
				*/
			} else {
				logger.ColorYellow("other state key: 0x%x", v.Key)
				if len(v.ActualValue) > 1024 || len(v.ExpectedValue) > 1024 {
					logger.ColorDebug("value too big, check json file")
				} else {
					logger.ColorDebug("actualVal: %+v", v.ActualValue)
					logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
				}
			}
		}
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
