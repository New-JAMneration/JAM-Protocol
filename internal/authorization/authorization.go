package authorization

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func updatePoolFromQueue(coreIndex types.CoreIndex, eg types.ReportGuarantee, alpha types.AuthPools) (types.AuthPools, error) {
	pool := alpha[coreIndex]
	if pool == nil {
		return nil, fmt.Errorf("alpha[%d] is nil", coreIndex)
	}
	authHashToRemoved := eg.Report.AuthorizerHash
	// Removed $authHashToRemoved from $coreIndex
	pool.RemovePairedValue(authHashToRemoved)

	alpha[coreIndex] = pool
	return alpha, nil
}

func STFAlpha2AlphaPrime(slot types.TimeSlot, guarantees types.GuaranteesExtrinsic, alpha types.AuthPools, varphi types.AuthQueues) (types.AuthPools, error) {
	// (8.3) Remove used authorizer from E_G
	for _, guarantee := range guarantees {
		updatedAlpha, err := updatePoolFromQueue(guarantee.Report.CoreIndex, guarantee, alpha)
		if err != nil || updatedAlpha == nil {
			return alpha, err
		}
		alpha = updatedAlpha
	}

	// (8.2) Append φ′[c][Ht↺] for each core
	// TODO: full mode we need to loop 341 times for each cores: optimization needed
	for coreIndex := range types.CoresCount {
		queue := varphi[coreIndex]
		if len(queue) == 0 {
			fmt.Printf("WARNING: varphi[%d] is empty, skipping append\n", coreIndex)
			continue
		}
		index := int(slot) % len(queue)
		alpha[coreIndex] = append(alpha[coreIndex], queue[index])

		if len(alpha[coreIndex]) > types.AuthPoolMaxSize {
			alpha[coreIndex] = alpha[coreIndex][len(alpha[coreIndex])-types.AuthPoolMaxSize:]
		}
	}
	if err := alpha.Validate(); err != nil {
		return nil, fmt.Errorf("post alpha validation failed: %v", err)
	}

	return alpha, nil
}

// Authorization performs the update of the core authorization pools α → α′,
// as defined in Graypaper §8.2–§8.3 (Authorization, Pool and Queue).
// key inputs:
// - H: current block header (to derive time slot t)
// - E_G: extrinsic guarantees in the block
// - φ′: posterior state of the authorizer queue (varphi)
// - α: prior authorization pool (alpha)
func Authorization() error {
	// Load state
	s := store.GetInstance()
	block := s.GetLatestBlock()
	slot := block.Header.Slot
	guarantees := block.Extrinsic.Guarantees
	if len(guarantees) == 0 {
		fmt.Println("No guarantees found in the block extrinsic, no authorization needed.")
	} else if err := guarantees.Validate(); err != nil {
		fmt.Printf("extrinsic_guarantee validation failed: %v\n", err)
	}
	// Validate input EG and φ′
	posteriorVarphi := s.GetPosteriorStates().GetVarphi()
	if err := posteriorVarphi.Validate(); err != nil {
		return fmt.Errorf("posterior_varphi validation failed: %v", err)
	}
	priorAlpha := s.GetPriorStates().GetAlpha()
	if err := priorAlpha.Validate(); err != nil {
		return fmt.Errorf("prior_alpha validation failed: %v", err)
	}
	// Apply STF transition: α′ ≺ (H, E_G, φ′, α)
	postAlpha, err := STFAlpha2AlphaPrime(slot, guarantees, priorAlpha, posteriorVarphi)
	if err != nil {
		return fmt.Errorf("stf_alpha_to_alpha_prime raised Error: %v", err)
	}

	s.GetPosteriorStates().SetAlpha(postAlpha)

	return nil
}
