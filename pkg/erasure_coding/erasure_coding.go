package erasurecoding

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

const (
	DataShards  = 342
	TotalShards = 1023 // 342:1023 (Appendix H)
	ChunkSize   = 2    // octect pairs
)

type Shard struct {
	Index int
	Data  []byte
}

type ErasureCoding struct {
	rs reedsolomon.Encoder
}

func NewErasureCoding() (*ErasureCoding, error) {
	rs, err := reedsolomon.New(DataShards, TotalShards-DataShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}
	return &ErasureCoding{
		rs: rs,
	}, nil
}

// Encode: encode data into shards
func (ec *ErasureCoding) Encode(data []byte) ([]Shard, error) {
	// Pad data if necessary
	paddedData := make([]byte, DataShards*ChunkSize)
	copy(paddedData, data)

	shards, err := ec.rs.Split(paddedData)
	if err != nil {
		return nil, fmt.Errorf("failed to split data into shards: %w", err)
	}

	if err := ec.rs.Encode(shards); err != nil {
		return nil, fmt.Errorf("failed to encode shards: %w", err)
	}

	// Convert to Shard struct with index
	result := make([]Shard, len(shards))
	for i, shard := range shards {
		result[i] = Shard{
			Index: i,
			Data:  shard,
		}
	}

	return result, nil
}

// Decode: decode shards into original data
func (ec *ErasureCoding) Decode(shards []Shard) ([]byte, error) {

	// Use index to reconstruct raw shards
	rawShards := make([][]byte, TotalShards)
	for _, shard := range shards {
		if shard.Index < 0 || shard.Index >= TotalShards {
			return nil, fmt.Errorf("invalid shard index: %d", shard.Index)
		}
		rawShards[shard.Index] = shard.Data
	}

	if err := ec.rs.Reconstruct(rawShards); err != nil {
		return nil, fmt.Errorf("failed to reconstruct shards: %w", err)
	}

	// Join raw shards into original data
	var buffer bytes.Buffer
	if err := ec.rs.Join(&buffer, rawShards, DataShards*ChunkSize); err != nil {
		return nil, fmt.Errorf("failed to join shards: %w", err)
	}

	return buffer.Bytes(), nil
}
