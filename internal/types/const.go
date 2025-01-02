package types

import "time"

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

var (
	SlotPeriod   = 6
	JamCommonEra = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
)
