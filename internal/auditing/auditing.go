package auditing

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// (17.1) Q ∈ ⟦W?⟧C(17.1)
// (17.2) Q ≡ ρ[c]w if ρ[c]w ∈ W
//
//	∅ otherwise 		c <− NC

func GetQ() (Q types.ReportsForAllCores) { // TODO, store where?
	var W []types.WorkReport // expected to get W* from store
	Q = make([]types.ReportsForCore, types.CoresCount)
	for _, report := range W {
		Q[report.CoreIndex] = append(Q[report.CoreIndex], report)
	}
	return Q
}

// (17.3) s0 ∈ F[]  ⟨XU ⌢ Y(Hv)⟩
//
//	κ[v]b
//
// (17.4) U = $jam_audit
func GetS0(validator_index int) types.BandersnatchVrfSignature { // TODO: get local validator index
	// Follow formula 6.22's implementation to get Y(H_v)
	s := store.GetInstance()

	posterior_state := s.GetPosteriorStates()
	header := s.GetIntermediateHeader()

	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch // Use local key??
	entropy_source := header.EntropySource
	handler, _ := safrole.CreateVRFHandler(public_key)
	vrfOutput, _ := handler.VRFIetfOutput(entropy_source[:])
	context := append(types.ByteSequence(types.JamAudit[:]), vrfOutput...)
	v := 0 // how to get validator index??
	local_key := posterior_state.GetKappa()[v].Bandersnatch

	handler2, _ := safrole.CreateVRFHandler(local_key)

	vrfOutput2, _ := handler2.VRFIetfOutput(context[:])
	return types.BandersnatchVrfSignature(vrfOutput2)
}

// (17.5)a0 = {(c, w) | (c, w) ∈ p⋅⋅⋅+10, w ≠ ∅}
// (17.6) where p = F([(c, Qc) | c c <− NC], r)
// (17.7) and r = Y(s0)

func Geta0(s0 types.BandersnatchVrfSignature, Q types.ReportsForAllCores) (output []types.CoreWorkReport) { // TODO: get s0 from local data?
	s := store.GetInstance()

	posterior_state := s.GetPosteriorStates()
	header := s.GetIntermediateHeader()
	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch // Use local key??
	handler, _ := safrole.CreateVRFHandler(public_key)
	r, _ := handler.VRFIetfOutput(s0[:])

	shuffle_array := make([]types.U32, types.CoresCount)
	for i := types.U32(0); i < types.U32(types.CoresCount); i++ {
		shuffle_array[i] = i
	}
	p := shuffle.Shuffle(shuffle_array, types.OpaqueHash(r))
	for _, idx := range p {
		if len(Q[idx]) > 0 {
			output = append(output, types.CoreWorkReport{
				Core:    types.CoreIndex(idx),
				Reports: Q[idx],
			})
		}
	}
	return output
}
