package stf

import "log"

// TODO: Implement the following functions to handle state transitions
// Each function should update the corresponding state in the data store
// The functions should validate inputs and handle errors appropriately
// Consider adding proper logging and metrics collection
func RunSTF() error {
	log.Println("Update Dispute")
	// Update Disputes
	err := UpdateDisputes()
	if err != nil {
		return err
	}
	log.Println("Update Safrole")
	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		return err
	}
	log.Println("Update Assurances")
	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		return err
	}
	log.Println("Update Reports")
	// Update Reports
	err = UpdateReports()
	if err != nil {
		return err
	}

	log.Println("Update Accumulate")
	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return err
	}
	log.Println("Update history")
	// Update History (beta^dagger -> beta^prime)
	log.Println("Update Preimages")
	// Update Preimages
	err = UpdatePreimages()
	if err != nil {
		return err
	}

	// Update Authorization
	log.Println("Update Authorization")
	// TODO: add authorization in sft

	// Update Statistics
	log.Println("Update statistics")
	err = UpdateStatistics()
	if err != nil {
		return err
	}

	return nil
}
