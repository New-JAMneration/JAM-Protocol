package header

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func TestCreateParentHash(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet block test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		"fallback",
		"orderedaccumulation",
		"safrole",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join(utilities.JAM_TEST_NET_DIR, "data", dirName, "blocks")

		files, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		file_index := 0
		parent_hash := types.HeaderHash{}
		for _, file := range files {
			// Read the binary file
			binPath := filepath.Join(dir, file)
			binData, err := utilities.GetBytesFromFile(binPath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Decode the binary data
			decoder := types.NewDecoder()
			block := &types.Block{}
			err = decoder.Decode(binData, block)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Create a header hash
			encoded_header, err := types.NewEncoder().Encode(&block.Header)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Hash the header encoded data
			headerHash := types.HeaderHash(hash.Blake2bHash(encoded_header))

			if file_index != 0 {
				// Compare the parent hash with the header hash
				if block.Header.Parent != parent_hash {
					t.Errorf("Error: %v", err)
				}
			}

			// Store the parent hash for the next iteration
			parent_hash = headerHash
			file_index++
		}
	}

	// Reset the test mode
	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

func TestCreateExtrinsicHash(t *testing.T) {
	BACKUP_TEST_MODE := types.TEST_MODE
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  jamtestnet block test cases only support tiny mode")
	}

	dirNames := []string{
		"assurances",
		"fallback",
		"orderedaccumulation",
		"safrole",
	}

	for _, dirName := range dirNames {
		dir := filepath.Join(utilities.JAM_TEST_NET_DIR, "data", dirName, "blocks")

		files, err := utilities.GetTargetExtensionFiles(dir, utilities.BIN_EXTENTION)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		for _, file := range files {
			binPath := filepath.Join(dir, file)
			binData, err := utilities.GetBytesFromFile(binPath)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			decoder := types.NewDecoder()
			block := &types.Block{}
			err = decoder.Decode(binData, block)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Create a extrinsic hash with header controller
			hc := NewHeaderController()
			hc.CreateExtrinsicHash(block.Extrinsic)
			if err != nil {
				t.Errorf("Error: %v", err)
			}

			// Get the extrinsic hash from the header controller
			// TODO: We have to get the extrinsic hash from the store
			extrinsicHash := hc.GetHeader().ExtrinsicHash

			// Compare the extrinsic hash with the block's extrinsic hash
			if block.Header.ExtrinsicHash != extrinsicHash {
				t.Errorf("Error: %v", err)
				extrinsicHashHex := fmt.Sprintf("0x%x", extrinsicHash)
				blockExtrinsicHashHex := fmt.Sprintf("0x%x", block.Header.ExtrinsicHash)
				fmt.Println("MyHash: ", extrinsicHashHex)
				fmt.Println("Answer: ", blockExtrinsicHashHex)
			}
		}
	}

	if BACKUP_TEST_MODE == "tiny" {
		types.SetTinyMode()
	} else {
		types.SetFullMode()
	}
}

// import (
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"os"
// 	"reflect"
// 	"testing"
// 	"time"

// 	"github.com/New-JAMneration/JAM-Protocol/internal/store"
// 	"github.com/New-JAMneration/JAM-Protocol/internal/types"
// )

// // ParentHeaderHash, ExtrinsicHash, ParentStateRoot will be tested in the
// // block serialization

// func strToHex(str string) []byte {
// 	hexStr, err := hex.DecodeString(str[2:]) // 0x prefix
// 	if err != nil {
// 		panic(err)
// 	}
// 	return hexStr
// }

// func bytesToHex(data []byte) string {
// 	// concat 0x prefix
// 	return "0x" + hex.EncodeToString(data)
// }

// func GetHeaderString(h types.Header) string {
// 	var output string
// 	output += fmt.Sprintf("Parent: %s\n", bytesToHex([]byte(h.Parent[:])))
// 	output += fmt.Sprintf("ParentStateRoot: %s\n", bytesToHex([]byte(h.ParentStateRoot[:])))
// 	output += fmt.Sprintf("ExtrinsicHash: %s\n", bytesToHex([]byte(h.ExtrinsicHash[:])))
// 	output += fmt.Sprintf("Slot: %d\n", h.Slot)
// 	output += fmt.Sprintf("EpochMark: %s\n", "")
// 	if h.EpochMark != nil {
// 		output += fmt.Sprintf("EpochMark.Entropy: %s\n", bytesToHex([]byte(h.EpochMark.Entropy[:])))
// 		output += fmt.Sprintf("EpochMark.TicketsEntropy: %s\n", bytesToHex([]byte(h.EpochMark.TicketsEntropy[:])))

