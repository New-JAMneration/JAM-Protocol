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
	"sort"
	"strconv"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/urfave/cli/v3"
)

var testStepFolderCmd = &cli.Command{
	Name:      "test_step_folder",
	Action:    testStepFolder,
	ArgsUsage: "<socket-addr> <folder-path>",
	Arguments: []cli.Argument{
		socketAddrArg,
		folderPathArg,
	},
}

type stepFile struct {
	step     int
	path     string
	action   string
	isFuzzer bool
}

// peerInfoFuzzerMsg represents the fuzzer message for peer_info action
type peerInfoFuzzerMsg struct {
	PeerInfo fuzz.PeerInfo `json:"peer_info"`
}

// peerInfoTargetMsg represents the target message for peer_info action
type peerInfoTargetMsg struct {
	PeerInfo fuzz.PeerInfo `json:"peer_info"`
}

// initializeData represents the initialize data structure
type initializeData struct {
	Header types.Header       `json:"header"`
	State  types.StateKeyVals `json:"state"`
}

// initializeFuzzerMsg represents the fuzzer message for initialize action
type initializeFuzzerMsg struct {
	Initialize initializeData `json:"initialize"`
}

// initializeTargetMsg represents the target message for initialize action
type initializeTargetMsg struct {
	StateRoot string `json:"state_root"`
}

// importBlockFuzzerMsg represents the fuzzer message for import_block action
type importBlockFuzzerMsg struct {
	ImportBlock types.Block `json:"import_block"`
}

// importBlockTargetMsg represents the target message for import_block action
type importBlockTargetMsg struct {
	StateRoot string `json:"state_root,omitempty"`
	Error     string `json:"error,omitempty"`
}

type stepTestProcessor struct {
	fuzzerData []byte
	targetData []byte
	stepFiles  []stepFile
}

func (p *stepTestProcessor) ScanFolder(folderPath string) ([]string, error) {
	var files []stepFile

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		baseName := filepath.Base(path)
		parts := strings.Split(baseName, "_")
		if len(parts) < 3 {
			return nil
		}

		step, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil
		}

		isFuzzer := strings.Contains(baseName, "_fuzzer_")
		var action string
		if isFuzzer {
			if len(parts) >= 3 && parts[1] == "fuzzer" {
				action = strings.TrimSuffix(strings.Join(parts[2:], "_"), ".json")
			}
		} else if strings.Contains(baseName, "_target_") {
			if len(parts) >= 3 && parts[1] == "target" {
				action = strings.TrimSuffix(strings.Join(parts[2:], "_"), ".json")
			}
		}

		if action == "" {
			return nil
		}

		files = append(files, stepFile{
			step:     step,
			path:     path,
			action:   action,
			isFuzzer: isFuzzer,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].step != files[j].step {
			return files[i].step < files[j].step
		}
		if files[i].isFuzzer != files[j].isFuzzer {
			return files[i].isFuzzer
		}
		return files[i].action < files[j].action
	})

	p.stepFiles = files
	var filePaths []string
	for _, step := range files {
		if step.isFuzzer {
			filePaths = append(filePaths, step.path)
		}
	}

	return filePaths, nil
}

func (p *stepTestProcessor) ProcessFile(client *fuzz.FuzzClient, filePath string) error {
	var step stepFile
	for _, s := range p.stepFiles {
		if s.path == filePath && s.isFuzzer {
			step = s
			break
		}
	}

	if step.path == "" {
		return fmt.Errorf("step file not found: %s", filePath)
	}

	var targetAction string
	switch step.action {
	case "peer_info":
		targetAction = "peer_info"
	case "initialize", "import_block":
		targetAction = "state_root"
	default:
		return fmt.Errorf("unknown action: %s", step.action)
	}

	dir := filepath.Dir(step.path)
	baseName := filepath.Base(step.path)
	stepNum := strings.Split(baseName, "_")[0]
	targetFileName := fmt.Sprintf("%s_target_%s.json", stepNum, targetAction)
	targetPath := filepath.Join(dir, targetFileName)

	if _, err := os.Stat(targetPath); err != nil {
		if targetAction == "state_root" {
			targetErrorPath := filepath.Join(dir, fmt.Sprintf("%s_target_error.json", stepNum))
			if _, err2 := os.Stat(targetErrorPath); err2 == nil {
				targetPath = targetErrorPath
			} else {
				return fmt.Errorf("target file not found: %s or %s", targetPath, targetErrorPath)
			}
		} else {
			return fmt.Errorf("target file not found: %s", targetPath)
		}
	}

	fuzzerData, err := os.ReadFile(step.path)
	if err != nil {
		return fmt.Errorf("error reading fuzzer file: %v", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("error reading target file: %v", err)
	}

	p.fuzzerData = fuzzerData
	p.targetData = targetData

	switch step.action {
	case "peer_info":
		return p.ProcessPeerInfo(client)
	case "initialize":
		return p.ProcessInitialize(client)
	case "import_block":
		return p.ProcessImportBlock(client)
	default:
		return fmt.Errorf("unknown action: %s", step.action)
	}
}

func (p *stepTestProcessor) ProcessInitialize(client *fuzz.FuzzClient) error {
	var fuzzerMsg initializeFuzzerMsg
	if err := json.Unmarshal(p.fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer initialize: %v", err)
	}

	var targetMsg initializeTargetMsg
	if err := json.Unmarshal(p.targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target state_root: %v", err)
	}

	expectedStateRoot, err := parseStateRoot(targetMsg.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing expected state_root: %v", err)
	}

	logger.ColorGreen("[SetState][Request] block_header_hash= 0x%v", hex.EncodeToString(fuzzerMsg.Initialize.Header.Parent[:]))
	actualStateRoot, err := client.SetState(fuzzerMsg.Initialize.Header, fuzzerMsg.Initialize.State)
	logger.ColorYellow("[SetState][Response] state_root= 0x%v", hex.EncodeToString(actualStateRoot[:]))
	if err != nil {
		return fmt.Errorf("SetState failed: %v", err)
	}

	if actualStateRoot != expectedStateRoot {
		logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x",
			expectedStateRoot, actualStateRoot)
		return fmt.Errorf("state_root mismatch")
	}

	return nil
}

