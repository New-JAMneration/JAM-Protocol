package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/fuzzenv"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/timing"
	jamteststrace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
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

	pvmBackendFlag = &cli.StringFlag{
		Name:  "pvm-backend",
		Usage: "PVM execution backend for the fuzz server: interpreter or recompiler",
		Value: PVM.BackendInterpreter,
	}

	socketAddrArg = &cli.StringArg{
		Name:  "socket-addr",
		Value: "/tmp/jam_target.sock",
	}

	jsonFileArg = &cli.StringArg{
		Name: "json-file",
	}

	folderPathArg = &cli.StringArg{
		Name: "folder-path",
	}
)

const (
	envFuzz         = "JAM_FUZZ"
	envFuzzSpec     = "JAM_FUZZ_SPEC"
	envFuzzDataPath = "JAM_FUZZ_DATA_PATH"
	envFuzzSockPath = "JAM_FUZZ_SOCK_PATH"
	envFuzzLogLevel = "JAM_FUZZ_LOG_LEVEL"
	envPVMBackend   = "JAM_PVM_BACKEND"
)

var (
	GP_VERSION     string
	TARGET_VERSION string
)

var cmd = cli.Command{
	Name:        "new-jamneration-target",
	Usage:       "New-JAMneration Fuzz Target",
	Description: `New-JAMneration Fuzz Target`,
	Authors:     []any{"New JAMneration"},
	Version:     fmt.Sprintf("[GP Version]: %s, [Target Version]: %s", GP_VERSION, TARGET_VERSION),
	Action:      serve,
	ArgsUsage:   "<socket-addr>",
	Arguments: []cli.Argument{
		socketAddrArg,
	},
	Flags: []cli.Flag{
		configPathFlag,
		modeFlag,
		pvmBackendFlag,
	},
	Commands: []*cli.Command{
		handshakeCmd,
		importBlockCmd,
		setStateCmd,
		getStateCmd,
		testFolderCmd,
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

	testFolderCmd = &cli.Command{
		Name:      "test_folder",
		Action:    testFolder,
		ArgsUsage: "<socket-addr> <folder-path>",
		Arguments: []cli.Argument{
			socketAddrArg,
			folderPathArg,
		},
		// Read .bin instead of .json, e.g.:
		//   go run ./cmd/fuzz/ test_folder --format=bin <sock> <dir>
		// --skip N drops the first N fixtures so SetState runs on a later block
		// (e.g. skip leading fork variants of a mid-chain dataset).
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Usage: "Fixture format to read: json or binary (.bin)",
				Value: "json",
			},
			&cli.IntFlag{
				Name:  "skip",
				Usage: "Drop the first N fixtures before replaying (SetState on the new first)",
				Value: 0,
			},
		},
	}
)

func main() {
	cli.VersionPrinter = func(c *cli.Command) {
		logger.Infof("[GP Version]: %s, [Target Version]: %s", GP_VERSION, TARGET_VERSION)
	}

	if GP_VERSION == "" {
		// Read the VERSION_GP file to get the GP version
		data, err := os.ReadFile("VERSION_GP")
		if err != nil {
			logger.Fatalf("error reading GP version file: %v", err)
		}
		GP_VERSION = strings.TrimSpace(string(data))
	}

	if TARGET_VERSION == "" {
		// Read the VERSION_TARGET file to get the Target version
		data, err := os.ReadFile("VERSION_TARGET")
		if err != nil {
			logger.Fatalf("error reading Target version file: %v", err)
		}
		TARGET_VERSION = strings.TrimSpace(string(data))
	}

	config.UpdateVersion(GP_VERSION, TARGET_VERSION)

	if stopProfile := startProfiling(); stopProfile != nil {
		defer stopProfile()
	}
	defer flushPVMProfile()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := cmd.Run(ctx, os.Args); err != nil {
		logger.Fatalf("error: %v", err)
	}
}

func jamFuzzEnvEnabled() bool {
	return fuzzenv.Enabled()
}

