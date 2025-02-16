package safrole

func SafroleIntegration() {
	KeyRotate()
	SealingByTickets()
	SealingByBandersnatchs()
	UpdateEtaPrime0()
	UpdateEntropy()
	UpdateHeaderEntropy()
	UpdateSlotKeySequence()
	CreateEpochMarker()
	CreateWinningTickets()
}

// // GetEpochIndex returns the epoch index of the most recent block't timeslot
// // \tau : The most recent block't timeslot
// // (6.2)
// func GetEpochIndex(t types.TimeSlot) types.TimeSlot {
// 	return t / types.TimeSlot(types.EpochLength)
// }

// // GetSlotIndex returns the slot index of the most recent block't timeslot
// // \tau : The most recent block't timeslot
// // (6.2)
// func GetSlotIndex(t types.TimeSlot) types.TimeSlot {
// 	return t % types.TimeSlot(types.EpochLength)
// }

// // R function return the epoch and slot index
// // Equation (6.2)
// func R(time types.TimeSlot) (epoch types.TimeSlot, slotIndex types.TimeSlot) {
// 	epoch = time / types.TimeSlot(types.EpochLength)
// 	slotIndex = time % types.TimeSlot(types.EpochLength)
// 	return epoch, slotIndex
// }

// // --- ticketbody_controller.go ---
// // (6.5) // (6.6)
// // TicketsBodiesController is a controller for TicketsBodies
// type TicketsBodiesController struct {
// 	TicketsBodies []types.TicketBody
// }

// // NewTicketsBodiesController returns a new TicketsBodiesController
// func NewTicketsBodiesController() *TicketsBodiesController {
// 	return &TicketsBodiesController{
// 		TicketsBodies: make([]types.TicketBody, 0),
// 	}
// }

// // Validate validates the controller
// func (tbc *TicketsBodiesController) Validate() error {
// 	if len(tbc.TicketsBodies) > types.EpochLength {
// 		return fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", types.EpochLength, len(tbc.TicketsBodies))
// 	}
// 	return nil
// }

// // (6.5)
// // AddTicketBody adds a ticket body to the controller
// func (tbc *TicketsBodiesController) AddTicketBody(ticketBody types.TicketBody) error {
// 	if len(tbc.TicketsBodies) < types.EpochLength {
// 		tbc.TicketsBodies = append(tbc.TicketsBodies, ticketBody)
// 		return nil
// 	}
// 	return fmt.Errorf("TicketsBodiesController must have less than %d entries, got %d", types.EpochLength, len(tbc.TicketsBodies))
// }

// // --- ticketbody_controller.go ---

// // --- safrole.go ---
// // ValidatorIsOffender checks if the validator is an offender
// // Equation (6.14) Phi(k)
// func ValidatorIsOffender(validator types.Validator, offendersMark types.OffendersMark) bool {
// 	return slices.Contains(offendersMark, validator.Ed25519)
// }

// // ReplaceOffenderKeys replaces the Ed25519 key of the validator with a null key
// // Equation (6.14) Phi(k)
// func ReplaceOffenderKeys(validators types.ValidatorsData) types.ValidatorsData {
// 	// Get offendersMark (Psi_O) from posterior state
// 	s := store.GetInstance()
// 	posteriorState := s.GetPosteriorStates()
// 	offendersMark := posteriorState.GetPsiO()

// 	for i, validator := range validators {
// 		if ValidatorIsOffender(validator, offendersMark) {
// 			// Replace the validator's Ed25519 key with a null key
// 			validators[i].Ed25519 = types.Ed25519Public{}
// 		}
// 	}

// 	return validators
// }

// // GetBandersnatchRingRootCommmitment returns the root commitment of the
// // Bandersnatch ring.
// // O function: The Bandersnatch ring root function.
// // See section 3.8 and appendix G.
// func GetBandersnatchRingRootCommmitment(bandersnatchKeys []types.BandersnatchPublic) (types.BandersnatchRingCommitment, error) {
// 	var proverIdx uint = 0
// 	vrfHandler, handlerErr := CreateRingVRFHandler(bandersnatchKeys, proverIdx)
// 	if handlerErr != nil {
// 		return types.BandersnatchRingCommitment{}, handlerErr
// 	}
// 	defer vrfHandler.Free()

