package types

import (
	"log"
	"os"
	"time"
)

var TEST_MODE = "tiny"

func SetTestMode() {
	TEST_MODE = os.Getenv("TEST_MODE")
	if TEST_MODE == "tiny" {
		log.Println("⚙️  Tiny mode activated")
		SetTinyMode()
		return
	}

	if TEST_MODE == "full" {
		log.Println("⚙️  Full mode activated")
		SetFullMode()
		return
	}

	log.Println("⚙️  Default(Tiny) mode activated")
	SetTinyMode()
}

func SetTinyMode() {
	log.Println("⚙️  Tiny mode activated")
	TEST_MODE = "tiny"
	ValidatorsCount = 6 // V
	CoresCount = 2 // C
	EpochLength = 12 // E 
	RotationPeriod = 4 // R
	MaxTicketsPerBlock = 3 // K
	TicketsPerValidator = 3 // N
	ValidatorsSuperMajority = 5 
	AvailBitfieldBytes = 1 
}

func SetFullMode() {
	log.Println("⚙️  Full mode activated")
	TEST_MODE = "full"
	ValidatorsCount = 1023
	CoresCount = 341
	EpochLength = 600
	SlotSubmissionEnd = 500
	RotationPeriod = 4
	MaxTicketsPerBlock = 16
	TicketsPerValidator = 2
	ValidatorsSuperMajority = 683
	AvailBitfieldBytes = 43
}

// changeable constants depends on chainspec

// tiny
var (
	// V
	ValidatorsCount = 6
	// C
	CoresCount = 2
	// N
	TicketsPerValidator = 3
	// E
	EpochLength = 12
	// Y: The number of slots into an epoch at which ticket-submission ends.
	SlotSubmissionEnd = 10
	// K
	MaxTicketsPerBlock = 3
	// R: The rotation period of validator-core assignments, in timeslots.
	RotationPeriod = 4

	MaxTicketsPerBlock  = 3
	TicketsPerValidator = 3

	ValidatorsSuperMajority = 5
	AvailBitfieldBytes      = 1
)

var JamCommonEra = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

// permanent constants
const (
	AdditionalMinBalancePerItem  = 10  // B_I
	AdditionalMinBalancePerOctet = 1   // B_L
	BasicMinBalance              = 100 // B_S
	SlotPeriod                   = 6
	SlotSubmissionEnd            = 10                  // Y = 500: The number of slots into an epoch at which ticket-submission ends.
	TranchePeriod                = 8                   // A
	JamEntropy                   = "jam_entropy"       // XE
	JamFallbackSeal              = "jam_fallback_seal" // XF
	JamTicketSeal                = "jam_ticket_seal"   // XT
	JamValid                     = "jam_valid"
	JamInvalid                   = "jam_invalid"
	JamAvailable                 = "jam_available"
	JamBeefy                     = "jam_beefy"
	JamGuarantee                 = "jam_guarantee"
	JamAnnounce                  = "jam_announce" // XI
	JamAudit                     = "jam_audit"    // XU
	AuthPoolMaxSize              = 8               // O: The maximum number of items in the authorizations pool
	AuthQueueSize                = 80              // Q: The maximum number of items in the authorizations queue
	MaxBlocksHistory             = 8               // H: The size of recent history, in blocks
)
// work item constants
const (
	MaximumWorkItems                 = 16        // I (graypaper 0.6.3)
	MaximumDependencyItems           = 8         // J
	WorkReportTimeout                = 5         // U
	WorkReportOutputBlobsMaximumSize = 48 * 1024 // W_R
	MaxLookupAge                     = 14400     // L
)

// work package constants
const (
	MaxTotalSize       = 12 * 1024 * 1024                 // W_B = 12 MB (14.6)
	MaxRefineGas       = 5_000_000_000                    // G_R v0.6.4
	MaxAccumulateGas   = 10_000_000                       // G_A v0.6.4
	IsAuthorizedGas    = 50_000_000                       // G_I v0.6.4 The gas allocated to invoke a work-package's Is-Authorized logic.
	TotalGas           = 3_500_000_000                    // G_T v0.6.4 The total gas allocated across for all Accumulation. Should be no smaller than GA ⋅ C + ∑g∈V(χg) (g).
	MaxImportCount     = 3072                             // W_M: The maximum number of import segments in a work package (14.4). graypaper v0.6.3
	MaxExportCount     = 3072                             // W_X: The maximum number of export segments in a work package (14.4). graypaper v0.6.5
	ECPiecesPerSegment = 6                                // W_P: The number of erasure-coded pieces in a segment
	ECBasicSize        = 684                              // W_E: The basic size of erasure-coded pieces in octets
	SegmentSize        = ECPiecesPerSegment * ECBasicSize // W_G = 4104: The size of a segment in octets
	MaxExtrinsics      = 128                              // T (14.4). graypaper 0.6.3
	MaxServiceCodeSize = 4_000_000                        // W_C v0.6.4
)

// erasure coding constants
// 342:1023 (Appendix H)
const (
	DataShards  = 342
	TotalShards = 1023
)

// genesis file path
const (
	GenesisBlockPath = "../../pkg/test_data/jamtestnet/chainspecs/blocks/genesis-tiny.bin"
	GenesisStatePath = "../../pkg/test_data/jamtestnet/chainspecs/state_snapshots/genesis-tiny.bin"
)

// PVM constants
const (
	UnreferencedPreimageTimeslots = LookupAnchorMaxAge + 4800 // D
	TransferMemoSize              = 128                       // W_T
	LookupAnchorMaxAge            = 14400                     // L
)

// Auditing (17.16)
const BiasFactor = 2

const SegmentErasureTTL = 28 * 24 * time.Hour // 28 days
