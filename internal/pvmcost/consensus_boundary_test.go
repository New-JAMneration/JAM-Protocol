package pvmcost_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/pvmcost"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// TestConsensusStructsDoNotContainCost asserts that none of the
// consensus-serialized structs in internal/types embed (transitively)
// any pvmcost type. The pvmcost package doc and #974 forbid this —
// cost values are observability output, not consensus state. A
// reflect-based walk catches accidental embedding before CI tests get
// confused by mismatched block hashes or conformance vector drift.
//
// To extend coverage: add new consensus structs to the
// consensusStructs slice below. The walk follows struct fields,
// slices, arrays, pointers, channels, and maps to arbitrary depth.
func TestConsensusStructsDoNotContainCost(t *testing.T) {
	forbidden := map[reflect.Type]bool{
		reflect.TypeOf(pvmcost.ExecCost{}):         true,
		reflect.TypeOf(pvmcost.IsAuthorizedCost{}): true,
		reflect.TypeOf(pvmcost.RefineCost{}):       true,
		reflect.TypeOf(pvmcost.AccumulateCost{}):   true,
	}

	// Consensus-serialized structs. New additions to internal/types
	// that participate in block hashing / conformance vectors should
	// be added here too.
	consensusStructs := []any{
		types.Block{},
		types.Header{},
		types.Extrinsic{},
		types.WorkPackage{},
		types.WorkItem{},
		types.WorkResult{},
		types.WorkExecResult{},
		types.WorkReport{},
		types.Privileges{},
		types.DisputesExtrinsic{},
	}

	for _, s := range consensusStructs {
		root := reflect.TypeOf(s)
		walk(t, root, forbidden, []string{root.Name()}, map[reflect.Type]bool{})
	}
}

// walk recursively inspects ty's fields / elements, failing the test
// when it lands on a type listed in forbidden. visited prevents
// infinite recursion on self-referential types (Mmr's tree etc.).
//
// Limitation: interface-typed fields are not walkable — reflect can't
// see the dynamic type behind an `any` field, so a pvmcost value
// smuggled through one would not be caught. Consensus structs have no
// interface fields today; if one ever appears, this guard needs a
// runtime-value check for it.
func walk(t *testing.T, ty reflect.Type, forbidden map[reflect.Type]bool, path []string, visited map[reflect.Type]bool) {
	t.Helper()
	if ty == nil || visited[ty] {
		return
	}
	visited[ty] = true

	if forbidden[ty] {
		t.Errorf("consensus struct contains forbidden pvmcost type %s at %s",
			ty.Name(), strings.Join(path, "."))
		return
	}

	// Copy before extending: plain append(path, x) can share the backing
	// array between sibling fields and corrupt the failure-message path.
	extend := func(seg string) []string {
		return append(append(make([]string, 0, len(path)+1), path...), seg)
	}

	switch ty.Kind() {
	case reflect.Struct:
		for i := 0; i < ty.NumField(); i++ {
			f := ty.Field(i)
			walk(t, f.Type, forbidden, extend(f.Name), visited)
		}
	case reflect.Slice, reflect.Array, reflect.Pointer, reflect.Chan:
		walk(t, ty.Elem(), forbidden, extend("[elem]"), visited)
	case reflect.Map:
		walk(t, ty.Key(), forbidden, extend("[key]"), visited)
		walk(t, ty.Elem(), forbidden, extend("[value]"), visited)
	}
}
