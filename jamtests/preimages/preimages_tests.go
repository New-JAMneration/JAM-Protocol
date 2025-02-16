package jamtests

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type PreimageTestCase struct {
	Input     PreimageInput  `json:"input"`
	PreState  PreimageState  `json:"pre_state"`
	Output    PreimageOutput `json:"output"`
	PostState PreimageState  `json:"post_state"`
}

type PreimageInput struct {
	Preimages types.PreimagesExtrinsic `json:"preimages"`
	Slot      types.TimeSlot           `json:"slot"`
}

type PreimageOutputData struct {
	Null *struct{}
}

type PreimageOutput struct {
	Ok  interface{}        `json:"ok,omitempty"`
	Err *PreimageErrorCode `json:"err,omitempty"`
}

type PreimagesMapEntry struct {
	Hash types.OpaqueHash   `json:"hash"`
	Blob types.ByteSequence `json:"blob"`
}

type LookupMetaMapkey struct {
	Hash   types.OpaqueHash `json:"hash"`
	Length types.U32        `json:"length"`
}

type LookupMetaMapEntry struct {
	Key LookupMetaMapkey `json:"key"`
	Val []types.TimeSlot `json:"value"`
}

type Account struct {
	Preimages  []PreimagesMapEntry  `json:"preimages"`
	LookupMeta []LookupMetaMapEntry `json:"lookup_meta"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Data Account         `json:"data"`
}

type PreimageState struct {
	Accounts []AccountsMapEntry `json:"accounts"`
}

type PreimageErrorCode types.ErrorCode

const (
	PreimageUnneeded         PreimageErrorCode = iota // 0
	PreimagesNotSortedUnique PreimageErrorCode = 1
)

var preimageErrorMap = map[string]PreimageErrorCode{
	"preimage_unneeded":           PreimageUnneeded,
	"preimages_not_sorted_unique": PreimagesNotSortedUnique,
}

func (e *PreimageErrorCode) UnmarshalJSON(data []byte) error {
	var str string

	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := preimageErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}

func (p *PreimagesMapEntry) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash string `json:"hash"`
		Blob string `json:"blob"`
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

func (l *LookupMetaMapkey) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash   string    `json:"hash"`
		Length types.U32 `json:"length"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}

	l.Hash = types.OpaqueHash(hashBytes)

	l.Length = temp.Length

	return nil
}
