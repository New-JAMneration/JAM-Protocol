package store

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestGetInitGenesisBlock(t *testing.T) {
	testFile := "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.json"
	block, err := GetInitGenesisBlock(testFile)
	if err != nil {
		t.Errorf("Error loading genesis block: %v", err)
		return
	}

	expectedBlock := types.Block{
		Header: types.Header{
			Parent:          types.HeaderHash(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			ParentStateRoot: types.StateRoot(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			ExtrinsicHash:   types.OpaqueHash(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			Slot:            types.TimeSlot(0),
			EpochMark: &types.EpochMark{
				Entropy:        types.Entropy(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
				TicketsEntropy: types.Entropy(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
				Validators: []types.BandersnatchPublic{
					types.BandersnatchPublic(hexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
					types.BandersnatchPublic(hexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
					types.BandersnatchPublic(hexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
					types.BandersnatchPublic(hexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
					types.BandersnatchPublic(hexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
					types.BandersnatchPublic(hexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
				},
			},
			TicketsMark:   nil,
			OffendersMark: nil,
			AuthorIndex:   types.ValidatorIndex(65535),
			EntropySource: types.BandersnatchVrfSignature(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Seal:          types.BandersnatchVrfSignature(hexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		},
		Extrinsic: types.Extrinsic{},
	}

	if !reflect.DeepEqual(block, expectedBlock) {
		t.Errorf("Genesis block does not match expected block")
	}
}

func hexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		fmt.Printf("failed to decode hex string: %v", err)
	}
	return bytes
}
