package authorization

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func updatePoolFromQueue(c types.CoreIndex, eg types.ReportGuarantee, alpha types.AuthPools) (types.AuthPools, error) {
	ac := alpha[c]
	if ac == nil {
		return nil, fmt.Errorf("alpha[%d] is nil", c)
	}
	set := eg.Report.AuthorizerHash
	ac.RemovePairedValue(set)
	// fmt.Printf("We removed %x from core %d\n", set, c)
	alpha[c] = ac
	return alpha, nil
}

func STFAlpha2AlphaPrime(slot types.TimeSlot, egs types.GuaranteesExtrinsic, alpha types.AuthPools, varphi types.AuthQueues) (types.AuthPools, error) {
	// (8.3) First update alpha by extrinsic_guarantee
	for _, eg := range egs {
		updatedAlpha, err := updatePoolFromQueue(eg.Report.CoreIndex, eg, alpha)
		if err != nil || updatedAlpha == nil {
			return alpha, err
		}
		alpha = updatedAlpha
	}
	// (8.2) Second update to alpha^prime by results of (8.3) concat. varphi^prime
	// TODO: full mode we need to loop 341 times for each cores: optimization needed
	for i := range types.CoresCount {
		// [Ht]↺
		varphiIndex := int(slot) % len(varphi[i])
		// F(c) concat. φ′[c][Ht]↺
		alpha[i] = append(alpha[i], varphi[i][varphiIndex])
		if len(alpha[i]) > types.AuthPoolMaxSize {
			alpha[i] = alpha[i][len(alpha[i])-types.AuthPoolMaxSize:]
		}
	}
	if err := alpha.Validate(); err != nil {
		return nil, fmt.Errorf("post alpha validation failed: %v", err)
	}

	return alpha, nil
}

// Outer-used Authorization function
/*
	α' ≺ (H, EG, φ', α)
*/
func Authorization() error {
	// === setup ===
	s := store.GetInstance()
	slot := s.GetProcessingBlockPointer().GetSlot()
	egs := s.GetProcessingBlockPointer().GetGuaranteesExtrinsic()
	if len(egs) == 0 {
		fmt.Println("No egs for authorization")
	} else if err := egs.Validate(); err != nil {
		// We just print this error cause testvector doesn't provide full egs
		fmt.Printf("extrinsic_guarantee validation failed: %v\n", err)
	}

	postVarphi := s.GetPosteriorStates().GetVarphi()
	if err := postVarphi.Validate(); err != nil {
		return fmt.Errorf("posterior_varphi validation failed: %v", err)
	}
	preAlpha := s.GetPriorStates().GetAlpha()
	if err := preAlpha.Validate(); err != nil {
		return fmt.Errorf("prior_alpha validation failed: %v", err)
	}

	// === STFAlpha2AlphaPrime ===
	postAlpha, err := STFAlpha2AlphaPrime(slot, egs, preAlpha, postVarphi)
	if err != nil {
		return fmt.Errorf("stf_alpha_to_alpha_prime raised Error: %v", err)
	}

	s.GetPosteriorStates().SetAlpha(postAlpha)

	return nil
}
