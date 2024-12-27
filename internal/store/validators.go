package store

import (
	"sync"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// kappa'
type PosteriorValidators struct {
	mu         sync.RWMutex
	Validators jamTypes.ValidatorsData
}

func NewPosteriorValidators() *PosteriorValidators {
	return &PosteriorValidators{
		Validators: jamTypes.ValidatorsData{},
	}
}

func (pv *PosteriorValidators) AddValidator(validator jamTypes.Validator) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	pv.Validators = append(pv.Validators, validator)
}

func (pv *PosteriorValidators) GetValidators() jamTypes.ValidatorsData {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	return pv.Validators
}

func (pv *PosteriorValidators) GetValidatorByIndex(index jamTypes.ValidatorIndex) jamTypes.Validator {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	return pv.Validators[index]
}
