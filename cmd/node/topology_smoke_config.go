package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	validatorpkg "github.com/New-JAMneration/JAM-Protocol/internal/networking/validator"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type smokeValidatorEntry struct {
	Index   int    `json:"index"`
	Ed25519 string `json:"ed25519"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

type smokeTopologyConfig struct {
	Validators []smokeValidatorEntry `json:"validators"`
}

// applyTopologySmokeConfig seeds kappa/lambda/gamma_k with mock validator metadata when
// JAM_TOPOLOGY_SMOKE_CONFIG points at a JSON file (local smoke harness only).
func applyTopologySmokeConfig(chain *blockchain.ChainState) error {
	cfgPath := os.Getenv("JAM_TOPOLOGY_SMOKE_CONFIG")
	if cfgPath == "" || chain == nil {
		return nil
	}

	payload, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("read topology smoke config: %w", err)
	}

	var cfg smokeTopologyConfig
	if err := json.Unmarshal(payload, &cfg); err != nil {
		return fmt.Errorf("parse topology smoke config: %w", err)
	}
	if len(cfg.Validators) == 0 {
		return fmt.Errorf("topology smoke config: no validators")
	}

	set := make(types.ValidatorsData, 0, len(cfg.Validators))
	for _, entry := range cfg.Validators {
		pubBytes, err := hex.DecodeString(entry.Ed25519)
		if err != nil {
			return fmt.Errorf("validator %d ed25519 hex: %w", entry.Index, err)
		}
		if len(pubBytes) != len(types.Ed25519Public{}) {
			return fmt.Errorf("validator %d ed25519: want %d bytes, got %d", entry.Index, len(types.Ed25519Public{}), len(pubBytes))
		}
		host := entry.Host
		if host == "" {
			host = "127.0.0.1"
		}
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, entry.Port))
		if err != nil {
			return fmt.Errorf("validator %d addr: %w", entry.Index, err)
		}
		var pub types.Ed25519Public
		copy(pub[:], pubBytes)
		v, err := validatorpkg.ValidatorWithUDPAddr(pub, addr)
		if err != nil {
			return fmt.Errorf("validator %d metadata: %w", entry.Index, err)
		}
		set = append(set, v)
	}

	prior := chain.GetPriorStates()
	prior.SetKappa(set)
	prior.SetLambda(set)
	prior.SetGammaK(set)
	log.Printf("topology smoke: seeded %d validators with QUIC metadata from %s", len(set), cfgPath)
	return nil
}