// 		output += fmt.Sprintf("EpochMark.Validators: %s\n", "")
// 		for _, v := range h.EpochMark.Validators {
// 			output += fmt.Sprintf("  %s\n", bytesToHex([]byte(v[:])))
// 		}
// 	}

// 	output += fmt.Sprintf("TicketsMark: %s\n", "")
// 	if h.TicketsMark != nil {
// 		for _, t := range *h.TicketsMark {
// 			output += fmt.Sprintf("  %s\n", bytesToHex([]byte(t.Id[:])))
// 		}
// 	}

// 	output += fmt.Sprintf("OffendersMark: %s\n", "")
// 	if h.OffendersMark != nil {
// 		for _, o := range h.OffendersMark {
// 			output += fmt.Sprintf("  %s\n", bytesToHex([]byte(o[:])))
// 		}
// 	}
// 	output += fmt.Sprintf("AuthorIndex: %d\n", h.AuthorIndex)
// 	output += fmt.Sprintf("EntropySource: %s\n", bytesToHex([]byte(h.EntropySource[:])))
// 	output += fmt.Sprintf("Seal: %s\n", bytesToHex([]byte(h.Seal[:])))

// 	return output
// }

// type HeaderJSON struct {
// 	Parent          string `json:"parent"`
// 	ParentStateRoot string `json:"parent_state_root"`
// 	ExtrinsicHash   string `json:"extrinsic_hash"`
// 	Slot            int    `json:"slot"`
// 	EpochMark       struct {
// 		Entropy        string   `json:"entropy"`
// 		TicketsEntropy string   `json:"tickets_entropy"`
// 		Validators     []string `json:"validators"`
// 	} `json:"epoch_mark"`
// 	TicketsMark   interface{} `json:"tickets_mark"`
// 	OffendersMark []string    `json:"offenders_mark"`
// 	AuthorIndex   int         `json:"author_index"`
// 	EntropySource string      `json:"entropy_source"`
// 	Seal          string      `json:"seal"`
// }

// func loadTestHeader0() types.Header {
// 	jsonFile := "../../pkg/test_data/jam-test-vectors/codec/data/header_0.json"

// 	// Open the JSON file
// 	file, err := os.Open(jsonFile)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	// Read the file content
// 	byteValue, err := io.ReadAll(file)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Unmarshal the JSON data
// 	var headerJSON HeaderJSON
// 	err = json.Unmarshal(byteValue, &headerJSON)
// 	if err != nil {
// 		panic(err)
// 	}

// 	header := convertHeaderJSONToHeader(headerJSON)

// 	return header
// }

// func convertHeaderJSONToHeader(headerJSON HeaderJSON) types.Header {
// 	header := types.Header{
// 		Parent:          types.HeaderHash(strToHex(headerJSON.Parent)),
// 		ParentStateRoot: types.StateRoot(strToHex(headerJSON.ParentStateRoot)),
// 		ExtrinsicHash:   types.OpaqueHash(strToHex(headerJSON.ExtrinsicHash)),
// 		Slot:            types.TimeSlot(headerJSON.Slot),
// 		EpochMark: &types.EpochMark{
// 			Entropy:        types.Entropy(strToHex(headerJSON.EpochMark.Entropy)),
// 			TicketsEntropy: types.Entropy(strToHex(headerJSON.EpochMark.TicketsEntropy)),
// 			Validators:     nil,
// 		},
// 		TicketsMark:   nil,
// 		OffendersMark: nil,
// 		AuthorIndex:   types.ValidatorIndex(headerJSON.AuthorIndex),
// 		EntropySource: types.BandersnatchVrfSignature(strToHex(headerJSON.EntropySource)),
// 		Seal:          types.BandersnatchVrfSignature(strToHex(headerJSON.Seal)),
// 	}

// 	validators := make([]types.BandersnatchPublic, len(headerJSON.EpochMark.Validators))
// 	for i, v := range headerJSON.EpochMark.Validators {
// 		validators[i] = types.BandersnatchPublic(strToHex(v))
// 	}

// 	offenders := make([]types.Ed25519Public, len(headerJSON.OffendersMark))
// 	for i, o := range headerJSON.OffendersMark {
// 		offenders[i] = types.Ed25519Public(strToHex(o))
// 	}

