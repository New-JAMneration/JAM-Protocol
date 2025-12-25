package main

import (
	"context"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	m "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
	"github.com/New-JAMneration/JAM-Protocol/logger"
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
			Usage:       "Test mode (accumulate, assurances, authorizations, disputes, history, preimages, reports, safrole, statistics, \nfallback, fuzzy, fuzzy_light, preimages, preimages_light, safrole, storage, storage_light)",
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
		// Initialize config
		config.InitConfig(configPath, testSize)

		// Validate inputs (these are user input errors, use Fatal)
		if err := validateTestType(testType); err != nil {
			logger.Fatal(err) // Fatal already calls os.Exit(1)
		}

		mode := testdata.TestMode(testMode)
		if err := validateTestMode(mode); err != nil {
			logger.Fatal(err)
		}

		dataFormat, err := validateTestFormat(testFileFormat)
		if err != nil {
			logger.Fatal(err)
		}

		// Create reader and runner
		reader, runner, err := createReaderAndRunner(testType, mode, testdata.TestSize(testSize), dataFormat)
		if err != nil {
			logger.Fatal(err)
		}

		// Read test data
		var testFiles []testdata.TestData
		testFiles, err = reader.ReadTestData()
		if err != nil {
			logger.Fatalf("Error reading test data: %v", err)
		}

		// Print results
		msg := fmt.Sprintf("Test Results for %s (type: %s) (format: %s) ", mode, testType, testFileFormat)
		if testType == "jam-test-vectors" {
			msg += fmt.Sprintf("(size: %s) ", testSize)
		}
		logger.Info(msg)

		passed := 0
		failed := 0

		if testType == "trace" {
			// get genesis file, genesis will be sorted at last
			genesisFile := testFiles[len(testFiles)-1]
			testFiles = testFiles[:len(testFiles)-1]

			// parse genesis data
			genesis, err := reader.ParseGenesis(genesisFile.Data)
			if err != nil {
				logger.Fatalf("error parsing genesis: %v", err)
			}

			genesisBlock := types.Block{
				Header: genesis.Header,
			}

			state, keyVals, err := m.StateKeyValsToState(genesis.State.KeyVals)
			if err != nil {
				logger.Fatalf("Failed to parse state key-vals to state: %v", err)
			}

			store.GetInstance().SetPriorStateUnmatchedKeyVals(keyVals)

			instance := store.GetInstance()
			instance.GenerateGenesisBlock(genesisBlock)
			instance.GenerateGenesisState(state)
		}

		for idx, testFile := range testFiles {
			logger.Infof("------------------{%v, %s}--------------------", idx+1, testFile.Name)
			if testType == "trace" {
				// post-state update to pre-state, tau_prime+1
				store.GetInstance().StateCommit()
			}

			data, err := reader.ParseTestData(testFile.Data)
			if err != nil {
				logger.Errorf("got error during parsing: %v", err)
				failed++
				continue
			}

			// Run the test
			outputErr := runner.Run(data, testRunSTF)

			if testType == "jam-test-vectors" {
				expectedErr := data.ExpectError()
				if expectedErr != nil {
					if outputErr == nil {
						logger.Errorf("Test %s failed: expected error %v but got none\n", testFile.Name, expectedErr)
						failed++
					}
					if outputErr.Error() != expectedErr.Error() {
						logger.Errorf("Test %s failed: expected error %v but got %v\n", testFile.Name, expectedErr, outputErr)
						failed++
					} else {
						logger.Infof("Test %s passed: expected: %v, got: %v\n", testFile.Name, expectedErr, outputErr)
						passed++
					}
					// Check the error message
				} else {
					if outputErr != nil {
						logger.Errorf("Test %s failed: expected no error but got %v\n", testFile.Name, outputErr)
						failed++
					} else {
						logger.Infof("Test %s passed", testFile.Name)
						passed++
					}
				}
			} else {
				// type = trace
				// stf occurs error
				if outputErr != nil {
					logger.Errorf("stf output error %v:", outputErr)
					priorState := store.GetInstance().GetPriorStates().GetState()
					store.GetInstance().GetPosteriorStates().SetState(priorState)
				}

				err := data.Validate()
				if err != nil {
					logger.Errorf("state root validate error: %v", err)
					failed++
					break
				}
				passed++
				logger.Infof("passed\n")

			}
		}

		logger.Info("----------------------------------------")
		logger.Infof("Total: %d, Passed: %d, Failed: %d\n", len(testFiles), passed, failed)

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
		testdata.FuzzyMode,         // trace
		testdata.FuzzyLightMode:    // trace
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
