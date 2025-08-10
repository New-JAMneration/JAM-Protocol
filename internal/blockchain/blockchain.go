package blockchain

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Just the template for the interface
// TODO: Implement the Blockchain interface.
// Blockchain defines the required interface for retrieving blocks.
type Blockchain interface {
	// GetBlockNumber returns the block number for the given block hash.
	GetBlockNumber(types.HeaderHash) (uint32, error)
	// GetBlockHashByNumber returns candidate block hashes for the specified block number.
	GetBlockHashByNumber(number uint32) ([]types.HeaderHash, error)
	// GetBlock returns a block for the given block hash.
	GetBlock(types.HeaderHash) (types.Block, error)
	// GenesisBlockHash returns the genesis block hash.
	GenesisBlockHash() types.HeaderHash
	// TODO: add a method to get the current head
	GetCurrentHead() (types.Block, error)
	// SetCurrentHead sets the current head.
	SetCurrentHead(types.HeaderHash)
	// GetStateAt returns the state at the given block hash
	GetStateAt(types.HeaderHash) (types.StateKeyVals, error)
	// GetStateRange returns state values in the given key range at the specified block hash
	GetStateRange(types.HeaderHash, types.StateKey, types.StateKey, uint32) (types.StateKeyVals, error)
	// GetBoundaryNodes returns the boundary nodes covering the paths from the root to the start key and to the last key/value pair included in the response.
	// The returned nodes should be ordered parent-first, with no duplicates.
	GetBoundaryNodes(headerHash types.HeaderHash, keyStart types.StateKey, keyEnd types.StateKey, maxSize uint32) ([]types.BoundaryNode, error)
}
