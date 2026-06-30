package validator

import "github.com/New-JAMneration/JAM-Protocol/internal/blockchain"

// RefreshFromChain reloads prev/current/next validator sets and self index from chain state.
func (vm *ValidatorManager) RefreshFromChain(chain *blockchain.ChainState) {
	if vm == nil || chain == nil {
		return
	}
	prior := chain.GetPriorStates()
	if vm.Grid == nil {
		vm.Grid = &GridMapper{}
	}
	vm.Grid.Previous = prior.GetLambda()
	vm.Grid.Current = prior.GetKappa()
	vm.Grid.Next = prior.GetGammaK()
	vm.SelfIndex, _ = vm.Grid.FindIndex(vm.SelfKey)
}
