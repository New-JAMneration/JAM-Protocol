package authorization

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func updatePoolFromQueue(eg types.ReportGuarantee, alpha types.AuthPools, c types.CoreIndex) (types.AuthPools, error) {
	ac := alpha[c]
	if ac == nil {
		return nil, errors.New("alpha[c] is nil")
	}
	set := eg.Report.AuthorizerHash
	ac.RemovePairedValue(set)
	fmt.Printf("We removed %x from core %d\n", set, c)
	alpha[c] = ac
	return alpha, nil
}

func updateAlpha(alpha types.AuthPools, varphi types.AuthQueues, egs types.GuaranteesExtrinsic, slot types.TimeSlot) (types.AuthPools, error) {
	// (8.3) First update alpha by extrinsic_guarantee
	newAlpha := alpha
	for _, eg := range egs {
		c := eg.Report.CoreIndex
		updatedAlpha, err := updatePoolFromQueue(eg, newAlpha, c)
		if err != nil || updatedAlpha == nil {
			return nil, err
		}
		newAlpha = updatedAlpha
	}
	// (8.2) Second update to alpha^prime by results of (8.3) and varphi^prime
	for i := range types.CoresCount {
		if i >= len(varphi) {
			return nil, fmt.Errorf("index %d is out of bound: %d", i, len(varphi))
		}
		varphiIndex := int(slot) % len(varphi[i])
		if varphiIndex >= len(varphi[i]) {
			return nil, fmt.Errorf("varphiIndex %d is out of bound: %d", varphiIndex, len(varphi[i]))
		}
		// fmt.Println("varphi index", varphi[i])
		// fmt.Println("varphi index", varphiIndex)
		// fmt.Printf("varphi index %x\n", varphi[i][varphiIndex])
		newAlpha[i] = append(newAlpha[i], varphi[i][varphiIndex])
		// fmt.Printf("We added %x to core %d\n", varphi[i][varphiIndex], i)
		if len(newAlpha[i]) > types.AuthPoolMaxSize {
			newAlpha[i] = newAlpha[i][len(newAlpha[i])-types.AuthPoolMaxSize:]
		}
	}

	return newAlpha, nil
}

// Outer used function
/*
	α' ≺ (H, EG, φ', α)
*/
func Authorization() error {
	s := store.GetInstance()
	slot := s.GetProcessingBlockPointer().GetSlot()
	egs := s.GetProcessingBlockPointer().GetGuaranteesExtrinsic()
	// fmt.Printf("Authrorization get egs %+v\n", egs)
	// if err := egs.Validate(); err != nil {
	// 	return fmt.Errorf("extrinsic_guarantee raised Error: %v", err)
	// }
	if len(egs) == 0 {
		fmt.Println("No extrinsic_guarantee")
	}
	priorAlpha := s.GetPriorStates().GetAlpha()
	if err := priorAlpha.Validate(); err != nil || len(priorAlpha) == 0 {
		return fmt.Errorf("prior_alpha raised Error: %v", err)
	}
	posteriorVarphi := s.GetPosteriorStates().GetVarphi()
	if err := posteriorVarphi.Validate(); err != nil || len(posteriorVarphi) == 0 {
		return fmt.Errorf("posterior_varphi raised Error: %v", err)
	}

	// Only update cores in Egs
	postAlpha, err := updateAlpha(priorAlpha, posteriorVarphi, egs, slot)
	if err != nil || len(postAlpha) == 0 {
		return fmt.Errorf("update_alpha raised Error: %v", err)
	}
	if reflect.DeepEqual(priorAlpha, postAlpha) {
		fmt.Println("prior_alpha and post_alpha are equal")
	} else if !reflect.DeepEqual(priorAlpha, postAlpha) {
		for i := range priorAlpha {
			fmt.Printf("priorAlpha[%d], %v\n", i, priorAlpha[i])
			fmt.Printf("postAlpha[%d], %v\n", i, postAlpha[i])
		}
	}

	s.GetPosteriorStates().SetAlpha(postAlpha)

	return nil
}
