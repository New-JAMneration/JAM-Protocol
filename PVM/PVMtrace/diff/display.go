package diff

import (
	"fmt"
	"io"
	"text/tabwriter"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
)

// PrintDivergence writes a human-readable divergence report to w.
func PrintDivergence(w io.Writer, result *DivergenceResult) {
	if result == nil {
		fmt.Fprintln(w, "no divergence found: traces are identical")
		return
	}

	fmt.Fprintf(w, "first divergence: step=%d stream=%s\n\n", result.Step, result.Stream)

	leftName := opcodeLabel(result.Left.Opcode)
	rightName := opcodeLabel(result.Right.Opcode)
	if result.Left.PC == result.Right.PC && result.Left.Opcode == result.Right.Opcode {
		fmt.Fprintf(w, "step=%d pc=0x%x opcode=%d (%s)\n",
			result.Step, result.Left.PC, result.Left.Opcode, leftName)
	} else {
		fmt.Fprintf(w, "step=%d\n", result.Step)
		fmt.Fprintf(w, "  left:  pc=0x%x opcode=%d (%s)\n",
			result.Left.PC, result.Left.Opcode, leftName)
		fmt.Fprintf(w, "  right: pc=0x%x opcode=%d (%s)\n",
			result.Right.PC, result.Right.Opcode, rightName)
	}

	fmt.Fprintln(w, "  left (interpreter):")
	printStepData(w, "    ", result.Left)
	fmt.Fprintln(w, "  right (recompiler):")
	printStepData(w, "    ", result.Right)
}

func opcodeLabel(op byte) string {
	name := PVM.OpcodeName(op)
	if name == "" {
		return "?"
	}
	return name
}

func printStepData(w io.Writer, prefix string, d StepData) {
	fmt.Fprintf(w, "%spc=0x%x opcode=%d gas=%d\n", prefix, d.PC, d.Opcode, d.Gas)
	fmt.Fprintf(w, "%sdst_val=0x%016x src1_val=0x%016x src2_val=0x%016x\n",
		prefix, d.DstVal, d.Src1Val, d.Src2Val)
	if d.LoadAddr != 0 || d.LoadVal != 0 {
		fmt.Fprintf(w, "%sload: addr=0x%08x val=0x%016x\n", prefix, d.LoadAddr, d.LoadVal)
	}
	if d.StoreAddr != 0 || d.StoreVal != 0 {
		fmt.Fprintf(w, "%sstore: addr=0x%08x val=0x%016x\n", prefix, d.StoreAddr, d.StoreVal)
	}
}

func stepsDiffer(left, right StepData) bool {
	return firstDifferingStream(left, right) != ""
}

// PrintStepTable writes a tabular comparison for a range of steps.
func PrintStepTable(w io.Writer, leftDir, rightDir string, from, limit int64) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "step\tL pc\tR pc\top\tgas\t| L dst_val\t| R dst_val\n")
	fmt.Fprintf(tw, "----\t----\t----\t-\t---\t| ---------\t| ---------\n")

	for step := from; step < from+limit; step++ {
		left, errL := readStepAt(leftDir, step)
		right, errR := readStepAt(rightDir, step)
		if errL != nil || errR != nil {
			break
		}

		marker := " "
		if stepsDiffer(left, right) {
			marker = "*"
		}

		fmt.Fprintf(tw, "%s%d\t0x%x\t0x%x\t%d\t%d\t| 0x%016x\t| 0x%016x\n",
			marker, step, left.PC, right.PC, left.Opcode, left.Gas, left.DstVal, right.DstVal)
	}
	return tw.Flush()
}
