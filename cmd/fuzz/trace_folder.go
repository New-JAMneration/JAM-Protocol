package main

import (
	"context"
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
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	jamtests_trace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	"github.com/urfave/cli/v3"
)

var (
	fromStepFlag = &cli.UintFlag{
		Name:  "from-step",
		Usage: "Replay from this step number (inclusive)",
	}

	toStepFlag = &cli.UintFlag{
		Name:  "to-step",
		Usage: "Replay up to this step number (inclusive)",
	}

	skipBootstrapFlag = &cli.BoolFlag{
		Name:  "skip-bootstrap",
		Usage: "Skip the one-time SetState bootstrap (for soak rounds after the first)",
	}

	testTraceFolderCmd = &cli.Command{
		Name:      "test_trace_folder",
		Usage:     "Replay TraceStep .bin fixtures (live fuzz: one SetState, then ImportBlock only)",
		Action:    testTraceFolder,
		ArgsUsage: "<socket-addr> <folder-path>",
		Arguments: []cli.Argument{
			socketAddrArg,
			folderPathArg,
		},
		Flags: []cli.Flag{
			fromStepFlag,
			toStepFlag,
			skipBootstrapFlag,
		},
	}
)

type traceBinFile struct {
	step int
	path string
}

func testTraceFolder(ctx context.Context, cmd *cli.Command) error {
	socketAddr := cmd.StringArg(socketAddrArg.Name)
	if socketAddr == "" {
		return errors.New("test_trace_folder requires a socket path argument")
	}
	folderPath := cmd.StringArg(folderPathArg.Name)
	if folderPath == "" {
		return errors.New("test_trace_folder requires a folder path argument")
	}

	fromStep := int(cmd.Uint(fromStepFlag.Name))
	toStep := int(cmd.Uint(toStepFlag.Name))

	headerMap, err := buildTraceHeaderMap(folderPath)
	if err != nil {
		return err
	}

	traceFiles, err := scanTraceBinFiles(folderPath, fromStep, toStep)
	if err != nil {
		return err
	}
	if len(traceFiles) == 0 {
		return errors.New("no trace .bin files found in the specified folder")
	}

	client, err := fuzz.NewFuzzClient("unix", socketAddr)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	defer client.Close()

	info, err := client.Handshake(fuzz.PeerInfo{AppName: "Fuzz-Trace-Replay"})
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}
	logger.Infof("handshake successful, fuzz-version: %d, jam-version: %v, app-name: %s",
		info.FuzzVersion, info.JamVersion, info.AppName)

	if !cmd.Bool(skipBootstrapFlag.Name) {
		if err := bootstrapLiveTraceReplay(client, folderPath, traceFiles, headerMap); err != nil {
			return fmt.Errorf("live bootstrap: %w", err)
		}
	} else {
		logger.Infof("live bootstrap skipped (--skip-bootstrap)")
	}

	logger.Infof("Found %d trace .bin files to replay (steps %08d .. %08d), live mode: ImportBlock only",
		len(traceFiles), traceFiles[0].step, traceFiles[len(traceFiles)-1].step)

	successCount := 0
	failureCount := 0

	for _, traceFile := range traceFiles {
		label := fmt.Sprintf("%08d.bin", traceFile.step)
		if err := replayTraceStepLive(client, traceFile.path, label); err != nil {
			logger.ColorRed("FAILED: %s - %v", label, err)
			failureCount++
			return fmt.Errorf("stopped at step %d: %w", traceFile.step, err)
		}
		logger.ColorGreen("PASSED: %s", label)
		successCount++
	}

	logger.Infof("Summary: %d passed, %d failed", successCount, failureCount)
	if failureCount > 0 {
		return fmt.Errorf("%d trace steps failed", failureCount)
	}
	return nil
}

// bootstrapLiveTraceReplay performs a single SetState like live fuzz (genesis / initialize).
// Subsequent steps must use ImportBlock only; sibling branches rely on target RestoreBlockAndState.
func bootstrapLiveTraceReplay(
	client *fuzz.FuzzClient,
	folderPath string,
	traceFiles []traceBinFile,
	headerMap map[types.HeaderHash]types.Header,
) error {
	genesisPath := filepath.Join(folderPath, "genesis.bin")
	if _, err := os.Stat(genesisPath); err == nil {
		logger.Infof("live bootstrap: genesis.bin")
		return replayGenesisBin(client, genesisPath)
	}

	first, err := loadTraceTestCase(traceFiles[0].path)
	if err != nil {
		return err
	}

	parentHash := first.Block.Header.Parent
	if parentHeader, ok := headerMap[parentHash]; ok {
		logger.Infof("live bootstrap: SetState pre-state before step %d (parent header in trace ring)",
			traceFiles[0].step)
		return setStateFromTraceState(client, parentHeader, first.PreState)
	}

	for _, tf := range traceFiles {
		tc, loadErr := loadTraceTestCase(tf.path)
		if loadErr != nil {
			return loadErr
		}
		if parentHeader, ok := headerMap[tc.Block.Header.Parent]; ok {
			logger.ColorYellow(
				"live bootstrap: ring edge — SetState pre-state before step %d (steps %d..%d may fail until chain reaches %d)",
				tf.step, traceFiles[0].step, tf.step-1, tf.step,
			)
			return setStateFromTraceState(client, parentHeader, tc.PreState)
		}
	}

	return fmt.Errorf(
		"need genesis.bin in %s (parent 0x%x of step %d not in trace ring); live fuzz uses one genesis SetState then ImportBlock only",
		folderPath, parentHash[:8], traceFiles[0].step,
	)
}

