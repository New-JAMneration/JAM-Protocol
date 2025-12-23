package safrole

import (
	"bytes"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/logger"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
	"github.com/google/go-cmp/cmp"
)

// TODO VERIFY 6.15 6.16
func SealingByTickets() error {
	/*
							  iy = Y(Hs)
		(6.15) γ′s ∈ ⟦C⟧ Hs ∈ F EU(H) Ha ⟨XT ⌢ η′3 ir⟩
	*/
	s := store.GetInstance()
	posteriorState := s.GetPosteriorStates()
	gammaSTickets := posteriorState.GetGammaS().Tickets
	header := s.GetLatestBlock().Header
	index := uint(header.Slot) % uint(len(gammaSTickets))
	ticket := gammaSTickets[index]
	publicKey := posteriorState.GetKappa()[header.AuthorIndex].Bandersnatch
	i_r := ticket.Attempt
	message, err := utilities.HeaderUSerialization(header)
	if err != nil {
		return err
	}
	eta_prime := posteriorState.GetEta()

	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamTicketSeal[:])...) // XT
	context = append(context, types.ByteSequence(eta_prime[3][:])...)        // η′3
	context = append(context, types.ByteSequence([]byte{uint8(i_r)})...)     // ir

	handler, err := CreateVRFHandler(publicKey)
	if err != nil {
		return err
	}
	signature, err := handler.IETFSign(context, message)
	if err != nil {
		return err
	}
	s.GetProcessingBlockPointer().SetSeal(types.BandersnatchVrfSignature(signature))
	return nil
}

func SealingByBandersnatchs() error {
	/*
		(6.16) γ′s ∈ ⟦HB⟧  Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
	*/
	/*
		public key: Ha
		message: EU (H)
		context: XF ⌢ η′3
	*/
	s := store.GetInstance()
	posterior_state := s.GetPosteriorStates()
	GammaSKeys := posterior_state.GetGammaS().Keys
	header := s.GetLatestBlock().Header
	index := uint(header.Slot) % uint(len(GammaSKeys))
	publicKey := GammaSKeys[index]
	message, err := utilities.HeaderUSerialization(header)
	if err != nil {
		return err
	}
	eta_prime := posterior_state.GetEta()

	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamFallbackSeal[:])...) // XF
	context = append(context, types.ByteSequence(eta_prime[3][:])...)          // η′3

	handler, err := CreateVRFHandler(publicKey)
	if err != nil {
		return err
	}
	signature, err := handler.IETFSign(context, message)
	if err != nil {
		return err
	}
	s.GetProcessingBlockPointer().SetSeal(types.BandersnatchVrfSignature(signature))
	return nil
}

// (6.15~6.16) Make H_s (seal for new header)
func SealingHeader() error {
	s := store.GetInstance()
	gammaS := s.GetPosteriorStates().GetGammaS()
	if len(gammaS.Keys) > 0 {
		err := SealingByBandersnatchs()
		if err != nil {
			return err
		}
	} else if len(gammaS.Tickets) > 0 {
		err := SealingByTickets()
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateEtaPrime0() error {
	// (6.22) η′0 ≡ H(η0 ⌢ Y(Hv))

	s := store.GetInstance()

	priorState := s.GetPriorStates()
	header := s.GetLatestBlock().Header

	entropySource := header.EntropySource
	eta := priorState.GetEta()

	vrfOutput, err := vrf.VRFIetfOutput(entropySource[:])
	if err != nil {
		return fmt.Errorf("VRFIetfOutput: %w", err)
	}
	hashInput := append(eta[0][:], vrfOutput...)
	s.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hashInput)))
	return nil
}

