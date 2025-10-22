package blockchain

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/pkg/errors"
)

// Just the template for the interface
// TODO: Implement the Blockchain interface.
// Blockchain defines the required interface for retrieving blocks.
type Blockchain interface {
	// GetBlockNumber returns the block number for the given block hash.
	GetBlockTimeSlot(types.HeaderHash) (types.TimeSlot, error)
	// GetBlockHashByTimeSlot returns candidate block hashes for the specified block number.
	GetBlockHashByTimeSlot(slot types.TimeSlot) ([]types.HeaderHash, error)
	// GetBlockByHash returns a block for the given block hash.
	GetBlockByHash(types.HeaderHash) (*types.Block, error)
	// GenesisBlockHash returns the genesis block hash.
	GenesisBlockHash() types.HeaderHash
}

type blockchain struct {
	db database.Database

	genesisHeader *types.Header
}

func NewBlockchain(db database.Database) (Blockchain, error) {
	genesis, err := GetGenesisBlock()
	if err != nil {
		return nil, err
	}

	return &blockchain{
		db:            db,
		genesisHeader: &genesis.Header,
	}, nil
}

func (bc *blockchain) GetBlockTimeSlot(hash types.HeaderHash) (types.TimeSlot, error) {

	return types.TimeSlot(0), nil
}

func (bc *blockchain) GetBlockHashByTimeSlot(slot types.TimeSlot) ([]types.HeaderHash, error) {

	return nil, nil
}

func (bc *blockchain) GetBlock(hash types.HeaderHash, slot types.TimeSlot) (*types.Block, error) {
	// TODO: cache blocks in memory if it is already finalized
	block, found, err := store.ReadBlock(bc.db, hash, slot)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Wrap(ErrBlockNotFound, fmt.Sprintf("hash: %s, timeslot: %d", hash, slot))
	}
	return block, nil
}

func (bc *blockchain) GetBlockByTimeSlot(slot types.TimeSlot) (*types.Block, error) {
	hash, found, err := store.ReadCanonicalHash(bc.db, slot)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Wrap(ErrBlockNotFound, fmt.Sprintf("timeslot: %d", slot))
	}

	return bc.GetBlock(hash, slot)
}

func (bc *blockchain) GetBlockByHash(hash types.HeaderHash) (*types.Block, error) {
	slot, found, err := store.ReadHeaderTimeSlot(bc.db, hash)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Wrap(ErrBlockNotFound, fmt.Sprintf("hash: %s", hash))
	}

	return bc.GetBlock(hash, slot)
}

func (bc *blockchain) GenesisBlockHash() types.HeaderHash {
	encoded, _ := types.NewEncoder().Encode(bc.genesisHeader)
	return types.HeaderHash(hash.Blake2bHash(encoded))
}