// 	commitment, commitmentErr := vrfHandler.GetCommitment()
// 	if commitmentErr != nil {
// 		return types.BandersnatchRingCommitment{}, commitmentErr
// 	}

// 	return types.BandersnatchRingCommitment(commitment), nil
// }

// // UpdateBandersnatchKeyRoot returns the root commitment of the Bandersnatch
// // ring
// // Equation (6.13)
// func UpdateBandersnatchKeyRoot(validators types.ValidatorsData) (types.BandersnatchRingCommitment, error) {
// 	bandersnatchKeys := []types.BandersnatchPublic{}
// 	for _, validator := range validators {
// 		bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
// 	}

// 	return GetBandersnatchRingRootCommmitment(bandersnatchKeys)
// }

// // keyRotation rotates the keys, it updates the state with the new Safrole state
// // Equation (6.13)
// /*
// func keyRotation(t types.TimeSlot, tPrime types.TimeSlot, safroleState types.State) (newSafroleState types.State) {
// 	e := GetEpochIndex(t)
// 	ePrime := GetEpochIndex(tPrime)

// 	if ePrime > e {
// 		// New epoch
// 		newSafroleState.Gamma.GammaK = ReplaceOffenderKeys(safroleState.Iota)
// 		newSafroleState.Kappa = safroleState.Gamma.GammaK
// 		newSafroleState.Lambda = safroleState.Kappa
// 		z, zErr := UpdateBandersnatchKeyRoot(safroleState.Gamma.GammaK)
// 		if zErr != nil {
// 			fmt.Printf("Error updating Bandersnatch key root: %v\n", zErr)
// 			return
// 		}
// 		newSafroleState.Gamma.GammaZ = z
// 		return newSafroleState
// 	} else {
// 		// Same epoch
// 		return safroleState
// 	}
// }
// */
// // KeyRotate rotates the keys
// // Update the state with the new Safrole state
// // (6.13)
// func KeyRotate() {
// 	s := store.GetInstance()

// 	// Get prior state
// 	priorState := s.GetPriorStates()

// 	// Get previous time slot index
// 	tau := priorState.GetTau()

// 	// Get current time slot
// 	now := time.Now().UTC()
// 	timeInSecond := uint64(now.Sub(types.JamCommonEra).Seconds())
// 	tauPrime := types.TimeSlot(timeInSecond / uint64(types.SlotPeriod))

// 	// Execute key rotation
// 	//newSafroleState := keyRotation(tau, tauPrime, priorState.GetState())
// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	if ePrime > e {
// 		z, zErr := UpdateBandersnatchKeyRoot(priorState.GetGammaK())
// 		if zErr != nil {
// 			fmt.Printf("Error updating Bandersnatch key root: %v\n", zErr)
// 			return
// 		}

// 		// Update state to posterior state
// 		s.GetPosteriorStates().SetGammaK(ReplaceOffenderKeys(priorState.GetIota()))
// 		s.GetPosteriorStates().SetKappa(priorState.GetGammaK())
// 		s.GetPosteriorStates().SetLambda(priorState.GetKappa())
// 		s.GetPosteriorStates().SetGammaZ(z)
// 	} else {
// 		s.GetPosteriorStates().SetGammaK(priorState.GetGammaK())
// 		s.GetPosteriorStates().SetKappa(priorState.GetKappa())
// 		s.GetPosteriorStates().SetLambda(priorState.GetLambda())
// 		s.GetPosteriorStates().SetGammaZ(priorState.GetGammaZ())
// 	}
// }
// // --- safrole.go ---

// // --- sealing.go ---

// // TODO VERIFY 6.15 6.16
// func SealingByTickets() {
// 	/*
// 							  iy = Y(Hs)
// 		(6.15) γ′s ∈ ⟦C⟧ Hs ∈ F EU(H) Ha ⟨XT ⌢ η′3 ir⟩
// 	*/
// 	s := store.GetInstance()
// 	posterior_state := s.GetPosteriorStates()
// 	gammaSTickets := posterior_state.GetGammaS().Tickets
// 	if len(gammaSTickets) == 0 {
// 		return
// 	}
// 	header := s.GetIntermediateHeader()
// 	index := uint(header.Slot) % uint(len(posterior_state.GetGammaS().Tickets))
// 	ticket := gammaSTickets[index]
// 	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch
// 	i_r := ticket.Attempt
// 	message := utils.HeaderUSerialization(header)
// 	eta_prime := posterior_state.GetEta()