func UpdateEntropy(e types.TimeSlot, ePrime types.TimeSlot) {
	/*
								(η0, η1, η2) if e′ > e
		(6.23) (η′1, η′2, η′3)
								(η1, η2, η3) otherwise
	*/

	s := store.GetInstance()
	eta := s.GetPriorStates().GetEta()
	if ePrime > e {
		for i := 2; i >= 0; i-- {
			eta[i+1] = eta[i]
		}
	}
	// This make sure we won't overwrite eta0
	eta[0] = s.GetPosteriorStates().GetEta0()
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
	var context types.ByteSequence                                        // context: XE ⌢ Y(Hs)
	context = append(context, types.ByteSequence(types.JamEntropy[:])...) // XE
	vrf, _ := handler.VRFIetfOutput(seal[:])
	context = append(context, types.ByteSequence(vrf)...) // Y(Hs)
	signature, _ := handler.IETFSign(context, message)    // F [] Ha ⟨XE ⌢ Y(Hs)⟩
	return signature
}

func ValidateHeaderEntropy(header types.Header, priorState *types.State) *types.ErrorCode {
	seal := header.Seal
	var message types.ByteSequence // message: []
	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamEntropy[:])...) // XE
	vrfOutput, err := vrf.VRFIetfOutput(seal[:])
	if err != nil {
		logger.Errorf("ValidateHeaderEntropy VRFIetfOutput: %v", err)
	}
	context = append(context, types.ByteSequence(vrfOutput)...) // Y(Hs)
	signature := header.EntropySource[:]                        // Hv
	signerKey := priorState.Kappa[header.AuthorIndex].Bandersnatch[:]
	_, err = vrf.IETFVerify(context, message, signature, signerKey)
	if err != nil {
		errCode := SafroleErrorCode.VrfEntropyInvalid
		return &errCode
	}
	return nil
}

func ValidateByBandersnatchs(header types.Header, state *types.State) *types.ErrorCode {
	gammaSKeys := state.Gamma.GammaS.Keys
	// logger.Debugf("length of gammaSKeys:", len(gammaSKeys))
	index := uint(header.Slot) % uint(len(gammaSKeys))

	expectedPublicKey := gammaSKeys[index]
	actualPublicKey := state.Kappa[header.AuthorIndex].Bandersnatch

	// i == H_A
	if expectedPublicKey != actualPublicKey {
		errCode := SafroleErrorCode.UnexpectedAuthor
		return &errCode
	}

	// Backup verification method
	message, err := utilities.HeaderUSerialization(header)
	if err != nil {
		logger.Errorf("ValidateByBandersnatchs HeaderUSerialization err: %v", err)
		return nil
	}
	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamFallbackSeal[:])...) // XF
	context = append(context, types.ByteSequence(state.Eta[3][:])...)
	signature := header.Seal[:]

	signerKey := state.Kappa[header.AuthorIndex].Bandersnatch[:]
	_, err = vrf.IETFVerify(context, message, signature, signerKey)
	if err != nil { // Backup verification method to find correct author
		// for i := 0; i < len(state.Kappa); i++ {
		// 	if i == int(header.AuthorIndex) {
		// 		continue
		// 	}
		// 	// logger.Debugf("verifying author index:", i)
		// 	signerKey = state.Kappa[i].Bandersnatch[:]
		// 	if _, err = vrf.IETFVerify(context, message, signature, signerKey); err == nil {
		// 		logger.Debugf("author index:", header.AuthorIndex, "is wrong but", i, "is valid")
		// 		// Found a valid signature, return Bad signature
		// 		errCode := SafroleErrorCode.UnexpectedAuthor
		// 		return &errCode
		// 	}
		// }
		logger.Errorf("ValidateByBandersnatchs IETFVerify err: %v", err)
		errCode := SafroleErrorCode.VrfSealInvalid
		return &errCode
	}
	return nil
}

