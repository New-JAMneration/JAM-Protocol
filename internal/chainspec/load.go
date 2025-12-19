package chainspec

import (
	"encoding/json"
	"fmt"
)

func LoadFromBytes(b []byte) (*ChainSpec, error) {
	var cs ChainSpec
	if err := json.Unmarshal(b, &cs); err != nil {
		return nil, fmt.Errorf("chainspec: unmarshal json: %w", err)
	}
	if err := cs.Validate(); err != nil {
		return nil, err
	}
	return &cs, nil
}
