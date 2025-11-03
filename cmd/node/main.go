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
	},
	Commands: []*cli.Command{
		exampleCmd,
		testCmd,
	},
}

var (
	configPath string
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
	SetupJAMProtocol()
	return nil
}

func SetupJAMProtocol() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
	log.Println("🚀 Start JAM Protocol")

	// Initialize global store
	s := store.GetInstance()

	// TODO: Read genesis block from jamtestvector bin file
	// Generate genesis block
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
