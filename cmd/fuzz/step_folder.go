package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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

func testStepFolder(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("test_step_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return errors.New("test_step_folder requires a folder path argument")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	stepFiles, err := scanStepFiles(folderPath)
	if err != nil {
		return err
	}

	if len(stepFiles) == 0 {
		return errors.New("no step files found in the specified folder")
	}

	logger.Infof("Found %d step files to process", len(stepFiles))

	successCount := 0
	failureCount := 0

	for _, step := range stepFiles {
		if err := processStep(client, step); err != nil {
			logger.ColorRed("FAILED: %s - %v", filepath.Base(step.path), err)
			failureCount++
		} else {
			logger.ColorGreen("PASSED: %s", filepath.Base(step.path))
			successCount++
		}
	}

	logger.Infof("Summary: %d passed, %d failed", successCount, failureCount)
	return nil
}

func scanStepFiles(folderPath string) ([]stepFile, error) {
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
		return nil, fmt.Errorf("error walking directory: %w", err)
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

	return files, nil
}

func processStep(client *fuzz.FuzzClient, step stepFile) error {
	if !step.isFuzzer {
		return nil
	}

	var targetAction string
	switch step.action {
	case "peer_info":
		targetAction = "peer_info"
	case "initialize", "import_block":
		targetAction = "state_root"
	case "get_state":
		targetAction = "get_state"
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
		} else if targetAction == "get_state" {
			targetPath = filepath.Join(dir, fmt.Sprintf("%s_target_state.json", stepNum))
		} else {
			return fmt.Errorf("target file not found: %s", targetPath)
		}
	}

	fuzzerData, err := os.ReadFile(step.path)
	if err != nil {
		return fmt.Errorf("error reading fuzzer file: %w", err)
	}

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("error reading target file: %w", err)
	}

	switch step.action {
	case "peer_info":
		return processPeerInfo(client, fuzzerData, targetData)
	case "initialize":
		return processInitialize(client, fuzzerData, targetData)
	case "import_block":
		return processImportBlock(client, fuzzerData, targetData)
	case "get_state":
		return processGetState(client, fuzzerData, targetData)
	default:
		return fmt.Errorf("unknown action: %s", step.action)
	}
}

func processPeerInfo(client *fuzz.FuzzClient, fuzzerData, targetData []byte) error {
	var fuzzerMsg struct {
		PeerInfo fuzz.PeerInfo `json:"peer_info"`
	}
	if err := json.Unmarshal(fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer peer_info: %w", err)
	}

	var targetMsg struct {
		PeerInfo fuzz.PeerInfo `json:"peer_info"`
	}
	if err := json.Unmarshal(targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target peer_info: %w", err)
	}

	actual, err := client.Handshake(fuzzerMsg.PeerInfo)
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
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

func processInitialize(client *fuzz.FuzzClient, fuzzerData, targetData []byte) error {
	var fuzzerMsg struct {
		Initialize struct {
			Header   types.Header       `json:"header"`
			State    types.StateKeyVals `json:"state"`
			Ancestry types.Ancestry     `json:"ancestry"`
		} `json:"initialize"`
	}
	if err := json.Unmarshal(fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer initialize: %w", err)
	}

	var targetMsg struct {
		StateRoot string `json:"state_root"`
	}
	if err := json.Unmarshal(targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target state_root: %w", err)
	}

	expectedStateRoot, err := parseStateRoot(targetMsg.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing expected state_root: %w", err)
	}

	logger.ColorGreen("[SetState][Request] block_header_hash= 0x%x", fuzzerMsg.Initialize.Header.Parent)
	actualStateRoot, err := client.SetState(fuzzerMsg.Initialize.Header, fuzzerMsg.Initialize.State, fuzzerMsg.Initialize.Ancestry)
	logger.ColorYellow("[SetState][Response] state_root= 0x%x", actualStateRoot)
	if err != nil {
		return fmt.Errorf("SetState failed: %w", err)
	}

	if actualStateRoot != expectedStateRoot {
		logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x",
			expectedStateRoot, actualStateRoot)
		return errors.New("state_root mismatch")
	}

	return nil
}

