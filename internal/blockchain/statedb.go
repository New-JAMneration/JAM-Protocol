package blockchain

import (
	"context"

	"github.com/New-JAMneration/JAM-Protocol/internal/db"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

var (
	HeaderKeyPrefix = []byte("h")
)

type StateDB struct {
	db db.KeyValueDB
}

func (s *StateDB) WriteHeader(ctx context.Context, head types.OpaqueHash) error {
	key := append(HeaderKeyPrefix, head[:]...)
	return s.db.Set(key, []byte{})
}