func (p *stepTestProcessor) ProcessImportBlock(client *fuzz.FuzzClient) error {
	var fuzzerMsg importBlockFuzzerMsg
	if err := json.Unmarshal(p.fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer import_block: %v", err)
	}

	var targetMsg importBlockTargetMsg
	if err := json.Unmarshal(p.targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target response: %v", err)
	}

	// Handle expected error case
	if targetMsg.Error != "" {
		_, errorMessage, err := client.ImportBlock(fuzzerMsg.ImportBlock)
		if err != nil {
			return fmt.Errorf("ImportBlock failed: %v", err)
		}
		if errorMessage == nil {
			return fmt.Errorf("expected error but got success")
		}
		if errorMessage.Error != targetMsg.Error {
			return fmt.Errorf("error message mismatch: expected %s, got %s", targetMsg.Error, errorMessage.Error)
		}
		return nil
	}

	expectedStateRoot, err := parseStateRoot(targetMsg.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing expected state_root: %v", err)
	}

	result, err := executeImportBlock(importBlockExecuteInput{
		Client:            client,
		Block:             fuzzerMsg.ImportBlock,
		ExpectedStateRoot: expectedStateRoot,
		EnableLogging:     false,
	})
	if err != nil {
		return fmt.Errorf("ImportBlock failed: %v", err)
	}

	if result.ErrorMessage != nil {
		return fmt.Errorf("unexpected error: %s", result.ErrorMessage.Error)
	}
	if result.HasMismatch {
		return fmt.Errorf("state_root mismatch")
	}

	return nil
}

func (p *stepTestProcessor) ProcessPeerInfo(client *fuzz.FuzzClient) error {
	var fuzzerMsg peerInfoFuzzerMsg
	if err := json.Unmarshal(p.fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer peer_info: %v", err)
	}

	var targetMsg peerInfoTargetMsg
	if err := json.Unmarshal(p.targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target peer_info: %v", err)
	}

	actual, err := client.Handshake(fuzzerMsg.PeerInfo)
	if err != nil {
		return fmt.Errorf("handshake failed: %v", err)
	}

	expected := targetMsg.PeerInfo
	if actual.FuzzVersion != expected.FuzzVersion ||
		actual.FuzzFeatures != expected.FuzzFeatures ||
		actual.AppName != expected.AppName ||
		actual.AppVersion != expected.AppVersion ||
		actual.JamVersion != expected.JamVersion {
		return fmt.Errorf("peer_info mismatch: expected %+v, got %+v", expected, actual)
	}

	return nil
}

func testStepFolder(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("test_step_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return fmt.Errorf("test_step_folder requires a folder path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	processor := &stepTestProcessor{}
	filePaths, err := processor.ScanFolder(folderPath)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no step files found in the specified folder")
	}

	log.Printf("Found %d step files to process\n", len(filePaths))

	successCount := 0
	failureCount := 0

	for _, filePath := range filePaths {
		if err := processor.ProcessFile(client, filePath); err != nil {
			logger.ColorRed("FAILED: %s - %v", filepath.Base(filePath), err)
			failureCount++
		} else {
			logger.ColorGreen("PASSED: %s", filepath.Base(filePath))
			successCount++
		}
	}

	log.Printf("Summary: %d passed, %d failed\n", successCount, failureCount)
	return nil
}
