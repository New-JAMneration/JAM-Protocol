package store

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func GetBlock(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Block, bool, error) {
	header, found, err := GetHeader(r, hash, slot)
	if err != nil {
		return nil, found, err
	}
	if !found {
		return nil, false, nil
	}

	extrinsic, found, err := GetExtrinsic(r, hash, slot)
	if err != nil {
		return nil, found, err
	}
	if !found {
		return nil, false, nil
	}

	return &types.Block{
		Header:    *header,
		Extrinsic: *extrinsic,
	}, true, nil
}

// GenesisBlock reads and returns the genesis block from the database.
// Genesis block must exist when this function is called.
// func GetGenesisBlock(r database.Reader) (*types.Block, error) {
// hash, _, err := GetCanonicalHash(r, 0)
// if err != nil {
// return nil, err
// }
// block, _, err := GetBlock(r, hash, 0)
// if err != nil {
// return nil, err
// }
// return block, nil
// }

func SaveBlock(w database.Writer, block *types.Block) error {
	// TODO: optimize header hashing by passing hash from outside
	if err := SaveHeader(w, &block.Header); err != nil {
		return err
	}

	encoded, err := types.NewEncoder().Encode(&block.Header)
	if err != nil {
		return err
	}
	headerHash := types.HeaderHash(hash.Blake2bHash(encoded))

	err = SaveExtrinsic(w, headerHash, block.Header.Slot, &block.Extrinsic)
	if err != nil {
		return err
	}

	return nil
}

func DeleteBlock(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	if err := DeleteHeader(w, hash, slot); err != nil {
		return err
	}
	if err := DeleteExtrinsic(w, hash, slot); err != nil {
		return err
	}
	return nil
}

func GetExtrinsic(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Extrinsic, bool, error) {
	encoded, found, err := r.Get(extrinsicKey(slot, hash))
	if err != nil {
		return nil, found, err
	}
	if !found {
		return nil, false, nil
	}

	extrinsic := &types.Extrinsic{}

	decoder := types.NewDecoder()
	if err := decoder.Decode(encoded, extrinsic); err != nil {
		return nil, false, err
	}

	return extrinsic, true, nil
}

func SaveExtrinsic(w database.Writer, hash types.HeaderHash, slot types.TimeSlot, extrinsic *types.Extrinsic) error {
	encoded, err := types.NewEncoder().Encode(extrinsic)
	if err != nil {
		return err
	}

	return w.Put(extrinsicKey(slot, hash), encoded)
}

func DeleteExtrinsic(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	return w.Delete(extrinsicKey(slot, hash))
}
