package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	jamteststrace "github.com/New-JAMneration/JAM-Protocol/jamtests/trace"
	"github.com/New-JAMneration/JAM-Protocol/testdata"
	jamtestvector "github.com/New-JAMneration/JAM-Protocol/testdata/jam_test_vector"
	jamtestnet "github.com/New-JAMneration/JAM-Protocol/testdata/jam_testnet"
	"github.com/New-JAMneration/JAM-Protocol/testdata/traces"
	"github.com/urfave/cli/v3"
)

var (
	testMode       string
	testSize       string
	testType       string
	testFileFormat string
	testRunSTF     bool
	testGenesis    string
)

var testCmd = &cli.Command{
	Name:  "test",
	Usage: "Run JAM Protocol tests",
	Description: `Run tests for the JAM Protocol. 
You can specify the test type (jam-test-vectors, jamtestnet, trace), mode (safrole, assurances, etc.), and size (tiny, full).
For example:
  go run ./cmd/node test --type jam-test-vectors --mode safrole --size tiny
  go run ./cmd/node test --type jamtestnet --mode assurances
  go run ./cmd/node test --type trace --mode safrole`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "type",
			Usage:       "Test data type (jam-test-vectors, jamtestnet, trace)",
			Value:       "jam-test-vectors",
			Destination: &testType,
		},
		&cli.StringFlag{
			Name:        "mode",
			Usage:       "Test mode (accumulate, assurances, authorizations, disputes, history, preimages, reports, safrole, statistics, \nfallback, fuzzy, preimages, preimages_light, safrole, storage, storage_light)",
			Value:       "safrole",
			Destination: &testMode,
		},
		&cli.StringFlag{
			Name:        "size",
			Usage:       "Test size (tiny, full, data) - only for jam-test-vectors",
			Value:       "tiny",
			Destination: &testSize,
		},
		&cli.StringFlag{
			Name:        "format",
			Usage:       "Test data format (json, binary) - only for jam-test-vectors",
			Value:       "binary",
			Destination: &testFileFormat,
		},
		&cli.BoolFlag{
			Name:        "stf",
			Usage:       "Run the whole STF (State Transition Function) instead of partial test",
			Value:       false,
			Destination: &testRunSTF,
		},
		&cli.StringFlag{
			Name:        "genesis",
			Usage:       "Trace specify genesis file",
			Value:       "genesis",
			Destination: &testGenesis,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
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
		msg := fmt.Sprintf("Test Results for %s (type: %s) (format: %s) ", mode, testType, testFileFormat)
		if testType == "jam-test-vectors" {
			msg += fmt.Sprintf("(size: %s) ", testSize)
		}
		log.Println(msg)

		passed := 0
		failed := 0

		if testType == "trace" {

			genesisFileFound := false
			var ext string
			if dataFormat == "json" {
				ext = ".json"
			} else {
				ext = ".bin"
			}

			// get genesis
			for _, testFile := range testFiles {
				if testFile.Name != testGenesis+ext {
					continue
				}

				genesisFileFound = true
				var state types.State
				var block types.Block

				if strings.Contains(testFile.Name, "genesis") { // genesis file
					var genesis jamteststrace.Genesis
					err := reader.ReadFile(testFile.Data, &genesis)
					if err != nil {
						log.Panicf("Failed to Read genesis: %v", err)
					}
					state, _, err = merklization.StateKeyValsToState(genesis.State.KeyVals)
					if err != nil {
						log.Panicf("Failed to parse state key-vals to state: %v", err)
					}
					block.Header = genesis.Header
				} else {
					var genesis jamteststrace.TraceTestCase
					err := reader.ReadFile(testFile.Data, &genesis)
					if err != nil {
						log.Panicf("Failed to Read genesis: %v", err)
					}
					state, _, err = merklization.StateKeyValsToState(genesis.PostState.KeyVals)
					if err != nil {
						log.Panicf("Failed to parse state key-vals to state: %v", err)
					}
					block.Header = genesis.Block.Header
					log.Println("genesis : ", testFile.Name)
				}

				instance := store.GetInstance()
				instance.GenerateGenesisBlock(block)
				instance.GenerateGenesisState(state)
			}

			if !genesisFileFound {
				log.Panicf("genesis file not found")
			}
		}

		for idx, testFile := range testFiles {
			// We've already set the genesis block, state
			if testType == "trace" && strings.Contains(testFile.Name, "genesis") {
				continue
			}

			log.Printf("------------------{%v, %s}--------------------", idx, testFile.Name)
			if testType == "trace" {
				// post-state update to pre-state, tau_prime+1
				store.GetInstance().StateCommit()
			}

			data, err := reader.ParseTestData(testFile.Data)
			if err != nil {
				log.Printf("got error during parsing: %v", err)
				failed++
				continue
			}

			// Run the test
			outputErr := runner.Run(data, testRunSTF)

			if testType == "jam-test-vectors" {
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
					if outputErr != nil {
						fmt.Printf("Test %s failed: expected no error but got %v\n", testFile.Name, outputErr)
						failed++
					} else {
						log.Printf("Test %s passed", testFile.Name)
						passed++
					}
				}
			} else {
				// type = trace
				// stf occurs error
				if outputErr != nil {
					log.Printf("stf output error %v:", outputErr)
					failed++
					break
				} else {
					err := data.Validate()
					if err != nil {
						log.Printf("state root validate error: %v", err)
						failed++
						break
					}
					passed++
					log.Printf("passed\n")
				}
			}
		}

		log.Printf("----------------------------------------")
		if testType == "trace" {
			// -1 : genesis file
			log.Printf("Total: %d, Passed: %d, Failed: %d\n", len(testFiles)-1, passed, failed)
		} else {
			log.Printf("Total: %d, Passed: %d, Failed: %d\n", len(testFiles), passed, failed)
		}
		return nil
	},
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
		testdata.FallbackMode,      // trace
		testdata.PreimageLightMode, // trace
		testdata.StorageLightMode,  // trace
		testdata.StorageMode,       // trace
		testdata.FuzzyMode:         // trace
		return nil
	default:
		return fmt.Errorf("invalid test mode '%s'", mode)
	}
}

func validateAndSetTestSize(size testdata.TestSize) error {
	switch size {
	case testdata.TinySize:
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
		reader = testdata.NewJamTestVectorsReader(mode, size, format)
		runner = jamtestvector.NewJamTestVectorsRunner(mode)
	case "jamtestnet":
		reader = testdata.NewJamTestNetReader(mode, format)
		runner = jamtestnet.NewJamTestNetRunner(mode)
	case "trace":
		if err := validateAndSetTestSize(size); err != nil {
			return nil, nil, err
		}
		reader = testdata.NewTracesReader(mode, format)
		runner = traces.NewTraceRunner()
	}

	return reader, runner, nil
}
