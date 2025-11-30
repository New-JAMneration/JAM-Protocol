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
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/urfave/cli/v3"
)

var testTraceFolderCmd = &cli.Command{
	Name:      "test_folder",
	Action:    testTraceFolder,
	ArgsUsage: "<socket-addr> <folder-path>",
	Arguments: []cli.Argument{
		socketAddrArg,
		folderPathArg,
	},
	Flags: []cli.Flag{
		folderWiseArg,
	},
}

// traceFixture represents the structure of trace test JSON files
type traceFixture struct {
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

type fixtureProbe struct {
	PreState json.RawMessage `json:"pre_state"`
	State    json.RawMessage `json:"state"`
}

type genesisState struct {
	StateRoot string             `json:"state_root"`
	KeyVals   types.StateKeyVals `json:"keyvals"`
}

type genesisFixture struct {
	Header types.Header `json:"header"`
	State  genesisState `json:"state"`
}

type traceTestProcessor struct {
	setStateRequired bool
	data             []byte
	firstFiles       map[string]string
}

func (p *traceTestProcessor) ScanFolder(folderPath string) ([]string, error) {
	var jsonFiles []string
	p.firstFiles = make(map[string]string)

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			jsonFiles = append(jsonFiles, path)
		} else {
			return nil
		}

		if config.Config.FolderWise {
			folderName := strings.Split(path, "/")
			folderIndex := folderName[len(folderName)-2]
			fileName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

			if _, ok := p.firstFiles[folderIndex]; !ok || fileName == "genesis" {
				p.firstFiles[folderIndex] = fileName
			}

			if fileName == "genesis" {
				index := findGroupFirstIndex(&jsonFiles, folderIndex)
				copy(jsonFiles[index+1:], jsonFiles[index:len(jsonFiles)-1])
				jsonFiles[index] = path
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return jsonFiles, nil
}

func (p *traceTestProcessor) ProcessFile(client *fuzz.FuzzClient, filePath string) error {
	var setStateRequired bool
	if config.Config.FolderWise {
		folderName := strings.Split(filePath, "/")
		fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		folderIndex := folderName[len(folderName)-2]

		firstFileName, ok := p.firstFiles[folderIndex]
		if ok && firstFileName == fileName {
			setStateRequired = true
		}
	}

	p.setStateRequired = setStateRequired

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}

	var probe fixtureProbe
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(probe.PreState) > 0 {
		return p.processTraceFixture(client, filePath, data)
	}

	if len(probe.State) > 0 {
		return p.processGenesisFixture(client, filePath, data)
	}

	return fmt.Errorf("unknown fixture format")
}

func (p *traceTestProcessor) processTraceFixture(client *fuzz.FuzzClient, jsonFile string, data []byte) error {
	fmt.Println("File: ", jsonFile)
	p.data = data

	if err := p.ProcessInitialize(client); err != nil {
		return err
	}

	if err := p.ProcessImportBlock(client); err != nil {
		return err
	}

	return nil
}

func (p *traceTestProcessor) processGenesisFixture(client *fuzz.FuzzClient, jsonFile string, data []byte) error {
	var fixture genesisFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	fmt.Println("File: ", jsonFile)

	expectedStateRoot, err := parseStateRoot(fixture.State.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing state state_root: %v", err)
	}

	logger.ColorGreen("[SetState][Request] block_header_hash= 0x%v", hex.EncodeToString(fixture.Header.Parent[:]))
	actualStateRoot, err := client.SetState(fixture.Header, fixture.State.KeyVals)
	logger.ColorYellow("[SetState][Response] state_root= 0x%v", hex.EncodeToString(actualStateRoot[:]))
	if err != nil {
		return fmt.Errorf("SetState failed: %v", err)
	}

	if actualStateRoot != expectedStateRoot {
		logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x", expectedStateRoot, actualStateRoot)
		return fmt.Errorf("state_root mismatch")
	}

	return nil
}

func (p *traceTestProcessor) ProcessInitialize(client *fuzz.FuzzClient) error {
	var fixture traceFixture
	if err := json.Unmarshal(p.data, &fixture); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	if !config.Config.FolderWise || (config.Config.FolderWise && p.setStateRequired) {
		expectedPreStateRoot, err := parseStateRoot(fixture.PreState.StateRoot)
		if err != nil {
			return fmt.Errorf("error parsing pre_state state_root: %v", err)
		}

		logger.ColorGreen("[SetState][Request] block_header_hash= 0x%v", hex.EncodeToString(fixture.Block.Header.Parent[:]))
		actualPreStateRoot, err := client.SetState(fixture.Block.Header, fixture.PreState.KeyVals)
		logger.ColorYellow("[SetState][Response] state_root= 0x%v", hex.EncodeToString(actualPreStateRoot[:]))
		if err != nil {
			return fmt.Errorf("SetState failed: %v", err)
		}

		if actualPreStateRoot != expectedPreStateRoot {
			logger.ColorBlue("[SetState][Check] state_root mismatch: expected 0x%x, got 0x%x",
				expectedPreStateRoot, actualPreStateRoot)
			return fmt.Errorf("state_root mismatch")
		}
	}

	return nil
}

func (p *traceTestProcessor) ProcessImportBlock(client *fuzz.FuzzClient) error {
	var fixture traceFixture
	if err := json.Unmarshal(p.data, &fixture); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	expectedPostStateRoot, err := parseStateRoot(fixture.PostState.StateRoot)
	if err != nil {
		return fmt.Errorf("error parsing post_state state_root: %v", err)
	}

	serializedHeader, err := utilities.HeaderSerialization(fixture.Block.Header)
	if err != nil {
		return fmt.Errorf("error serializing header: %v", err)
	}
	blockHeaderHashHex := types.HeaderHash(hash.Blake2bHash(serializedHeader))

	result, err := executeImportBlock(importBlockExecuteInput{
		Client:            client,
		Block:             fixture.Block,
		ExpectedStateRoot: expectedPostStateRoot,
		BlockHeaderHash:   blockHeaderHashHex,
		ExpectedPostState: fixture.PostState.KeyVals,
		EnableLogging:     true,
	})
	if err != nil {
		return err
	}
	if result.HasMismatch {
		return fmt.Errorf("state_root mismatch")
	}

	return nil
}

func testTraceFolder(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return fmt.Errorf("test_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return fmt.Errorf("test_folder requires a json file path argument")
	}
	config.Config.FolderWise = cmd.Bool(folderWiseArg.Name)

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	processor := &traceTestProcessor{}
	jsonFiles, err := processor.ScanFolder(folderPath)
	if err != nil {
		return err
	}

	if len(jsonFiles) == 0 {
		return fmt.Errorf("no JSON files found in the specified folder")
	}

	log.Printf("Found %d JSON files to test\n", len(jsonFiles))

	successCount := 0
	failureCount := 0

	info, err := client.Handshake(fuzz.PeerInfo{AppName: "Fuzz-Test"})
	if err != nil {
		return fmt.Errorf("error doing handshake: %v", err)
	}
	log.Printf("handshake successful, fuzz-version: %d, fuzz-features: %d, jam-version: %v, app-version: %v, app-name: %s\n", info.FuzzVersion, info.FuzzFeatures, info.JamVersion, info.AppVersion, info.AppName)

	for _, jsonFile := range jsonFiles {
		if err := processor.ProcessFile(client, jsonFile); err != nil {
			logger.ColorRed("FAILED!!: %s - %v", jsonFile, err)
			failureCount++
		} else {
			logger.ColorGreen("PASSED: %s", jsonFile)
			successCount++
		}
	}

	return nil
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