// 	var context types.ByteSequence
// 	context = append(context, types.ByteSequence(types.JamTicketSeal[:])...) // XT
// 	context = append(context, types.ByteSequence(eta_prime[3][:])...)        // η′3
// 	context = append(context, types.ByteSequence([]byte{uint8(i_r)})...)     // ir

// 	handler, _ := CreateVRFHandler(public_key)
// 	signature, _ := handler.IETFSign(context, message)

// 	s.GetIntermediateHeaders().SetSeal(types.BandersnatchVrfSignature(signature))
// }

// func SealingByBandersnatchs() {
// 	/*
// 		(6.16) γ′s ∈ ⟦HB⟧  Hs ∈ F EU(H) Ha ⟨XF ⌢ η′3⟩
// 	*/
// 	/*
// 		public key: Ha
// 		message: EU (H)
// 		context: XF ⌢ η′3
// 	*/
// 	s := store.GetInstance()
// 	posterior_state := s.GetPosteriorStates()
// 	GammaSKeys := posterior_state.GetGammaS().Keys
// 	if len(GammaSKeys) == 0 {
// 		return
// 	}
// 	header := s.GetIntermediateHeader()
// 	index := uint(header.Slot) % uint(len(GammaSKeys))
// 	public_key := GammaSKeys[index]
// 	message := utils.HeaderUSerialization(header)
// 	eta_prime := posterior_state.GetEta()

// 	var context types.ByteSequence
// 	context = append(context, types.ByteSequence(types.JamFallbackSeal[:])...) // XF
// 	context = append(context, types.ByteSequence(eta_prime[3][:])...)          // η′3

// 	handler, _ := CreateVRFHandler(public_key)
// 	signature, _ := handler.IETFSign(context, message)
// 	s.GetIntermediateHeaders().SetSeal(types.BandersnatchVrfSignature(signature))
// }

// func UpdateEtaPrime0() {
// 	// (6.22) η′0 ≡ H(η0 ⌢ Y(Hv))

// 	s := store.GetInstance()

// 	posterior_state := s.GetPosteriorStates()
// 	prior_state := s.GetPriorStates()
// 	header := s.GetIntermediateHeader()

// 	//public_key := posterior_state.Kappa[header.AuthorIndex].Bandersnatch
// 	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch
// 	entropy_source := header.EntropySource
// 	eta := prior_state.GetEta()
// 	handler, _ := CreateVRFHandler(public_key)
// 	vrfOutput, _ := handler.VRFIetfOutput(entropy_source[:])
// 	hash_input := append(eta[0][:], vrfOutput...)
// 	s.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hash_input)))
// }

// func UpdateEntropy() {
// 	/*
// 								(η0, η1, η2) if e′ > e
// 		(6.23) (η′1, η′2, η′3)
// 								(η1, η2, η3) otherwise
// 	*/

// 	s := store.GetInstance()

// 	prior_state := s.GetPriorStates()

// 	posterior_state := s.GetPosteriorStates()

// 	tau := prior_state.GetTau()

// 	tauPrime := posterior_state.GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)
// 	eta := prior_state.GetEta()
// 	if ePrime > e {
// 		for i := 2; i >= 0; i-- {
// 			eta[i+1] = eta[i]
// 		}
// 	}
// 	posterior_state.SetEta(eta)
// }

// func CalculateHeaderEntropy(public_key types.BandersnatchPublic, seal types.BandersnatchVrfSignature) (sign []byte) {
// 	/*
// 		F M K ⟨C⟩: The set of Bandersnatch signatures of the public key K, context C and message M. A subset of F.
// 		See section 3.8.
// 	*/
// 	/*
// 		(6.17) Hv ∈ F [] Ha ⟨XE ⌢ Y(Hs)⟩
// 	*/
// 	handler, _ := CreateVRFHandler(public_key)
// 	var message types.ByteSequence                                        // message: []
// 	var context types.ByteSequence                                        //context: XE ⌢ Y(Hs)
// 	context = append(context, types.ByteSequence(types.JamEntropy[:])...) // XE
// 	vrf, _ := handler.VRFIetfOutput(seal[:])
// 	context = append(context, types.ByteSequence(vrf)...) // Y(Hs)
// 	signature, _ := handler.IETFSign(context, message)    // F [] Ha ⟨XE ⌢ Y(Hs)⟩
// 	return signature
// }

