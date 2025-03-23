package PolkaVM_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/PolkaVM"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

const hostCallTestVectorsDir = "../pkg/test_data/host_function/host_function"

var hostCallMap = map[string]PolkaVM.OperationType{
	"Eject":             PolkaVM.EjectOp,
	"Export":            PolkaVM.ExportOp,
	"Expunge":           PolkaVM.ExpungeOp,
	"Forget":            PolkaVM.ForgetOp,
	"Historical_lookup": PolkaVM.HistoricalLookupOp,
	"Import":            PolkaVM.FetchOp, // by process of elimination, Import has to be FetchOp.
	"Info":              PolkaVM.InfoOp,
	"Invoke":            PolkaVM.InvokeOp,
	"Lookup":            PolkaVM.LookupOp,
	"Machine":           PolkaVM.MachineOp,
	"New":               PolkaVM.NewOp,
	"Peek":              PolkaVM.PeekOp,
	"Poke":              PolkaVM.PokeOp,
	"Query":             PolkaVM.QueryOp,
	"Solicit":           PolkaVM.SolicitOp,
	"Transfer":          PolkaVM.TransferOp,
	"Upgrade":           PolkaVM.UpgradeOp,
	"Void":              PolkaVM.VoidOp,
	"Write":             PolkaVM.WriteOp,
	"Yield":             PolkaVM.YieldOp,
	"Zero":              PolkaVM.ZeroOp,
}

func TestHostCall(t *testing.T) {
	entries, err := os.ReadDir(hostCallTestVectorsDir)
	if err != nil {
		t.Fatalf("failed to read the host_function directory: %v", err)
	}

	for _, entry := range entries {
		functionName := entry.Name()

		if !entry.IsDir() {
			continue
		}

		if _, found := hostCallMap[functionName]; !found {
			continue
		}

		t.Run(functionName, func(t *testing.T) {
			testFunction(t, functionName)
		})
	}
}

func testFunction(t *testing.T, functionName string) {
	dirPath := path.Join(hostCallTestVectorsDir, functionName)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Errorf("failed to read the %v directory: %v", functionName, err)
		return
	}

	for _, entry := range entries {
		filename := entry.Name()
		filePath := path.Join(dirPath, filename)

		if testName, match := strings.CutSuffix(filename, ".json"); match {
			t.Run(testName, func(t *testing.T) {
				testOneCase(t, functionName, filePath)
			})
		}
	}
}

func testOneCase(t *testing.T, functionName string, filePath string) {
	t.Logf("running host call test in file: %v", filePath)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("failed to read file %v: %v", filePath, err)
	}

	var setup HostCallTestSetup
	err = json.Unmarshal(fileData, &setup)
	if err != nil {
		t.Errorf("failed to load test setup from file %v: %v", filePath, err)
	}

	t.Logf("successfully loaded file %v", filePath)

	operation := hostCallMap[functionName]
	omega := PolkaVM.HostCallFunctions[operation]
	if omega == nil {
		t.Skipf("skipping test: %v/%v", functionName, filePath)
	}

	input := PolkaVM.OmegaInput{
		Operation: operation,
		Gas:       setup.InitialGas,
		Registers: setup.InitialRegisters.ToRegisters(),
		Memory:    setup.InitialMemory.ToMemory(),
		Addition:  setup.GetHostCallArgs(),
	}

	output := omega(input)

	t.Logf("test setup: %v", setup)
	t.Logf("input: %v", input)
	t.Logf("output: %v", output)

	checkValue(t, "registers", setup.ExpectedRegisters.ToRegisters(), output.NewRegisters)
	checkMemory(t, setup.ExpectedMemory.ToMemory(), output.NewMemory)
	checkValue(t, "gas", PolkaVM.Gas(setup.ExpectedGas), output.NewGas)
	checkValue(t, "delta", setup.ExpectedDelta.ToServiceAccountState(), output.Addition.ServiceAccountState)
	checkValue(t, "service account", setup.ExpectedServiceAccount.ToServiceAccount(), output.Addition.ServiceAccount)
}

func checkValue(t *testing.T, name string, expected any, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s mismatch found:\nexpected %v\n     got %v",
			name, expected, actual)
	}
}

func checkMemory(t *testing.T, expected PolkaVM.Memory, actual PolkaVM.Memory) {
	for pageNumber, page := range expected.Pages {
		if !reflect.DeepEqual(page, actual.Pages[pageNumber]) {
			t.Errorf("memory mismatch found on page %v\nexpected %v,\n     got %v",
				pageNumber, page, actual.Pages[pageNumber])
		}
	}
}

