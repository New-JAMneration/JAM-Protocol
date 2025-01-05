package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// kappa'
type PosteriorCurrentValidators struct {
	mu         sync.RWMutex
	Validators types.ValidatorsData
}

func NewPosteriorValidators() *PosteriorCurrentValidators {
	return &PosteriorCurrentValidators{
		Validators: types.ValidatorsData{},
	}
}

func (pv *PosteriorCurrentValidators) AddValidator(validator types.Validator) error {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if len(pv.Validators) == types.ValidatorsCount {
		// show validates count in error message printf
		errMessage := fmt.Sprintf("The number of validators is %d", types.ValidatorsCount)
		return errors.New(errMessage)
	}

	pv.Validators = append(pv.Validators, validator)

	return nil
}

func (pv *PosteriorCurrentValidators) GetValidators() types.ValidatorsData {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	return pv.Validators
}

func (pv *PosteriorCurrentValidators) GetValidatorByIndex(index types.ValidatorIndex) types.Validator {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	return pv.Validators[index]
}