func setStateFromTraceState(client *fuzz.FuzzClient, header types.Header, state jamtests_trace.TraceState) error {
	expectedStateRoot := state.StateRoot
	headerHash, err := hash.ComputeBlockHeaderHash(header)
	if err != nil {
		return fmt.Errorf("compute header hash: %w", err)
	}

	logger.ColorGreen("[SetState][Request] header_hash= 0x%x state_root= 0x%x", headerHash[:8], expectedStateRoot)
	actualStateRoot, err := client.SetState(header, state.KeyVals, types.Ancestry{})
	logger.ColorYellow("[SetState][Response] state_root= 0x%x", actualStateRoot)
	if err != nil {
		return fmt.Errorf("SetState failed: %w", err)
	}
	if actualStateRoot != expectedStateRoot {
		return fmt.Errorf("state_root mismatch: expected 0x%x, got 0x%x", expectedStateRoot, actualStateRoot)
	}
	return nil
}

func buildTraceHeaderMap(folderPath string) (map[types.HeaderHash]types.Header, error) {
	headerMap := make(map[types.HeaderHash]types.Header)

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".bin") {
			return nil
		}

		baseName := filepath.Base(path)
		if baseName == "report.bin" || baseName == "genesis.bin" {
			return nil
		}

		stepText := strings.TrimSuffix(baseName, ".bin")
		if _, err := strconv.Atoi(stepText); err != nil || len(stepText) != 8 {
			return nil
		}

		testCase, err := loadTraceTestCase(path)
		if err != nil {
			return fmt.Errorf("%s: %w", baseName, err)
		}

		headerHash, err := hash.ComputeBlockHeaderHash(testCase.Block.Header)
		if err != nil {
			return fmt.Errorf("%s: compute header hash: %w", baseName, err)
		}
		headerMap[types.HeaderHash(headerHash)] = testCase.Block.Header
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("build header map: %w", err)
	}

	return headerMap, nil
}

func scanTraceBinFiles(folderPath string, fromStep, toStep int) ([]traceBinFile, error) {
	files := make([]traceBinFile, 0, 1024)

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".bin") {
			return nil
		}

		baseName := filepath.Base(path)
		if baseName == "report.bin" || baseName == "genesis.bin" {
			return nil
		}

		stepText := strings.TrimSuffix(baseName, ".bin")
		step, err := strconv.Atoi(stepText)
		if err != nil || len(stepText) != 8 {
			return nil
		}

		if fromStep > 0 && step < fromStep {
			return nil
		}
		if toStep > 0 && step > toStep {
			return nil
		}

		files = append(files, traceBinFile{step: step, path: path})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].step < files[j].step
	})

	return files, nil
}

func loadTraceTestCase(path string) (*jamtests_trace.TraceTestCase, error) {
	testCase := &jamtests_trace.TraceTestCase{}
	if err := utilities.GetTestFromBin(path, testCase); err != nil {
		return nil, fmt.Errorf("decode trace bin: %w", err)
	}
	return testCase, nil
}

func replayGenesisBin(client *fuzz.FuzzClient, path string) error {
	genesis := &jamtests_trace.Genesis{}
	if err := utilities.GetTestFromBin(path, genesis); err != nil {
		return fmt.Errorf("decode genesis bin: %w", err)
	}
	return setStateFromTraceState(client, genesis.Header, genesis.State)
}

func replayTraceStepLive(client *fuzz.FuzzClient, path, label string) error {
	testCase, err := loadTraceTestCase(path)
	if err != nil {
		return err
	}

	expectedPreStateRoot := testCase.PreState.StateRoot
	expectedPostStateRoot := testCase.PostState.StateRoot

	headerHash, err := hash.ComputeBlockHeaderHash(testCase.Block.Header)
	if err != nil {
		return fmt.Errorf("compute header hash: %w", err)
	}

	logger.ColorBlue("File: %s", label)

	logger.ColorGreen("[ImportBlock][Request] header_hash= 0x%x...", headerHash[:8])
	actualPostStateRoot, errorMessage, err := client.ImportBlock(testCase.Block, expectedPreStateRoot)
	if err != nil {
		logger.ColorYellow("[ImportBlock][Response] error= %v", err)
		return err
	}
	if errorMessage != nil {
		logger.ColorYellow("[ImportBlock][Response] error message= %v", errorMessage.Error)
	} else {
		logger.ColorYellow("[ImportBlock][Response] state_root= 0x%x", actualPostStateRoot)
	}

	mismatch, err := compareImportBlockState(importBlockCompareInput{
		Client:            client,
		Block:             testCase.Block,
		BlockHeaderHash:   types.HeaderHash(headerHash),
		ExpectedStateRoot: expectedPostStateRoot,
		ActualStateRoot:   actualPostStateRoot,
		ExpectedPostState: testCase.PostState.KeyVals,
	})
	if err != nil {
		return err
	}
	if mismatch {
		return errors.New("state_root mismatch")
	}

	return nil
}
