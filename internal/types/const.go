package types

import "time"

// changeable constants depends on chainspec

// tiny
var (
	ValidatorsCount = 6
	CoresCount      = 2
	EpochLength     = 12

	MaxTicketsPerBlock  = 3
	TicketsPerValidator = 3

	MaxBlocksHistory = 8

	AuthPoolMaxSize = 8
	AuthQueueSize   = 80

	ValidatorsSuperMajority = 5
	AvailBitfieldBytes      = 1
)

// permanent constants
var (
	AdditionalMinBalancePerItem  = 10  // B_I
	AdditionalMinBalancePerOctet = 1   // B_L
	BasicMinBalance              = 100 // B_S

	SlotPeriod   = 6
	JamCommonEra = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
)
