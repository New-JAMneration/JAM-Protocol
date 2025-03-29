package accumulation

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func updatePartialStateSetToPosteriorState(store *store.Store, o types.PartialStateSet) {
	// (12.22)
	postChi := o.Privileges
	deltaDagger := o.ServiceAccounts
	postIota := o.ValidatorKeys
	postVarphi := o.Authorizers

	// Update the posterior state
	store.GetPosteriorStates().SetChi(postChi)
	store.GetIntermediateStates().SetDeltaDagger(deltaDagger)
	store.GetPosteriorStates().SetIota(postIota)
	store.GetPosteriorStates().SetVarphi(postVarphi)
}
