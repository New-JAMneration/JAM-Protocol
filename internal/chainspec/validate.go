package chainspec

import (
	"fmt"
	"strings"
)

func (cs *ChainSpec) Validate() error {
	if cs == nil {
		return fmt.Errorf("chainspec: nil")
	}

	if cs.ID == "" {
		return fmt.Errorf("chainspec: id is required")
	}

	// bootnodes (optional, but validate format if present)
	for i, bn := range cs.Bootnodes {
		if _, err := ParseBootnode(bn); err != nil {
			return fmt.Errorf("chainspec: bootnodes[%d]: %w", i, err)
		}
	}

	if strings.TrimSpace(cs.GenesisHeader) == "" {
		return fmt.Errorf("chainspec: genesis_header is required")
	}

	if _, err := parseHexString(cs.GenesisHeader); err != nil {
		return fmt.Errorf("chainspec: genesis_header: %w", err)
	}

	if strings.TrimSpace(cs.ProtocolParamsHex) == "" {
		return fmt.Errorf("chainspec: protocol_parameters is required")
	}

	if _, err := parseHexString(cs.ProtocolParamsHex); err != nil {
		return fmt.Errorf("chainspec: protocol_parameters: %w", err)
	}

	if len(cs.GenesisState) == 0 {
		return fmt.Errorf("chainspec: genesis_state is required")
	}

	for k, v := range cs.GenesisState {
		keyBytes, err := parseHexString(k)
		if err != nil {
			return fmt.Errorf("chainspec: genesis_state key %q: %w", k, err)
		}
		if len(keyBytes) != 31 {
			return fmt.Errorf(
				"chainspec: genesis_state key must be 31 bytes (62 hex chars), got %d bytes for key %q",
				len(keyBytes), k,
			)
		}

		if _, err := parseHexString(v); err != nil {
			return fmt.Errorf("chainspec: genesis_state[%q] value: %w", k, err)
		}
	}

	return nil
}
