package testdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
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
func (r *TestDataReader) ParseTestData(data []byte) (result interface{}, err error) {
	fmt.Printf("Data type: %T\n", data)
	// Handle different test types
	switch r.dataType {
	case "jam-test-vectors":
		// For jam-test-vectors, we need to handle different test modes
		switch r.mode {
		case SafroleMode:
			var safroleState jamtests.SafroleTestCase
			decoder := types.NewDecoder()
			if err := decoder.Decode(data, &safroleState); err != nil {
				return nil, fmt.Errorf("failed to decode safrole test data: %v", err)
			}
			result = safroleState

		case AssurancesMode:
			// Use the assurances test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal assurances test data: %v", err)
			}
		case PreimagesMode:
			// Use the preimages test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal preimages test data: %v", err)
			}
		case DisputesMode:
			// Use the disputes test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal disputes test data: %v", err)
			}
		case HistoryMode:
			// Use the history test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal history test data: %v", err)
			}
		case AccumulateMode:
			// Use the accumulate test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal accumulate test data: %v", err)
			}
		case AuthorizationsMode:
			// Use the authorizations test case parser
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal authorizations test data: %v", err)
			}
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

	return &result, nil
}

// TODO: Implement the test data reader for jamtestnet
// Currently we only read the data from file, but next step we should implement the
// logic to parse the data into our data store
func SetTestDataToDataStore(testData interface{}) {
	// TODO: Implement the logic to set test data to data store
}
