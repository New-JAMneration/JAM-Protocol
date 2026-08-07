//go:build linux && amd64 && cgo

package PVM_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/interpreter"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler"
	"github.com/New-JAMneration/JAM-Protocol/internal/service_account"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const (
	backendConsistencyBlobDir   = "testdata/psi_a_consistency/blobs"
	backendConsistencyMinBlobs  = 30
	backendConsistencyGas       = types.Gas(50_000_000)
	backendConsistencyEntry     = PVM.ProgramCounter(5) // Ψ_A entry
)

// TestInterpreterVsRecompilerProgramBlobs runs each extracted MetaCode program
// through Psi_M_OnBackend on both backends and requires matching Gas +
// ReasonOrBytes (no process-global ExecutionBackend swap).
//
// Blob corpus is not committed: 0.7.2 traces are the wrong generation for
// 0.8.0 semantics. Regenerate locally when suitable traces exist:
//
//	python3 scripts/scan_psi_a_program_blobs.py …  # see script help / JSON outs
//
// then place MetaCode .bin files under testdata/psi_a_consistency/blobs/.
func TestInterpreterVsRecompilerProgramBlobs(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	if PVM.Psi_M_interpreterHook == nil {
		t.Fatal("interpreter backend not linked")
	}
	if PVM.Psi_M_recompilerHook == nil {
		t.Fatal("recompiler backend not linked")
	}

	codes, err := loadProgramCodes(backendConsistencyBlobDir)
	if err != nil {
		t.Skipf("no local blob corpus (%v); regenerate with scripts/scan_psi_a_program_blobs.py when 0.8.0 traces are available", err)
	}
	if len(codes) < backendConsistencyMinBlobs {
		t.Skipf("need >= %d decodable program blobs, found %d in %s (local corpus only)",
			backendConsistencyMinBlobs, len(codes), backendConsistencyBlobDir)
	}

	arg := accumulateEmptyArgument(t)
	for _, tc := range codes {
		t.Run(tc.name, func(t *testing.T) {
			gotI, panicI := runPsiM(t, PVM.BackendInterpreter, tc.code, arg)
			gotR, panicR := runPsiM(t, PVM.BackendRecompiler, tc.code, arg)

			if panicI != panicR {
				t.Fatalf("panic mismatch\n  interpreter: %v\n  recompiler:  %v", panicI, panicR)
			}
			if panicI != "" {
				t.Logf("both panicked: %v", panicI)
				return
			}
			if gotI.Gas != gotR.Gas {
				t.Fatalf("Gas: interpreter=%d recompiler=%d", gotI.Gas, gotR.Gas)
			}
			if !reasonEqual(gotI.ReasonOrBytes, gotR.ReasonOrBytes) {
				t.Fatalf("ReasonOrBytes mismatch\n  interpreter: %#v\n  recompiler:  %#v",
					gotI.ReasonOrBytes, gotR.ReasonOrBytes)
			}
		})
	}
}

type programCase struct {
	name string
	code []byte
}

func loadProgramCodes(dir string) ([]programCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w (extract blobs first)", dir, err)
	}
	var out []programCase
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".bin" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		_, code, err := service_account.DecodeMetaCode(raw)
		if err != nil || len(code) == 0 {
			continue
		}
		out = append(out, programCase{name: e.Name(), code: []byte(code)})
	}
	return out, nil
}

func accumulateEmptyArgument(t *testing.T) PVM.Argument {
	t.Helper()
	enc := types.NewEncoder()
	var serialized []byte
	for _, v := range []uint64{0, 1, 0} { // timeslot, serviceId, |operands|
		b, err := enc.EncodeUint(v)
		if err != nil {
			t.Fatalf("EncodeUint: %v", err)
		}
		serialized = append(serialized, b...)
	}
	return PVM.Argument(serialized)
}

func runPsiM(t *testing.T, backend string, code []byte, arg PVM.Argument) (got PVM.Psi_M_ReturnType, panicMsg string) {
	t.Helper()
	addition := minimalAccumulateHostArgs()
	defer func() {
		if r := recover(); r != nil {
			panicMsg = fmt.Sprint(r)
		}
	}()
	var err error
	got, err = PVM.Psi_M_OnBackend(
		backend,
		PVM.StandardCodeFormat(code),
		backendConsistencyEntry,
		backendConsistencyGas,
		arg,
		PVM.AccumulateOmegas,
		addition,
	)
	if err != nil {
		t.Fatalf("Psi_M_OnBackend(%s): %v", backend, err)
	}
	return got, panicMsg
}

func minimalAccumulateHostArgs() PVM.HostCallArgs {
	sid := types.ServiceID(1)
	acct := types.ServiceAccount{
		PreimageLookup: types.PreimagesMapEntry{},
		LookupDict:     types.LookupMetaMapEntry{},
		StorageDict:    types.Storage{},
	}
	state := types.ServiceAccountState{sid: acct}
	storage := types.StateKeyVals{}
	partial := types.PartialStateSet{
		ServiceAccounts: state,
		AlwaysAccum:     types.AlwaysAccumulateMap{},
	}
	return PVM.HostCallArgs{
		GeneralArgs: PVM.GeneralArgs{
			ServiceAccount:      &acct,
			ServiceID:           &sid,
			ServiceAccountState: &state,
			StorageKeyVal:       &storage,
		},
		AccumulateArgs: PVM.AccumulateArgs{
			ResultContextX: PVM.ResultContext{
				ServiceID:         sid,
				PartialState:      partial,
				DeferredTransfers: []types.DeferredTransfer{},
				ServiceBlobs:      map[types.OpaqueHash]types.ServiceBlob{},
				StorageKeyVal:     &storage,
			},
			ResultContextY: PVM.ResultContext{
				ServiceID:         sid,
				PartialState:      partial.DeepCopy(),
				DeferredTransfers: []types.DeferredTransfer{},
				ServiceBlobs:      map[types.OpaqueHash]types.ServiceBlob{},
				StorageKeyVal:     &storage,
			},
			OperandOrDeferredTransfers: nil,
			Timeslot:                   0,
		},
	}
}

func reasonEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	// Normalize []byte vs nil empty.
	ab, aOK := a.([]byte)
	bb, bOK := b.([]byte)
	if aOK && bOK {
		return string(ab) == string(bb)
	}
	return false
}
