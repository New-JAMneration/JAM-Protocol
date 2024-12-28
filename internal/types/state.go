package types

// (4.4)
type State struct {
	Alpha  AuthPools               `json:"alpha"`
	Beta   BlockInfo               `json:"beta"`
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

// (9.2)
type ServiceAccountState map[ServiceId]ServiceAccount

// (9.3)
type ServiceAccount struct {
	StorageDict    map[OpaqueHash]ByteSequence
	PreimageLookup map[OpaqueHash]ByteSequence
	LookupDict     map[OpaqueHash]ByteSequence
	CodeHash       OpaqueHash
	Balance        U64
	MinItemGas     Gas
	MinMemoGas     Gas
}

// (9.9)
type PrivilegedServices struct {
	ManagerServiceIndex     U32         `json:"chi_m"`
	AlterPhiServiceIndex    U32         `json:"chi_a"`
	AlterIotaServiceIndex   U32         `json:"chi_v"`
	AutoAccumulateGasLimits map[U32]U64 `json:"chi_g"`
}

// (12.3)
type UnaccumulateWorkReports []UnaccumulateWorkReport
type UnaccumulateWorkReport struct {
	WorkReport        WorkReport
	UnaccumulatedDeps []WorkPackageHash
}

// (12.1)
type AccumulatedHistories []AccumulatedHistory
type AccumulatedHistory []WorkPackageHash