// 	header.EpochMark.Validators = validators
// 	header.OffendersMark = offenders

// 	return header
// }

// func TestNewHeaderController(t *testing.T) {
// 	t.Run("NewHeaderController", func(t *testing.T) {
// 		hc := NewHeaderController()
// 		if hc == nil {
// 			t.Errorf("NewHeaderController() = %v; want not nil", hc)
// 		}
// 	})
// }

// func TestHeaderController_SetHeader(t *testing.T) {
// 	t.Run("SetHeader", func(t *testing.T) {
// 		hc := NewHeaderController()

// 		inputHeader := types.Header{}
// 		hc.SetHeader(inputHeader)

// 		outputHeader := hc.GetHeader()

// 		// compare struct values
// 		if !reflect.DeepEqual(inputHeader, outputHeader) {
// 			t.Errorf("SetHeader() = %v; want %v", outputHeader, inputHeader)
// 		}
// 	})
// }

// func TestHeaderController_GetHeader(t *testing.T) {
// 	t.Run("GetHeader", func(t *testing.T) {
// 		hc := NewHeaderController()

// 		inputHeader := types.Header{}
// 		hc.SetHeader(inputHeader)

// 		outputHeader := hc.GetHeader()

// 		// compare struct values
// 		if !reflect.DeepEqual(inputHeader, outputHeader) {
// 			t.Errorf("GetHeader() = %v; want %v", outputHeader, inputHeader)
// 		}
// 	})
// }

// func TestValidateTimeSlot(t *testing.T) {
// 	now := time.Now().UTC()
// 	secondsSinceJam := uint64(now.Sub(types.JamCommonEra).Seconds())
// 	testCases := []struct {
// 		parentHeader types.Header
// 		slot         types.TimeSlot
// 		expected     error
// 	}{
// 		{
// 			parentHeader: types.Header{
// 				Slot: 0,
// 			},
// 			slot:     1,
// 			expected: nil,
// 		},
// 		{
// 			parentHeader: types.Header{
// 				Slot: 1,
// 			},
// 			slot:     2,
// 			expected: nil,
// 		},
// 		{
// 			parentHeader: types.Header{
// 				Slot: 2,
// 			},
// 			slot:     1,
// 			expected: fmt.Errorf("The time slot of the header is always larger than the parent header's time slot."),
// 		},
// 		{
// 			parentHeader: types.Header{
// 				Slot: 2,
// 			},
// 			slot:     types.TimeSlot((secondsSinceJam + 100) / uint64(types.SlotPeriod)),
// 			expected: fmt.Errorf("The time slot of the header is always smaller than the current time."),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("ValidateTimeSlot", func(t *testing.T) {
// 			hc := NewHeaderController()

// 			err := hc.ValidateTimeSlot(tc.parentHeader, tc.slot)

// 			if err != nil && tc.expected != nil {
// 				if err.Error() != tc.expected.Error() {
// 					t.Errorf("ValidateTimeSlot() = %v; want %v", err, tc.expected)
// 				}
// 			} else {
// 				if err != tc.expected {
// 					t.Errorf("ValidateTimeSlot() = %v; want %v", err, tc.expected)
// 				}
// 			}
// 		})
// 	}
// }

// func TestCreateHeaderSlot(t *testing.T) {
// 	testCases := []struct {
// 		parentHeader    types.Header
// 		currentTimeslot types.TimeSlot
// 		error           error
// 	}{
// 		{
// 			parentHeader: types.Header{
// 				Slot: 0,
// 			},
// 			currentTimeslot: 1,
// 			error:           nil,
// 		},
// 		{
// 			parentHeader: types.Header{
// 				Slot: 2,
// 			},
// 			currentTimeslot: 5,
// 			error:           nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("CreateHeaderSlot", func(t *testing.T) {
// 			hc := NewHeaderController()

// 			err := hc.CreateHeaderSlot(tc.parentHeader, tc.currentTimeslot)

// 			if err == nil {
// 				if hc.GetHeader().Slot != tc.currentTimeslot {
// 					t.Errorf("CreateHeaderSlot() = %d; want %d", hc.GetHeader().Slot, tc.currentTimeslot)
// 				}
// 			}
// 		})
// 	}
// }

