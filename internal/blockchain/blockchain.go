package blockchain

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/pkg/errors"
)

// Just the template for the interface
// Blockchain defines the required interface for retrieving blocks.
type Blockchain interface {
	// GetBlockTimeSlot returns the block number for the given block hash.
	GetBlockTimeSlot(types.HeaderHash) (types.TimeSlot, error)
	// GetBlockHashByTimeSlot returns candidate block hashes for the specified block number.
	GetBlockHashesByTimeSlot(slot types.TimeSlot) ([]types.HeaderHash, error)
	// GetBlockByHash returns a block for the given block hash.
	GetBlockByHash(types.HeaderHash) (*types.Block, error)
	// GenesisBlockHash returns the genesis block hash.
	GenesisBlockHash() types.HeaderHash
}

type blockchain struct {
	db database.Database

	// genesisHeader *types.Header
}

func NewBlockchain(db database.Database) (Blockchain, error) {
	// genesis, err := GetGenesisBlock()
	// if err != nil {
	// return nil, err
	// }
	// genesisHash, found, err := store.ReadCanonicalHash(db, 0)
	// if err != nil {
	// return nil, err
	// }
	// if found {
	// // verify the genesis block hash matches the stored one
	// headerHash := func(header *types.Header) types.HeaderHash {
	// encoded, _ := types.NewEncoder().Encode(header)
	// return types.HeaderHash(hash.Blake2bHash(encoded))
	// }
	// if genesisHash != headerHash(&genesis.Header) {
	// return nil, errors.New("genesis block hash mismatch with the stored one")
	// }
	// } else {
	// batch := db.NewBatch()
	// if err = store.WriteBlock(batch, genesis); err != nil {
	// return nil, err
	// }
	// if err = store.WriteCanonicalHash(batch, genesisHash, 0); err != nil {
	// return nil, err
	// }
	// if err = batch.Commit(); err != nil {
	// return nil, err
	// }
	// }

	return &blockchain{
		db: db,
		// genesisHeader: &genesis.Header,
	}, nil
}

func (bc *blockchain) GetBlockTimeSlot(hash types.HeaderHash) (types.TimeSlot, error) {
	slot, found, err := store.ReadHeaderTimeSlot(bc.db, hash)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, errors.Wrap(ErrBlockNotFound, fmt.Sprintf("hash: %s", hash))
	}
	return slot, nil
}

func (bc *blockchain) GetBlockHashesByTimeSlot(slot types.TimeSlot) ([]types.HeaderHash, error) {
	hashes, err := store.ReadHeaderHashesByTimeSlot(bc.db, slot)
	if err != nil {
		return nil, err
	}
	return hashes, nil
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
	// TODO: implement

	// encoded, _ := types.NewEncoder().Encode(bc.genesisHeader)
	// return types.HeaderHash(hash.Blake2bHash(encoded))

	return types.HeaderHash{}
}