// func UpdateHeaderEntropy() {
// 	s := store.GetInstance()

// 	// Get prior state
// 	posterior_state := s.GetPosteriorStates()

// 	header := s.GetIntermediateHeader()

// 	public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch // Ha
// 	seal := header.Seal                                                       // Hs
// 	s.GetIntermediateHeaders().SetEntropySource(types.BandersnatchVrfSignature(CalculateHeaderEntropy(public_key, seal)))
// }

// func UpdateSlotKeySequence() {
// 	/*
// 		Slot Key Sequence Update
// 						Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
// 		(6.24) γ′s ≡    γs if e′ = e
// 						F(η′2, κ′) otherwise
// 	*/
// 	// CalculateNewEntropy
// 	s := store.GetInstance()

// 	// Get prior state
// 	priorState := s.GetPriorStates()

// 	// Get posterior state
// 	posteriorState := s.GetPosteriorStates()

// 	// Get previous time slot index
// 	tau := priorState.GetTau()

// 	// Get current time slot index
// 	tauPrime := posteriorState.GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)
// 	eta_prime := posteriorState.GetEta()

// 	slot_index := GetSlotIndex(tau)
// 	var new_GammaS types.TicketsOrKeys
// 	if ePrime == e+1 {
// 		gammaA := priorState.GetGammaA()
// 		if len(priorState.GetGammaA()) == types.EpochLength && int(slot_index) >= types.SlotSubmissionEnd { // Z(γa) if e′ = e + 1 ∧ m ≥ Y ∧ ∣γa∣ = E
// 			new_GammaS.Tickets = OutsideInSequencer(&gammaA)
// 		} else { //F(η′2, κ′) otherwise
// 			new_GammaS.Keys = FallbackKeySequence(eta_prime[2], posteriorState.GetKappa())
// 		}
// 	}
// 	posteriorState.SetGammaS(new_GammaS)

// }

// // --- sealing.go ---

// // --- slot_key_sequence.go ---
// // OutsideInSequencer re-order the slice of ticketsBodies as in GP Eq. 6.25
// func OutsideInSequencer(t *types.TicketsAccumulator) types.TicketsAccumulator {
// 	left := 0
// 	right := types.EpochLength - 1

// 	out := make(types.TicketsAccumulator, types.EpochLength)

// 	for i := 0; i < types.EpochLength; i++ {
// 		if i%2 == 0 {
// 			out[i] = (*t)[left]
// 			left++
// 		} else {
// 			out[i] = (*t)[right]
// 			right--
// 		}
// 	}

// 	return out
// }

// // FallbackKeySequence implements the fallback key sequence in GP Eq. 6.26
// func FallbackKeySequence(entropy types.Entropy, validators types.ValidatorsData) []types.BandersnatchPublic {
// 	// require globa variable entropy
// 	// type EpochKeys []BandersnatchKey
// 	keys := make([]types.BandersnatchPublic, types.EpochLength)
// 	var i types.U32
// 	var epochLength types.U32 = types.U32(types.EpochLength)

// 	for i = 0; i < epochLength; i++ {
// 		// Get E_4(i)
// 		serial := utils.SerializeFixedLength(i, 4)
// 		// Concatenate  entropy with E_4(i)
// 		concatenation := append(entropy[:], serial...)
// 		// H4 : Keccak256(serializedBytes) -> See section 3.8 , take only the first 4 octets of the hash,
// 		hash := hash.Blake2bHashPartial(concatenation, 4)
// 		// E^(-1) deserialization
// 		validatorIndex, _ := utils.DeserializeFixedLength(types.ByteSequence(hash), types.U32(4))
// 		// validatorIndex : jamtypes.U64
// 		validatorIndex %= (types.U32(len(validators)))
// 		// k[]_b : validatorData -> bandersnatch
// 		keys[i] = validators[validatorIndex].Bandersnatch
// 	}

// 	return keys
// }

// // --- slot_key_sequence.go ---
// // --- markers.go ---