type HostCallTestSetup struct {
	Name              string      `json:"name"`
	InitialGas        PolkaVM.Gas `json:"initial-gas"`
	InitialRegisters  Registers   `json:"initial-regs"`
	InitialMemory     Memory      `json:"initial-memory"`
	ExpectedGas       PolkaVM.Gas `json:"expected-gas"`
	ExpectedRegisters Registers   `json:"expected-regs"`
	ExpectedMemory    Memory      `json:"expected-memory"`

	HostCallArgs
}

type Registers map[string]uint64

func (r *Registers) ToRegisters() PolkaVM.Registers {
	var result PolkaVM.Registers

	for registerNumberStr, registerValue := range *r {
		registerNumber, err := strconv.ParseUint(registerNumberStr, 10, 64)
		if err != nil {
			panic(err)
		}

		if registerNumber >= 13 {
			panic(fmt.Errorf("invalid register number: %v", registerNumber))
		}

		result[registerNumber] = registerValue
	}

	return result
}

type Memory struct {
	Pages map[string]Page `json:"pages"`
}

func (m *Memory) ToMemory() PolkaVM.Memory {
	pages := make(map[uint32]*PolkaVM.Page)

	for pageNumberStr, pageSetup := range m.Pages {
		pageNumber, err := strconv.ParseUint(pageNumberStr, 10, 32)
		if err != nil {
			panic(err)
		}

		page := pageSetup.ToPage()
		pages[uint32(pageNumber)] = &page
	}

	return PolkaVM.Memory{
		Pages: pages,
	}
}

type Page struct {
	Value  []byte       `json:"value"`
	Access MemoryAccess `json:"access"`
}

func (p *Page) ToPage() PolkaVM.Page {
	value := make([]byte, PolkaVM.ZP)
	copy(value, p.Value)

	return PolkaVM.Page{
		Value:  value,
		Access: p.Access.ToMemoryAccess(),
	}
}

type MemoryAccess struct {
	Inaccessible bool `json:"inaccessible"`
	Writable     bool `json:"writable"`
	Readable     bool `json:"readable"`
}

func (a *MemoryAccess) ToMemoryAccess() PolkaVM.MemoryAccess {
	switch {
	case a.Inaccessible:
		return PolkaVM.MemoryInaccessible
	case a.Writable: // ignores a.Readable in this case
		return PolkaVM.MemoryReadWrite
	case a.Readable:
		return PolkaVM.MemoryReadOnly
	default:
		return PolkaVM.MemoryInaccessible
	}
}

type HostCallArgs struct {
	GeneralArgs
	RefineArgs
	AccumulateArgs
}

func (h HostCallArgs) GetHostCallArgs() PolkaVM.HostCallArgs {
	var result PolkaVM.HostCallArgs

	// General args
	if h.InitialServiceAccount != nil {
		result.ServiceAccount = h.InitialServiceAccount.ToServiceAccount()
	}

	if h.InitialServiceIndex != nil {
		result.ServiceId = *h.InitialServiceIndex
	}

	result.ServiceAccountState = h.InitialDelta.ToServiceAccountState()

	// Refine args
	result.RefineMap = h.InitialRefineMap
	result.ImportSegment = h.InitialImportSegment
	result.ExportSegment = h.InitialExportSegment
	result.ExportSegmentIndex = h.InitialExportSegmentIndex

	// Accumulate args
	if h.InitialTimeslot != nil {
		result.Timeslot = types.TimeSlot(*h.InitialTimeslot)
	}

	if h.InitialXContentX != nil {
		result.ResultContextX = h.InitialXContentX.ToResultContext()
	}

	if h.InitialXContentY != nil {
		result.ResultContextY = h.InitialXContentY.ToResultContext()
	}

	return result
}

type GeneralArgs struct {
	InitialDelta           Delta            `json:"initial-delta"`
	InitialServiceAccount  *D               `json:"initial-service-account"`
	InitialServiceIndex    *types.ServiceId `json:"initial-service-index"`
	ExpectedDelta          Delta            `json:"expected-delta"`
	ExpectedServiceAccount *D               `json:"expected-service-account"`
}

type RefineArgs struct {
	InitialRefineMap          map[any]any `json:"initial-refine-map"`
	InitialImportSegment      any         `json:"initial-import-segment"`
	InitialExportSegment      any         `json:"initial-export-segment"`
	InitialExportSegmentIndex int         `json:"initial-export-segment-index"`
	ExpectedRefineMap         map[any]any `json:"expected-refine-map"`
	ExpectedExportSegment     any         `json:"expected-export-segment"`
}

type AccumulateArgs struct {
	InitialTimeslot   *uint32   `json:"initial-timeslot"`
	InitialXContentX  *XContent `json:"initial-xcontent-x"`
	InitialXContentY  *XContent `json:"initial-xcontent-y"`
	ExpectedXContentX *XContent `json:"expected-xcontent-x"`
	ExpectedXContentY *XContent `json:"expected-xcontent-y"`
}

