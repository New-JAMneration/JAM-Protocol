package node

import (
	"errors"
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/stf"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// ErrParentMismatch is returned when a block's parent hash does not match local head.
var ErrParentMismatch = errors.New("block parent does not match local head")

// ImportBlock validates and imports a block through STF (#966).
// On success the posterior state is committed and the imported header hash is returned.
// Protocol validation failures are logged and returned without StateCommit.
func ImportBlock(cs *blockchain.ChainState, block types.Block) (types.HeaderHash, error) {
	if cs == nil {
		return types.HeaderHash{}, fmt.Errorf("nil chain state")
	}

	head, err := cs.GetCurrentHead()
	if err != nil {
		return types.HeaderHash{}, fmt.Errorf("current head: %w", err)
	}
	headHash, err := hash.ComputeBlockHeaderHash(head.Header)
	if err != nil {
		return types.HeaderHash{}, fmt.Errorf("compute head hash: %w", err)
	}
	if block.Header.Parent != headHash {
		return types.HeaderHash{}, fmt.Errorf("%w: head=0x%x parent=0x%x",
			ErrParentMismatch, headHash[:8], block.Header.Parent[:8])
	}

	headerHash, err := hash.ComputeBlockHeaderHash(block.Header)
	if err != nil {
		return types.HeaderHash{}, fmt.Errorf("compute block hash: %w", err)
	}

	cs.AddBlock(block)

	isProtocolError, err := stf.RunSTF()
	if err != nil {
		if isProtocolError {
			log.Printf("block import: protocol error slot=%d hash=0x%x: %v",
				block.Header.Slot, headerHash[:8], err)
			return types.HeaderHash{}, fmt.Errorf("protocol: %w", err)
		}
		return types.HeaderHash{}, fmt.Errorf("STF runtime: %w", err)
	}

	cs.StateCommit()
	return types.HeaderHash(headerHash), nil
}
