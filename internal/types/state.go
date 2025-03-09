package types

import "fmt"

// (4.4)
type State struct {
	Alpha  AuthPools               `json:"alpha"`
	Varphi AuthQueues              `json:"varphi"`
	Beta   BlocksHistory           `json:"beta"`
	Gamma  Gamma                   `json:"gamma"`
	Psi    DisputesRecords         `json:"psi"`
	Eta    EntropyBuffer           `json:"eta"`
	Iota   ValidatorsData          `json:"iota"`
	Kappa  ValidatorsData          `json:"kappa"`
	Lambda ValidatorsData          `json:"lambda"`
	Rho    AvailabilityAssignments `json:"rho"`
	Tau    TimeSlot                `json:"tau"`
	Chi    Privileges              `json:"chi"`
	Pi     Statistics              `json:"pi"`
	Theta  ReadyQueue              `json:"theta"`
	Xi     AccumulatedHistories    `json:"xi"`
	Delta  Accounts                `json:"accounts"`
}

type TestState struct {
	Alpha  AuthPools               `json:"alpha"`
	Varphi AuthQueues              `json:"varphi"`
	Beta   BlocksHistory           `json:"beta"`
	Gamma  Gamma                   `json:"gamma"`
	Psi    DisputesRecords         `json:"psi"`
	Eta    EntropyBuffer           `json:"eta"`
	Iota   ValidatorsData          `json:"iota"`
	Kappa  ValidatorsData          `json:"kappa"`
	Lambda ValidatorsData          `json:"lambda"`
	Rho    AvailabilityAssignments `json:"rho"`
	Tau    TimeSlot                `json:"tau"`
	Chi    Privileges              `json:"chi"`
	Pi     Statistics              `json:"pi"`
	Theta  ReadyQueue              `json:"theta"`
	Xi     AccumulatedHistories    `json:"xi"`
}

// (6.3)
type Gamma struct {
	GammaK ValidatorsData             `json:"gamma_k"`
	GammaZ BandersnatchRingCommitment `json:"gamma_z"`
	GammaS TicketsOrKeys              `json:"gamma_s"`
	GammaA TicketsAccumulator         `json:"gamma_a"`
}

// INFO: New Account Struct for jamtestnet

// from davxy asn
type PreimagesMapEntry struct {
	Hash OpaqueHash   `json:"hash"`
	Blob ByteSequence `json:"blob"`
}

// from davxy asn
type LookupMetaMapkey struct {
	Hash   OpaqueHash `json:"hash"`
	Length U32        `json:"length"`
}

// from davxy asn
type LookupMetaMapEntry struct {
	Key LookupMetaMapkey `json:"key"`
	Val []TimeSlot       `json:"value"`
}

type Storage map[OpaqueHash]ByteSequence

type AccountData struct {
	Service    ServiceInfo          `json:"service"`
	Preimages  []PreimagesMapEntry  `json:"preimages"`
	LookupMeta []LookupMetaMapEntry `json:"lookup_meta"`
	Storage    Storage              `json:"storage"`
}

type Account struct {
	Id   ServiceId   `json:"id"`
	Data AccountData `json:"data"`
}

type Accounts []Account

// (9.2) delta
type ServiceAccountState map[ServiceId]ServiceAccount

// (9.3)
type ServiceAccount struct {
	StorageDict    map[OpaqueHash]ByteSequence   // a_s
	PreimageLookup map[OpaqueHash]ByteSequence   // a_p
	LookupDict     map[DictionaryKey]TimeSlotSet // a_l
	CodeHash       OpaqueHash                    // a_c
	Balance        U64                           // a_b
	MinItemGas     Gas                           // a_g
	MinMemoGas     Gas                           // a_m
}

type DictionaryKey struct {
	Hash   OpaqueHash
	Length U32
}

type TimeSlotSet []TimeSlot

// (9.9)
type PrivilegedServices struct {
	ManagerServiceIndex     U32         `json:"chi_m"`
	AlterPhiServiceIndex    U32         `json:"chi_a"`
	AlterIotaServiceIndex   U32         `json:"chi_v"`
	AutoAccumulateGasLimits map[U32]U64 `json:"chi_g"`
}

// (12.3)
type (
	AccumulationQueue       []UnaccumulateWorkReports
	UnaccumulateWorkReports []UnaccumulateWorkReport
)

func (accumulationQueue AccumulationQueue) Validate() error {
	if len(accumulationQueue) != EpochLength {
		return fmt.Errorf("AccumulationQueue must have exactly %d items, but got %d", EpochLength, len(accumulationQueue))
	}
	return nil
}

type UnaccumulateWorkReport struct {
	WorkReport        WorkReport        `json:"report"`
	UnaccumulatedDeps []WorkPackageHash `json:"dependencies"`
}

// (12.1)
type (
	AccumulatedHistories []AccumulatedHistory
	AccumulatedHistory   []WorkPackageHash
)

func (accumulatedHistories AccumulatedHistories) Validate() error {
	if len(accumulatedHistories) != EpochLength {
		return fmt.Errorf("AccumulatedHistories must have exactly %d items, but got %d", EpochLength, len(accumulatedHistories))
	}
	return nil
}
