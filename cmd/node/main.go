package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/urfave/cli/v3"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	AppVersion = "v0.1.0"
)

var cmd = &cli.Command{
	Name:        "node",
	Usage:       "JAM Node Command Line Interface",
	Description: "JAM Node Command Line Interface",
	Version:     AppVersion,
	Authors:     []any{"New JAMneration"},
	Action:      node,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to configuration file",
			Value:       "example.json",
			Destination: &configPath,
		},
		&cli.StringFlag{
			Name:        "mode",
			Aliases:     []string{"m"},
			Usage:       "Node mode: tiny or full or custom",
			Value:       "tiny",
			Destination: &mode,
		},
		&cli.StringFlag{
			Name:        "chain",
			Usage:       "Path to chainspec JSON",
			Destination: &chainPath,
			// example cmd line: --chain cmd/node/test_data/dev.chainspec.json
		},
	},
	Commands: []*cli.Command{
		exampleCmd,
		testCmd,
	},
}

var (
	configPath string
	chainPath  string
	mode       string
)

func init() {
	cmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		// Set Linux container-aware GOMAXPROCS. No-op on non-Linux systems.
		// Ref: https://go.dev/blog/container-aware-gomaxprocs
		// This is needed until upgrading Go version 1.25 or higher.
		maxprocs.Set()
		return ctx, nil
	}
}

func node(ctx context.Context, cmd *cli.Command) error {
	config.InitConfig(configPath, mode)
	SetupJAMProtocol(chainPath)
	return nil
}

func SetupJAMProtocol(chainPath string) {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
	log.Println("🚀 Start JAM Protocol")

	s := store.GetInstance()
	ctx := context.Background()

	if chainPath != "" {
		spec, err := store.GetChainSpecFromJson(chainPath)
		if err != nil {
			log.Fatalf("failed to load chainspec: %v", err)
		}

		// log.Printf("types protocol params (before apply):\n%s", types.SnapshotProtocolParams().JSON())
		pp, err := spec.ParseProtocolParameters()
		if err != nil {
			log.Fatalf("failed to decode protocol_parameters into struct: %v", err)
		}

		if err := types.ApplyProtocolParameters(pp); err != nil {
			log.Fatalf("invalid protocol_parameters: %v", err)
		}
		// log.Printf("types protocol params (after apply):\n%s", types.SnapshotProtocolParams().JSON())
		hdrBytes, err := spec.GenesisHeaderBytes()
		if err != nil {
			log.Fatalf("failed to decode genesis_header hex: %v", err)
		}

		hdr, err := store.DecodeHeaderFromBin(hdrBytes)
		if err != nil {
			log.Fatalf("failed to decode genesis_header bytes: %v", err)
		}

		kvs, err := spec.GenesisStateKeyVals()
		if err != nil {
			log.Fatalf("failed to decode genesis_state: %v", err)
		}

		genesisHash, stateRoot, err := s.SeedGenesisToBackend(ctx, *hdr, kvs)
		if err != nil {
			log.Fatalf("failed to seed genesis to backend: %v", err)
		}

		genesisBlock := types.Block{
			Header: *hdr,
			Extrinsic: types.Extrinsic{
				Tickets:    types.TicketsExtrinsic{},
				Preimages:  types.PreimagesExtrinsic{},
				Guarantees: types.GuaranteesExtrinsic{},
				Assurances: types.AssurancesExtrinsic{},
				Disputes:   types.DisputesExtrinsic{},
			},
		}

		s.GenerateGenesisBlock(genesisBlock)

		log.Printf("✅ Genesis seeded")
		log.Printf("  genesis_hash: 0x%s", hex.EncodeToString(genesisHash[:]))
		log.Printf("  genesis_state_root: 0x%s", hex.EncodeToString(stateRoot[:]))

		log.Printf("  header.parent: 0x%s", hex.EncodeToString(genesisBlock.Header.Parent[:]))
		log.Printf("  header.parent_state_root: 0x%s", hex.EncodeToString(genesisBlock.Header.ParentStateRoot[:]))

		log.Printf("  redis: state_root:%s -> %s",
			hex.EncodeToString(genesisHash[:]),
			hex.EncodeToString(stateRoot[:]),
		)
		log.Printf("  redis: state_data:%s -> (encoded StateKeyVals merkle input)",
			hex.EncodeToString(stateRoot[:]),
		)

		return
	}

	// no --chain: dummy genesis
	hash := "5c743dbc514284b2ea57798787c5a155ef9d7ac1e9499ec65910a7a3d65897b7"
	byteArray, _ := hex.DecodeString(hash)
	genesisBlock := types.Block{
		Header: types.Header{
			// hash string to jamTypes.HeaderHash
			Parent:          types.HeaderHash(byteArray),
			ParentStateRoot: types.StateRoot{},
			ExtrinsicHash:   types.OpaqueHash{},
			Slot:            0,
			EpochMark:       nil,
			TicketsMark:     nil,
			OffendersMark:   types.OffendersMark{},
			AuthorIndex:     0,
			EntropySource:   types.BandersnatchVrfSignature{},
			Seal:            types.BandersnatchVrfSignature{},
		},
		Extrinsic: types.Extrinsic{
			Tickets:    types.TicketsExtrinsic{},
			Preimages:  types.PreimagesExtrinsic{},
			Guarantees: types.GuaranteesExtrinsic{},
			Assurances: types.AssurancesExtrinsic{},
			Disputes:   types.DisputesExtrinsic{},
		},
	}

	s.GenerateGenesisBlock(genesisBlock)

	log.Println("Genesis block parent header hash:")
	log.Printf("0x%s\n", hex.EncodeToString(s.GetBlocks()[0].Header.Parent[:]))
}

func main() {
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
