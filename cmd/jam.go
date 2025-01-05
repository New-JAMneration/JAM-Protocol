package cmd

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
)

func init() {
	jamCmd := &cli.Command{
		Use:   "start",
		Short: "Start JAM Protocol",
		Long:  "Start JAM Protocol",
		Run: func(args []string) {
			SetupJAMProtocol()
		},
	}

	cli.AddCommand(jamCmd)
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
