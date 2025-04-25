package authorization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func f(eg types.ReportGuarantee, alpha types.AuthPools, c types.CoreIndex) (types.AuthPools, error) {
	ac := alpha[c]
	set := eg.Report.AuthorizerHash
	ac.RemovePairedValue(set)
	return alpha, nil
}

func updateAlpha(alpha types.AuthPools, varphi types.AuthQueues, egs types.GuaranteesExtrinsic, slot types.TimeSlot) (types.AuthPools, error) {
	for _, eg := range egs {
		c := eg.Report.CoreIndex
		newAlpha, err := f(eg, alpha, c)
		if err != nil {
			return nil, err
		}
		alpha = newAlpha
	}
	for i := range types.ValidatorsCount {
		varphiIndex := int(slot) % len(varphi[i])
		alpha[i] = append(alpha[i], varphi[i][varphiIndex])
		if len(alpha[i]) > types.AuthPoolMaxSize {
			alpha[i] = alpha[i][:types.AuthPoolMaxSize]
		}
	}

	return alpha, nil
}

// Outer used function
func Authorization() error {
	s := store.GetInstance()
	priorAlpha := s.GetPriorStates().GetAlpha()
	if err := priorAlpha.Validate(); err != nil {
		return err
	}
	posteriorVarphi := s.GetPosteriorStates().GetVarphi()
	if err := posteriorVarphi.Validate(); err != nil {
		return err
	}
	block := s.GetBlock()
	egs := block.Extrinsic.Guarantees
	if err := egs.Validate(); err != nil {
		return err
	}
	slot := block.Header.Slot

	// Only update cores in Egs
	postAlpha, err := updateAlpha(priorAlpha, posteriorVarphi, egs, slot)
	if err != nil {
		return err
	}

	s.GetPosteriorStates().SetAlpha(postAlpha)

	return nil
}
