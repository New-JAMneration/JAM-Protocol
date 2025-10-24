package store

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func ReadCanonicalHash(r database.Reader, slot types.TimeSlot) (types.HeaderHash, bool, error) {
	data, found, err := r.Get(canocicalHeaderHashKey(slot))
	if err != nil {
		return types.HeaderHash{}, found, err
	}
	if !found {
		return types.HeaderHash{}, found, nil
	}
	return types.HeaderHash(data), found, nil
}

func WriteCanonicalHash(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	return w.Put(canocicalHeaderHashKey(slot), hash[:])
}

func ReadHeader(r database.Reader, hash types.HeaderHash, slot types.TimeSlot) (*types.Header, bool, error) {
	encoded, found, err := r.Get(headerKey(slot, hash))
	if err != nil {
		return nil, found, err
	}
	if !found {
		return nil, false, nil
	}

	header := &types.Header{}
	decoder := types.NewDecoder()
	if err := decoder.Decode(encoded, header); err != nil {
		return nil, false, err
	}

	return header, true, nil
}

func WriteHeader(w database.Writer, header *types.Header) error {
	encoded, err := types.NewEncoder().Encode(header)
	if err != nil {
		return err
	}

	var (
		slot = header.Slot
		// TODO: implement Hash method for Header type
		hash = types.HeaderHash(hash.Blake2bHash(encoded))
	)

	if err := WriteHeaderTimeSlot(w, hash, slot); err != nil {
		return err
	}
	if err := w.Put(headerKey(slot, hash), encoded); err != nil {
		return err
	}
	return nil
}

func DeleteHeader(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	if err := DeleteHeaderTimeSlot(w, hash); err != nil {
		return err
	}

	return w.Delete(headerKey(slot, hash))
}

func ReadHeaderTimeSlot(r database.Reader, hash types.HeaderHash) (types.TimeSlot, bool, error) {
	encoded, found, err := r.Get(headerTimeSlotKey(hash))
	if err != nil {
		return types.TimeSlot(0), found, err
	}
	if !found {
		return types.TimeSlot(0), false, nil
	}

	var slot types.TimeSlot

	decoder := types.NewDecoder()
	if err := decoder.Decode(encoded, &slot); err != nil {
		return types.TimeSlot(0), false, err
	}

	return slot, true, nil
}

func WriteHeaderTimeSlot(w database.Writer, hash types.HeaderHash, slot types.TimeSlot) error {
	encoded, err := types.NewEncoder().Encode(&slot)
	if err != nil {
		return err
	}
	return w.Put(headerTimeSlotKey(hash), encoded)
}

func DeleteHeaderTimeSlot(w database.Writer, hash types.HeaderHash) error {
	return w.Delete(headerTimeSlotKey(hash))
}

func ReadHeaderHashesByTimeSlot(r database.Iterable, slot types.TimeSlot) ([]types.HeaderHash, error) {
	encoded, _ := types.NewEncoder().Encode(&slot)
	iter, err := r.NewIterator(headerPrefix, encoded)
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
