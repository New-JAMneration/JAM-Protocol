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
// through Psi_M on both backends (via PVM.WithExecutionBackend) and requires
// matching Gas + ReasonOrBytes. Blobs are 0.7.2 programs under 0.8.0 semantics;
// the assertion is backend consistency, not GP correctness.
func TestInterpreterVsRecompilerProgramBlobs(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	if err := PVM.SetExecutionBackend(PVM.BackendInterpreter); err != nil {
		t.Fatalf("interpreter: %v", err)
	}
	if err := PVM.SetExecutionBackend(PVM.BackendRecompiler); err != nil {
		t.Fatalf("recompiler: %v", err)
	}
	// Restore default for other tests in the package.
	t.Cleanup(func() { _ = PVM.SetExecutionBackend(PVM.BackendInterpreter) })

	codes, err := loadProgramCodes(backendConsistencyBlobDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) < backendConsistencyMinBlobs {
		t.Fatalf("need >= %d decodable program blobs, found %d in %s",
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
	err := PVM.WithExecutionBackend(backend, func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = fmt.Sprint(r)
			}
		}()
		got = PVM.Psi_M(
			PVM.StandardCodeFormat(code),
			backendConsistencyEntry,
			backendConsistencyGas,
			arg,
			PVM.AccumulateOmegas,
			addition,
		)
	})
	if err != nil {
		t.Fatalf("WithExecutionBackend(%s): %v", backend, err)
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
