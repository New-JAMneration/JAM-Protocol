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
You can specify the test type (jam-test-vectors, jamtestnet), mode (safrole, assurances, etc.), and size (tiny, full).
For example:
  jam test --type jam-test-vectors --mode safrole --size tiny
  jam test --type jamtestnet --mode assurances`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:         "type",
				Usage:        "Test data type (jam-test-vectors, jamtestnet)",
				DefaultValue: "jam-test-vectors",
				Destination:  &testType,
			},
			&cli.StringFlag{
				Name:         "mode",
				Usage:        "Test mode (safrole, assurances, preimages, disputes, history, accumulate, authorizations, statistics)",
				DefaultValue: "safrole",
				Destination:  &testMode,
			},
			&cli.StringFlag{
				Name:         "size",
				Usage:        "Test size (tiny, full) - only for jam-test-vectors",
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
			// Validate test type
			if testType != "jam-test-vectors" && testType != "jamtestnet" {
				fmt.Printf("Error: Invalid test type '%s'\n", testType)
				fmt.Println("Valid types are: jam-test-vectors, jamtestnet")
				os.Exit(1)
			}

			// Validate test mode
			mode := testdata.TestMode(testMode)
			switch mode {
			case testdata.SafroleMode, testdata.AssurancesMode, testdata.PreimagesMode,
				testdata.DisputesMode, testdata.HistoryMode, testdata.AccumulateMode,
				testdata.AuthorizationsMode, testdata.StatisticsMode:
				// Valid mode
			default:
				fmt.Printf("Error: Invalid test mode '%s'\n", testMode)
				fmt.Println("Valid modes are: safrole, assurances, preimages, disputes, history, accumulate, authorizations, statistics")
				os.Exit(1)
			}

			// Validate format
			var dataFormat testdata.DataFormat
			switch testFileFormat {
			case "binary":
				dataFormat = testdata.BinaryFormat
			case "json":
				dataFormat = testdata.JSONFormat
			default:
				fmt.Printf("Error: Invalid format '%s'\n", testFileFormat)
				fmt.Println("Valid formats are: binary, json")
				os.Exit(1)
			}

			// Create test data reader based on type
			var reader *testdata.TestDataReader
			var runner testdata.TestRunner
			if testType == "jam-test-vectors" {
				// Validate test size for jam-test-vectors
				size := testdata.TestSize(testSize)
				switch size {
				case testdata.TinySize:
					types.SetTinyMode()
				case testdata.FullSize:
					types.SetFullMode()
				default:
					fmt.Printf("Error: Invalid test size '%s'\n", testSize)
					fmt.Println("Valid sizes are: tiny, full")
					os.Exit(1)
				}
				reader = testdata.NewTestDataReader(mode, size, dataFormat)
				runner = jamtestvector.NewJamTestVectorsRunner(mode)
			} else {
				reader = testdata.NewJamTestNetReader(mode, dataFormat)
				runner = jamtestnet.NewJamTestNetRunner(mode)
			}

			// Read test data
			testFiles, err := reader.ReadTestData()
			if err != nil {
				fmt.Printf("Error reading test data: %v\n", err)
				os.Exit(1)
			}

			// Print results
			fmt.Printf("\nTest Results for %s (%s):\n", mode, testType)
			if testType == "jam-test-vectors" {
				fmt.Printf("Size: %s\n", testSize)
			}
			fmt.Println("----------------------------------------")
			passed := 0
			failed := 0
			for idx, testFile := range testFiles {
				log.Printf("------------------{%v}--------------------", idx)
				// Parse the test data
				data, err := reader.ParseTestData(testFile.Data)
				if err != nil {
					log.Printf("got error: %v", err)
				}

				// Run the est

				outputErr := runner.Run(data, testRunSTF)
				expectedErr := data.ExpectError()

				if expectedErr != nil {
					if outputErr == nil {
						fmt.Printf("Test %s failed: expected error but got none\n", testFile.Name)
						failed++
					} else {
						fmt.Printf("Test %s passed (expected error: %v)\n", testFile.Name, expectedErr)
						passed++
					}
					// Check the error message
				} else {
					if outputErr != nil {
						fmt.Printf("Test %s failed: unexpected error: %v\n", testFile.Name, outputErr)
						failed++
					} else {
						err = runner.Verify(data)
						if err != nil {
							fmt.Printf("Test %s failed: %v\n", testFile.Name, err)
							failed++
						} else {
							fmt.Printf("Test %s passed\n", testFile.Name)
							passed++
						}
					}
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
