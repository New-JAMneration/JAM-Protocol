package types

import "fmt"

// (4.4)
type State struct {
	Alpha  AuthPools               `json:"alpha"`
	Beta   BlocksHistory           `json:"beta"`
	Gamma  Gamma                   `json:"gamma"`
	Delta  ServiceAccountState     `json:"delta"`
	Eta    EntropyBuffer           `json:"eta"`
	Iota   ValidatorsData          `json:"iota"`
	Kappa  ValidatorsData          `json:"kappa"`
	Lambda ValidatorsData          `json:"lambda"`
	Rho    AvailabilityAssignments `json:"rho"`
	Tau    TimeSlot                `json:"tau"`
	Varphi AuthQueues              `json:"varphi"`
	Chi    PrivilegedServices      `json:"chi"`
	Psi    DisputesRecords         `json:"psi"`
	Pi     Statistics              `json:"pi"`
	Theta  UnaccumulateWorkReports `json:"theta"`
	Xi     AccumulatedHistories    `json:"xi"`
}

// (6.3)
type Gamma struct {
	GammaK ValidatorsData             `json:"gamma_k"`
	GammaZ BandersnatchRingCommitment `json:"gamma_z"`
	GammaS TicketsOrKeys              `json:"gamma_s"`
	GammaA TicketsAccumulator         `json:"gamma_a"`
}

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

type ServiceAccountStateDerivatives map[ServiceId]ServiceAccountDerivatives
type ServiceAccountDerivatives struct {
	Items      U32 `json:"items,omitempty"` // a_i
	Bytes      U64 `json:"bytes,omitempty"` // a_o
	Minbalance U64 // a_t
}

// (9.9)
type PrivilegedServices struct {
	ManagerServiceIndex     U32         `json:"chi_m"`
	AlterPhiServiceIndex    U32         `json:"chi_a"`
	AlterIotaServiceIndex   U32         `json:"chi_v"`
	AutoAccumulateGasLimits map[U32]U64 `json:"chi_g"`
}

// (12.3)
type AccumulationQueue []UnaccumulateWorkReports
type UnaccumulateWorkReports []UnaccumulateWorkReport

func (accumulationQueue AccumulationQueue) Validate() error {
	if len(accumulationQueue) != EpochLength {
		return fmt.Errorf("AccumulationQueue must have exactly %d items, but got %d", EpochLength, len(accumulationQueue))
	}
	return nil
}

type UnaccumulateWorkReport struct {
	WorkReport        WorkReport
	UnaccumulatedDeps []WorkPackageHash
}

// (12.1)
type AccumulatedHistories []AccumulatedHistory
type AccumulatedHistory []WorkPackageHash

func (accumulatedHistories AccumulatedHistories) Validate() error {
	if len(accumulatedHistories) != EpochLength {
		return fmt.Errorf("AccumulatedHistories must have exactly %d items, but got %d", EpochLength, len(accumulatedHistories))
	}
	return nil
}
