//go:build linux && amd64 && cgo

package PVM_test

import (
	"os"
	"path/filepath"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/interpreter"
	"github.com/New-JAMneration/JAM-Protocol/PVM/recompiler"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type psiMWorkload struct {
	code []byte
	arg  PVM.Argument
	gas  types.Gas
	hash types.OpaqueHash
}

func loadPsiMBenchWorkload(b *testing.B) psiMWorkload {
	b.Helper()
	types.SetTinyMode()
	dumpDirs, err := resolvePsiADumpDirs()
	if err != nil {
		b.Skipf("no psi_a dumps: %v", err)
	}
	for _, dumpDir := range dumpDirs {
		records, err := loadPsiADumpRecords(dumpDir)
		if err != nil {
			b.Fatal(err)
		}
		for _, rec := range records {
			if rec.ExitReasonType != "Halt" || rec.GasIn == 0 {
				continue
			}
			code, err := os.ReadFile(filepath.Join(dumpDir, rec.CodeRelPath))
			if err != nil {
				b.Fatal(err)
			}
			argBytes, err := os.ReadFile(filepath.Join(dumpDir, rec.ArgumentRelPath))
			if err != nil {
				b.Fatal(err)
			}
			b.Logf("bench workload dump=%s rec=%s code=%dB arg=%dB gas_in=%d gas_consumed=%d code_hash=0x%x",
				filepath.Base(dumpDir), filepath.Base(rec.path), len(code), len(argBytes), rec.GasIn, rec.GasIn-rec.GasOut, rec.CodeHash[:])
			return psiMWorkload{
				code: code,
				arg:  PVM.Argument(argBytes),
				gas:  types.Gas(rec.GasIn),
				hash: rec.CodeHash,
			}
		}
	}
	b.Skip("no Halt dump record")
	return psiMWorkload{}
}

func runPsiMBench(b *testing.B, backend string, w psiMWorkload, hash types.OpaqueHash) {
	b.Helper()
	addition := minimalAccumulateHostArgs(hash)
	got, err := PVM.Psi_M_OnBackend(
		backend,
		PVM.StandardCodeFormat(w.code),
		backendConsistencyEntry,
		w.gas,
		w.arg,
		PVM.AccumulateOmegas,
		addition,
	)
	if err != nil {
		b.Fatalf("Psi_M_OnBackend(%s): %v", backend, err)
	}
	if got.ReasonOrBytes == nil {
		b.Fatalf("Psi_M_OnBackend(%s): empty result", backend)
	}
}

func resetCachesForCold(b *testing.B) {
	b.Helper()
	PVM.ResetProgramCacheForTest()
	recompiler.ResetProgramStoreForTest()
}

func BenchmarkPsiM(b *testing.B) {
	w := loadPsiMBenchWorkload(b)
	var zero types.OpaqueHash

	benches := []struct {
		name    string
		backend string
		hash    types.OpaqueHash
		cold    bool
		prewarm bool
	}{
		{"Interpreter/UncachedZeroHash", PVM.BackendInterpreter, zero, false, false},
		{"Recompiler/UncachedZeroHash", PVM.BackendRecompiler, zero, false, false},
		{"Interpreter/ColdNamedHash", PVM.BackendInterpreter, w.hash, true, false},
		{"Recompiler/ColdNamedHash", PVM.BackendRecompiler, w.hash, true, false},
		{"Interpreter/WarmNamedHash", PVM.BackendInterpreter, w.hash, false, true},
		{"Recompiler/WarmNamedHash", PVM.BackendRecompiler, w.hash, false, true},
	}

	for _, bc := range benches {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			if bc.prewarm {
				runPsiMBench(b, bc.backend, w, bc.hash)
			}
			b.ResetTimer()
			for b.Loop() {
				if bc.cold {
					b.StopTimer()
					resetCachesForCold(b)
					b.StartTimer()
				}
				runPsiMBench(b, bc.backend, w, bc.hash)
			}
		})
	}
}