type Delta map[string]D

func (d Delta) ToServiceAccountState() types.ServiceAccountState {
	result := make(types.ServiceAccountState)

	for serviceIdStr, serviceAccount := range d {
		serviceId, err := strconv.ParseUint(serviceIdStr, 10, 32)
		if err != nil {
			panic(err)
		}

		result[types.ServiceId(serviceId)] = serviceAccount.ToServiceAccount()
	}

	return result
}

type D struct {
	SMap     SMap      `json:"s_map"`
	LMap     LMap      `json:"l_map"`
	PMap     PMap      `json:"p_map"`
	CodeHash CodeHash  `json:"code_hash"`
	Balance  types.U64 `json:"balance"`
	G        types.Gas `json:"g"`
	M        types.Gas `json:"m"`
}

func (d D) ToServiceAccount() types.ServiceAccount {
	var result types.ServiceAccount

	result.ServiceInfo.CodeHash = d.CodeHash.ToOpaqueHash()
	result.ServiceInfo.Balance = d.Balance
	result.ServiceInfo.MinItemGas = d.G
	result.ServiceInfo.MinMemoGas = d.M
	result.StorageDict = d.SMap.ToStorageDict()
	result.PreimageLookup = d.PMap.ToPreimagesLookup()
	result.LookupDict = d.LMap.ToLookupDict()

	return result
}

type CodeHash string

func (h CodeHash) ToOpaqueHash() types.OpaqueHash {
	if len(h) == 0 {
		return types.OpaqueHash{}
	}

	hash, match := strings.CutPrefix(string(h), "0x")
	if !(match && len(hash) == 64) {
		panic(fmt.Errorf("invalid code hash: %v", h))
	}

	decoded, err := hex.DecodeString(hash)
	if err != nil {
		panic(fmt.Errorf("failed to decode code hash: %v", err))
	}

	var result types.OpaqueHash
	copy(result[:], decoded)
	return result
}

type SMap map[CodeHash][]byte

func (s SMap) ToStorageDict() types.Storage {
	result := make(types.Storage)

	for k, v := range s {
		hash := k.ToOpaqueHash()
		result[hash] = v
	}

	return result
}

type LMap map[CodeHash]L

func (l LMap) ToLookupDict() types.LookupMetaMapEntry {
	result := make(types.LookupMetaMapEntry)

	for k, v := range l {
		hash := k.ToOpaqueHash()

		key := types.LookupMetaMapkey{
			Hash:   hash,
			Length: 32,
		}

		t := make(types.TimeSlotSet, len(v.T))
		for index, e := range v.T {
			t[index] = types.TimeSlot(e)
		}

		result[key] = t
	}

	return result
}

type PMap map[CodeHash][]byte

func (p PMap) ToPreimagesLookup() types.PreimagesMapEntry {
	result := make(types.PreimagesMapEntry)

	for k, v := range p {
		hash := k.ToOpaqueHash()
		result[hash] = v
	}

	return result
}

type L struct {
	T []uint32 `json:"t"`
	L uint32   `json:"l"`
}

// result content B.6
type XContent struct {
	I types.ServiceId `json:"I"`
	S types.ServiceId `json:"S"`
	U U               `json:"U"`
	T []T             `json:"T"` // DeferredTransfer
	Y *CodeHash       `json:"Y"`
}

func (x XContent) ToResultContext() PolkaVM.ResultContext {
	var result PolkaVM.ResultContext

	result.ServiceId = x.S
	result.PartialState = x.U.ToPartialStateSet()
	result.ImportServiceId = x.I
	result.DeferredTransfers = make([]types.DeferredTransfer, len(x.T))

	for i, t := range x.T {
		result.DeferredTransfers[i] = t.ToDeferredTransfer()
	}

	if x.Y != nil {
		result.Exception = x.Y.ToOpaqueHash()
	}

	return result
}

// this should be like DeferredTransfer
type T struct {
}

func (t T) ToDeferredTransfer() types.DeferredTransfer {
	return types.DeferredTransfer{
		/*
			SenderID:   t.SenderID,
			ReceiverID: t.ReceiverID,
			Balance:    t.Balance,
			Memo:       t.Memo,
			GasLimit:   t.GasLimit,
		*/
	}
}

// 12.13 partial stateset
// look into internal/types.go
type U struct {
	D Delta            `json:"D"`
	I []any            `json:"I"` // Validator
	Q types.AuthQueues `json:"Q"`
	X types.Privileges `json:"X"`
}

func (u U) ToPartialStateSet() types.PartialStateSet {
	return types.PartialStateSet{
		ServiceAccounts: u.D.ToServiceAccountState(),
		//		ValidatorKeys:   u.I,
		Authorizers: u.Q,
		Privileges:  u.X,
	}
}
