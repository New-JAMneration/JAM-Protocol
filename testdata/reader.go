package testdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtestsaccumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtestsassurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	jamtestsauth "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	jamtestsdisputes "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
	jamtestshistory "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	jamtestspreimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	jamtestssafrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
)

// TestMode represents the type of test to run
type TestMode string

const (
	SafroleMode        TestMode = "safrole"
	AssurancesMode     TestMode = "assurances"
	PreimagesMode      TestMode = "preimages"
	DisputesMode       TestMode = "disputes"
	HistoryMode        TestMode = "history"
	AccumulateMode     TestMode = "accumulate"
	AuthorizationsMode TestMode = "authorizations"
)

// TestSize represents the size of the test data
type TestSize string

const (
	TinySize TestSize = "tiny"
	FullSize TestSize = "full"
)

// DataFormat represents the format of the test data
type DataFormat string

const (
	JSONFormat   DataFormat = "json"
	BinaryFormat DataFormat = "binary"
)

// TestData represents a single test case
type TestData struct {
	Name string
	Data []byte
}

// TestDataReader handles reading test data from different sources
type TestDataReader struct {
	dataType string
	mode     TestMode
	size     TestSize
	format   DataFormat
	basePath string
}

// NewTestDataReader creates a new TestDataReader
func NewTestDataReader(mode TestMode, size TestSize, format DataFormat) *TestDataReader {
	reader := &TestDataReader{
		dataType: "jam-test-vectors",
		mode:     mode,
		size:     size,
		format:   format,
	}

	// Construct the base path based on the data type and size
	reader.basePath = filepath.Join("pkg", "test_data", "jam-test-vectors", string(mode), string(size))

	return reader
}

// NewJamTestNetReader creates a new TestDataReader for jamtestnet
func NewJamTestNetReader(mode TestMode, format DataFormat) *TestDataReader {
	reader := &TestDataReader{
		dataType: "jamtestnet",
		mode:     mode,
		format:   format,
	}

	// Construct the base path for jamtestnet
	reader.basePath = filepath.Join("pkg", "test_data", "jamtestnet", "state_transitions")

	return reader
}

// ReadTestData reads all test files from the configured directory
func (r *TestDataReader) ReadTestData() ([]TestData, error) {
	var testFiles []TestData

	// Read all files in the directory
	err := filepath.Walk(r.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension based on format
		ext := filepath.Ext(path)
		if (r.format == JSONFormat && ext != ".json") || (r.format == BinaryFormat && ext != ".bin") {
			return nil
		}

		// Read the file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read test file %s: %v", path, err)
		}

		testFiles = append(testFiles, TestData{
			Name: filepath.Base(path),
			Data: data,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read test files: %v", err)
	}

	return testFiles, nil
}

// ParseTestData parses the test data into the specified type based on the test type
func (r *TestDataReader) ParseTestData(data []byte) (result Testable, err error) {
	fmt.Printf("Data type: %T\n", data)

	decoder := types.NewDecoder()

	switch r.dataType {
	case "jam-test-vectors":
		// For jam-test-vectors, we need to handle different test modes
		switch r.mode {
		case SafroleMode:
			var safroleTestCase jamtestssafrole.SafroleTestCase
			if err := decoder.Decode(data, &safroleTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode safrole test data: %v", err)
			}
			result = safroleTestCase
		case AssurancesMode:
			var assuranceTestCase jamtestsassurances.AssuranceTestCase
			if err := decoder.Decode(data, &assuranceTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode safrole test data: %v", err)
			}
			result = assuranceTestCase
		case PreimagesMode:
			var preimageTestCase jamtestspreimages.PreimageTestCase
			if err := decoder.Decode(data, &preimageTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode preimages test data: %v", err)
			}
			result = preimageTestCase
		case DisputesMode:
			var disputeTestCase jamtestsdisputes.DisputeTestCase
			if err := decoder.Decode(data, &disputeTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode disputes test data: %v", err)
			}
			result = disputeTestCase
		case HistoryMode:
			var historyTestCase jamtestshistory.HistoryTestCase
			if err := decoder.Decode(data, &historyTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode history test data: %v", err)
			}
			result = historyTestCase
		case AccumulateMode:
			var accumlateTestCase jamtestsaccumulate.AccumulateTestCase
			if err := decoder.Decode(data, &accumlateTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode accumulate test data: %v", err)
			}
			result = accumlateTestCase
		case AuthorizationsMode:
			var authorizationsTestCase jamtestsauth.AuthorizationTestCase
			if err := decoder.Decode(data, &authorizationsTestCase); err != nil {
				return nil, fmt.Errorf("failed to decode authorization test data: %v", err)
			}
			result = authorizationsTestCase
		default:
			return nil, fmt.Errorf("unsupported test mode: %s", r.mode)
		}
	case "jamtestnet":
		// For jamtestnet, we need to handle state transitions
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal jamtestnet state transition data: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported test type: %s", r.dataType)
	}

	return result, nil
}

// TODO: Implement the test data reader for jamtestnet
// Currently we only read the data from file, but next step we should implement the
// logic to parse the data into our data store
func SetTestDataToDataStore(testData interface{}) {
	// TODO: Implement the logic to set test data to data store
}
