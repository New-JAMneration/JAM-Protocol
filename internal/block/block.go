package block

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// BlockController is a struct that contains a Header and an Extrinsic
type BlockController struct {
	Header    *types.Header
	Extrinsic *extrinsic.ExtrinsicController
}

// NewBlockController returns a new BlockController
func NewBlockController(headers *types.Header, extrinsic *extrinsic.ExtrinsicController) *BlockController {
	return &BlockController{
		Header:    headers,
		Extrinsic: extrinsic,
	}
}
