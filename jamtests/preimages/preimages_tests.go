package jamtests

import (
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

type PreimageOutput struct {
	Ok  interface{}       // output is nil, so use interface since there is no nil type
	Err PreimageErrorCode `json:"err,omitempty"`
}

type PreimageState struct {
	Delta types.ServiceAccountState `json:"accounts"`
}

type PreimageErrorCode types.ErrorCode

const (
	PreimageUnneeded PreimageErrorCode = iota // 0
)
