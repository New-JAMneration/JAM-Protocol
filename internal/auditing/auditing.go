package auditing

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// (17.1) Q ∈ ⟦W?⟧C(17.1)
// (17.2) Q ≡ ρ[c]w if ρ[c]w ∈ W
//
//	∅ otherwise 		c <− NC

func getQ() (Q []*types.WorkReport) { // TODO, store where?
	rho := store.GetInstance().GetPosteriorStates().GetState().Rho

	// 建立固定長度 slice，預設每個元素是 nil（代表 ∅）
	Q = make([]*types.WorkReport, len(rho))

	for index, assignment := range rho {
		if assignment != nil {
			Q[index] = &assignment.Report
			// 否則保持 nil，表示 ∅
		} else {
			Q[index] = nil
		}
	}
	return Q
}

// (17.3) s0 ∈ F[]  ⟨XU ⌢ Y(Hv)⟩
//
//	κ[v]b
//
// (17.4) U = $jam_audit
func getS0(validator_index int) types.BandersnatchVrfSignature { // TODO: get local validator index
	// Follow formula 6.22's implementation to get Y(H_v)
	s := store.GetInstance()

	posterior_state := s.GetPosteriorStates()
	header := s.GetIntermediateHeader()

	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch // Use local key??
	entropy_source := header.EntropySource
	handler, _ := safrole.CreateVRFHandler(public_key)
	vrfOutput, _ := handler.VRFIetfOutput(entropy_source[:])
	context := append(types.ByteSequence(types.JamAudit[:]), vrfOutput...)
	v := validator_index // how to get validator index??
	local_key := posterior_state.GetKappa()[v].Bandersnatch

	handler2, _ := safrole.CreateVRFHandler(local_key)

	vrfOutput2, _ := handler2.VRFIetfOutput(context[:])
	return types.BandersnatchVrfSignature(vrfOutput2)
}

// (17.5)a0 = {(c, w) | (c, w) ∈ p⋅⋅⋅+10, w ≠ ∅}
// (17.6) where p = F([(c, Qc) | c c <− NC], r)
// (17.7) and r = Y(s0)

func getA0(s0 types.BandersnatchVrfSignature, Q []*types.WorkReport) (output []types.CoreIndexReport) { // TODO: get s0 from local data?
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
		if Q[idx] != nil {
			output = append(output, types.CoreIndexReport{
				CoreID: types.CoreIndex(idx),
				Report: *Q[idx],
			})
		}
	}
	return output
}

// (17.8) let n = (T − P ⋅ Ht) / A
// - T  = current wall-clock time (seconds)
// - P  = SlotPeriod (e.g. 6s, constant)
// - Ht = slot number from block header
// - A  = TranchePeriod (e.g. 8s, constant)
func getTranchIndex(T types.U32) types.U32 { // TODO: how to get T, type?
	H_t := store.GetInstance().GetIntermediateHeader().Slot
	n := (T - types.U32(types.SlotPeriod)*types.U32(H_t)) / types.TranchePeriod
	return n
}

// (17.9) S ≡ Eκ[v]e ⟨XI + n ⌢ xn ⌢ H(H)⟩
// (17.10) where xn = E([E2(c) ⌢ H(w) S(c, w) ∈ an])
// (17.11) XI = $jam_announce

func BuildAnnouncement(tranche types.U32, reports []types.CoreIndexReport, hashFunc func(types.ByteSequence) types.OpaqueHash) []types.BandersnatchVrfSignature {
	var output types.ByteSequence

	for _, value := range reports {
		output = append(output, utilities.SerializeFixedLength(types.U64(value.CoreID), 2)...)
		H_w := hashFunc(utilities.WorkReportSerialization(value.Report))
		output = append(output, H_w[:]...)
	}
	xn := utilities.SerializeByteSequence(output)
	H := store.GetInstance().GetIntermediateHeader()
	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamAnnounce[:])...)   // XI
	context = append(context, types.ByteSequence([]byte{uint8(tranche)})...) // n
	context = append(context, xn...)                                         // xn
	context = append(context, utilities.HeaderSerialization(H)...)           // H(H)
	return signature
}
