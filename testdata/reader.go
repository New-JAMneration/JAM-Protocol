package testdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtestsaccumulate "github.com/New-JAMneration/JAM-Protocol/jamtests/accumulate"
	jamtestsassurances "github.com/New-JAMneration/JAM-Protocol/jamtests/assurances"
	jamtestsauth "github.com/New-JAMneration/JAM-Protocol/jamtests/authorizations"
	jamtestsdisputes "github.com/New-JAMneration/JAM-Protocol/jamtests/disputes"
	jamtestshistory "github.com/New-JAMneration/JAM-Protocol/jamtests/history"
	jamtestspreimages "github.com/New-JAMneration/JAM-Protocol/jamtests/preimages"
	jamtestsreports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtestssafrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
	jamtestsstatistics "github.com/New-JAMneration/JAM-Protocol/jamtests/statistics"
	jamteststrace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
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
	StatisticsMode     TestMode = "statistics"
	ReportsMode        TestMode = "reports"

	// Trace
	FallbackMode      TestMode = "fallback"
	StorageMode       TestMode = "storage"
	StorageLightMode  TestMode = "storage_light"
	PreimageLightMode TestMode = "preimages_light"
	FuzzyMode         TestMode = "fuzzy"
	FuzzyLightMode    TestMode = "fuzzy_light"
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

// NewJamTestVectorsReader creates a new TestDataReader
func NewJamTestVectorsReader(mode TestMode, size TestSize, format DataFormat) *TestDataReader {
	reader := &TestDataReader{
		dataType: "jam-test-vectors",
		mode:     mode,
		size:     size,
		format:   format,
	}

	// jam-test-vectors/stf/mode/size
	reader.basePath = filepath.Join("pkg", "test_data", "jam-test-vectors", "stf", string(mode), string(size))

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

func NewTracesReader(mode TestMode, format DataFormat) *TestDataReader {
	reader := &TestDataReader{
		dataType: "trace",
		mode:     mode,
		format:   format,
	}

	// TODO: we could suppport the different project traces
	folderName := mode
	if mode == ReportsMode {
		folderName = "reports-l0"
	}
	reader.basePath = filepath.Join("pkg", "test_data", "jam-test-vectors", "traces", string(folderName))
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

	// Sort the test files by name
	sort.Slice(testFiles, func(i, j int) bool {
		return testFiles[i].Name < testFiles[j].Name
	})

	return testFiles, nil
}

// ParseTestData parses the test data into the specified type based on the test type
func (r *TestDataReader) ParseTestData(data []byte) (result Testable, err error) {
	switch r.dataType {
	case "jam-test-vectors":
		// For jam-test-vectors, we need to handle different test modes
		switch r.mode {
		case SafroleMode:
			var safroleTestCase jamtestssafrole.SafroleTestCase
			if err := r.ReadFile(data, &safroleTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode safrole test data: %v", err)
			}
			result = &safroleTestCase
		case AssurancesMode:
			var assuranceTestCase jamtestsassurances.AssuranceTestCase
			if err := r.ReadFile(data, &assuranceTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode assurances test data: %v", err)
			}
			result = &assuranceTestCase
		case PreimagesMode:
			var preimageTestCase jamtestspreimages.PreimageTestCase
			if err := r.ReadFile(data, &preimageTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode preimages test data: %v", err)
			}
			result = &preimageTestCase
		case DisputesMode:
			var disputeTestCase jamtestsdisputes.DisputeTestCase
			if err := r.ReadFile(data, &disputeTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode disputes test data: %v", err)
			}
			result = &disputeTestCase
		case HistoryMode:
			var historyTestCase jamtestshistory.HistoryTestCase
			if err := r.ReadFile(data, &historyTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode history test data: %v", err)
			}
			result = &historyTestCase
		case AccumulateMode:
			var accumulateTestCase jamtestsaccumulate.AccumulateTestCase
			if err := r.ReadFile(data, &accumulateTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode accumulate test data: %v", err)
			}
			result = &accumulateTestCase
		case AuthorizationsMode:
			var authorizationsTestCase jamtestsauth.AuthorizationTestCase
			if err := r.ReadFile(data, &authorizationsTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode authorization test data: %v", err)
			}
			result = &authorizationsTestCase
		case StatisticsMode:
			var statisticsTestCase jamtestsstatistics.StatisticsTestCase
			if err := r.ReadFile(data, &statisticsTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode statistics test data: %v", err)
			}
			result = &statisticsTestCase
		case ReportsMode:
			var reportsTestCase jamtestsreports.ReportsTestCase
			if err := r.ReadFile(data, &reportsTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode reports test data: %v", err)
			}
			result = &reportsTestCase
		default:
			return nil, fmt.Errorf("unsupported test mode: %s", r.mode)
		}
	case "jamtestnet":
		// For jamtestnet, we need to handle state transitions
		return nil, fmt.Errorf("work in progress: %s", r.dataType)
	case "trace":
		switch r.mode {
		case SafroleMode, FallbackMode, ReportsMode, PreimagesMode, PreimageLightMode, StorageMode, StorageLightMode, FuzzyMode, FuzzyLightMode:
			var traceTestCase jamteststrace.TraceTestCase
			if err := r.ReadFile(data, &traceTestCase); err != nil {
				return nil, fmt.Errorf("failed to unmarshal/decode trace test data: %v", err)
			}

			result = &traceTestCase
		default:
			return nil, fmt.Errorf("unsupported test mode for trace: %s", r.mode)
		}

	default:
		return nil, fmt.Errorf("unsupported test type: %s", r.dataType)
	}

	err = result.Dump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump test data: %v", err)
	}

	return result, nil
}

func (r *TestDataReader) ReadFile(data []byte, result interface{}) error {
	if r.format == JSONFormat {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("failed to unmarshal JSON data: %v", err)
		}
	} else {
		decoder := types.NewDecoder()
		if err := decoder.Decode(data, result); err != nil {
			return fmt.Errorf("failed to decode data: %v", err)
		}
	}

	return nil
}

func (r *TestDataReader) ParseGenesis(data []byte) (*jamteststrace.Genesis, error) {
	var genesis jamteststrace.Genesis
	if err := r.ReadFile(data, &genesis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal/decode reports test data: %v", err)
	}

	return &genesis, nil
}