// Need test vectors to verify correctness
func ValidateByTickets(header types.Header, state *types.State) *types.ErrorCode {
	gammaSTickets := state.Gamma.GammaS.Tickets

	index := uint(header.Slot) % uint(len(gammaSTickets))
	ticket := gammaSTickets[index]

	// i_y == Y(Hs)
	vrfOutput, err := vrf.VRFIetfOutput(header.Seal[:])
	if err != nil {
		logger.Errorf("ValidateByTickets VRFIetfOutput err: %v", err)
		return nil
	}
	if !bytes.Equal(vrfOutput, ticket.Id[:]) {
		logger.Errorf(cmp.Diff(vrfOutput, ticket.Id[:]))
		errCode := SafroleErrorCode.VrfSealInvalid
		return &errCode
	}

	// Backup verification method
	message, err := utilities.HeaderUSerialization(header)
	if err != nil {
		logger.Errorf("ValidateByTickets HeaderUSerialization err: %v", err)
		return nil
	}
	eta_prime := state.Eta

	var context types.ByteSequence
	context = append(context, types.ByteSequence(types.JamTicketSeal[:])...) // XT
	context = append(context, types.ByteSequence(eta_prime[3][:])...)        // η′3
	context = append(context, byte(ticket.Attempt))                          // ir (uint8)

	signature := header.Seal[:]

	signerKey := state.Kappa[header.AuthorIndex].Bandersnatch[:]
	_, err = vrf.IETFVerify(context, message, signature, signerKey)
	if err != nil { // Backup verification method to find correct author
		// logger.Debugf("Validate ticket by author index:", header.AuthorIndex, "failed")
		// for i := 0; i < len(state.Kappa); i++ {
		// 	if i == int(header.AuthorIndex) {
		// 		continue
		// 	}
		// 	signerKey = state.Kappa[i].Bandersnatch[:]
		// 	// logger.Debugf("Validating by author index:", i)
		// 	if _, err = vrf.IETFVerify(context, message, signature, signerKey); err == nil {
		// 		logger.Debugf("author index:", header.AuthorIndex, "is wrong but", i, "is valid")
		// 		// Found a valid signature, return Bad signature
		// 		errCode := SafroleErrorCode.UnexpectedAuthor
		// 		return &errCode
		// 	}
		// }
		// All verifiers failed, return Bad signature
		logger.Errorf("ValidateByTickets IETFVerify err: %v", err)
		errCode := SafroleErrorCode.VrfSealInvalid
		return &errCode
	}

	return nil
}

func ValidateHeaderSeal(header types.Header, state *types.State) *types.ErrorCode {
	gammaS := state.Gamma.GammaS
	if len(gammaS.Tickets) > 0 {
		// logger.Debugf("Validating by tickets")
		return ValidateByTickets(header, state)
	} else {
		// logger.Debugf("Validating by bandersnatchs")
		return ValidateByBandersnatchs(header, state)
	}
}

// NO REFERENCES
func UpdateHeaderEntropy() {
	s := store.GetInstance()

	// Get prior state
	posterior_state := s.GetPosteriorStates()

	header := s.GetProcessingBlockPointer().GetHeader()

	publicKey := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch // Ha
	seal := header.Seal                                                      // Hs
	s.GetProcessingBlockPointer().SetEntropySource(types.BandersnatchVrfSignature(CalculateHeaderEntropy(publicKey, seal)))
}

// Calculate gamma^prime_s
func UpdateSlotKeySequence(e types.TimeSlot, ePrime types.TimeSlot, slotIndex types.TimeSlot) {
	/*
		Slot Key Sequence Update
						Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
		(6.24) γ′s ≡    γs if e′ = e
						F(η′2, κ′) otherwise
	*/
	s := store.GetInstance()

	// Get prior state
	priorState := s.GetPriorStates()
	gammaA := priorState.GetGammaA()

	// Get posterior state
	posteriorState := s.GetPosteriorStates()
	etaPrime := posteriorState.GetEta()

	var newGammaS types.TicketsOrKeys

	if ePrime == e+1 && len(gammaA) == types.EpochLength && int(slotIndex) >= types.SlotSubmissionEnd { // Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
		newGammaS.Tickets = OutsideInSequencer(&gammaA)
	} else if ePrime == e { // γs if e′ = e
		newGammaS = priorState.GetGammaS()
	} else { // F(η′2, κ′) otherwise
		newGammaS.Keys = FallbackKeySequence(etaPrime[2], posteriorState.GetKappa())
	}
	posteriorState.SetGammaS(newGammaS)
}
