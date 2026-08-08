//go:build linux && amd64 && cgo

package PVM_test

// Transitional Ψ_A dual-backend consistency harness (0.7.2 → 0.8.0).
//
// There is not yet a formal 0.8.0 jam-conformance corpus for accumulate / Ψ_A
// backend checks. This file (and PVM/consistent-testdata/) exists only to bridge
// that gap: we extract workable cases from 0.7.2 conformance fuzz traces by
// dumping services that actually enter accumulate / Psi_M (code blob +
// serialized Ψ_A argument + gas), then compare interpreter vs recompiler via
// Psi_M_OnBackend.
//
// When official 0.8.0 test data lands, discard this file and
// consistent-testdata/ in favour of that corpus.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	PVM "github.com/New-JAMneration/JAM-Protocol/PVM"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/interpreter"
	_ "github.com/New-JAMneration/JAM-Protocol/PVM/recompiler"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const (
	// Default dump root under this package (each dump_<folder>/ has records/).
	// Override single dump with JAM_PSI_A_DUMP_DIR, or root with JAM_PSI_A_DUMP_ROOT.
	defaultPsiADumpRoot     = "consistent-testdata"
	backendConsistencyEntry = PVM.ProgramCounter(5) // Ψ_A entry
)

type psiADumpRecordFile struct {
	ServiceID       uint64 `json:"service_id"`
	Timeslot        uint64 `json:"timeslot"`
	GasIn           uint64 `json:"gas_in"`
	CodeSHA256      string `json:"code_sha256"`
	CodeRelPath     string `json:"code_relpath"`
	ArgumentRelPath string `json:"argument_relpath"`
	ExitReasonType  string `json:"exit_reason_type"`
	ResultNonNil    bool   `json:"result_non_nil"`
	Folder          string `json:"folder"`
	Block           string `json:"block"`
	Seq             uint64 `json:"seq"`
}

// TestInterpreterVsRecompilerPsiADump runs dual-backend Psi_M on the
// transitional dumps under consistent-testdata/ (see file comment). Expect
// matching Gas + ReasonOrBytes. Not a full STF host-state replay.
func TestInterpreterVsRecompilerPsiADump(t *testing.T) {
	types.SetTinyMode()
	t.Cleanup(types.SetTinyMode)

	if PVM.Psi_M_interpreterHook == nil {
		t.Fatal("interpreter backend not linked")
	}
	if PVM.Psi_M_recompilerHook == nil {
		t.Fatal("recompiler backend not linked")
	}

	dumpDirs, err := resolvePsiADumpDirs()
	if err != nil {
		t.Skipf("no psi_a dumps: %v", err)
	}

	var haltNotes []string
	for _, dumpDir := range dumpDirs {
		dumpName := filepath.Base(dumpDir)
		t.Run(dumpName, func(t *testing.T) {
			records, err := loadPsiADumpRecords(dumpDir)
			if err != nil {
				t.Fatalf("load records: %v", err)
			}
			if len(records) == 0 {
				t.Skip("no records")
			}
			for _, rec := range records {
				name := filepath.Base(rec.path)
				t.Run(name, func(t *testing.T) {
					code, err := os.ReadFile(filepath.Join(dumpDir, rec.CodeRelPath))
					if err != nil {
						t.Fatalf("read code: %v", err)
					}
					argBytes, err := os.ReadFile(filepath.Join(dumpDir, rec.ArgumentRelPath))
					if err != nil {
						t.Fatalf("read argument: %v", err)
					}
					if !isPsiAConsistencyExit(rec.ExitReasonType) {
						t.Skipf("skip exit=%s (need Halt / Panic / Page Fault; OOG does not count)", rec.ExitReasonType)
					}
					gas := types.Gas(rec.GasIn)
					if gas <= 0 {
						t.Skipf("skip gas_in=%d (svc=%d exit=%s)", rec.GasIn, rec.ServiceID, rec.ExitReasonType)
					}

					t.Logf("folder=%s block=%s svc=%d gas_in=%d dump_exit=%s result_non_nil=%v code=%dB arg=%dB",
						rec.Folder, rec.Block, rec.ServiceID, rec.GasIn, rec.ExitReasonType, rec.ResultNonNil, len(code), len(argBytes))

					gotI, panicI := runPsiMWithGas(t, PVM.BackendInterpreter, code, PVM.Argument(argBytes), gas)
					gotR, panicR := runPsiMWithGas(t, PVM.BackendRecompiler, code, PVM.Argument(argBytes), gas)

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
					t.Logf("ok Gas=%d ReasonOrBytes=%#v (backends match)", gotI.Gas, gotI.ReasonOrBytes)
					if rec.ExitReasonType == "Halt" {
						note := fmt.Sprintf("%s/%s svc=%d block=%s result_non_nil=%v gas_in=%d matched_gas=%d matched_reason=%#v",
							dumpName, name, rec.ServiceID, rec.Block, rec.ResultNonNil, rec.GasIn, gotI.Gas, gotI.ReasonOrBytes)
						haltNotes = append(haltNotes, note)
						t.Logf("HALT dump record: %s", note)
					}
				})
			}
		})
	}
	if len(haltNotes) > 0 {
		t.Logf("Halt dump records (%d):", len(haltNotes))
		for _, n := range haltNotes {
			t.Logf("  HALT: %s", n)
		}
	}
}

func isPsiAConsistencyExit(exit string) bool {
	switch exit {
	case "Halt", "Panic", "Page Fault":
		return true
	default:
		return false
	}
}

func resolvePsiADumpDirs() ([]string, error) {
	if single := os.Getenv("JAM_PSI_A_DUMP_DIR"); single != "" {
		return []string{single}, nil
	}
	root := os.Getenv("JAM_PSI_A_DUMP_ROOT")
	if root == "" {
		root = defaultPsiADumpRoot
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "dump_") {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "records")); err != nil {
			continue
		}
		dirs = append(dirs, dir)
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("no dump_* under %s", root)
	}
	return dirs, nil
}

type loadedDumpRecord struct {
	psiADumpRecordFile
	path string
}

func loadPsiADumpRecords(dumpDir string) ([]loadedDumpRecord, error) {
	recDir := filepath.Join(dumpDir, "records")
	entries, err := os.ReadDir(recDir)
	if err != nil {
		return nil, err
	}
	var out []loadedDumpRecord
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(recDir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var rec psiADumpRecordFile
		if err := json.Unmarshal(raw, &rec); err != nil {
			return nil, err
		}
		out = append(out, loadedDumpRecord{psiADumpRecordFile: rec, path: path})
	}
	return out, nil
}

func runPsiMWithGas(t *testing.T, backend string, code []byte, arg PVM.Argument, gas types.Gas) (got PVM.Psi_M_ReturnType, panicMsg string) {
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
		gas,
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
