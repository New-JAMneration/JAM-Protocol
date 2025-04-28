package erasurecoding

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

type ErasureCoding struct {
	rs          reedsolomon.Encoder
	dataShards  int
	totalShards int
	chunkSize   int
}

func NewErasureCoding(dataShards int, totalShards int, k int) (*ErasureCoding, error) {
	if k <= 0 {
		return nil, fmt.Errorf("invalid k")
	}

	chunkSize := 2 * k
	rs, err := reedsolomon.New(dataShards, totalShards-dataShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}
	return &ErasureCoding{
		rs:          rs,
		dataShards:  dataShards,
		totalShards: totalShards,
		chunkSize:   chunkSize,
	}, nil
}

// Encode: encode data into shards
func (ec *ErasureCoding) Encode(data []byte) ([][]byte, error) {
	paddedData := make([]byte, ec.dataShards*ec.chunkSize)
	copy(paddedData, data)

	shards, err := ec.rs.Split(paddedData)
	if err != nil {
		return nil, fmt.Errorf("failed to split data into shards: %w", err)
	}

	if err := ec.rs.Encode(shards); err != nil {
		return nil, fmt.Errorf("failed to encode shards: %w", err)
	}

	return shards, nil
}

// Decode: decode shards into original data
func (ec *ErasureCoding) Decode(shards [][]byte) ([]byte, error) {
	if len(shards) != ec.totalShards {
		return nil, fmt.Errorf("invalid number of shards: expected %d, got %d", ec.totalShards, len(shards))
	}

	if err := ec.rs.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("failed to reconstruct shards: %w", err)
	}

	var buffer bytes.Buffer
	if err := ec.rs.Join(&buffer, shards, ec.dataShards*ec.chunkSize); err != nil {
		return nil, fmt.Errorf("failed to join shards: %w", err)
	}

	return buffer.Bytes(), nil
}
