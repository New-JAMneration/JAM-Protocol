package types

import "time"

var (
	ValidatorsCount = 6
	CoresCount      = 2
	EpochLength     = 12

	// R: The rotation period of validator-core assignments, in timeslots.
	RotationPeriod = 4

	MaxTicketsPerBlock  = 3
	TicketsPerValidator = 3

	MaxBlocksHistory = 8

	AuthPoolMaxSize = 8
	AuthQueueSize   = 80

	ValidatorsSuperMajority = 5
	AvailBitfieldBytes      = 1
)

var (
	SlotPeriod        = 6
	JamCommonEra      = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	SlotSubmissionEnd = 10                  // Y = 500: The number of slots into an epoch at which ticket-submission ends.
	JamEntropy        = "jam_entropy"       // XE
	JamFallbackSeal   = "jam_fallback_seal" // XF
	JamTicketSeal     = "jam_ticket_seal"   // XT
)

var (
	MaximumWorkItems                 = 4 // I
	MaximumDependencyItems           = 8 // J
	WorkReportTimeout                = 5 // U
	WorkReportOutputBlobsMaximumSize = 48 * 1024
)
