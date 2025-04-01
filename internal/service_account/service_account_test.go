package service_account

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	jamtests_preimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func TestFetchCodeByHash(t *testing.T) {
	// set up test data
	var (
		mockMetadata = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
		mockCode     = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
	)

	// encode metaCode
	testMetaCode := types.MetaCode{
		Metadata: mockMetadata,
		Code:     mockCode,
	}
	encoder := types.NewEncoder()
	encodedMetaCode, err := encoder.Encode(&testMetaCode)
	if err != nil {
		t.Errorf("Error encoding MetaCode: %v", err)
	}

	mockCodeHash := hash.Blake2bHash(encodedMetaCode)

	// create ServiceAccount
	mockAccount := types.ServiceAccount{
		PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
			mockCodeHash: encodedMetaCode,
		},
	}

	// fetch code by hash
	metadata, code := FetchCodeByHash(mockAccount, mockCodeHash)

	// check if code is equal to mockCode
	if code == nil {
		t.Errorf("FetchCodeByHash failed: non exist code for codeHash %v", mockCodeHash)
	} else if !reflect.DeepEqual(code, mockCode) {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockCode, code)
	}
	if metadata == nil {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockMetadata, metadata)
	} else if !reflect.DeepEqual(metadata, mockMetadata) {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockMetadata, metadata)
	}

}

func TestValidatePreimageLookupDict(t *testing.T) {
	// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)

	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		// mockCode     = types.ByteSequence("0x92cdf578c47085a5992256f0dcf97d0b19f1")
		mockCode     = types.ByteSequence("0x5b5477bef56d05dd59b758c2c4672c88aa8a71a2949f3921f37a25a9a167aeba")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		// mockCodeHash_bs  = types.ByteSequence(mockCodeHash[:])
		// mockCodeHash_str = hex.EncodeToString(mockCodeHash_bs)
		preimage = mockCode

		// create ServiceAccount
		mockAccount = types.ServiceAccount{
			// h = H(p)
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			// (h, |p|) ∈ K(a_l)
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(preimage))}: {},
			},
		}
	)

	// test ValidateAccount
	err := ValidatePreimageLookupDict(mockAccount)
	if err != nil {
		t.Errorf("ValidateAccount failed: %v", err)
	}
}

func TestHistoricalLookupFunction(t *testing.T) {
	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode     = types.ByteSequence("0x123456789")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		preimage     = mockCode

		mockTimestamp = types.TimeSlot(42)

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(preimage))}: {mockTimestamp},
			},
		}
	)
	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// test HistoricalLookupFunction
	result, _ := HistoricalLookupFunction(mockAccount, mockTimestamp, mockCodeHash)
	if !reflect.DeepEqual(result, preimage) {
		t.Errorf("HistoricalLookupFunction failed: expected %v, got %v", preimage, result)
	}
}

func TestGetSerivecAccountDerivatives(t *testing.T) {
	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode        = types.ByteSequence("0x123456789")
		mockCodeHash    = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		preimage        = mockCode
		mockPreimageLen = types.U32(len(preimage))

		mockTimestamp = types.TimeSlot(42)

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			StorageDict: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: mockPreimageLen}: {mockTimestamp},
			},
		}
	)
	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// test GetSerivecAccountDerivatives
	accountDer := GetSerivecAccountDerivatives(mockId)
	t.Log("accountDer:", accountDer)
	t.Logf("a_i=2*|a_l|+|a_s|\n LHS: %v, RHS: %v", accountDer.Items, 2*len(mockAccount.LookupDict)+len(mockAccount.StorageDict))
	var totalZ types.U32
	for key := range mockAccount.LookupDict {
		totalZ += key.Length
	}
	var totalX int
	for _, value := range mockAccount.StorageDict {
		totalX += len(value)
	}
	t.Logf("a_o=[ ∑_{(h,z)∈Key(a_l)}  81 + z ] + [ ∑_{x∈Value(a_s)} 32 + |x| ]\n LHS: %v, RHS: %v + %v", accountDer.Bytes, 81+totalZ, 32+totalX)
	t.Logf("a_t = B_S + B_I*a_i + B_L*a_o\n LHS: %v, RHS: %v+%v+%v", accountDer.Minbalance, types.BasicMinBalance, types.U32(types.AdditionalMinBalancePerItem)*accountDer.Items, types.U64(types.AdditionalMinBalancePerOctet)*accountDer.Bytes)
}

// Constants
const (
	MODE                 = "full" // tiny or full
	JSON_EXTENTION       = ".json"
	BIN_EXTENTION        = ".bin"
	JAM_TEST_VECTORS_DIR = "../../pkg/test_data/jam-test-vectors/"
	JAM_TEST_NET_DIR     = "../../pkg/test_data/jamtestnet/"
)

func TestMain(m *testing.M) {
	// Set the test mode
	types.SetTestMode()

	// Run the tests
	os.Exit(m.Run())
}

func LoadJAMTestJsonCase(filename string, structType reflect.Type) (interface{}, error) {
	// Create a new instance of the struct
	structValue := reflect.New(structType).Elem()

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data
	err = json.Unmarshal(byteValue, structValue.Addr().Interface())
	if err != nil {
		return nil, err
	}

	// Return the struct
	return structValue.Interface(), nil
}

func LoadJAMTestBinaryCase(filename string) ([]byte, error) {
	// read binary file and return byte array
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

func GetTargetExtensionFiles(dir string, extension string) ([]string, error) {
	// Get all files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Get all files with the target extension
	var targetFiles []string
	for _, file := range files {
		fileName := file.Name()
		if fileName[len(fileName)-len(extension):] == extension {
			targetFiles = append(targetFiles, fileName)
		}
	}

	return targetFiles, nil
}

func GetJsonFilename(filename string) string {
	return filename + JSON_EXTENTION
}

func GetBinFilename(filename string) string {
	return filename + BIN_EXTENTION
}

// preimages
func TestDecodeJamTestVectorsPreimages(t *testing.T) {
	if types.TEST_MODE != "tiny" {
		types.SetTinyMode()
		log.Println("⚠️  Preimages test cases only support tiny mode")
	}

	dir := filepath.Join(JAM_TEST_VECTORS_DIR, "preimages", "data")

	// Read binary files
	binFiles, err := GetTargetExtensionFiles(dir, BIN_EXTENTION)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	for _, binFile := range binFiles {
		// Read the binary file
		binPath := filepath.Join(dir, binFile)
		binData, err := LoadJAMTestBinaryCase(binPath)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// Decode the binary data
		decoder := types.NewDecoder()
		preimages := &jamtests_preimages.PreimageTestCase{}
		err = decoder.Decode(binData, preimages)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		// ---
		// TODO: implement comparison
	}
}
