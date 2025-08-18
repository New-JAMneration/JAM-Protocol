package types

import "fmt"

// (4.4)
type State struct {
	Alpha  AuthPools               `json:"alpha"`
	Varphi AuthQueues              `json:"varphi"`
	Beta   RecentBlocks            `json:"beta"`
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
	// TODO: rename LastAccOut to Theta, and Theta to Vartheta
	LastAccOut LastAccOut
	Xi         AccumulatedQueue    `json:"xi"`
	Delta      ServiceAccountState `json:"accounts"`
}

// (6.3)
type Gamma struct {
	GammaK ValidatorsData             `json:"gamma_k"`
	GammaZ BandersnatchRingCommitment `json:"gamma_z"`
	GammaS TicketsOrKeys              `json:"gamma_s"`
	GammaA TicketsAccumulator         `json:"gamma_a"`
}

// We use this type to parse json file
type PreimagesMapEntryDTO struct {
	Hash OpaqueHash   `json:"hash"`
	Blob ByteSequence `json:"blob"`
}

// We use this type to parse json file
type LookupMetaMapEntryDTO struct {
	Key LookupMetaMapkey `json:"key"`
	Val []TimeSlot       `json:"value"`
}

// We use this type to parse json file
type AccountDataDTO struct {
	Service    ServiceInfo             `json:"service"`
	Preimages  []PreimagesMapEntryDTO  `json:"preimages"`
	LookupMeta []LookupMetaMapEntryDTO `json:"lookup_meta"`
	Storage    Storage                 `json:"storage"`
}

// We use this type to parse json file
type AccountDTO struct {
	Id   ServiceId      `json:"id"`
	Data AccountDataDTO `json:"data"`
}

type (
	Storage            map[string]ByteSequence
	PreimagesMapEntry  map[OpaqueHash]ByteSequence
	LookupMetaMapEntry map[LookupMetaMapkey]TimeSlotSet
)

// (9.2) delta
type ServiceAccountState map[ServiceId]ServiceAccount

// (9.3)
type ServiceAccount struct {
	ServiceInfo    ServiceInfo
	PreimageLookup PreimagesMapEntry  // a_p
	LookupDict     LookupMetaMapEntry // a_l
	StorageDict    Storage            // a_s
}

type LookupMetaMapkey struct {
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
// type (
// 	AccumulatedQueue []AccumulatedQueueItem
// 	AccumulatedQueueItem   []WorkPackageHash
// )

func (AccumulatedQueue AccumulatedQueue) Validate() error {
	if len(AccumulatedQueue) != EpochLength {
		return fmt.Errorf("AccumulatedQueue must have exactly %d items, but got %d", EpochLength, len(AccumulatedQueue))
	}
	return nil
}
