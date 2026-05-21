// Package diff implements offline streaming comparison of two PVM trace directories.
package diff

import (
	"encoding/binary"
	"path/filepath"

	"github.com/New-JAMneration/JAM-Protocol/PVM/PVMtrace"
)

// DivergenceResult holds information about the first detected divergence.
type DivergenceResult struct {
	Step   int64
	Stream string
	Left   StepData
	Right  StepData
}

// StepData holds decoded values for one step from one side.
type StepData struct {
	PC       uint64
	Opcode   byte
	Gas      int64
	DstVal   uint64
	Src1Val  uint64
	Src2Val  uint64
	LoadAddr uint32
	LoadVal  uint64
	StoreAddr uint32
	StoreVal  uint64
}

// FindFirstDivergence walks both traces step-by-step and returns the first
// semantic mismatch. When PC and opcode match at a step, register/memory/gas
// fields are compared before treating a later PC skew as the root cause.
func FindFirstDivergence(leftDir, rightDir string) (*DivergenceResult, error) {
	for step := int64(0); ; step++ {
		left, errL := readStepAt(leftDir, step)
		right, errR := readStepAt(rightDir, step)
		if errL != nil && errR != nil {
			return nil, nil
		}
		if errL != nil || errR != nil {
			stream := "length"
			if errL == nil {
				stream = "left_longer"
			} else if errR == nil {
				stream = "right_longer"
			}
			return &DivergenceResult{
				Step:   step,
				Stream: stream,
				Left:   left,
				Right:  right,
			}, nil
		}

		if stream := firstDifferingStream(left, right); stream != "" {
			return &DivergenceResult{
				Step:   step,
				Stream: stream,
				Left:   left,
				Right:  right,
			}, nil
		}
	}
}

func firstDifferingStream(l, r StepData) string {
	if l.PC == r.PC && l.Opcode == r.Opcode {
		if l.DstVal != r.DstVal {
			return "dst_val"
		}
		if l.Src1Val != r.Src1Val {
			return "src1_val"
		}
		if l.Src2Val != r.Src2Val {
			return "src2_val"
		}
		if l.Gas != r.Gas {
			return "gas"
		}
		if l.LoadAddr != r.LoadAddr || l.LoadVal != r.LoadVal {
			return "loads"
		}
		if l.StoreAddr != r.StoreAddr || l.StoreVal != r.StoreVal {
			return "stores"
		}
		return ""
	}

	if l.PC != r.PC {
		return "pc"
	}
	if l.Opcode != r.Opcode {
		return "opcode"
	}
	if l.DstVal != r.DstVal {
		return "dst_val"
	}
	if l.Src1Val != r.Src1Val {
		return "src1_val"
	}
	if l.Src2Val != r.Src2Val {
		return "src2_val"
	}
	if l.Gas != r.Gas {
		return "gas"
	}
	if l.LoadAddr != r.LoadAddr || l.LoadVal != r.LoadVal {
		return "loads"
	}
	if l.StoreAddr != r.StoreAddr || l.StoreVal != r.StoreVal {
		return "stores"
	}
	return ""
}

// ReadStepAtPublic is the exported version of readStepAt for CLI use.
func ReadStepAtPublic(dir string, step int64) (StepData, error) {
	return readStepAt(dir, step)
}

// readStepAt reads all stream values at a given step index.
func readStepAt(dir string, step int64) (StepData, error) {
	var d StepData

	if v, err := readU64At(filepath.Join(dir, PVMtrace.StreamPC), step); err == nil {
		d.PC = v
	}
	if v, err := readU8At(filepath.Join(dir, PVMtrace.StreamOpcode), step); err == nil {
		d.Opcode = v
	}
	if v, err := readI64At(filepath.Join(dir, PVMtrace.StreamGas), step); err == nil {
		d.Gas = v
	}
	if v, err := readU64At(filepath.Join(dir, PVMtrace.StreamDstVal), step); err == nil {
		d.DstVal = v
	}
	if v, err := readU64At(filepath.Join(dir, PVMtrace.StreamSrc1Val), step); err == nil {
		d.Src1Val = v
	}
	if v, err := readU64At(filepath.Join(dir, PVMtrace.StreamSrc2Val), step); err == nil {
		d.Src2Val = v
	}
	if addr, val, err := readMemAt(filepath.Join(dir, PVMtrace.StreamLoads), step); err == nil {
		d.LoadAddr = addr
		d.LoadVal = val
	}
	if addr, val, err := readMemAt(filepath.Join(dir, PVMtrace.StreamStores), step); err == nil {
		d.StoreAddr = addr
		d.StoreVal = val
	}
	return d, nil
}

func readU64At(path string, step int64) (uint64, error) {
	r, err := PVMtrace.NewGzipRecordReader(path, 8)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	var buf [8]byte
	for i := int64(0); i <= step; i++ {
		if err := r.ReadRecord(buf[:]); err != nil {
			return 0, err
		}
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func readI64At(path string, step int64) (int64, error) {
	v, err := readU64At(path, step)
	return int64(v), err
}

func readU8At(path string, step int64) (byte, error) {
	r, err := PVMtrace.NewGzipRecordReader(path, 1)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	var buf [1]byte
	for i := int64(0); i <= step; i++ {
		if err := r.ReadRecord(buf[:]); err != nil {
			return 0, err
		}
	}
	return buf[0], nil
}

func readMemAt(path string, step int64) (uint32, uint64, error) {
	r, err := PVMtrace.NewGzipRecordReader(path, 12)
	if err != nil {
		return 0, 0, err
	}
	defer r.Close()
	var buf [12]byte
	for i := int64(0); i <= step; i++ {
		if err := r.ReadRecord(buf[:]); err != nil {
			return 0, 0, err
		}
	}
	addr := binary.LittleEndian.Uint32(buf[0:4])
	val := binary.LittleEndian.Uint64(buf[4:12])
	return addr, val, nil
}
