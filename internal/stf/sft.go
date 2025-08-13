package stf

import (
	"fmt"
	"log"
)

// TODO: Implement the following functions to handle state transitions
// Each function should update the corresponding state in the data store
// The functions should validate inputs and handle errors appropriately
// Consider adding proper logging and metrics collection
func RunSTF() error {
	log.Println("Update Dispute")
	// Update Disputes
	err := UpdateDisputes()
	if err != nil {
		return fmt.Errorf("update disputes error: %v", err)
	}
	log.Println("Update Safrole")
	// Update Safrole
	err = UpdateSafrole()
	if err != nil {
		return fmt.Errorf("update safrole error: %v", err)
	}
	log.Println("Update Assurances")
	// Update Assurances
	err = UpdateAssurances()
	if err != nil {
		return fmt.Errorf("update assurances error: %v", err)
	}
	// Update Reports
	err = UpdateReports()
	if err != nil {
		return fmt.Errorf("update reports error: %v", err)
	}

	log.Println("Update Accumulate")
	// Update Accumlate
	err = UpdateAccumlate()
	if err != nil {
		return fmt.Errorf("update accumulate error: %v", err)
	}
	log.Println("Update history")
	// Update History (beta^dagger -> beta^prime)
	err = UpdateHistory()
	if err != nil {
		return fmt.Errorf("update histroy error: %v", err)
	}

	// Update Preimages
	err = UpdatePreimages()
	if err != nil {
		return fmt.Errorf("update preimages error: %v", err)
	}

	// Update Authorization
	err = UpdateAuthorizations()
	if err != nil {
		return fmt.Errorf("update authorization error: %v", err)
	}

	// Update Statistics
	log.Println("Update statistics")
	err = UpdateStatistics()
	if err != nil {
		return fmt.Errorf("update statistics error: %v", err)
	}

	return nil
}
