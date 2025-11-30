package main

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// importBlockExecuteInput contains parameters for executing ImportBlock
type importBlockExecuteInput struct {
	Client            *fuzz.FuzzClient
	Block             types.Block
	ExpectedStateRoot types.StateRoot
	BlockHeaderHash   types.HeaderHash   // optional, will be computed if empty
	ExpectedPostState types.StateKeyVals // optional, for detailed diff logging
	EnableLogging     bool               // whether to log request/response
}

// importBlockExecuteResult contains the result of executing ImportBlock
type importBlockExecuteResult struct {
	ActualStateRoot types.StateRoot
	ErrorMessage    *fuzz.ErrorMessage
	HasMismatch     bool
}

type importBlockCompareInput struct {
	Client            *fuzz.FuzzClient
	Block             types.Block
	BlockHeaderHash   types.HeaderHash
	ExpectedStateRoot types.StateRoot
	ActualStateRoot   types.StateRoot
	ExpectedPostState types.StateKeyVals
}

func compareImportBlockState(input importBlockCompareInput) (bool, error) {
	if input.ActualStateRoot == input.ExpectedStateRoot {
		return false, nil
	}

	headerHash, err := input.ensureBlockHeaderHash()
	if err != nil {
		return false, fmt.Errorf("error preparing block header hash: %w", err)
	}

	logger.ColorBlue("[ImportBlock][Check] state_root mismatch: expected 0x%x, got 0x%x",
		input.ExpectedStateRoot, input.ActualStateRoot)

	if input.Client == nil || len(input.ExpectedPostState) == 0 {
		return true, nil
	}

	if err := logImportBlockDiff(input.Client, headerHash, input.ExpectedStateRoot, input.ActualStateRoot, input.ExpectedPostState); err != nil {
		return true, err
	}

	return true, nil
}

func (input importBlockCompareInput) ensureBlockHeaderHash() (types.HeaderHash, error) {
	if input.BlockHeaderHash != (types.HeaderHash{}) {
		return input.BlockHeaderHash, nil
	}

	serializedHeader, err := utilities.HeaderSerialization(input.Block.Header)
	if err != nil {
		return types.HeaderHash{}, err
	}

	return types.HeaderHash(hash.Blake2bHash(serializedHeader)), nil
}

func logImportBlockDiff(client *fuzz.FuzzClient, headerHash types.HeaderHash, expectedRoot, actualRoot types.StateRoot, expectedKeyVals types.StateKeyVals) error {
	logger.ColorGreen("[GetState][Request] header_hash= 0x%v", hex.EncodeToString(headerHash[:]))
	actualStateKeyVal, err := client.GetState(headerHash)
	if err != nil {
		return fmt.Errorf("error sending get_state request: %w", err)
	}

	diffs, err := merklization.GetStateKeyValsDiff(expectedKeyVals, actualStateKeyVal)
	if err != nil {
		return fmt.Errorf("fuzzer GetStateKeyValsDiff error: %w", err)
	}

	logger.ColorYellow("[GetState][Response] %d different key-val: ", len(diffs))
	nullStateRoot := types.StateRoot{}
	if actualRoot == nullStateRoot || expectedRoot == nullStateRoot || len(actualStateKeyVal) == 0 {
		logger.ColorYellow("[GetState][Response] nil state root, imply protocol error")
		return nil // no diffs
	}
	err = merklization.DebugStateKeyValsDiff(diffs)
	if err != nil {
		return fmt.Errorf("fuzzer DebugStateKeyValsDiff error: %w", err)
	}

	return nil
}

// executeImportBlock executes ImportBlock and compares the result with expected state root
func executeImportBlock(input importBlockExecuteInput) (*importBlockExecuteResult, error) {
	if input.EnableLogging {
		blockHeaderHash := input.BlockHeaderHash
		if blockHeaderHash == (types.HeaderHash{}) {
			serializedHeader, err := utilities.HeaderSerialization(input.Block.Header)
			if err != nil {
				return nil, fmt.Errorf("error serializing header: %v", err)
			}
			blockHeaderHash = types.HeaderHash(hash.Blake2bHash(serializedHeader))
		}
		logger.ColorGreen("[ImportBlock][Request] block_header_hash= 0x%x", blockHeaderHash)
	}

	actualStateRoot, errorMessage, err := input.Client.ImportBlock(input.Block)
	if err != nil {
		if input.EnableLogging {
			logger.ColorYellow("[ImportBlock][Response] error= %v", err)
		}
		return nil, err
	}

	result := &importBlockExecuteResult{
		ActualStateRoot: actualStateRoot,
		ErrorMessage:    errorMessage,
	}

	if errorMessage != nil {
		if input.EnableLogging {
			logger.ColorYellow("[ImportBlock][Response] error message= %v", errorMessage.Error)
		}
		// Don't check mismatch if there's an error message
		return result, nil
	}

	if input.EnableLogging {
		logger.ColorYellow("[ImportBlock][Response] state_root= 0x%v", hex.EncodeToString(actualStateRoot[:]))
	}

	// Compare state roots
	mismatch, err := compareImportBlockState(importBlockCompareInput{
		Client:            input.Client,
		Block:             input.Block,
		BlockHeaderHash:   input.BlockHeaderHash,
		ExpectedStateRoot: input.ExpectedStateRoot,
		ActualStateRoot:   actualStateRoot,
		ExpectedPostState: input.ExpectedPostState,
	})
	if err != nil {
		return nil, err
	}

	result.HasMismatch = mismatch
	return result, nil
}
