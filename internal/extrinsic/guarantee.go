package extrinsic

import (
	"errors"
	"strconv"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	jamtests "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
)

func Guarantee() error {
	// for test
	s := store.GetInstance()
	extrinsic := s.GetProcessingBlockPointer().GetGuaranteesExtrinsic()

	// GP 0.6.6 Eqs
	guarantees := NewGuaranteeController()
	guarantees.Guarantees = extrinsic

	// 11.23
	err := guarantees.Validate()
	if err != nil {
		err = transform(err)
		return err
	}
	// 11.24-11.25
	err = guarantees.Sort()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.26
	err = guarantees.ValidateSignatures()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.29-11.30
	err = guarantees.ValidateWorkReports()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.32
	err = guarantees.CardinalityCheck()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.33-11.35
	err = guarantees.ValidateContexts()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.36-11.38
	err = guarantees.ValidateWorkPackageHashes()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.39
	err = guarantees.CheckExtrinsicOrRecentHistory()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.40-11.41
	err = guarantees.CheckSegmentRootLookup()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.42
	err = guarantees.CheckWorkResult()
	if err != nil {
		err = transform(err)
		return err
	}

	// 11.43
	guarantees.TransitionWorkReport()

	return nil
}

func transform(outputError error) error {
	// if error code is defined in jamtests error map, transform the outputError to errorCode
	if reportsErrCode, errCodeExists := jamtests.ReportsErrorMap[outputError.Error()]; errCodeExists {
		return errors.New(strconv.Itoa(int(reportsErrCode)))
	}

	return outputError
}
