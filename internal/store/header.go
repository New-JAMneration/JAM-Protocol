package store

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func (repo *Repository) GetCanonicalHash(r database.Reader, slot types.TimeSlot) (types.HeaderHash, error) {
	data, found, err := r.Get(canocicalHeaderHashKey(repo.encoder, slot))
	if err != nil {
		return types.HeaderHash{}, err
	}
	if !found {
		return types.HeaderHash{}, fmt.Errorf("canonical hash not found for slot %d", slot)
	}
	return types.HeaderHash(data), nil
}

func (repo *Repository) SaveCanonicalHash(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	return w.Put(canocicalHeaderHashKey(repo.encoder, slot), hash[:])
}

func (repo *Repository) GetFinalizedHash(r database.Reader) (types.HeaderHash, error) {
	data, _, err := r.Get(finalizedHeaderHashPrefix)
	if err != nil {
		return types.HeaderHash{}, err
	}
	return types.HeaderHash(data), nil
}

func (repo *Repository) SaveFinalizedHash(w database.Writer, hash types.HeaderHash) error {
	return w.Put(finalizedHeaderHashPrefix, hash[:])
}

func (repo *Repository) GetHeader(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Header, error) {
	encoded, found, err := r.Get(headerKey(repo.encoder, slot, hash))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("header not found for hash %x and slot %d", hash, slot)
	}

	header := &types.Header{}
	if err := repo.decoder.Decode(encoded, header); err != nil {
		return nil, err
	}

	return header, nil
}

func (repo *Repository) SaveHeader(w database.Writer, header *types.Header) (types.HeaderHash, error) {
	encoded, err := repo.encoder.Encode(header)
	if err != nil {
		return types.HeaderHash{}, err
	}

	var (
		slot = header.Slot
		hash = types.HeaderHash(hash.Blake2bHash(encoded))
	)

	if err := repo.SaveHeaderTimeSlot(w, hash, slot); err != nil {
		return types.HeaderHash{}, err
	}

	if err := w.Put(headerKey(repo.encoder, slot, hash), encoded); err != nil {
		return types.HeaderHash{}, err
	}

	return hash, nil
}

func (repo *Repository) DeleteHeader(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	if err := repo.DeleteHeaderTimeSlot(w, hash); err != nil {
		return err
	}

	return w.Delete(headerKey(repo.encoder, slot, hash))
}

func (repo *Repository) GetHeaderTimeSlot(r database.Reader, hash types.HeaderHash) (types.TimeSlot, error) {
	encoded, found, err := r.Get(headerTimeSlotKey(hash))
	if err != nil {
		return types.TimeSlot(0), err
	}
	if !found {
		return types.TimeSlot(0), fmt.Errorf("time slot not found for header %x", hash)
	}

	var slot types.TimeSlot

	if err := repo.decoder.Decode(encoded, &slot); err != nil {
		return types.TimeSlot(0), err
	}

	return slot, nil
}

func (repo *Repository) SaveHeaderTimeSlot(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	encoded, err := repo.encoder.Encode(&slot)
	if err != nil {
		return err
	}
	return w.Put(headerTimeSlotKey(hash), encoded)
}

func (repo *Repository) DeleteHeaderTimeSlot(w database.Writer, hash types.HeaderHash) error {
	return w.Delete(headerTimeSlotKey(hash))
}

func (repo *Repository) GetHeaderHashesByTimeSlot(r database.Iterable, slot types.TimeSlot) ([]types.HeaderHash, error) {
	headerKeyPrefix := headerKeyPrefix(repo.encoder, slot)
	iter, err := r.NewIterator(headerKeyPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var hashes []types.HeaderHash
	for iter.Next() {
		hash := hash.Blake2bHash(iter.Value())
		hashes = append(hashes, types.HeaderHash(hash))
	}

	return hashes, nil
}
