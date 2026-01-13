package store

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func (repo *Repository) GetBlock(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Block, error) {
	header, err := repo.GetHeader(r, hash, slot)
	if err != nil {
		return nil, err
	}

	extrinsic, err := repo.GetExtrinsic(r, hash, slot)
	if err != nil {
		return nil, err
	}

	return &types.Block{
		Header:    *header,
		Extrinsic: *extrinsic,
	}, nil
}

func (repo *Repository) SaveBlock(w database.Writer, block *types.Block) error {
	headerHash, err := repo.SaveHeader(w, &block.Header)
	if err != nil {
		return err
	}

	err = repo.SaveExtrinsic(w, headerHash, block.Header.Slot, &block.Extrinsic)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) DeleteBlock(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	if err := repo.DeleteHeader(w, hash, slot); err != nil {
		return err
	}
	if err := repo.DeleteExtrinsic(w, hash, slot); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) GetExtrinsic(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Extrinsic, error) {
	encoded, found, err := r.Get(extrinsicKey(repo.encoder, slot, hash))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("extrinsic not found for hash %x and slot %d", hash, slot)
	}

	extrinsic := &types.Extrinsic{}
	if err := repo.decoder.Decode(encoded, extrinsic); err != nil {
		return nil, err
	}

	return extrinsic, nil
}

func (repo *Repository) SaveExtrinsic(w database.Writer, hash types.HeaderHash, slot types.TimeSlot, extrinsic *types.Extrinsic) error {
	encoded, err := types.NewEncoder().Encode(extrinsic)
	if err != nil {
		return err
	}

	return w.Put(extrinsicKey(repo.encoder, slot, hash), encoded)
}

func (repo *Repository) DeleteExtrinsic(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	return w.Delete(extrinsicKey(repo.encoder, slot, hash))
}
