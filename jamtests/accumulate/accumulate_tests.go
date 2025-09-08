package jamtests

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/statistics"
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

type StorageMapEntry struct {
	Key   types.ByteSequence `json:"key"`
	Value types.ByteSequence `json:"value"`
}

type PreimagesMapEntry struct {
	Hash types.OpaqueHash   `json:"hash"`
	Blob types.ByteSequence `json:"blob"`
}

type Account struct {
	Service   types.ServiceInfo   `json:"service"`
	Storage   []StorageMapEntry   `json:"storage"`
	Preimages []PreimagesMapEntry `json:"preimages"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Data Account         `json:"data"`
}

type AccumulateState struct {
	Slot        types.TimeSlot           `json:"slot"`
	Entropy     types.Entropy            `json:"entropy"`
	ReadyQueue  types.ReadyQueue         `json:"ready_queue"`
	Accumulated types.AccumulatedQueue   `json:"accumulated"`
	Privileges  types.Privileges         `json:"privileges"`
	Statistics  types.ServicesStatistics `json:"statistics"`
	Accounts    []AccountsMapEntry       `json:"accounts"`
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
		Slot        types.TimeSlot           `json:"slot"`
		Entropy     types.Entropy            `json:"entropy"`
		ReadyQueue  types.ReadyQueue         `json:"ready_queue"`
		Accumulated types.AccumulatedQueue   `json:"accumulated"`
		Privileges  types.Privileges         `json:"privileges"`
		Statistics  types.ServicesStatistics `json:"statistics"`
		Accounts    []AccountsMapEntry       `json:"accounts"`
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

	a.Statistics = temp.Statistics

	a.Accounts = temp.Accounts

	return nil
}

// Unmarshal json StorageMapEntry
func (s *StorageMapEntry) UnmarshalJSON(data []byte) error {
	cLog(Cyan, "Unmarshalling StorageMapEntry")

	var temp struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	keyBytes, err := hex.DecodeString(temp.Key[2:])
	if err != nil {
		return err
	}

	s.Key = types.ByteSequence(keyBytes)

	valueBytes, err := hex.DecodeString(temp.Value[2:])
	if err != nil {
		return err
	}

	s.Value = types.ByteSequence(valueBytes)

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
		Storage   []StorageMapEntry   `json:"storage"`
		Preimages []PreimagesMapEntry `json:"preimages"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Service = temp.Service

	if len(temp.Storage) != 0 {
		a.Storage = temp.Storage
	} else {
		a.Storage = nil
	}

	if len(temp.Preimages) != 0 {
		a.Preimages = temp.Preimages
	} else {
		a.Preimages = nil
	}

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

// StorageMapEntry
func (s *StorageMapEntry) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding StorageMapEntry")
	var err error

	if err = s.Key.Decode(d); err != nil {
		return err
	}

	if err = s.Value.Decode(d); err != nil {
		return err
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

	// Storage
	storageLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if storageLength == 0 {
		a.Storage = nil
	} else {
		a.Storage = make([]StorageMapEntry, storageLength)
		for i := uint64(0); i < storageLength; i++ {
			if err = a.Storage[i].Decode(d); err != nil {
				return err
			}
		}
	}

	preimageLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if preimageLength == 0 {
		a.Preimages = nil
	} else {
		a.Preimages = make([]PreimagesMapEntry, preimageLength)
		for i := uint64(0); i < preimageLength; i++ {
			if err = a.Preimages[i].Decode(d); err != nil {
				return err
			}
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

	if err = a.Statistics.Decode(d); err != nil {
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

// StorageMapEntry
func (s *StorageMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding StorageMapEntry")

	var err error

	if err = s.Key.Encode(e); err != nil {
		return err
	}

	if err = s.Value.Encode(e); err != nil {
		return err
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

	// Storage
	if err = e.EncodeLength(uint64(len(a.Storage))); err != nil {
		return err
	}

	for _, storage := range a.Storage {
		if err = storage.Encode(e); err != nil {
			return err
		}
	}

	// Preimages
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

	if err = a.Statistics.Encode(e); err != nil {
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
func ParseAccountToServiceAccountState(input []AccountsMapEntry) (output types.ServiceAccountState) {
	output = make(types.ServiceAccountState)
	for _, delta := range input {
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

		// Fill Storage
		for _, storageEntry := range delta.Data.Storage {
			serviceAccount.StorageDict[string(storageEntry.Key)] = storageEntry.Value
		}

		// Store ServiceAccount into inputDelta
		serviceId := types.ServiceId(delta.Id)
		output[serviceId] = serviceAccount
	}
	return output
}

// TODO: Implement Dump method
func (a *AccumulateTestCase) Dump() error {
	s := store.GetInstance()

	// Set time slot
	s.GetPriorStates().SetTau(a.PreState.Slot)

	// Add block with header slot
	block := types.Block{
		Header: types.Header{
			Slot: a.Input.Slot,
		},
	}
	s.AddBlock(block)
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
	inputDelta := ParseAccountToServiceAccountState(a.PreState.Accounts)
	s.GetPriorStates().SetDelta(inputDelta)

	sort.Slice(a.Input.Reports, func(i, j int) bool {
		return a.Input.Reports[i].CoreIndex < a.Input.Reports[j].CoreIndex
	})
	s.GetIntermediateStates().SetAvailableWorkReports(a.Input.Reports)
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

	// Validate time slot
	if s.GetPosteriorStates().GetTau() != a.PostState.Slot {
		log.Printf(Red+"Time slot does not match expected: %v, but got %v"+Reset, a.PostState.Slot, s.GetPosteriorStates().GetTau())
		return fmt.Errorf("time slot does not match expected: %v, but got %v", a.PostState.Slot, s.GetPosteriorStates().GetTau())
	}

	// Validate entropy
	if !reflect.DeepEqual(s.GetPosteriorStates().GetEta(), types.EntropyBuffer{a.PostState.Entropy}) {
		log.Printf(Red+"Entropy does not match expected: %v, but got %v"+Reset, a.PostState.Entropy, s.GetPosteriorStates().GetEta())
		return fmt.Errorf("entropy does not match expected: %v, but got %v", a.PostState.Entropy, s.GetPosteriorStates().GetEta())
	}

	// Validate ready queue reports (passed expect nil and [])
	ourTheta := s.GetPosteriorStates().GetTheta()
	if !reflect.DeepEqual(ourTheta, a.PostState.ReadyQueue) {
		// log.Printf("len of queue reports expected: %d, got: %d", len(a.PostState.ReadyQueue), len(s.GetPosteriorStates().GetTheta()))
		for i := range ourTheta {
			if a.PostState.ReadyQueue[i] == nil {
				a.PostState.ReadyQueue[i] = []types.ReadyRecord{}
			}
			if ourTheta[i] == nil {
				ourTheta[i] = []types.ReadyRecord{}
			}
			diff := cmp.Diff(ourTheta[i], a.PostState.ReadyQueue[i])
			if len(diff) != 0 {
				log.Printf(Red+"Ready queue reports do not match expected:\n%v,\nbut got %v\nDiff:\n%v"+Reset, a.PostState.ReadyQueue[i], ourTheta[i], diff)
				return fmt.Errorf("theta[%d] diff:\n%v", i, diff)
			}
		}
	}

	// Validate accumulated reports (passed by implementing sort)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetXi(), a.PostState.Accumulated) {
		diff := cmp.Diff(s.GetPosteriorStates().GetXi(), a.PostState.Accumulated)
		log.Printf(Red+"Accumulated reports do not match expected:\n%v,\nbut got %v\nDiff:\n%v"+Reset, a.PostState.Accumulated, s.GetPosteriorStates().GetXi(), diff)
		return fmt.Errorf("accumulated reports do not match expected:\n%v,but got \n%v\nDiff:\n%v", a.PostState.Accumulated, s.GetPosteriorStates().GetXi(), diff)
	}

	// Validate privileges (passed)
	if !reflect.DeepEqual(s.GetPosteriorStates().GetChi(), a.PostState.Privileges) {
		log.Printf(Red+"Privileges do not match expected:\n%v,\nbut got %v"+Reset, a.PostState.Privileges, s.GetPosteriorStates().GetChi())
		return fmt.Errorf("privileges do not match expected:\n%v,\nbut got %v", a.PostState.Privileges, s.GetPosteriorStates().GetChi())
	}

	// FIXME: Before validating statistics, we need to execute the update_preimage and update statistics functions

	// Validate Statistics (types.Statistics.Services, PI_S)
	// Calculate the actual statistics
	// INFO: This step will be executed in the UpdateStatistics function, but we can do it here for validation
	serviceIds := make([]types.ServiceId, len(a.PostState.Accounts))
	for i, account := range a.PostState.Accounts {
		serviceIds[i] = account.Id
	}

	ourStatisticsServices := s.GetPosteriorStates().GetServicesStatistics()
	accumulationStatisitcs := s.GetIntermediateStates().GetAccumulationStatistics()
	for _, serviceId := range serviceIds {
		accumulateCount, accumulateGasUsed := statistics.CalculateAccumulationStatistics(serviceId, accumulationStatisitcs)

		// Update the statistics for the service
		thisServiceActivityRecord, ok := ourStatisticsServices[serviceId]
		if ok {
			thisServiceActivityRecord.AccumulateCount = accumulateCount
			thisServiceActivityRecord.AccumulateGasUsed = accumulateGasUsed
			ourStatisticsServices[serviceId] = thisServiceActivityRecord
		} else {
			newServiceActivityRecord := types.ServiceActivityRecord{
				AccumulateCount:   accumulateCount,
				AccumulateGasUsed: accumulateGasUsed,
			}
			ourStatisticsServices[serviceId] = newServiceActivityRecord
		}
	}

	// Update the statistics in the PosteriorStates
	// s.GetPosteriorStates().SetServicesStatistics(ourStatisticsServices)

	// Validate statistics
	if !reflect.DeepEqual(s.GetPosteriorStates().GetServicesStatistics(), a.PostState.Statistics) {
		log.Printf(Red+"Statistics do not match expected:\n%v,\nbut got %v"+Reset, a.PostState.Statistics, s.GetPosteriorStates().GetServicesStatistics())
		diff := cmp.Diff(s.GetPosteriorStates().GetServicesStatistics(), a.PostState.Statistics)
		log.Printf("Diff:\n%v", diff)
		return fmt.Errorf("statistics do not match expected:\n%v,\nbut got %v", a.PostState.Statistics, s.GetPosteriorStates().GetPi())
	}

	// Validate Accounts (AccountsMapEntry)
	// INFO:
	// The type of state.Delta is ServiceAccountState
	// The type of a.PostState.Accounts is []AccountsMapEntry

	// Validate Delta
	// FIXME: Review after PVM stable
	expectedDelta := ParseAccountToServiceAccountState(a.PostState.Accounts)
	actualDelta := s.GetIntermediateStates().GetDeltaDoubleDagger()

	for key, expectedAcc := range expectedDelta {
		actualAcc, ok := actualDelta[key]
		if !ok {
			return fmt.Errorf("serviceId %v missing in actualDelta", key)
		}

		// ServiceInfo
		if !reflect.DeepEqual(expectedAcc.ServiceInfo, actualAcc.ServiceInfo) {
			return fmt.Errorf("mismatch in ServiceInfo for serviceId %v:\n expected=%+v\n actual=%+v",
				key, expectedAcc.ServiceInfo, actualAcc.ServiceInfo)
		}

		// PreimageLookup
		for h, expectedBlob := range expectedAcc.PreimageLookup {
			actualBlob, ok := actualAcc.PreimageLookup[h]
			if !ok {
				return fmt.Errorf("serviceId %v missing Preimage hash %x in actualDelta", key, h)
			}
			if !bytes.Equal(expectedBlob, actualBlob) {
				return fmt.Errorf("mismatch for serviceId %v, Preimage hash %x:\n expected=%x\n actual=%x",
					key, h, expectedBlob, actualBlob)
			}
		}
		for h := range actualAcc.PreimageLookup {
			if _, ok := expectedAcc.PreimageLookup[h]; !ok {
				return fmt.Errorf("serviceId %v has extra Preimage hash %x in actualDelta", key, h)
			}
		}

		// StorageDict
		for storageKey, expectedValue := range expectedAcc.StorageDict {
			actualValue, ok := actualAcc.StorageDict[storageKey]
			if !ok {
				return fmt.Errorf("serviceId %v missing Storage key %q in actualDelta", key, storageKey)
			}
			if !bytes.Equal(expectedValue, actualValue) {
				return fmt.Errorf("mismatch for serviceId %v, Storage key %q:\n expected=%x\n actual=%x",
					key, storageKey, expectedValue, actualValue)
			}
		}
		for storageKey := range actualAcc.StorageDict {
			if _, ok := expectedAcc.StorageDict[storageKey]; !ok {
				return fmt.Errorf("serviceId %v has extra Storage key %q in actualDelta", key, storageKey)
			}
		}
	}
	return nil
}