// // CreateEpochMarker creates the epoch marker
// // (6.27)
// func CreateEpochMarker() *types.ErrorCode {
// 	s := store.GetInstance()

// 	// Get previous time slot index
// 	tau := s.GetPriorStates().GetTau()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	// prior time slot must be less than posterior time slot
// 	if tau >= tauPrime {
// 		err := SafroleErrorCode.BadSlot
// 		return &err
// 	}

// 	if ePrime > e {
// 		// New epoch, create epoch marker
// 		// Get eta_0, eta_1
// 		eta := s.GetPriorStates().GetEta()

// 		// Get gamma_k from posterior state
// 		gammaK := s.GetPosteriorStates().GetGammaK()

// 		// Get bandersnatch key from gamma_k
// 		bandersnatchKeys := []types.BandersnatchPublic{}
// 		for _, validator := range gammaK {
// 			bandersnatchKeys = append(bandersnatchKeys, validator.Bandersnatch)
// 		}

// 		epochMarker := &types.EpochMark{
// 			Entropy:        eta[0],
// 			TicketsEntropy: eta[1],
// 			Validators:     bandersnatchKeys,
// 		}

// 		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
// 	} else {
// 		// The epoch is the same
// 		var epochMarker *types.EpochMark = nil
// 		s.GetIntermediateHeaderPointer().SetEpochMark(epochMarker)
// 	}

// 	return nil
// }

// // CreateWinningTickets creates the winning tickets
// // (6.28)
// func CreateWinningTickets() {
// 	s := store.GetInstance()

// 	// Get previous time slot index
// 	tau := s.GetPriorStates().GetTau()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	m := GetSlotIndex(tau)
// 	mPrime := GetSlotIndex(tauPrime)

// 	gammaA := s.GetPriorStates().GetGammaA()

// 	condition1 := ePrime == e
// 	condition2 := m < types.TimeSlot(types.SlotSubmissionEnd) && mPrime >= types.TimeSlot(types.SlotSubmissionEnd)
// 	condition3 := len(gammaA) == types.EpochLength

// 	if condition1 && condition2 && condition3 {
// 		// Z(gamma_a)
// 		ticketsMark := types.TicketsMark(OutsideInSequencer(&gammaA))
// 		s.GetIntermediateHeaderPointer().SetTicketsMark(&ticketsMark)
// 	} else {
// 		// The epoch is the same
// 		var ticketsMark *types.TicketsMark = nil
// 		s.GetIntermediateHeaderPointer().SetTicketsMark(ticketsMark)
// 	}
// }

// // --- markers.go ---

// // --- extrinsic_tickets.go ---

// // (6.30)
// // If the current time slot is in the epoch tail, we should not receive any
// // tickets.
// // Return error code: UnexpectedTicket
// func VerifyEpochTail(tickets types.TicketsExtrinsic) *types.ErrorCode {
// 	s := store.GetInstance()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	// m'
// 	mPrime := GetSlotIndex(tauPrime)

// 	// m' < Y => |E_T| <= K
// 	if mPrime < types.TimeSlot(types.SlotSubmissionEnd) {
// 		if len(tickets) > types.ValidatorsCount {
// 			err := SafroleErrorCode.UnexpectedTicket
// 			return &err
// 		}
// 	} else {
// 		if len(tickets) != 0 {
// 			err := SafroleErrorCode.UnexpectedTicket
// 			return &err
// 		}
// 	}

// 	return nil
// }

// // (6.31)
// // VerifyTicketsProof verifies the proof of the tickets
// // If the proof is valid, return the ticket bodies
// func VerifyTicketsProof(tickets types.TicketsExtrinsic) (types.TicketsAccumulator, *types.ErrorCode) {
// 	s := store.GetInstance()
// 	gammaK := s.GetPosteriorStates().GetGammaK()
// 	ring := []byte{}
// 	for _, validator := range gammaK {
// 		ring = append(ring, []byte(validator.Bandersnatch[:])...)
// 	}
// 	ringSize := uint(len(gammaK))

// 	verifier, err := vrf.NewVerifier(ring, ringSize)
// 	if err != nil {
// 		fmt.Printf("Failed to create verifier: %v\n", err)
// 	}
// 	defer verifier.Free()

