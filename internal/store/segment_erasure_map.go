package store

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type segmentErasureValue struct {
	ErasureRoot types.OpaqueHash
	ExpiresAt   int64
}

func (repo *Repository) GetSegmentErasureMap(r database.Reader, segmentRoot types.OpaqueHash) (types.OpaqueHash, error) {
	data, found, err := r.Get(segmentErasureKey(segmentRoot))
	if err != nil {
		return types.OpaqueHash{}, err
	}
	if !found || data == nil {
		return types.OpaqueHash{}, nil
	}

	if len(data) < 40 {
		return types.OpaqueHash{}, fmt.Errorf("invalid segment erasure data length: got %d, want at least 40", len(data))
	}

	var value segmentErasureValue
	copy(value.ErasureRoot[:], data[:32])
	value.ExpiresAt = int64(binary.BigEndian.Uint64(data[32:40]))

	if time.Now().Unix() > value.ExpiresAt {
		return types.OpaqueHash{}, nil
	}

	return value.ErasureRoot, nil
}

func (repo *Repository) SetSegmentErasureMap(w database.Writer, segmentRoot, erasureRoot types.OpaqueHash, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl).Unix()

	data := make([]byte, 40)
	copy(data[:32], erasureRoot[:])
	binary.BigEndian.PutUint64(data[32:40], uint64(expiresAt))

	return w.Put(segmentErasureKey(segmentRoot), data)
}
