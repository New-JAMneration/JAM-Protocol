package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/pkg/cli"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
	jamtestvector "github.com/New-JAMneration/JAM-Protocol/testdata/jam_test_vector"
	jamtestnet "github.com/New-JAMneration/JAM-Protocol/testdata/jam_testnet"
	"github.com/New-JAMneration/JAM-Protocol/testdata/traces"
)

var (
	testMode       string
	testSize       string
	testType       string
	testFileFormat string
	testRunSTF     bool
)

// TestCommand returns the test command
func TestCommand() *cli.Command {
	return &cli.Command{
		Use:   "test",
		Short: "Run JAM Protocol tests",
		Long: `Run tests for the JAM Protocol. 
You can specify the test type (jam-test-vectors, jamtestnet, trace), mode (safrole, assurances, etc.), and size (tiny, full).
For example:
  jam test --type jam-test-vectors --mode safrole --size tiny
  jam test --type jamtestnet --mode assurances
  jam test --type trace --mode safrole`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:         "type",
				Usage:        "Test data type (jam-test-vectors, jamtestnet, trace)",
				DefaultValue: "jam-test-vectors",
				Destination:  &testType,
			},
			&cli.StringFlag{
				Name:         "mode",
				Usage:        "Test mode (safrole, assurances, preimages, disputes, history, accumulate, authorizations, statistics, reports)",
				DefaultValue: "safrole",
				Destination:  &testMode,
			},
			&cli.StringFlag{
				Name:         "size",
				Usage:        "Test size (tiny, full, data) - only for jam-test-vectors",
				DefaultValue: "tiny",
				Destination:  &testSize,
			},
			&cli.StringFlag{
				Name:         "format",
				Usage:        "Test data format (json, binary) - only for jam-test-vectors",
				DefaultValue: "binary",
				Destination:  &testFileFormat,
			},
			&cli.BooleanFlag{
				Name:         "stf",
				Usage:        "Run the whole STF (State Transition Function) instead of partial test",
				DefaultValue: false,
				Destination:  &testRunSTF,
			},
		},
		Run: func(args []string) {
			// Validate inputs
			if err := validateTestType(testType); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			mode := testdata.TestMode(testMode)
			if err := validateTestMode(mode); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			dataFormat, err := validateTestFormat(testFileFormat)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Create reader and runner
			reader, runner, err := createReaderAndRunner(testType, mode, testdata.TestSize(testSize), dataFormat)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Read test data
			var testFiles []testdata.TestData
			testFiles, err = reader.ReadTestData()
			if err != nil {
				fmt.Printf("Error reading test data: %v\n", err)
				os.Exit(1)
			}

			// Print results
			fmt.Printf("\nTest Results for %s (type: %s) (format: %s): \n", mode, testType, testFileFormat)
			if testType == "jam-test-vectors" {
				fmt.Printf("Size: %s\n", testSize)
			}
			passed := 0
			failed := 0

			for idx, testFile := range testFiles {
				log.Printf("------------------{%v}--------------------", idx)
				// Parse the test data
				data, err := reader.ParseTestData(testFile.Data)
				if err != nil {
					log.Printf("got error: %v", err)
					failed++
					continue
				}

				// Run the test

				outputErr := runner.Run(data, testRunSTF)
				expectedErr := data.ExpectError()

				if expectedErr != nil {
					if outputErr == nil {
						fmt.Printf("Test %s failed: expected error %v but got none\n", testFile.Name, expectedErr)
						failed++
					}
					if outputErr.Error() != expectedErr.Error() {
						fmt.Printf("Test %s failed: expected error %v but got %v\n", testFile.Name, expectedErr, outputErr)
						failed++
					} else {
						fmt.Printf("Test %s passed: expected: %v, got: %v\n", testFile.Name, expectedErr, outputErr)
						passed++
					}
					// Check the error message
				} else {
					fmt.Printf("Test %s passed\n", testFile.Name)
					passed++
				}
			}

			fmt.Println("----------------------------------------")
			fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", len(testFiles), passed, failed)
		},
	}
}

func init() {
	testCmd := TestCommand()

	cli.AddCommand(testCmd)
}

// Encapsulate validation logic into separate functions
func validateTestType(testType string) error {
	if testType != "jam-test-vectors" && testType != "jamtestnet" && testType != "trace" {
		return fmt.Errorf("invalid test type '%s'", testType)
	}
	return nil
}

func validateTestMode(mode testdata.TestMode) error {
	switch mode {
	case testdata.SafroleMode, testdata.AssurancesMode, testdata.PreimagesMode,
		testdata.DisputesMode, testdata.HistoryMode, testdata.AccumulateMode,
		testdata.AuthorizationsMode, testdata.StatisticsMode, testdata.ReportsMode,
		testdata.FallbackMode:
		return nil
	default:
		return fmt.Errorf("invalid test mode '%s'", mode)
	}
}

func validateAndSetTestSize(size testdata.TestSize) error {
	switch size {
	case testdata.TinySize, testdata.DataSize:
		types.SetTinyMode()
		return nil
	case testdata.FullSize:
		types.SetFullMode()
		return nil
	default:
		return fmt.Errorf("invalid test size '%s'", size)
	}
}

func validateTestFormat(format string) (testdata.DataFormat, error) {
	switch format {
	case "binary":
		return testdata.BinaryFormat, nil
	case "json":
		return testdata.JSONFormat, nil
	default:
		return "", fmt.Errorf("invalid format '%s'", format)
	}
}

// Factory functions for creating readers and runners
func createReaderAndRunner(testType string, mode testdata.TestMode, size testdata.TestSize, format testdata.DataFormat) (*testdata.TestDataReader, testdata.TestRunner, error) {
	var reader *testdata.TestDataReader
	var runner testdata.TestRunner

	switch testType {
	case "jam-test-vectors":
		if err := validateAndSetTestSize(size); err != nil {
			return nil, nil, err
		}
		reader = testdata.NewTestDataReader(mode, size, format)
		runner = jamtestvector.NewJamTestVectorsRunner(mode)
	case "jamtestnet":
		reader = testdata.NewJamTestNetReader(mode, format)
		runner = jamtestnet.NewJamTestNetRunner(mode)
	case "trace":
		// Only suuport the json type for now
		if format != "json" {
			format = "json"
		}

		reader = testdata.NewTracesReader(mode, format)
		runner = traces.NewTraceRunner()
	}

	return reader, runner, nil
}
