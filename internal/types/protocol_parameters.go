package types

import "fmt"

const maxInt = int(^uint(0) >> 1)

func assignIntFromU64(name string, v uint64) (int, error) {
	if v > uint64(maxInt) {
		return 0, fmt.Errorf("%s overflow: %d > MaxInt(%d)", name, v, maxInt)
	}
	return int(v), nil
}

func assignIntFromU32(name string, v uint32) (int, error) {
	if uint64(v) > uint64(maxInt) {
		return 0, fmt.Errorf("%s overflow: %d > MaxInt(%d)", name, v, maxInt)
	}
	return int(v), nil
}

// TODO: overwrite all value by chainspec ProtocolParameters
func ApplyProtocolParameters(pp ProtocolParameters) error {
	// all const must match chainspec ProtocolParameters
	if uint64(pp.BI) != uint64(AdditionalMinBalancePerItem) {
		return fmt.Errorf("pp.B_I mismatch: got %d want %d", uint64(pp.BI), AdditionalMinBalancePerItem)
	}
	if uint64(pp.BL) != uint64(AdditionalMinBalancePerOctet) {
		return fmt.Errorf("pp.B_L mismatch: got %d want %d", uint64(pp.BL), AdditionalMinBalancePerOctet)
	}
	if uint64(pp.BS) != uint64(BasicMinBalance) {
		return fmt.Errorf("pp.B_S mismatch: got %d want %d", uint64(pp.BS), BasicMinBalance)
	}

	if uint16(pp.P) != uint16(SlotPeriod) {
		return fmt.Errorf("pp.P (SlotPeriod) mismatch: got %d want %d", uint16(pp.P), SlotPeriod)
	}
	if uint16(pp.H) != uint16(MaxBlocksHistory) {
		return fmt.Errorf("pp.H (MaxBlocksHistory) mismatch: got %d want %d", uint16(pp.H), MaxBlocksHistory)
	}

	if uint16(pp.O) != uint16(AuthPoolMaxSize) {
		return fmt.Errorf("pp.O (AuthPoolMaxSize) mismatch: got %d want %d", uint16(pp.O), AuthPoolMaxSize)
	}
	if uint16(pp.Q) != uint16(AuthQueueSize) {
		return fmt.Errorf("pp.Q (AuthQueueSize) mismatch: got %d want %d", uint16(pp.Q), AuthQueueSize)
	}
	if uint16(pp.I) != uint16(MaximumWorkItems) {
		return fmt.Errorf("pp.I (MaximumWorkItems) mismatch: got %d want %d", uint16(pp.I), MaximumWorkItems)
	}
	if uint16(pp.J) != uint16(MaximumDependencyItems) {
		return fmt.Errorf("pp.J (MaximumDependencyItems) mismatch: got %d want %d", uint16(pp.J), MaximumDependencyItems)
	}
	if uint16(pp.U) != uint16(WorkReportTimeout) {
		return fmt.Errorf("pp.U (WorkReportTimeout) mismatch: got %d want %d", uint16(pp.U), WorkReportTimeout)
	}

	if uint64(pp.GA) != uint64(MaxAccumulateGas) {
		return fmt.Errorf("pp.G_A (MaxAccumulateGas) mismatch: got %d want %d", uint64(pp.GA), MaxAccumulateGas)
	}
	if uint64(pp.GI) != uint64(IsAuthorizedGas) {
		return fmt.Errorf("pp.G_I (IsAuthorizedGas) mismatch: got %d want %d", uint64(pp.GI), IsAuthorizedGas)
	}

	if uint64(pp.WA) != uint64(MaxIsAuthorizedCodeSize) {
		return fmt.Errorf("pp.W_A (MaxIsAuthorizedCodeSize) mismatch: got %d want %d", uint64(pp.WA), MaxIsAuthorizedCodeSize)
	}
	if uint64(pp.WB) != uint64(MaxTotalSize) {
		return fmt.Errorf("pp.W_B (MaxTotalSize) mismatch: got %d want %d", uint64(pp.WB), MaxTotalSize)
	}
	if uint64(pp.WC) != uint64(MaxServiceCodeSize) {
		return fmt.Errorf("pp.W_C (MaxServiceCodeSize) mismatch: got %d want %d", uint64(pp.WC), MaxServiceCodeSize)
	}
	if uint64(pp.WR) != uint64(WorkReportOutputBlobsMaximumSize) {
		return fmt.Errorf("pp.W_R (WorkReportOutputBlobsMaximumSize) mismatch: got %d want %d", uint64(pp.WR), WorkReportOutputBlobsMaximumSize)
	}
	if uint32(pp.WT) != uint32(TransferMemoSize) {
		return fmt.Errorf("pp.W_T (TransferMemoSize) mismatch: got %d want %d", uint32(pp.WT), TransferMemoSize)
	}

	if uint32(pp.WM) != uint32(MaxImportCount) {
		return fmt.Errorf("pp.W_M (MaxImportCount) mismatch: got %d want %d", uint32(pp.WM), MaxImportCount)
	}
	if uint32(pp.WX) != uint32(MaxExportCount) {
		return fmt.Errorf("pp.W_X (MaxExportCount) mismatch: got %d want %d", uint32(pp.WX), MaxExportCount)
	}

	if uint16(pp.T) != uint16(MaxExtrinsics) {
		return fmt.Errorf("pp.T (MaxExtrinsics) mismatch: got %d want %d", uint16(pp.T), MaxExtrinsics)
	}

	// overwrite var by chainspec ProtocolParameters
	var err error

	CoresCount = int(uint16(pp.C))

	if UnreferencedPreimageTimeslots, err = assignIntFromU32("pp.D", uint32(pp.D)); err != nil {
		return err
	}
	if EpochLength, err = assignIntFromU32("pp.E", uint32(pp.E)); err != nil {
		return err
	}

	if MaxRefineGas, err = assignIntFromU64("pp.G_R", uint64(pp.GR)); err != nil {
		return err
	}
	if TotalGas, err = assignIntFromU64("pp.G_T", uint64(pp.GT)); err != nil {
		return err
	}

	MaxTicketsPerBlock = int(uint16(pp.K))

	if MaxLookupAge, err = assignIntFromU32("pp.L", uint32(pp.L)); err != nil {
		return err
	}

	TicketsPerValidator = int(uint16(pp.N))
	RotationPeriod = int(uint16(pp.R))
	ValidatorsCount = int(uint16(pp.V))

	if ECBasicSize, err = assignIntFromU32("pp.W_E", uint32(pp.WE)); err != nil {
		return err
	}
	if ECPiecesPerSegment, err = assignIntFromU32("pp.W_P", uint32(pp.WP)); err != nil {
		return err
	}
	if SlotSubmissionEnd, err = assignIntFromU32("pp.Y", uint32(pp.Y)); err != nil {
		return err
	}

	return nil
}
