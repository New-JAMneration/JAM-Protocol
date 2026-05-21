// Command pvmtrace provides offline PVM trace analysis tools.
//
// Subcommands:
//
//	pvmtrace deblob program <target-file> <codehash>
//	pvmtrace pvm-diff find-diff --meta <dir> --left <dir> --right <dir>
//	pvmtrace pvm-diff detail --meta <dir> --left <dir> --right <dir> --step <N>
//	pvmtrace pvm-diff show --meta <dir> --left <dir> --right <dir> --from <N> --limit <N>
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "deblob":
		if err := runDeblob(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pvmtrace deblob: %v\n", err)
			os.Exit(1)
		}
	case "pvm-diff":
		if err := runPVMDiff(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "pvmtrace pvm-diff: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  pvmtrace deblob program <target-file> <codehash>
  pvmtrace pvm-diff find-diff --left <dir> --right <dir>
  pvmtrace pvm-diff show --left <dir> --right <dir> --from <N> --limit <N>
  pvmtrace pvm-diff detail --left <dir> --right <dir> --step <N>
`)
}
