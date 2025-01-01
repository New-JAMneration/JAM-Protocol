package header

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// {
//   "parent": "0x5c743dbc514284b2ea57798787c5a155ef9d7ac1e9499ec65910a7a3d65897b7",
//   "parent_state_root": "0x2591ebd047489f1006361a4254731466a946174af02fe1d86681d254cfd4a00b",
//   "extrinsic_hash": "0x74a9e79d2618e0ce8720ff61811b10e045c02224a09299f04e404a9656e85c81",
//   "slot": 42,
//   "epoch_mark": {
//     "entropy": "0xae85d6635e9ae539d0846b911ec86a27fe000f619b78bcac8a74b77e36f6dbcf",
//     "tickets_entropy": "0x333a7e328f0c4183f4b947e1d8f68aa4034f762e5ecdb5a7f6fbf0afea2fd8cd",
//     "validators": [
//       "0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d",
//       "0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0",
//       "0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc",
//       "0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33",
//       "0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3",
//       "0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d"
//     ]
//   },
//   "tickets_mark": null,
//   "offenders_mark": [
//     "0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
//   ],
//   "author_index": 3,
//   "entropy_source": "0xae85d6635e9ae539d0846b911ec86a27fe000f619b78bcac8a74b77e36f6dbcf49a52360f74a0233cea0775356ab0512fafff0683df08fae3cb848122e296cbc50fed22418ea55f19e55b3c75eb8b0ec71dcae0d79823d39920bf8d6a2256c5f",
//   "seal": "0x31dc5b1e9423eccff9bccd6549eae8034162158000d5be9339919cc03d14046e6431c14cbb172b3aed702b9e9869904b1f39a6fe1f3e904b0fd536f13e8cac496682e1c81898e88e604904fa7c3e496f9a8771ef1102cc29d567c4aad283f7b0"
// }

type headerJSON struct {
	Parent          string `json:"parent"`
	ParentStateRoot string `json:"parent_state_root"`
	ExtrinsicHash   string `json:"extrinsic_hash"`
	Slot            int    `json:"slot"`
	EpochMark       struct {
		Entropy        string   `json:"entropy"`
		TicketsEntropy string   `json:"tickets_entropy"`
		Validators     []string `json:"validators"`
	} `json:"epoch_mark"`
	TicketsMark   interface{} `json:"tickets_mark"`
	OffendersMark []string    `json:"offenders_mark"`
	AuthorIndex   int         `json:"author_index"`
	EntropySource string      `json:"entropy_source"`
}

func loadTestData() types.Header {
	jsonFile := "test_data/header.json"

	// Open the JSON file
	file, err := os.Open(jsonFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Decode the JSON file
	var header headerJSON
	err = json.Unmarshal(file, &header)
}

func TestNewHeaderController(t *testing.T) {
	t.Run("NewHeaderController", func(t *testing.T) {
		hc := NewHeaderController()
		if hc == nil {
			t.Errorf("NewHeaderController() = %v; want not nil", hc)
		}
	})
}

func TestHeaderController_SetHeader(t *testing.T) {
	t.Run("SetHeader", func(t *testing.T) {
		hc := NewHeaderController()

		inputHeader := types.Header{}
		hc.SetHeader(inputHeader)

		outputHeader := hc.GetHeader()

		// compare struct values
		if !reflect.DeepEqual(inputHeader, outputHeader) {
			t.Errorf("SetHeader() = %v; want %v", outputHeader, inputHeader)
		}
	})
}

func TestHeaderController_GetHeader(t *testing.T) {
	t.Run("GetHeader", func(t *testing.T) {
		hc := NewHeaderController()

		inputHeader := types.Header{}
		hc.SetHeader(inputHeader)

		outputHeader := hc.GetHeader()

		// compare struct values
		if !reflect.DeepEqual(inputHeader, outputHeader) {
			t.Errorf("GetHeader() = %v; want %v", outputHeader, inputHeader)
		}
	})
}

func TestCreateParentHeaderHash(t *testing.T) {
	t.Run("CreateParentHeaderHash", func(t *testing.T) {
		hc := NewHeaderController()

		// TODO: Load Parent Header from test data
		parentHeader := types.Header{}
		hc.CreateParentHeaderHash(parentHeader)
	})
}