func processImportBlock(client *fuzz.FuzzClient, fuzzerData, targetData []byte) error {
	var fuzzerMsg struct {
		ImportBlock types.Block `json:"import_block"`
	}
	if err := json.Unmarshal(fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer import_block: %w", err)
	}

	var targetMsg struct {
		StateRoot string `json:"state_root,omitempty"`
		Error     string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target response: %w", err)
	}

	// Use block's ParentStateRoot as priorStateRoot for protocol error fallback
	priorStateRoot := fuzzerMsg.ImportBlock.Header.ParentStateRoot
	
	if targetMsg.Error != "" {
		_, errorMessage, err := client.ImportBlock(fuzzerMsg.ImportBlock, priorStateRoot)
		if err != nil {
			return fmt.Errorf("ImportBlock failed: %w", err)
		}
		if errorMessage == nil {
			return errors.New("expected error but got success")
		}
		if errorMessage.Error != targetMsg.Error {
			if fuzzyMatchErrorMessage(targetMsg.Error, errorMessage.Error) {
				return nil
			}
			return fmt.Errorf("error message mismatch: expected %s, got %s", targetMsg.Error, errorMessage.Error)
		}
		return nil
	}

	expectedStateRoot, err := parseStateRoot(targetMsg.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing expected state_root: %w", err)
	}

	actualStateRoot, errorMessage, err := client.ImportBlock(fuzzerMsg.ImportBlock, priorStateRoot)
	if err != nil {
		return fmt.Errorf("ImportBlock failed: %w", err)
	}
	if errorMessage != nil {
		return fmt.Errorf("unexpected error: %s", errorMessage.Error)
	}

	mismatch, err := compareImportBlockState(importBlockCompareInput{
		Client:            client,
		Block:             fuzzerMsg.ImportBlock,
		ExpectedStateRoot: expectedStateRoot,
		ActualStateRoot:   actualStateRoot,
	})
	if err != nil {
		return err
	}
	if mismatch {
		return errors.New("state_root mismatch")
	}

	return nil
}

func processGetState(client *fuzz.FuzzClient, fuzzerData, targetData []byte) error {
	var fuzzerMsg struct {
		GetState types.HeaderHash `json:"get_state"`
	}

	if err := json.Unmarshal(fuzzerData, &fuzzerMsg); err != nil {
		return fmt.Errorf("error parsing fuzzer get_state: %w", err)
	}

	var targetMsg struct {
		State types.StateKeyVals `json:"state"`
	}

	if err := json.Unmarshal(targetData, &targetMsg); err != nil {
		return fmt.Errorf("error parsing target state: %w", err)
	}

	actualState, err := client.GetState(fuzzerMsg.GetState)
	if err != nil {
		return fmt.Errorf("GetState failed: %w", err)
	}

	expectedState := targetMsg.State

	if len(actualState) != len(expectedState) {
		return fmt.Errorf("state length mismatch: expected %d, got %d", len(expectedState), len(actualState))
	}

	// Compare expected and actual state key-values
	for i := range expectedState {
		expectedKey := expectedState[i].Key
		expectedValue := expectedState[i].Value
		actualKey := actualState[i].Key
		actualValue := actualState[i].Value

		if expectedKey != actualKey {
			return fmt.Errorf("state key mismatch at index %d: expected key %s, got key %s", i, expectedKey, actualKey)
		}
		if !bytes.Equal(expectedValue, actualValue) {
			return fmt.Errorf("state value mismatch at index %d for key %s: expected value %x, got value %x", i, expectedKey, expectedValue, actualValue)
		}
	}

	return nil
}

func fuzzyMatchErrorMessage(expected, actual string) bool {
	if strings.Contains(expected, actual) {
		return true
	}

	// Map our error messages to possible aliases
	errorAliasMap := map[string]string{
		"invalid epoch mark":   "InvalidEpochMark",
		"invalid tickets mark": "InvalidTicketsMark",
		"unexpected author":    "UnexpectedAuthor",
	}

	if alias, exists := errorAliasMap[actual]; exists {
		if strings.Contains(expected, alias) {
			return true
		}
	}

	return false
}
