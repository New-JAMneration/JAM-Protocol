package store

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/chainspec"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func DecodeHeaderFromBin(data []byte) (*types.Header, error) {
	h := &types.Header{}
	decoder := types.NewDecoder()
	if err := decoder.Decode(data, h); err != nil {
		return nil, fmt.Errorf("failed to decode header: %v", err)
	}
	return h, nil
}

func GetChainSpecFromJson(filename string) (*chainspec.ChainSpec, error) {
	data, err := GetBytesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from file: %v", err)
	}
	spec, err := chainspec.LoadFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load chainspec: %v", err)
	}
	return spec, nil
}
