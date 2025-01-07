package safrole

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

var kJamEntropy = "jam_entropy"            // XE
var kJamFallbackSeal = "jam_fallback_seal" // XF
var kJamTicketSeal = "jam_ticket_seal"     // XT
var kSlotSubmissionEnd = 500               // Y

func SealingByTickets(state types.State, header types.Header, eta_p types.EntropyBuffer) (sign []byte, vrf []byte) {
	/*
				F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
							  iy = Y(Hs)
		(6.15) γ′s ∈ ⟦C⟧ Hs ∈ F EU(H) Ha ⟨XT ⌢ η′3 ir⟩
	*/
	public_key := state.Kappa[header.AuthorIndex].Bandersnatch
	i_r := state.Gamma.GammaS.Tickets[GetSlotIndex(header.Slot)].Attempt // header.Slot or  GetSlotIndex(state.Tau) or ????
	message := utilities.HeaderUSerialization(header)
	var context types.ByteSequence
	context = append(context, types.ByteSequence(kJamTicketSeal[:])...)  // XT
	context = append(context, types.ByteSequence(eta_p[3][:])...)        // η′3
	context = append(context, types.ByteSequence([]byte{uint8(i_r)})...) // ir

	bandersnatchKeys := []types.BandersnatchPublic{public_key}
	handler, _ := CreateVRFHandler(bandersnatchKeys)
	signature, _ := handler.IETFSign(context, message)
	vrf_output, _ := handler.VRFOutput(signature)

	sign = signature
	vrf = vrf_output
	return sign, vrf
}

func SealingByBandersnatchs(state types.State, header types.Header, eta_p types.EntropyBuffer) (sign []byte) {
	/*
		F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
		(6.16) γ′s ∈ ⟦HB⟧  Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
	*/
	/*
		public key: Ha  Bandersnatch key of the block author  header.AuthorIndex
		message: EU (H)
		context: XF ⌢ η′3
	*/
	public_key := state.Kappa[header.AuthorIndex].Bandersnatch
	message := utilities.HeaderUSerialization(header)
	var context types.ByteSequence
	context = append(context, types.ByteSequence(kJamFallbackSeal[:])...) // XF
	context = append(context, types.ByteSequence(eta_p[3][:])...)         // η′3

	bandersnatchKeys := []types.BandersnatchPublic{public_key}
	handler, _ := CreateVRFHandler(bandersnatchKeys)
	signature, _ := handler.IETFSign(context, message)

	sign = signature
	return sign
}

func UpdateEntropy(Eta types.EntropyBuffer, validators types.ValidatorsData) types.EntropyBuffer {
	/*
		Entropy Update
		(6.21) η ∈ ⟦H⟧4
		(6.22) η′0 ≡ H(η0 ⌢ Y(Hv))
								(η0, η1, η2) if e′ > e
		(6.23) (η′1, η′2, η′3)
								(η1, η2, η3) otherwise
	*/
	// TODO: check the correct usage of header state, vrf
	inter := store.IntermediateHeader{}
	header := inter.GetHeader()
	for i := 2; i >= 0; i-- {
		Eta[i+1] = Eta[i]
	}

	bandersnatchKeys := []types.BandersnatchPublic{}
	for _, validator := range validators {
		bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
	}
	handler, _ := CreateVRFHandler(bandersnatchKeys)
	vrfOutput, _ := handler.VRFOutput(header.EntropySource[:])
	hash_input := append(Eta[1][:], vrfOutput...)
	Eta[0] = types.Entropy(hash.Blake2bHash(hash_input))
	return Eta
}

func UpdateHeaderEntropy(state types.State, header types.Header) (sign []byte) {
	/*
		F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
		(6.17) Hv ∈ F [] Ha ⟨XE ⌢ Y(Hs)⟩
	*/
	/*
		public key: Ha  Bandersnatch key of the block author  header.AuthorIndex
		message: []
		context: XE ⌢ Y(Hs)
	*/
	public_key := state.Kappa[header.AuthorIndex].Bandersnatch
	var message types.ByteSequence
	var context types.ByteSequence
	context = append(context, types.ByteSequence(kJamEntropy[:])...) // XE

	bandersnatchKeys := []types.BandersnatchPublic{public_key}
	handler, _ := CreateVRFHandler(bandersnatchKeys)
	signature, _ := handler.IETFSign(context, message)

	sign = signature
	return sign
}

func Sealing() {
	s := store.GetInstance()
	state := s.GetPriorState()
	inter := store.IntermediateHeader{}
	header := inter.GetHeader()
	e := GetEpochIndex(state.Tau)
	e_prime := GetEpochIndex(header.Slot)
	m := GetSlotIndex(state.Tau)
	eta_prime := state.Eta
	if e_prime > e { // e′ > e
		/*
			Entropy Update
			(6.21) η ∈ ⟦H⟧4
			(6.22) η′0 ≡ H(η0 ⌢ Y(Hv))
									(η0, η1, η2) if e′ > e
			(6.23) (η′1, η′2, η′3)
									(η1, η2, η3) otherwise
		*/
		s.GetPosteriorStates().SetEta(UpdateEntropy(state.Eta, state.Gamma.GammaK))
	}

	if len(state.Gamma.GammaS.Tickets) > 0 {
		sign, _ := SealingByTickets(state, header, eta_prime)
		inter.SetSeal(types.BandersnatchVrfSignature(sign))
	} else if len(state.Gamma.GammaS.Keys) > 0 {
		sign := SealingByBandersnatchs(state, header, eta_prime)
		inter.SetSeal(types.BandersnatchVrfSignature(sign))
	}
	sign := UpdateHeaderEntropy(state, header)
	inter.SetEntropySource(types.BandersnatchVrfSignature(sign))
	/*
		Slot Key Sequence Update
						Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
		(6.24) γ′s ≡    γs if e′ = e
						F(η′2, κ′) otherwise
	*/
	if e_prime > e {
		s.GetPosteriorStates().SetEta(eta_prime)
	}
	if e_prime == e+1 {
		if len(state.Gamma.GammaA) == types.EpochLength && int(m) >= kSlotSubmissionEnd { // Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
			s.GetPosteriorStates().SetGammaSTickets(OutsideInSequencer(&state.Gamma.GammaA))
		} else { //F(η′2, κ′) otherwise
			s.GetPosteriorStates().SetGammaSKeys(FallbackKeySequence(eta_prime[2], state.Kappa))
		}
	}
}
