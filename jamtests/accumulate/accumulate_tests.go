package jamtests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/google/go-cmp/cmp"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

type PreimagesMapEntry struct {
	Hash types.OpaqueHash   `json:"hash"`
	Blob types.ByteSequence `json:"blob"`
}

type Account struct {
	Service   types.ServiceInfo   `json:"service"`
	Preimages []PreimagesMapEntry `json:"preimages"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Data Account         `json:"data"`
}

type AccumulateState struct {
	Slot        types.TimeSlot         `json:"slot"`
	Entropy     types.Entropy          `json:"entropy"`
	ReadyQueue  types.ReadyQueue       `json:"ready_queue"`
	Accumulated types.AccumulatedQueue `json:"accumulated"`
	Privileges  types.Privileges       `json:"privileges"`
	Accounts    []AccountsMapEntry     `json:"accounts"`
}

type AccumulateInput struct {
	Slot    types.TimeSlot     `json:"slot"`
	Reports []types.WorkReport `json:"reports"`
}

type AccumulateOutput struct {
	Ok  *types.AccumulateRoot `json:"ok,omitempty"`
	Err *AccumulatedErrorCode `json:"err,omitempty"` // err NULL
}

type AccumulateTestCase struct {
	Input     AccumulateInput  `json:"input"`
	PreState  AccumulateState  `json:"pre_state"`
	Output    AccumulateOutput `json:"output"`
	PostState AccumulateState  `json:"post_state"`
}

type AccumulatedErrorCode types.ErrorCode

func (a *AccumulatedErrorCode) Error() string {
	if a == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *a)
}

// AccumulateInput
func (a *AccumulateInput) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AccumulateInput")

	var temp struct {
		Slot    types.TimeSlot     `json:"slot"`
		Reports []types.WorkReport `json:"reports"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Slot = temp.Slot

	if len(temp.Reports) != 0 {
		a.Reports = temp.Reports
	}

	return nil
}

// AccumulateState
func (a *AccumulateState) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AccumulateState")

	var temp struct {
		Slot        types.TimeSlot         `json:"slot"`
		Entropy     types.Entropy          `json:"entropy"`
		ReadyQueue  types.ReadyQueue       `json:"ready_queue"`
		Accumulated types.AccumulatedQueue `json:"accumulated"`
		Privileges  types.Privileges       `json:"privileges"`
		Accounts    []AccountsMapEntry     `json:"accounts"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Slot = temp.Slot
	a.Entropy = temp.Entropy

	if len(temp.ReadyQueue) != 0 {
		a.ReadyQueue = temp.ReadyQueue
	}

	if len(temp.Accumulated) != 0 {
		a.Accumulated = temp.Accumulated
	}

	a.Privileges = temp.Privileges
	a.Accounts = temp.Accounts

	return nil
}

func (p *PreimagesMapEntry) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash string `json:"hash,omitempty"`
		Blob string `json:"blob,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}

	p.Hash = types.OpaqueHash(hashBytes)

	blobBytes, err := hex.DecodeString(temp.Blob[2:])
	if err != nil {
		return err
	}

	p.Blob = types.ByteSequence(blobBytes)

	return nil
}

// Unmarshal json Account
func (a *Account) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling Account")

	var temp struct {
		Service   types.ServiceInfo   `json:"service"`
		Preimages []PreimagesMapEntry `json:"preimages"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Service = temp.Service
	a.Preimages = temp.Preimages

	return nil
}

// Unmarshal json AccountsMapEntry
func (a *AccountsMapEntry) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling AccountsMapEntry")

	var temp struct {
		Id   types.ServiceId `json:"id"`
		Data Account         `json:"data"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Id = temp.Id
	a.Data = temp.Data

	return nil
}