// 	newTickets := types.TicketsAccumulator{}
// 	posteriorEta := s.GetPosteriorStates().GetEta()
// 	for _, ticket := range tickets {
// 		// print eta3 hex string
// 		context := createSignatureContext(types.JamTicketSeal, posteriorEta[2], ticket.Attempt)
// 		message := []byte{}
// 		signature := ticket.Signature[:]
// 		output, verifyErr := verifier.RingVerify(context, message, signature)

// 		if verifyErr != nil {
// 			err := SafroleErrorCode.BadTicketProof
// 			return nil, &err
// 		}

// 		// If the proof is valid, append the ticket body to the new tickets
// 		newTickets = append(newTickets, types.TicketBody{
// 			Id:      types.TicketId(output),
// 			Attempt: ticket.Attempt,
// 		})
// 	}

// 	return newTickets, nil
// }

// // (6.32)
// // Tickets must be sorted by ticket signature
// func VerifyTicketsOrder(tickets types.TicketsAccumulator) *types.ErrorCode {
// 	for i := 1; i < len(tickets); i++ {
// 		if bytes.Compare(tickets[i-1].Id[:], tickets[i].Id[:]) > 0 {
// 			err := SafroleErrorCode.BadTicketOrder
// 			return &err
// 		}
// 	}

// 	return nil
// }

// // (6.32) The extrinsic tickets must not contain any duplicate tickets
// // (6.33) The new ticket accumulator must not contain any duplicate tickets
// // (Validators should not submit the same ticket)
// func VerifyTicketsDuplicate(tickets types.TicketsAccumulator) *types.ErrorCode {
// 	for i := 1; i < len(tickets); i++ {
// 		if bytes.Equal(tickets[i-1].Id[:], tickets[i].Id[:]) {
// 			err := SafroleErrorCode.DuplicateTicket
// 			return &err
// 		}
// 	}

// 	return nil
// }

// // Tickets Attempt must be less than or equal to TicketsPerValidator
// func VerifyTicketsAttempt(tickets types.TicketsExtrinsic) *types.ErrorCode {
// 	for _, ticket := range tickets {
// 		// ticket.Attempt is an entry index (0-based)
// 		if ticket.Attempt >= types.TicketAttempt(types.TicketsPerValidator) {
// 			err := SafroleErrorCode.BadTicketAttempt
// 			return &err
// 		}
// 	}

// 	return nil
// }

// // createSignatureContext creates the context for the VRF signature
// func createSignatureContext(_X_T string, _posteriorEta2 types.Entropy, _r types.TicketAttempt) []byte {
// 	X_T := []byte(_X_T)
// 	posteriorEta2 := _posteriorEta2[:]
// 	r := []byte{byte(_r)}

// 	context := []byte{}
// 	context = append(context, X_T...)
// 	context = append(context, posteriorEta2...)
// 	context = append(context, r...)

// 	return context
// }

// // (6.32)
// // This function is not used in the current implementation, becasue we throw an
// // error if we find a duplicate ticket in the new ticket (from extrinsic
// // tickets).
// func RemoveAndSortDuplicateTickets(tickets types.TicketsAccumulator) types.TicketsAccumulator {
// 	if len(tickets) == 0 {
// 		return tickets
// 	}

// 	sort.Slice(tickets, func(i, j int) bool {
// 		return bytes.Compare(tickets[i].Id[:], tickets[j].Id[:]) < 0
// 	})

// 	j := 0
// 	for i := 1; i < len(tickets); i++ {
// 		if !bytes.Equal(tickets[i].Id[:], tickets[j].Id[:]) {
// 			j++
// 			tickets[j] = tickets[i]
// 		}
// 	}

// 	return tickets[:j+1]
// }

// func Contains(tickets types.TicketsAccumulator, ticketId types.TicketId) bool {
// 	for _, ticket := range tickets {
// 		if bytes.Equal(ticket.Id[:], ticketId[:]) {
// 			return true
// 		}
// 	}

// 	return false
// }

// // (6.33)
// // This function is not used in the current implementation, becasue we throw an
// // error if we find a duplicate ticket in the new ticket accumulator.
// func RemoveTicketsInGammaA(tickets, gammaA types.TicketsAccumulator) types.TicketsAccumulator {
// 	result := types.TicketsAccumulator{}
// 	for _, ticket := range tickets {
// 		if !Contains(gammaA, ticket.Id) {
// 			result = append(result, ticket)
// 		}
// 	}

