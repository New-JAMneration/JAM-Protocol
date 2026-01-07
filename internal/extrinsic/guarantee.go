package extrinsic

import "github.com/New-JAMneration/JAM-Protocol/internal/blockchain"

func Guarantee() error {
	// for test
	cs := blockchain.GetInstance()

	// GP 0.6.6 Eqs
	guarantees := NewGuaranteeController()
	guarantees.Guarantees = cs.GetLatestBlock().Extrinsic.Guarantees

	// 11.23
	err := guarantees.Validate()
	if err != nil {
		return err
	}
	// 11.24-11.25
	err = guarantees.Sort()
	if err != nil {
		return err
	}

	// 11.26
	err = guarantees.ValidateSignatures()
	if err != nil {
		return err
	}

	// 11.29-11.30
	err = guarantees.ValidateWorkReports()
	if err != nil {
		return err
	}

	// 11.32
	err = guarantees.CardinalityCheck()
	if err != nil {
		return err
	}

	// 11.33-11.35
	err = guarantees.ValidateContexts()
	if err != nil {
		return err
	}

	// 11.36-11.38
	err = guarantees.ValidateWorkPackageHashes()
	if err != nil {
		return err
	}

	// 11.39
	err = guarantees.CheckExtrinsicOrRecentHistory()
	if err != nil {
		return err
	}

	// 11.40-11.41
	err = guarantees.CheckSegmentRootLookup()
	if err != nil {
		return err
	}

	// 11.42
	err = guarantees.CheckWorkResult()
	if err != nil {
		return err
	}

	// 11.43
	guarantees.TransitionWorkReport()

	return nil
}
