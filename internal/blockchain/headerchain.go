package blockchain

import (
	"sync/atomic"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type HeaderChain struct {
	database database.Database

	genesisHeader     *types.Header
	currentHeaderHash atomic.Pointer[types.HeaderHash]
	currentHeader     atomic.Pointer[types.Header]

	// headerCache *lru.Cache[common.Hash, *types.Header]
	// numberCache *lru.Cache[common.Hash, uint64] // most recent block numbers
}

// func NewHeaderChain(db database.Database) *HeaderChain {
// hc := &HeaderChain{
// database: db,
// }

// hc.genesisHeader = hc.GetHeaderByNumber(0)

// return hc
// }

// func (hc *HeaderChain) GetHeaderByNumber(number uint32) *types.Header {
// hash := hc.ReadCanonicalHash(number)
// if hash == (common.Hash{}) {
// return nil
// }
// return hc.GetHeader(hash, number)
// }

// func (hc *HeaderChain) GetHeader(hash common.Hash, number uint64) *types.Header {
// header := hc.ReadHeader(hc.chainDb, hash, number)
// if header == nil {
// return nil
// }
// // Cache the found header for next time and return
// hc.headerCache.Add(hash, header)
// return header
// }