func serve(ctx context.Context, cmd *cli.Command) error {
	if !jamFuzzEnvEnabled() {
		return fmt.Errorf("fuzz server requires %s plus %s, %s, %s (see fuzz-proto README / docker -e)",
			envFuzz, envFuzzSpec, envFuzzDataPath, envFuzzSockPath)
	}

	configPath := cmd.String(configPathFlag.Name)

	spec := strings.TrimSpace(os.Getenv(envFuzzSpec))
	if spec == "" {
		return errors.New(envFuzz + ": " + envFuzzSpec + " must be set (tiny or full)")
	}
	lowSpec := strings.ToLower(spec)
	if lowSpec != "tiny" && lowSpec != "full" {
		return fmt.Errorf("%s must be tiny or full, got %q", envFuzzSpec, spec)
	}
	mode := lowSpec

	dataPath := strings.TrimSpace(os.Getenv(envFuzzDataPath))
	if dataPath == "" {
		return errors.New(envFuzz + ": " + envFuzzDataPath + " must be set")
	}
	if err := os.MkdirAll(dataPath, 0o755); err != nil {
		return fmt.Errorf("failed to create fuzz data directory: %w", err)
	}

	config.InitConfig(configPath, mode)
	config.UpdateVersion(GP_VERSION, TARGET_VERSION)
	if err := applyPVMBackend(cmd); err != nil {
		return err
	}
	applyFuzzLogLevelOverride(strings.TrimSpace(os.Getenv(envFuzzLogLevel)))

	if timing.Enabled {
		fuzz.ResetImportBlockTimings()
		defer fuzz.PrintImportBlockTimingSummary()
	}

	socketAddr, err := fuzzServerSocketAddr(cmd)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(socketAddr), 0o755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	server, err := fuzz.NewFuzzServer("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	err = server.ListenAndServe(ctx)
	if err != nil {
		return fmt.Errorf("error running server: %w", err)
	}

	return nil
}

func applyPVMBackend(cmd *cli.Command) error {
	// --pvm-backend has a non-empty default ("interpreter"), which would otherwise
	// shadow JAM_PVM_BACKEND entirely (cmd.String never returns ""). Only treat the
	// flag as authoritative when it was passed explicitly; otherwise fall back to
	// the env var, then the default. Precedence: explicit flag > env var > default.
	backend := strings.TrimSpace(cmd.String(pvmBackendFlag.Name))
	if !cmd.IsSet(pvmBackendFlag.Name) {
		if env := strings.TrimSpace(os.Getenv(envPVMBackend)); env != "" {
			backend = env
		}
	}
	if backend == "" {
		backend = PVM.BackendInterpreter
	}

	switch backend {
	case PVM.BackendInterpreter:
		PVM.ExecutionBackend = PVM.BackendInterpreter
	case PVM.BackendRecompiler:
		if PVM.Psi_M_recompilerHook == nil {
			return fmt.Errorf("pvm-backend %q is not available in this build (requires linux/amd64 with cgo and recompiler linked)", backend)
		}
		PVM.ExecutionBackend = PVM.BackendRecompiler
	default:
		return fmt.Errorf("pvm-backend must be %q or %q, got %q",
			PVM.BackendInterpreter, PVM.BackendRecompiler, backend)
	}

	logger.Infof("PVM ExecutionBackend: %s", PVM.ExecutionBackend)
	return nil
}

func fuzzServerSocketAddr(cmd *cli.Command) (string, error) {
	arg := strings.TrimSpace(cmd.StringArg(socketAddrArg.Name))
	env := strings.TrimSpace(os.Getenv(envFuzzSockPath))

	if arg != "" && arg != "/tmp/jam_target.sock" {
		return arg, nil
	}
	if env != "" {
		return env, nil
	}
	if arg != "" {
		return arg, nil
	}
	return "", errors.New(envFuzz + ": socket path required via positional argument or " + envFuzzSockPath)
}

func applyFuzzLogLevelOverride(logLevel string) {
	if logLevel == "" {
		return
	}

	config.Config.Log.Level = logLevel
	logger.ConfigureLogger("main", logger.LoggerConfig{
		Level:      logLevel,
		Enabled:    config.Config.Log.Enabled,
		Color:      config.Config.Log.Color,
		TimeFormat: config.Config.Log.TimeFormat,
	})
	logger.ConfigureLogger("pvm", logger.LoggerConfig{
		Level:      logLevel,
		Enabled:    config.Config.Log.PVM,
		Color:      config.Config.Log.Color,
		TimeFormat: config.Config.Log.TimeFormat,
	})
}

func handshake(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("handshake requires a socket path argument")
	}

	configPath := cmd.String(configPathFlag.Name)
	mode := cmd.String(modeFlag.Name)

	config.InitConfig(configPath, mode)
	config.UpdateVersion(GP_VERSION, TARGET_VERSION)

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	var info fuzz.PeerInfo

	if err := info.FromConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	resp, err := client.Handshake(info)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	logger.Info("received handshake response:")
	logger.Infof("  fuzz-version: %d", resp.FuzzVersion)
	logger.Infof("  fuzz-features: %d", resp.FuzzFeatures)
	logger.Infof("  jam-version: %v", resp.JamVersion)
	logger.Infof("  app-version: %v", resp.AppVersion)
	logger.Infof("  app-name: %s", resp.AppName)

	return nil
}

