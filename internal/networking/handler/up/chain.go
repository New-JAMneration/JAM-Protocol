package up

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// ChainView indexes known blocks by header hash for ancestry and leaf queries.
type ChainView struct {
	hashToRef  map[types.HeaderHash]BlockRef
	parentOf   map[types.HeaderHash]types.HeaderHash
	childrenOf map[types.HeaderHash][]types.HeaderHash
}

// NewChainView builds an index from known blocks (genesis through tips).
func NewChainView(blocks []types.Block) (ChainView, error) {
	cv := ChainView{
		hashToRef:  make(map[types.HeaderHash]BlockRef, len(blocks)),
		parentOf:   make(map[types.HeaderHash]types.HeaderHash, len(blocks)),
		childrenOf: make(map[types.HeaderHash][]types.HeaderHash),
	}
	for _, block := range blocks {
		h, err := hash.ComputeBlockHeaderHash(block.Header)
		if err != nil {
			return ChainView{}, fmt.Errorf("hash block slot %d: %w", block.Header.Slot, err)
		}
		cv.hashToRef[h] = BlockRef{Hash: h, Slot: block.Header.Slot}
		cv.parentOf[h] = block.Header.Parent
		cv.childrenOf[block.Header.Parent] = append(cv.childrenOf[block.Header.Parent], h)
	}
	return cv, nil
}

// ViewAtFinalized builds a chain view and returns the finalized block ref.
func ViewAtFinalized(blocks []types.Block, finalized types.HeaderHash) (ChainView, BlockRef, error) {
	cv, err := NewChainView(blocks)
	if err != nil {
		return ChainView{}, BlockRef{}, err
	}
	finalRef, ok := cv.hashToRef[finalized]
	if !ok {
		return ChainView{}, BlockRef{}, fmt.Errorf("finalized block %x not in chain view", finalized[:4])
	}
	return cv, finalRef, nil
}

// IsDescendantOf reports whether descendant is a strict or non-strict descendant of ancestor.
func (cv ChainView) IsDescendantOf(descendant, ancestor types.HeaderHash) bool {
	if descendant == ancestor {
		return true
	}
	for cur := descendant; ; {
		parent, ok := cv.parentOf[cur]
		if !ok {
			return false
		}
		if parent == ancestor {
			return true
		}
		cur = parent
	}
}

// CollectLeaves returns tips that are descendants of finalized and have no known children.
func (cv ChainView) CollectLeaves(finalized types.HeaderHash) []BlockRef {
	leaves := make([]BlockRef, 0)
	for h, ref := range cv.hashToRef {
		if h == finalized {
			continue
		}
		if !cv.IsDescendantOf(h, finalized) {
			continue
		}
		if len(cv.childrenOf[h]) == 0 {
			leaves = append(leaves, ref)
		}
	}
	return leaves
}

// ShouldSkipAnnouncement applies JAMNP-S UP 0 skip rules for announcing block.
func ShouldSkipAnnouncement(
	block types.HeaderHash,
	finalized types.HeaderHash,
	cv ChainView,
	announcedByUs map[types.HeaderHash]struct{},
	announcedByPeer map[types.HeaderHash]struct{},
) bool {
	if !cv.IsDescendantOf(block, finalized) {
		return true
	}
	for h := range announcedByUs {
		if h != block && cv.IsDescendantOf(h, block) {
			return true
		}
	}
	for h := range announcedByPeer {
		if cv.IsDescendantOf(h, block) {
			return true
		}
	}
	return false
}
