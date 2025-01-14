package safrole

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// TODO VERIFY 6.15 6.16
func SealingByTickets() {
	/*
							  iy = Y(Hs)
		(6.15) γ′s ∈ ⟦C⟧ Hs ∈ F EU(H) Ha ⟨XT ⌢ η′3 ir⟩
	*/
	s := store.GetInstance()
	posterior_state := s.GetPosteriorState()
	if len(posterior_state.Gamma.GammaS.Tickets) == 0 {
		return
	}
	header := s.GetIntermediateHeader()
	index := uint(header.Slot) % uint(len(posterior_state.Gamma.GammaS.Keys))
	ticket := posterior_state.Gamma.GammaS.Tickets[index]
	public_key := posterior_state.Kappa[header.AuthorIndex].Bandersnatch
	i_r := ticket.Attempt
	message := utilities.HeaderUSerialization(header)
	eta_prime := s.GetPosteriorState().Eta

	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamTicketSeal[:])...) // XT
	context = append(context, types.ByteSequence(eta_prime[3][:])...)        // η′3
	context = append(context, types.ByteSequence([]byte{uint8(i_r)})...)     // ir

	handler, _ := CreateVRFHandler(public_key)
	signature, _ := handler.IETFSign(context, message)

	s.GetIntermediateHeaders().SetSeal(types.BandersnatchVrfSignature(signature))
}

func SealingByBandersnatchs() {
	/*
		(6.16) γ′s ∈ ⟦HB⟧  Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
	*/
	/*
		public key: Ha
		message: EU (H)
		context: XF ⌢ η′3
	*/
	s := store.GetInstance()
	posterior_state := s.GetPosteriorState()
	if len(posterior_state.Gamma.GammaS.Keys) == 0 {
		return
	}
	header := s.GetIntermediateHeader()
	index := uint(header.Slot) % uint(len(posterior_state.Gamma.GammaS.Keys))
	public_key := posterior_state.Gamma.GammaS.Keys[index]
	message := utilities.HeaderUSerialization(header)
	eta_prime := s.GetPosteriorState().Eta

	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamFallbackSeal[:])...) // XF
	context = append(context, types.ByteSequence(eta_prime[3][:])...)          // η′3

	handler, _ := CreateVRFHandler(public_key)
	signature, _ := handler.IETFSign(context, message)
	s.GetIntermediateHeaders().SetSeal(types.BandersnatchVrfSignature(signature))
}

func UpdateEtaPrime0() {
	// (6.22) η′0 ≡ H(η0 ⌢ Y(Hv))

	s := store.GetInstance()

	posterior_state := s.GetPosteriorState()
	prior_state := s.GetPriorState()
	header := s.GetIntermediateHeader()

	public_key := posterior_state.Kappa[header.AuthorIndex].Bandersnatch
	entropy_source := header.EntropySource
	eta := prior_state.Eta
	handler, _ := CreateVRFHandler(public_key)
	vrfOutput, _ := handler.VRFIetfOutput(entropy_source[:])
	hash_input := append(eta[0][:], vrfOutput...)
	s.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hash_input)))
}

func UpdateEntropy() {
	/*
								(η0, η1, η2) if e′ > e
		(6.23) (η′1, η′2, η′3)
								(η1, η2, η3) otherwise
	*/

	s := store.GetInstance()

	prior_state := s.GetPriorState()

	posterior_state := s.GetPosteriorState()

	tau := prior_state.Tau

	tauPrime := posterior_state.Tau

	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)
	eta := prior_state.Eta
	if ePrime > e {
		for i := 2; i >= 0; i-- {
			eta[i+1] = eta[i]
		}
	}
	s.GetPosteriorStates().SetEta(eta)
}

func CalculateHeaderEntropy(public_key types.BandersnatchPublic, seal types.BandersnatchVrfSignature) (sign []byte) {
	/*
		F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
		(6.17) Hv ∈ F [] Ha ⟨XE ⌢ Y(Hs)⟩
	*/
	handler, _ := CreateVRFHandler(public_key)
	var message types.ByteSequence                                        // message: []
	var context types.ByteSequence                                        //context: XE ⌢ Y(Hs)
	context = append(context, types.ByteSequence(types.JamEntropy[:])...) // XE
	vrf, _ := handler.VRFIetfOutput(seal[:])
	context = append(context, types.ByteSequence(vrf)...) // Y(Hs)
	signature, _ := handler.IETFSign(context, message)    // F [] Ha ⟨XE ⌢ Y(Hs)⟩
	return signature
}

func UpdateHeaderEntropy() {
	s := store.GetInstance()

	// Get prior state
	posterior_state := s.GetPosteriorState()

	header := s.GetIntermediateHeader()

	public_key := posterior_state.Kappa[header.AuthorIndex].Bandersnatch // Ha
	seal := header.Seal                                                  // Hs
	s.GetIntermediateHeaders().SetEntropySource(types.BandersnatchVrfSignature(CalculateHeaderEntropy(public_key, seal)))
}

func UpdateSlotKeySequence() {
	/*
		Slot Key Sequence Update
						Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
		(6.24) γ′s ≡    γs if e′ = e
						F(η′2, κ′) otherwise
	*/
	// CalculateNewEntropy
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorState()

	// Get posterior state
	posteriorState := s.GetPosteriorState()

	// Get previous time slot index
	tau := priorState.Tau

	// Get current time slot index
	tauPrime := posteriorState.Tau

	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)
	eta_prime := posteriorState.Eta

	slot_index := GetSlotIndex(priorState.Tau)
	var new_GammaS types.TicketsOrKeys
	if ePrime == e+1 {
		if len(priorState.Gamma.GammaA) == types.EpochLength && int(slot_index) >= types.SlotSubmissionEnd { // Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
			new_GammaS.Tickets = OutsideInSequencer(&priorState.Gamma.GammaA)
		} else { //F(η′2, κ′) otherwise
			new_GammaS.Keys = FallbackKeySequence(eta_prime[2], posteriorState.Kappa)
		}
	}
	s.GetPosteriorStates().SetGammaS(new_GammaS)

}
