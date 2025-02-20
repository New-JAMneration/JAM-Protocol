package store

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestGetInitGenesisBlock(t *testing.T) {
	testFile := "../../pkg/test_data/jamtestnet/assurances/blocks/407413_000.json"
	block, err := GetInitGenesisBlock(testFile)

	if err != nil {
		t.Errorf("Error loading genesis block: %v", err)
		return
	}

	expectedBlock := types.Block{
		Header: types.Header{
			Parent:          types.HeaderHash(hexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
			ParentStateRoot: types.StateRoot(hexToBytes("0x57b3782db21156854f214dc7cc88da9bdad250e90813437a1279a9c4af9b5003")),
			ExtrinsicHash:   types.OpaqueHash(hexToBytes("0xdc080ad182cb9ff052a1ca8ecbc51164264efc7dd6debaaa45764950f843acb8")),
			Slot:            types.TimeSlot(4888956),
			EpochMark: &types.EpochMark{
				Entropy:        types.Entropy(hexToBytes("0x6f6ad2224d7d58aec6573c623ab110700eaca20a48dc2965d535e466d524af2a")),
				TicketsEntropy: types.Entropy(hexToBytes("0x835ac82bfa2ce8390bb50680d4b7a73dfa2a4cff6d8c30694b24a605f9574eaf")),
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
			OffendersMark: types.OffendersMark{},
			AuthorIndex:   types.ValidatorIndex(2),
			EntropySource: types.BandersnatchVrfSignature(hexToBytes("0x8fc34b4c24f74f16ef7b13fff47ab7b84e6ce6ccae23870ee63d0b9c39261eb84dfed75e12c38d7599bf39f443f5fa5ce583a10b96d1184b450752a93c8c6a1566a41c9aa8ac0645aef85b6f6c0c6c00bc7421cd472cac7011abc190116e8b0d")),
			Seal:          types.BandersnatchVrfSignature(hexToBytes("0xc0909e69767bdce7476a497eabc11c3ab0892703755c7ed244c8144424fa397163ece7a751fc0b845ce6b597950fc2fae904564b08c066b33670a8c60d6bf8074ae2fd6da177b29f9cb7fdc97cd30b57263a469c58749c672c43b5cf5f178c10")),
		},
		Extrinsic: types.Extrinsic{},
	}

	//copmare the two blocks with deep equal reflect.DeepEqual()
	if reflect.DeepEqual(block, expectedBlock) {
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