func importBlock(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("import_block requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return errors.New("import_block requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		logger.Fatalf("error creating client: %v", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Parse JSON data into Block structure
	var block types.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	// Send import_block request
	// Use block's ParentStateRoot as priorStateRoot for protocol error fallback
	stateRoot, errorMessage, err := client.ImportBlock(block, block.Header.ParentStateRoot)
	if err != nil {
		return fmt.Errorf("error sending import_block request: %w", err)
	} else if errorMessage != nil {
		return fmt.Errorf("error sending import_block request: %v", errorMessage.Error)
	}

	logger.Infof("import_block successful, state root: 0x%x", stateRoot)

	return nil
}

func setState(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("set_state requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return errors.New("set_state requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Parse JSON data into header and state structures
	var requestData struct {
		Header   types.Header       `json:"header"`
		State    types.StateKeyVals `json:"state"`
		Ancestry types.Ancestry     `json:"ancestry"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	// Send set_state request
	stateRoot, err := client.SetState(requestData.Header, requestData.State, requestData.Ancestry)
	if err != nil {
		return fmt.Errorf("error sending set_state request: %w", err)
	}

	logger.Infof("set_state successful, state root: 0x%x", stateRoot)

	return nil
}

func getState(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("get_state requires a socket path argument")
	}
	jsonFile := cmd.StringArg(jsonFileArg.Name)
	if jsonFile == "" {
		return errors.New("get_state requires a json file path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}

	// Parse JSON data into header hash
	var requestData struct {
		HeaderHash string `json:"header_hash"`
	}

	if err := json.Unmarshal(data, &requestData); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
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
		return fmt.Errorf("error parsing header hash: %w", err)
	}

	if len(hashBytes) != 32 {
		return fmt.Errorf("header hash must be 32 bytes, got %d bytes", len(hashBytes))
	}

	copy(headerHash[:], hashBytes)

	// Send get_state request
	state, err := client.GetState(headerHash)
	if err != nil {
		return fmt.Errorf("error sending get_state request: %w", err)
	}

	logger.Infof("get_state successful, retrieved %d key-value pairs", len(state))

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
		return errors.New("test_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return errors.New("test_folder requires a json file path argument")
	}

	// Decoding fixtures (.bin or .json) needs the chainspec, so the client inits
	// config too. Default tiny; set JAM_FUZZ_SPEC=full for full-spec data.
	spec := strings.TrimSpace(os.Getenv(envFuzzSpec))
	if spec == "" {
		spec = "tiny"
	}
	spec = strings.ToLower(spec)
	if spec != "tiny" && spec != "full" {
		return fmt.Errorf("%s must be tiny or full, got %q", envFuzzSpec, spec)
	}
	config.InitConfig(cmd.String(configPathFlag.Name), spec)
	config.UpdateVersion(GP_VERSION, TARGET_VERSION)

	// Connect to server
	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	// Pick a single fixture format for the whole folder (never mixed).
	fixtureExt := ".json"
	if f := strings.ToLower(strings.TrimSpace(cmd.String("format"))); f == "bin" || f == "binary" {
		fixtureExt = ".bin"
	}
	// report.bin is fuzz-session metadata (FuzzerReport), not a fixture.
	isFixture := func(path string, d fs.DirEntry) bool {
		return !d.IsDir() && strings.HasSuffix(strings.ToLower(path), fixtureExt) && filepath.Base(path) != "report.bin"
	}

	// Count first so the slice is sized exactly — sessions can hold ~1000 files,
	// and a small fixed cap would re-allocate repeatedly on append.
	fileCount := 0
	if err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err == nil && isFixture(path, d) {
			fileCount++
		}
		return err
	}); err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Read all fixture files in the folder
	jsonFiles := make([]string, 0, fileCount)
	firstFiles := make(map[string]string)
	err = filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !isFixture(path, d) {
			return nil
		}
		jsonFiles = append(jsonFiles, path)

		folderName := strings.Split(path, "/")
		folderIndex := folderName[len(folderName)-2]
		fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		// 1. record each group of the first-data, genesis might not be the first file to be read
		if _, ok := firstFiles[folderIndex]; !ok || fileName == "genesis" {
			firstFiles[folderIndex] = fileName
		}

		// 2. re-order genesis, assume each test-data index is unique, genesis might not be the first data to be append in jsonFiles
		if fileName == "genesis" {
			index := findGroupFirstIndex(&jsonFiles, folderIndex) // find how many test-data in a group has appended in the jsonFiles
			copy(jsonFiles[index+1:], jsonFiles[index:len(jsonFiles)-1])
			jsonFiles[index] = path
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	if len(jsonFiles) == 0 {
		return errors.New("no JSON files found in the specified folder")
	}

	// --skip N drops the leading N fixtures; the new first per group then does
	// SetState. Rebuild firstFiles so setStateRequired tracks the new boundary.
	if skip := cmd.Int("skip"); skip > 0 {
		if skip >= len(jsonFiles) {
			return fmt.Errorf("--skip %d drops all %d fixtures", skip, len(jsonFiles))
		}
		jsonFiles = jsonFiles[skip:]
		firstFiles = make(map[string]string)
		for _, f := range jsonFiles {
			parts := strings.Split(f, "/")
			folderIndex := parts[len(parts)-2]
			if _, ok := firstFiles[folderIndex]; !ok {
				firstFiles[folderIndex] = strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
			}
		}
		logger.Infof("Skipped first %d fixtures", skip)
	}

	logger.Infof("Found %d JSON files to test", len(jsonFiles))

	successCount := 0
	failureCount := 0

	// Do the handshake
	info, err := client.Handshake(fuzz.PeerInfo{AppName: "Fuzz-Test"})
	if err != nil {
		return fmt.Errorf("error doing handshake: %w", err)
	}
	logger.Infof("handshake successful, fuzz-version: %d, fuzz-features: %d, jam-version: %v, app-version: %v, app-name: %s", info.FuzzVersion, info.FuzzFeatures, info.JamVersion, info.AppVersion, info.AppName)

	for _, jsonFile := range jsonFiles {
		var setStateRequired bool
		folderName := strings.Split(jsonFile, "/")
		fileName := strings.TrimSuffix(filepath.Base(jsonFile), filepath.Ext(jsonFile))
		folderIndex := folderName[len(folderName)-2]

		firstFileName, ok := firstFiles[folderIndex]
		if ok && firstFileName == fileName {
			setStateRequired = true
		}

		if err := testFixtureFile(client, jsonFile, setStateRequired); err != nil {
			logger.ColorRed("FAILED!!: %s - %v", jsonFile, err)
			failureCount++
		} else {
			logger.ColorGreen("PASSED: %s", jsonFile)
			successCount++
		}
	}

	return nil
}

// traceFixture is the format-agnostic input to the replay logic, populated from
// either a JSON fixture or a binary TraceTestCase.
type traceFixture struct {
	Block         types.Block
	PreStateRoot  types.StateRoot
	PostStateRoot types.StateRoot
	PostKeyVals   types.StateKeyVals
}

func testFixtureFile(client *fuzz.FuzzClient, path string, setStateRequired bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Binary: dispatch by filename — genesis.bin is a Genesis, NNNNNNNN.bin a TraceStep.
	if strings.HasSuffix(strings.ToLower(path), ".bin") {
		if filepath.Base(path) == "genesis.bin" {
			return testGenesisBin(client, path, data)
		}
		return testTraceBin(client, path, data, setStateRequired)
	}

	var probe struct {
		PreState json.RawMessage `json:"pre_state"`
		State    json.RawMessage `json:"state"`
	}

	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	if len(probe.PreState) > 0 {
		return testTraceFixture(client, path, data, setStateRequired)
	}

	if len(probe.State) > 0 {
		return testGenesisFixture(client, path, data)
	}

	return errors.New("unknown fixture format")
}

// testTraceFixture loads a JSON trace fixture and replays it.
func testTraceFixture(client *fuzz.FuzzClient, jsonFile string, data []byte, setStateRequired bool) error {
	var jsonData TestData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	expectedPreStateRoot, err := parseStateRoot(jsonData.PreState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing pre_state state_root: %w", err)
	}

	expectedPostStateRoot, err := parseStateRoot(jsonData.PostState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing post_state state_root: %w", err)
	}

	return replayTrace(client, jsonFile, traceFixture{
		Block:         jsonData.Block,
		PreStateRoot:  expectedPreStateRoot,
		PostStateRoot: expectedPostStateRoot,
		PostKeyVals:   jsonData.PostState.KeyVals,
	}, setStateRequired)
}

// testTraceBin loads a binary TraceStep (NNNNNNNN.bin) and replays it.
func testTraceBin(client *fuzz.FuzzClient, file string, data []byte, setStateRequired bool) error {
	var tc jamteststrace.TraceTestCase
	n, err := types.NewDecoder().DecodeWithConsumed(data, &tc)
	if err != nil {
		return fmt.Errorf("decode TraceTestCase: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("partial decode: consumed %d of %d bytes (spec/layout mismatch)", n, len(data))
	}

	return replayTrace(client, file, traceFixture{
		Block:         tc.Block,
		PreStateRoot:  tc.PreState.StateRoot,
		PostStateRoot: tc.PostState.StateRoot,
		PostKeyVals:   tc.PostState.KeyVals,
	}, setStateRequired)
}

// replayTrace drives a single trace fixture against the target: SetState for the
// first fixture in a group, otherwise ImportBlock + state-root compare.
func replayTrace(client *fuzz.FuzzClient, jsonFile string, testData traceFixture, setStateRequired bool) error {
	mismatchCount := 0
	logger.ColorBlue("File: %s", jsonFile)

	expectedPreStateRoot := testData.PreStateRoot
	expectedPostStateRoot := testData.PostStateRoot
	/*
		Step 1: Initialization (SetState) to the post_state
	*/
	// folder-wise: only when the data is the first data will do SetState
	if setStateRequired {
		decoder := types.NewDecoder()
		recentBlocks := &types.RecentBlocks{}
		recentBlocksKeyVal := types.StateKeyVal{}
		for _, kv := range testData.PostKeyVals {
			if len(kv.Key) > 0 && kv.Key[0] == 0x03 {
				recentBlocksKeyVal = kv
				// Decode the recent history value
				err := decoder.Decode(recentBlocksKeyVal.Value, recentBlocks)
				if err != nil {
					logger.Debugf("error decoding recent blocks: %v", err)
					// Service related key might be misleaded as recent blocks
					continue
				}
				break
			}
		}

		// Create ancestry from recent blocks
		ancestry := types.Ancestry{}
		for index, blockInfo := range recentBlocks.History {
			// Calculate the mock slot for each ancestry item
			mockSlot := testData.Block.Header.Slot - types.TimeSlot(len(recentBlocks.History)-index)

			ancestry = append(ancestry, types.AncestryItem{
				Slot:       mockSlot,
				HeaderHash: blockInfo.HeaderHash,
			})
		}

		// Print Sending SetState
		// NOTE: default ancestry is empty, cause the current test-data is not provided with ancestry
		logger.ColorGreen("[SetState][Request] state_root= 0x%x", expectedPostStateRoot)
		actualPostStateRoot, err := client.SetState(testData.Block.Header, testData.PostKeyVals, types.Ancestry{})
		logger.ColorYellow("[SetState][Response] state_root= 0x%x", actualPostStateRoot)
		if err != nil {
			return fmt.Errorf("SetState failed: %w", err)
		}

		if actualPostStateRoot != expectedPostStateRoot {
			logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x",
				expectedPostStateRoot, actualPostStateRoot)
			mismatchCount++
		}
		return nil
	}
	/*
		Step 2: ImportBlock
	*/
	// Print Sending ImportBlock
	headerHash, err := hash.ComputeBlockHeaderHash(testData.Block.Header)
	if err != nil {
		return fmt.Errorf("error computing header hash: %w", err)
	}
	logger.ColorGreen("[ImportBlock][Request] header_hash= 0x%x...", headerHash[:8])

	// Print ImportBlock Response
	actualPostStateRoot, errorMessage, err := client.ImportBlock(testData.Block, expectedPreStateRoot)
	if err != nil {
		logger.ColorYellow("[ImportBlock][Response] error= %v", err)
		return err
	} else if errorMessage != nil {
		logger.ColorYellow("[ImportBlock][Response] error message= %v", errorMessage.Error)
	} else {
		logger.ColorYellow("[ImportBlock][Response] state_root= 0x%x", actualPostStateRoot)
	}

	mismatch, err := compareImportBlockState(importBlockCompareInput{
		Client:            client,
		Block:             testData.Block,
		BlockHeaderHash:   types.HeaderHash(headerHash),
		ExpectedStateRoot: expectedPostStateRoot,
		ActualStateRoot:   actualPostStateRoot,
		ExpectedPostState: testData.PostKeyVals,
	})
	if err != nil {
		return err
	}
	if mismatch {
		mismatchCount++
	}

	if mismatchCount > 0 {
		return fmt.Errorf("mismatch count: %d", mismatchCount)
	}

	return nil
}

// testGenesisFixture loads a JSON genesis fixture and replays it (SetState).
func testGenesisFixture(client *fuzz.FuzzClient, jsonFile string, data []byte) error {
	var genesisData struct {
		Header types.Header `json:"header"`
		State  struct {
			StateRoot string             `json:"state_root"`
			KeyVals   types.StateKeyVals `json:"keyvals"`
		} `json:"state"`
	}

	if err := json.Unmarshal(data, &genesisData); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	expectedStateRoot, err := parseStateRoot(genesisData.State.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing state state_root: %w", err)
	}

	return replayGenesis(client, jsonFile, genesisData.Header, genesisData.State.KeyVals, expectedStateRoot)
}

// testGenesisBin loads a binary Genesis (genesis.bin) and replays it (SetState).
func testGenesisBin(client *fuzz.FuzzClient, file string, data []byte) error {
	var genesisData jamteststrace.Genesis
	n, err := types.NewDecoder().DecodeWithConsumed(data, &genesisData)
	if err != nil {
		return fmt.Errorf("decode Genesis: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("partial decode: consumed %d of %d bytes (spec/layout mismatch)", n, len(data))
	}

	return replayGenesis(client, file, genesisData.Header, genesisData.State.KeyVals, genesisData.State.StateRoot)
}

// replayGenesis sets the genesis state on the target and checks the state root.
func replayGenesis(client *fuzz.FuzzClient, jsonFile string, header types.Header, keyVals types.StateKeyVals, expectedStateRoot types.StateRoot) error {
	logger.ColorBlue("File: %s", jsonFile)

	headerHash, err := hash.ComputeBlockHeaderHash(header)
	if err != nil {
		return fmt.Errorf("error computing header hash: %w", err)
	}
	logger.ColorGreen("[SetState][Request] genesis header_hash= 0x%x and state_root= 0x%x", headerHash[:8], expectedStateRoot)
	// Genesis state does not have ancestry
	actualStateRoot, err := client.SetState(header, keyVals, types.Ancestry{})
	logger.ColorYellow("[SetState][Response] genesis state_root= 0x%x", actualStateRoot)
	if err != nil {
		return fmt.Errorf("SetState failed: %w", err)
	}

	if actualStateRoot != expectedStateRoot {
		logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x", expectedStateRoot, actualStateRoot)
		return errors.New("state_root mismatch")
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

func findGroupFirstIndex(jsonFiles *[]string, groupIndex string) int {
	var firstFileIndex int

	for i := len(*jsonFiles) - 1; i >= 0; i-- {
		folderName := strings.Split((*jsonFiles)[i], "/")
		folderIndex := folderName[len(folderName)-2]
		if folderIndex != groupIndex {
			firstFileIndex = i + 1
			break
		}
	}
	return firstFileIndex
}
