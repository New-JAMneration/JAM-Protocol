package jamtests

import (
	"encoding/hex"
	"encoding/json"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PreimagesMapEntry struct {
	Hash types.OpaqueHash   `json:"hash"`
	Blob types.ByteSequence `json:"blob"`
}

type Account struct {
	Service   types.ServiceInfo   `json:"service"`
	Preimages []PreimagesMapEntry `json:"preimages"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Data Account         `json:"data"`
}

type AccumulateState struct {
	Slot        types.TimeSlot         `json:"slot"`
	Entropy     types.Entropy          `json:"entropy"`
	ReadyQueue  types.ReadyQueue       `json:"ready_queue"`
	Accumulated types.AccumulatedQueue `json:"accumulated"`
	Privileges  types.Privileges       `json:"privileges"`
	Accounts    []AccountsMapEntry     `json:"accounts"`
}

type AccumulateInput struct {
	Slot    types.TimeSlot     `json:"slot"`
	Reports []types.WorkReport `json:"reports"`
}

type AccumulateOutput struct {
	Ok  *types.AccumulateRoot `json:"ok,omitempty"`
	Err interface{}           `json:"err,omitempty"` // err NULL
}

type AccumulateTestCase struct {
	Input     AccumulateInput  `json:"input"`
	PreState  AccumulateState  `json:"pre_state"`
	Output    AccumulateOutput `json:"output"`
	PostState AccumulateState  `json:"post_state"`
}

func (p *PreimagesMapEntry) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash string `json:"hash,omitempty"`
		Blob string `json:"blob,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}

	p.Hash = types.OpaqueHash(hashBytes)

	blobBytes, err := hex.DecodeString(temp.Blob[2:])
	if err != nil {
		return err
	}

	p.Blob = types.ByteSequence(blobBytes)

	return nil
}
