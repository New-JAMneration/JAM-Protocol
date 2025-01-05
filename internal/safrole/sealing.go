package safrole

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

var kJamEntropy = "jam_entropy"            // XE
var kJamFallbackSeal = "jam_fallback_seal" // XF
var kJamTicketSeal = "jam_ticket_seal"     // XT
var kSlotSubmissionEnd = 500               // Y
/*
- [x] $\mathbf{H}_p$ : parent hash
- [x] $\mathbf{H}_r$ : prior state root
- [ ] $\mathbf{H}_x$ : extrinsic hash
- [x] $\mathbf{H}_t$ : a time slot index
- [x] $\mathbf{H}_e$ : the epoch
- [x] $\mathbf{H}_w$ : winning tickets
- [x] $\mathbf{H}_o$ : offenders markders
- [x] $\mathbf{H}_i$ : a Bandersnatch block author index
- [x] $\mathbf{H}_v$ : the entropy-rielding VRF signatrue
- [ ] $\mathbf{H}_s$ : a block seal
*/
func SealingByTickets(state types.State, header types.Header, eta_p types.EntropyBuffer) types.State {
	/*
				F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
							  iy = Y(Hs)
		(6.15) γ′s ∈ ⟦C⟧ Hs ∈ F EU(H) Ha ⟨XT ⌢ η′3 ir⟩
	*/
	// TODO ir EU H Ha
	var concat types.ByteSequence
	concat = append(concat, types.ByteSequence(kJamTicketSeal[:])...) // XT
	concat = append(concat, types.ByteSequence(eta_p[3][:])...)       // η′3
	//
	fmt.Print(concat)
	return state
}
func SealingByBandersnatchs(state types.State, header types.Header, eta_p types.EntropyBuffer) types.State {
	/*
				F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
		See section 3.8.
	*/
	/*
		(6.16) γ′s ∈ ⟦HB⟧  Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
	*/
	// TODO EU H Ha
	var concat types.ByteSequence
	concat = append(concat, types.ByteSequence(kJamFallbackSeal[:])...) // XF
	concat = append(concat, types.ByteSequence(eta_p[3][:])...)         // η′3
	// TODO
	// header.Seal = Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
	// header.
	fmt.Print(concat)
	return state
}
func Sealing(state types.State, header types.Header) {
	// TODO using global
	/*
		type State struct {
			Tau           TimeSlot                   `json:"tau"`            // Most recent block's timeslot
			Eta           EntropyBuffer              `json:"eta"`            // Entropy accumulator and epochal randomness
			Lambda        ValidatorsData             `json:"lambda"`         // Validator keys and metadata which were active in the prior epoch
			Kappa         ValidatorsData             `json:"kappa"`          // Validator keys and metadata currently active
			GammaK        ValidatorsData             `json:"gamma_k"`        // Validator keys for the following epoch
			Iota          ValidatorsData             `json:"iota"`           // Validator keys and metadata to be drawn from next
			GammaA        TicketsAccumulator         `json:"gamma_a"`        // Sealing-key contest ticket accumulator
			GammaS        TicketsOrKeys              `json:"gamma_s"`        // Sealing-key series of the current epoch
			GammaZ        BandersnatchRingCommitment `json:"gamma_z"`        // Bandersnatch ring commitment
			PostOffenders []Ed25519Public            `json:"post_offendors"` // Posterior offenders sequence
		}
	*/
	/*
		Entropy Update
		(6.21) η ∈ ⟦H⟧4
		(6.22) η′0 ≡ H(η0 ⌢ Y(Hv))
								(η0, η1, η2) if e′ > e
		(6.23) (η′1, η′2, η′3)
								(η1, η2, η3) otherwise
	*/
	e := GetEpochIndex(state.Tau)
	e_prime := GetEpochIndex(header.Slot)
	m := GetSlotIndex(state.Tau)
	eta_prime := state.Eta
	if e_prime > e { // e′ > e
		for i := 2; i >= 0; i-- {
			eta_prime[i+1] = eta_prime[i]
		}
		eta_prime[0] = types.Entropy(hash.Blake2bHash(eta_prime[1][:])) // TODO concat Y(Hv)
	}

	if len(state.Gamma.GammaS.Tickets) > 0 {
		state = SealingByTickets(state, header, eta_prime)
	} else if len(state.Gamma.GammaS.Keys) > 0 {
		state = SealingByBandersnatchs(state, header, eta_prime)
	}
	/*
		TODO (6.17) Hv ∈ F[] Ha ⟨XE ⌢ Y(Hs)⟩
	*/
	/*
		Slot Key Sequence Update
						Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
		(6.24) γ′s ≡    γs if e′ = e
						F(η′2, κ′) otherwise
	*/
	if e_prime > e {
		state.Eta = eta_prime
	}
	if e_prime == e+1 {
		if len(state.Gamma.GammaA) == types.EpochLength && int(m) >= kSlotSubmissionEnd { // Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
			state.Gamma.GammaS.Tickets = OutsideInSequencer(&state.Gamma.GammaA)
		} else { //F(η′2, κ′) otherwise
			state.Gamma.GammaS.Keys = FallbackKeySequence(types.Entropy(eta_prime[2]), state.Kappa)
		}
	}
}
