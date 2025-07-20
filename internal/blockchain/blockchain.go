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
	// StoreBlockByHash stores a block by its hash
	StoreBlockByHash(types.HeaderHash, *types.Block) error
	// GetLatestBlock returns the latest block.
	GetLatestBlock() types.Block
}
