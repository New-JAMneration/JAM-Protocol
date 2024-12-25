package cmd

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
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
	genesisBlock := jamTypes.Block{
		Header: jamTypes.Header{
			// hash string to jamTypes.HeaderHash
			Parent:          jamTypes.HeaderHash(byteArray),
			ParentStateRoot: jamTypes.StateRoot{},
			ExtrinsicHash:   jamTypes.OpaqueHash{},
			Slot:            0,
			EpochMark:       nil,
			TicketsMark:     nil,
			OffendersMark:   jamTypes.OffendersMark{},
			AuthorIndex:     0,
			EntropySource:   jamTypes.BandersnatchVrfSignature{},
			Seal:            jamTypes.BandersnatchVrfSignature{},
		},
		Extrinsic: jamTypes.Extrinsic{
			Tickets:    jamTypes.TicketsExtrinsic{},
			Preimages:  jamTypes.PreimagesExtrinsic{},
			Guarantees: jamTypes.GuaranteesExtrinsic{},
			Assurances: jamTypes.AssurancesExtrinsic{},
			Disputes:   jamTypes.DisputesExtrinsic{},
		},
	}

	s.GenerateGenesisBlock(genesisBlock)

	log.Println("Genesis block parent header hash:")
	log.Printf("0x%s\n", hex.EncodeToString(s.GetBlocks()[0].Header.Parent[:]))
}
