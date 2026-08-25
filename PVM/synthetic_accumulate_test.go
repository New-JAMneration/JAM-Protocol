//go:build linux && amd64 && cgo

package PVM_test

import (
	"fmt"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/interpreter"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestSynthGrowHeapTargetWithinRange(t *testing.T) {
	types.SetTinyMode()
	for _, kind := range []synthKind{synthSmall, synthLarge} {
		t.Run(string(kind), func(t *testing.T) {
			prog := buildSynthProgram(kind)
			if !prog.lay.growInRange() {
				t.Fatalf("grow target n=%d not in (h=%d, b=%d]",
					prog.lay.targetPages, prog.lay.heapPages, prog.lay.maxPages)
			}
			headroom := prog.lay.maxPages - prog.lay.targetPages
			if headroom < 1024 {
				t.Fatalf("grow target too close to max: n=%d b=%d headroom=%d",
					prog.lay.targetPages, prog.lay.maxPages, headroom)
			}
			t.Logf("\n%s", formatSynthInfo(prog))
		})
	}
}

func TestInterpreterVsRecompilerSynthetic(t *testing.T) {
	types.SetTinyMode()
	if PVM.Psi_M_interpreterHook == nil || PVM.Psi_M_recompilerHook == nil {
		t.Fatal("both backends must be linked")
	}
	for _, kind := range []synthKind{synthSmall, synthLarge} {
		t.Run(string(kind), func(t *testing.T) {
			prog := buildSynthProgram(kind)
			t.Logf("\n%s", formatSynthInfo(prog))

			gotI, panicI := runSynthPsiM(t, PVM.BackendInterpreter, prog)
			gotR, panicR := runSynthPsiM(t, PVM.BackendRecompiler, prog)
			if panicI != panicR {
				t.Fatalf("Go panic mismatch\n  interpreter: %v\n  recompiler:  %v", panicI, panicR)
			}
			if panicI != "" {
				t.Fatalf("both Go-panicked: %v", panicI)
			}
			if gotI.Gas != gotR.Gas {
				t.Fatalf("Gas: interpreter=%d recompiler=%d", gotI.Gas, gotR.Gas)
			}
			if !reasonEqual(gotI.ReasonOrBytes, gotR.ReasonOrBytes) {
				t.Fatalf("ReasonOrBytes mismatch\n  interpreter: %#v\n  recompiler:  %#v",
					gotI.ReasonOrBytes, gotR.ReasonOrBytes)
			}
			if _, ok := gotI.ReasonOrBytes.([]byte); !ok && gotI.ReasonOrBytes != nil {
				t.Fatalf("expected Halt payload ([]byte or nil), got %T %#v", gotI.ReasonOrBytes, gotI.ReasonOrBytes)
			}
			t.Logf("ok Gas=%d ReasonOrBytes=%#v", gotI.Gas, gotI.ReasonOrBytes)
		})
	}
}

func runSynthPsiM(t *testing.T, backend string, prog synthProgram) (got PVM.Psi_M_ReturnType, panicMsg string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			panicMsg = fmt.Sprint(r)
		}
	}()
	got, err := PVM.Psi_M_OnBackend(
		backend,
		PVM.StandardCodeFormat(prog.code),
		synthEntryPC,
		synthGas,
		prog.arg,
		PVM.AccumulateOmegas,
		prog.hostArgs(),
	)
	if err != nil {
		t.Fatalf("Psi_M_OnBackend(%s): %v", backend, err)
	}
	return got, panicMsg
}

func BenchmarkPsiMSynthetic(b *testing.B) {
	types.SetTinyMode()
	var zero types.OpaqueHash
	backends := []struct {
		name string
		id   string
	}{
		{"Interpreter", PVM.BackendInterpreter},
		{"Recompiler", PVM.BackendRecompiler},
	}
	for _, kind := range []synthKind{synthSmall, synthLarge} {
		prog := buildSynthProgram(kind)
		b.Logf("\n%s", formatSynthInfo(prog))
		for _, backend := range backends {
			// Warm then Cold on the same backend so the pair is adjacent in the table.
			b.Run(string(kind)+"/"+backend.name, func(b *testing.B) {
				b.Run("Warm", func(b *testing.B) {
					runSynthCacheBench(b, backend.id, prog, prog.hash, false, true)
				})
				b.Run("Cold", func(b *testing.B) {
					runSynthCacheBench(b, backend.id, prog, prog.hash, true, false)
				})
				b.Run("Uncached", func(b *testing.B) {
					runSynthCacheBench(b, backend.id, prog, zero, false, false)
				})
			})
		}
	}
}

func runSynthCacheBench(
	b *testing.B,
	backend string,
	prog synthProgram,
	hash types.OpaqueHash,
	cold, prewarm bool,
) {
	b.Helper()
	b.ReportAllocs()
	if prewarm {
		runSynthPsiMBench(b, backend, prog, hash)
	}
	b.ResetTimer()
	for b.Loop() {
		if cold {
			b.StopTimer()
			resetCachesForCold(b)
			b.StartTimer()
		}
		runSynthPsiMBench(b, backend, prog, hash)
	}
}

func runSynthPsiMBench(b *testing.B, backend string, prog synthProgram, hash types.OpaqueHash) {
	b.Helper()
	addition := prog.hostArgs()
	addition.CodeHash = hash
	got, err := PVM.Psi_M_OnBackend(
		backend,
		PVM.StandardCodeFormat(prog.code),
		synthEntryPC,
		synthGas,
		prog.arg,
		PVM.AccumulateOmegas,
		addition,
	)
	if err != nil {
		b.Fatalf("Psi_M_OnBackend(%s): %v", backend, err)
	}
	switch got.ReasonOrBytes {
	case PVM.PANIC, PVM.OUT_OF_GAS:
		b.Fatalf("Psi_M_OnBackend(%s): %v", backend, got.ReasonOrBytes)
	}
}