// 	return result
// }

// // (6.34)
// func GetPreviousTicketsAccumulator() types.TicketsAccumulator {
// 	s := store.GetInstance()

// 	// Get previous time slot index
// 	tau := s.GetPriorStates().GetTau()

// 	// Get current time slot index
// 	tauPrime := s.GetPosteriorStates().GetTau()

// 	e := GetEpochIndex(tau)
// 	ePrime := GetEpochIndex(tauPrime)

// 	if ePrime > e {
// 		return types.TicketsAccumulator{}
// 	} else {
// 		gammaA := s.GetPriorStates().GetGammaA()
// 		return gammaA
// 	}
// }

// // (6.34)
// // create gamma_a'(New ticket accumulator)
// func CreateNewTicketAccumulator() *types.ErrorCode {
// 	// 1. Verify the epoch tail
// 	// 2. Verify the attempt of the tickets
// 	// 3. Verify the tickets proof (return the new tickets)
// 	// 4. Verify the new tickets order
// 	// 5. Verify the new tickets duplicate
// 	// 6. Get the previous ticket accumulator
// 	// 7. Concatenate the new tickets and the previous ticket accumulator
// 	// 8. Sort the tickets by ticket id
// 	// 9. Select E tickets from the sorted tickets for the new ticket accumulator
// 	// 10. Set the new ticket accumulator to the posterior state

// 	// Get extrinsic tickets
// 	s := store.GetInstance()
// 	extrinsicTickets := s.GetProcessingBlockPointer().GetTicketsExtrinsic()

// 	// (6.30) Verify the epoch tail
// 	err := VerifyEpochTail(extrinsicTickets)
// 	if err != nil {
// 		return err
// 	}

// 	// Verify the attempt of the tickets
// 	err = VerifyTicketsAttempt(extrinsicTickets)
// 	if err != nil {
// 		// Extrinsic tickets attempt is invalid
// 		return err
// 	}

// 	// (6.31) Verify the tickets proof
// 	newTickets, err := VerifyTicketsProof(extrinsicTickets)
// 	if err != nil {
// 		// Extrinsic tickets proof is invalid
// 		return err
// 	}

// 	// (6.32) Verify the new tickets order
// 	err = VerifyTicketsOrder(newTickets)
// 	if err != nil {
// 		// Extrinsic tickets order is invalid
// 		return err
// 	}

// 	// (6.32) Verify the new tickets duplicate
// 	err = VerifyTicketsDuplicate(newTickets)
// 	if err != nil {
// 		// Extrinsic tickets duplicate is invalid
// 		return err
// 	}

// 	// (6.34) Get previous ticket accumulator
// 	previousTicketsAccumulator := GetPreviousTicketsAccumulator()

// 	// (6.34) Concatenate the new tickets and the previous ticket accumulator
// 	newTicketsAccumulator := append(newTickets, previousTicketsAccumulator...)

// 	// (6.34) sort the tickets by ticket id
// 	// We already verified the duplicate tickets, so the newTicketsAccumulator
// 	// should not contain any duplicate tickets
// 	sort.Slice(newTicketsAccumulator, func(i, j int) bool {
// 		return bytes.Compare(newTicketsAccumulator[i].Id[:], newTicketsAccumulator[j].Id[:]) < 0
// 	})

// 	// (6.33) Verify the new tickets accmuulator
// 	err = VerifyTicketsDuplicate(newTicketsAccumulator)
// 	if err != nil {
// 		// Found a ticket duplicate (Someone submitted the same ticket)
// 		return err
// 	}

// 	// (6.34) select E tickets from the sorted tickets for the new ticket accumulator
// 	maxTicketsAccumulatorSize := types.EpochLength
// 	if len(newTicketsAccumulator) > maxTicketsAccumulatorSize {
// 		newTicketsAccumulator = newTicketsAccumulator[:maxTicketsAccumulatorSize]
// 	}

// 	// (6.34) set the new ticket accumulator to the posterior state
// 	s.GetPosteriorStates().SetGammaA(newTicketsAccumulator)

// 	return nil
// }

// // --- extrinsic_tickets.go ---