// AccumulateInput
func (a *AccumulateInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccumulateInput")

	var err error

	if err = a.Slot.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	a.Reports = make([]types.WorkReport, length)
	for i := uint64(0); i < length; i++ {
		if err = a.Reports[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// AccumulateOutput
func (a *AccumulateOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccumulateOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	if err != nil {
		return err
	}

	isOk := okOrErr == 0
	if isOk {
		cLog(Cyan, "AccumulateOutput is ok")

		if a.Ok == nil {
			a.Ok = &types.AccumulateRoot{}
		}

		if err = a.Ok.Decode(d); err != nil {
			return err
		}

		return nil
	} else {
		cLog(Cyan, "AccumulateOutput is err")
		cLog(Yellow, "AccumulateOutput.Err is nil")

		// AccumulateOutput.Err is NULL
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Decode(d *types.Decoder) error {
	var err error

	if err = p.Hash.Decode(d); err != nil {
		return err
	}

	if err = p.Blob.Decode(d); err != nil {
		return err
	}

	return nil
}

// Account
func (a *Account) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding Account")
	var err error

	if err = a.Service.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	a.Preimages = make([]PreimagesMapEntry, length)
	for i := uint64(0); i < length; i++ {
		if err = a.Preimages[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccountsMapEntry")
	var err error

	if err = a.Id.Decode(d); err != nil {
		return err
	}

	if err = a.Data.Decode(d); err != nil {
		return err
	}

	return nil
}

// AccumulateState
func (a *AccumulateState) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccumulateState")
	var err error

	if err = a.Slot.Decode(d); err != nil {
		return err
	}

	if err = a.Entropy.Decode(d); err != nil {
		return err
	}

	if err = a.ReadyQueue.Decode(d); err != nil {
		return err
	}

	if err = a.Accumulated.Decode(d); err != nil {
		return err
	}

	if err = a.Privileges.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	a.Accounts = make([]AccountsMapEntry, length)
	for i := uint64(0); i < length; i++ {
		if err = a.Accounts[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// AccumulateTestCase
func (t *AccumulateTestCase) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AccumulateTestCase")
	var err error

	if err = t.Input.Decode(d); err != nil {
		return err
	}

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Output.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// AccumulateInput
func (a *AccumulateInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccumulateInput")
	var err error

	if err = a.Slot.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(a.Reports))); err != nil {
		return err
	}

	for _, report := range a.Reports {
		if err = report.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding PreimagesMapEntry")
	var err error

	if err = p.Hash.Encode(e); err != nil {
		return err
	}

	if err = p.Blob.Encode(e); err != nil {
		return err
	}

	return nil
}

// Account
func (a *Account) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding Account")
	var err error

	if err = a.Service.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(a.Preimages))); err != nil {
		return err
	}

	for _, preimage := range a.Preimages {
		if err = preimage.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccountsMapEntry")
	var err error

	if err = a.Id.Encode(e); err != nil {
		return err
	}

	if err = a.Data.Encode(e); err != nil {
		return err
	}

	return nil
}

// AccumulateState
func (a *AccumulateState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccumulateState")
	var err error

	if err = a.Slot.Encode(e); err != nil {
		return err
	}

	if err = a.Entropy.Encode(e); err != nil {
		return err
	}

	if err = a.ReadyQueue.Encode(e); err != nil {
		return err
	}

	if err = a.Accumulated.Encode(e); err != nil {
		return err
	}

	if err = a.Privileges.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(a.Accounts))); err != nil {
		return err
	}

	for _, account := range a.Accounts {
		if err = account.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AccumulateOutput
func (a *AccumulateOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccumulateOutput")
	var err error

	if a.Ok != nil {
		cLog(Yellow, "AccumulateOutput is ok")
		if err := e.WriteByte(0); err != nil {
			return err
		}

		// Encode AccumulateOutput
		if err = a.Ok.Encode(e); err != nil {
			return err
		}

		return nil
	}

	// AccumulateOutput.Err is NULL

	return nil
}

// AccumulateTestCase
func (t *AccumulateTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccumulateTestCase")
	var err error

	if err = t.Input.Encode(e); err != nil {
		return err
	}

	if err = t.PreState.Encode(e); err != nil {
		return err
	}

	if err = t.Output.Encode(e); err != nil {
		return err
	}

	if err = t.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

// TODO: Implement Dump method
func (a *AccumulateTestCase) Dump() error {
	s := store.GetInstance()

	// Set time slot
	s.GetPriorStates().SetTau(a.PreState.Slot)
	s.GetIntermediateHeaderPointer().SetSlot(a.Input.Slot)
	s.GetPosteriorStates().SetTau(a.Input.Slot)

	// Set entropy
	s.GetPriorStates().SetEta(types.EntropyBuffer{a.PreState.Entropy})
	s.GetPosteriorStates().SetEta0(a.PreState.Entropy)

	// Set ready queue
	s.GetPriorStates().SetTheta(a.PreState.ReadyQueue)

	// Set accumulated reports
	s.GetPriorStates().SetXi(a.PreState.Accumulated)

	// Set privileges
	s.GetPriorStates().SetChi(a.PreState.Privileges)

	// Set accounts
	inputDelta := make(types.ServiceAccountState)
	for serviceId, delta := range a.PreState.Accounts {
		// Create or get ServiceAccount, ensure its internal maps are initialized
		serviceAccount := types.ServiceAccount{
			ServiceInfo:    delta.Data.Service,
			PreimageLookup: make(types.PreimagesMapEntry),
			LookupDict:     make(types.LookupMetaMapEntry),
			StorageDict:    make(types.Storage),
		}

		// Fill PreimageLookup
		for _, preimage := range delta.Data.Preimages {
			serviceAccount.PreimageLookup[preimage.Hash] = preimage.Blob
		}

		// Store ServiceAccount into inputDelta
		inputDelta[types.ServiceId(serviceId)] = serviceAccount
	}
	s.GetPriorStates().SetDelta(inputDelta)

	sort.Slice(a.Input.Reports, func(i, j int) bool {
		return a.Input.Reports[i].CoreIndex < a.Input.Reports[j].CoreIndex
	})
	s.GetAccumulatableWorkReportsPointer().SetAccumulatableWorkReports(a.Input.Reports)
	return nil
}

func (a *AccumulateTestCase) GetPostState() interface{} {
	return a.PostState
}

func (a *AccumulateTestCase) GetOutput() interface{} {
	return a.Output
}

func (a *AccumulateTestCase) ExpectError() error {
	if a.Output.Err == nil {
		return nil
	}
	return a.Output.Err
}

func (a *AccumulateTestCase) Validate() error {
	s := store.GetInstance()

	if s.GetPosteriorStates().GetTau() != a.PostState.Slot {
		fmt.Errorf("Time slot does not match expected: %v, but got %v", a.PostState.Slot, s.GetPosteriorStates().GetTau())
	}

	if !reflect.DeepEqual(s.GetPosteriorStates().GetEta(), types.EntropyBuffer{a.PostState.Entropy}) {
		fmt.Errorf("Entropy does not match expected: %v, but got %v", a.PostState.Entropy, s.GetPosteriorStates().GetEta())
	}

	// Validate ready queue reports (passed expect nil and [])
	ourTheta := s.GetPosteriorStates().GetTheta()
	if !reflect.DeepEqual(ourTheta, a.PostState.ReadyQueue) {
		log.Printf("len of queue reports expected: %d, got: %d", len(a.PostState.ReadyQueue), len(s.GetPosteriorStates().GetTheta()))
		for i := range ourTheta {
			diff := cmp.Diff(ourTheta[i], a.PostState.ReadyQueue[i])
			fmt.Errorf("Theta[%d] Diff:\n%v", i, diff)
		}

	}

	// Validate accumulated reports (passed by implementing sort)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetXi(), a.PostState.Accumulated) {
		diff := cmp.Diff(s.GetPosteriorStates().GetXi(), a.PostState.Accumulated)
		fmt.Errorf("Accumulated reports do not match expected:\n%v,but got \n%v\nDiff:\n%v", a.PostState.Accumulated, s.GetPosteriorStates().GetXi(), diff)
	}

	// Validate privileges (passed)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetChi(), a.PostState.Privileges) {
		fmt.Errorf("Privileges do not match expected:\n%v,\nbut got %v", a.PostState.Privileges, s.GetPosteriorStates().GetChi())
	}
	return nil
}