// func TestCreateBlockAuthorIndex(t *testing.T) {
// 	testCases := []struct {
// 		authorIndex types.ValidatorIndex
// 		expected    types.ValidatorIndex
// 	}{
// 		{
// 			authorIndex: 0,
// 			expected:    0,
// 		},
// 		{
// 			authorIndex: 123,
// 			expected:    123,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		hc := NewHeaderController()
// 		hc.CreateBlockAuthorIndex(types.ValidatorIndex(tc.authorIndex))
// 		if hc.Header.AuthorIndex != tc.expected {
// 			t.Errorf("CreateBlockAuthorIndex() = %d; want %d", hc.Header.AuthorIndex, tc.expected)
// 		}
// 	}
// }

// func TestGetAuthorBandersnatchKey(t *testing.T) {
// 	testCases := []struct {
// 		bandersnatchKey types.BandersnatchPublic
// 		authorIndex     types.ValidatorIndex
// 		expected        types.BandersnatchPublic
// 	}{
// 		{
// 			bandersnatchKey: types.BandersnatchPublic(strToHex("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
// 			authorIndex:     0,
// 			expected:        types.BandersnatchPublic(strToHex("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
// 		},
// 		{
// 			bandersnatchKey: types.BandersnatchPublic(strToHex("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
// 			authorIndex:     1,
// 			expected:        types.BandersnatchPublic(strToHex("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
// 		},
// 		{
// 			bandersnatchKey: types.BandersnatchPublic(strToHex("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
// 			authorIndex:     2,
// 			expected:        types.BandersnatchPublic(strToHex("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
// 		},
// 	}

// 	s := store.GetInstance()
// 	// Initialize the store
// 	for _, tc := range testCases {
// 		s.AddPosteriorCurrentValidator(types.Validator{
// 			Bandersnatch: tc.bandersnatchKey,
// 		})
// 	}

// 	for _, tc := range testCases {
// 		thisHeader := types.Header{
// 			AuthorIndex: tc.authorIndex,
// 		}

// 		hc := NewHeaderController()
// 		thisBandersnatch := hc.GetAuthorBandersnatchKey(thisHeader)

// 		if thisBandersnatch != tc.expected {
// 			t.Errorf("GetAuthorBandersnatchKey() = %s; want %s", thisBandersnatch, tc.expected)
// 		}
// 	}
// }

// func TestGetAncenstorHeaders(t *testing.T) {
// 	now := time.Now().UTC()
// 	secondsSinceJam := uint64(now.Sub(types.JamCommonEra).Seconds())

// 	s := store.GetInstance()
// 	s.AddAncestorHeader(types.Header{
// 		Slot: types.TimeSlot((secondsSinceJam + 100) / uint64(types.SlotPeriod)),
// 	})
// 	s.AddAncestorHeader(types.Header{
// 		Slot: types.TimeSlot((secondsSinceJam + 101) / uint64(types.SlotPeriod)),
// 	})
// 	s.AddAncestorHeader(types.Header{
// 		Slot: types.TimeSlot((secondsSinceJam + 102) / uint64(types.SlotPeriod)),
// 	})

// 	hc := NewHeaderController()
// 	ancestorHeaders := hc.GetAncestorHeaders()

// 	if len(ancestorHeaders) != 3 {
// 		t.Errorf("GetAncestorHeaders() = %d; want 3", len(ancestorHeaders))
// 	}
// }

// func TestGetAncenstorHeadersTooOld(t *testing.T) {
// 	s := store.GetInstance()
// 	s.AddAncestorHeader(types.Header{
// 		Slot: 0,
// 	})
// 	s.AddAncestorHeader(types.Header{
// 		Slot: 1,
// 	})
// 	s.AddAncestorHeader(types.Header{
// 		Slot: 2,
// 	})

// 	hc := NewHeaderController()
// 	ancestorHeaders := hc.GetAncestorHeaders()

// 	if len(ancestorHeaders) != 0 {
// 		t.Errorf("GetAncestorHeaders() = %d; want 0", len(ancestorHeaders))
// 	}
// }

// // INFO: We have to test the merklization function in the merklization package.
// // This funciton only checks if the length of the parentStateRoot is correct.
// func TestCreateParentStateRoot(t *testing.T) {
// 	testParentState := types.State{}

// 	hc := NewHeaderController()

// 	hc.CreateStateRootHash(testParentState)

// 	parentStateRoot := hc.GetHeader().ParentStateRoot

// 	if len(parentStateRoot) != len(types.StateRoot{}) {
// 		t.Errorf("CreateStateRootHash() = %s; want %s", parentStateRoot, types.StateRoot{})
// 	}
// }
