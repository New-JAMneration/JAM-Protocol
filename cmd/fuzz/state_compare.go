package main

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/fuzz"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/New-JAMneration/JAM-Protocol/logger"
)

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

	for _, v := range diffs {
		if actualRoot == nullStateRoot || expectedRoot == nullStateRoot || len(actualStateKeyVal) == 0 {
			break
		}

		if state, keyExists := jamtests.KeyValMap[v.Key]; keyExists {
			logger.ColorYellow("state: %s, key: %v", state, v.Key)
			if len(v.ActualValue) > 256 || len(v.ExpectedValue) > 256 {
				logger.ColorDebug("value too big, only check diff")
			} else {
				logger.ColorDebug("actualVal: %+v", v.ActualValue)
				logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
			}
			continue
		}

		if v.Key[0] == byte(255) {
			serviceID, err := merklization.DecodeServiceIdFromType2(v.Key)
			if err != nil {
				return fmt.Errorf("fuzzer DecodeServiceIdFromType2 error: %w", err)
			}
			logger.ColorYellow("service: %d", serviceID)
			logger.ColorDebug("actualVal: %+v", v.ActualValue)
			logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
			continue
		}

		logger.ColorYellow("other state key: 0x%x", v.Key)
		if len(v.ActualValue) > 256 || len(v.ExpectedValue) > 256 {
			logger.ColorDebug("value too big, check json file")
		} else {
			logger.ColorDebug("actualVal: %+v", v.ActualValue)
			logger.ColorDebug("expectVal: %+v", v.ExpectedValue)
		}
	}

	return nil
}
