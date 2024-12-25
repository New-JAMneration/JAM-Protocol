package block

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/extrinsic"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

// BlockController is a struct that contains a Header and an Extrinsic
type BlockController struct {
	Header    *jamTypes.Header
	Extrinsic *extrinsic.ExtrinsicController
}

// NewBlockController returns a new BlockController
func NewBlockController(headers *jamTypes.Header, extrinsic *extrinsic.ExtrinsicController) *BlockController {
	return &BlockController{
		Header:    headers,
		Extrinsic: extrinsic,
	}
}
