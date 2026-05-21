package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace/diff"
)

func runPVMDiff(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing subcommand: find-diff | show | detail")
	}

	sub := args[0]
	fs := flag.NewFlagSet("pvm-diff "+sub, flag.ExitOnError)
	leftDir := fs.String("left", "", "left trace directory (interpreter)")
	rightDir := fs.String("right", "", "right trace directory (recompiler)")
	metaDir := fs.String("meta", "", "deblob metadata directory (optional)")
	step := fs.Int64("step", 0, "step index for detail view")
	from := fs.Int64("from", 0, "start step for show view")
	limit := fs.Int64("limit", 50, "number of steps for show view")
	fs.Parse(args[1:])

	_ = metaDir // will be used for InstrMeta lookup in future

	if *leftDir == "" || *rightDir == "" {
		return fmt.Errorf("--left and --right are required")
	}

	switch sub {
	case "find-diff":
		result, err := diff.FindFirstDivergence(*leftDir, *rightDir)
		if err != nil {
			return err
		}
		diff.PrintDivergence(os.Stdout, result)
		return nil

	case "show":
		return diff.PrintStepTable(os.Stdout, *leftDir, *rightDir, *from, *limit)

	case "detail":
		left, _ := diff.ReadStepAtPublic(*leftDir, *step)
		right, _ := diff.ReadStepAtPublic(*rightDir, *step)
		result := &diff.DivergenceResult{
			Step:   *step,
			Stream: "detail",
			Left:   left,
			Right:  right,
		}
		diff.PrintDivergence(os.Stdout, result)
		return nil

	default:
		return fmt.Errorf("unknown subcommand: %s (use find-diff, show, or detail)", sub)
	}
}
