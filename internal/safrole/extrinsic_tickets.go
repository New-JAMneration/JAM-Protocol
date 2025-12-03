// 6.7. The Extrinsic and Tickets (graypaper 0.5.4)
package safrole

import (
	"bytes"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	SafroleErrorCode "github.com/New-JAMneration/JAM-Protocol/internal/types/error_codes/safrole"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

// (6.30)
// If the current time slot is in the epoch tail, we should not receive any
// tickets.
// Return error code: UnexpectedTicket
func VerifyEpochTail(tickets types.TicketsExtrinsic) *types.ErrorCode {
	s := store.GetInstance()

	// Get current time slot index
	tauPrime := s.GetPosteriorStates().GetTau()

	// m'
	mPrime := GetSlotIndex(tauPrime)

	// m' < Y => |E_T| <= K
	if mPrime < types.TimeSlot(types.SlotSubmissionEnd) {
		if len(tickets) > types.ValidatorsCount {
			err := SafroleErrorCode.UnexpectedTicket
			return &err
		}
	} else {
		if len(tickets) != 0 {
			err := SafroleErrorCode.UnexpectedTicket
			return &err
		}
	}

	return nil
}

// (6.31)
// VerifyTicketsProof verifies the proof of the tickets
// If the proof is valid, return the ticket bodies
func VerifyTicketsProof(ringVerifier *vrf.Verifier, tickets types.TicketsExtrinsic) (types.TicketsAccumulator, *types.ErrorCode) {
	s := store.GetInstance()

	newTickets := types.TicketsAccumulator{}
	posteriorEta := s.GetPosteriorStates().GetEta()
	for _, ticket := range tickets {
		// print eta3 hex string
		context := createSignatureContext(types.JamTicketSeal, posteriorEta[2], ticket.Attempt)
		message := []byte{}
		signature := ticket.Signature[:]
		output, verifyErr := ringVerifier.RingVerify(context, message, signature)
		if verifyErr != nil {
			err := SafroleErrorCode.BadTicketProof
			return nil, &err
		}

		// If the proof is valid, append the ticket body to the new tickets
		newTickets = append(newTickets, types.TicketBody{
			Id:      types.TicketId(output),
			Attempt: ticket.Attempt,
		})
	}

	return newTickets, nil
}

// (6.32)
// Tickets must be sorted by ticket signature
func VerifyTicketsOrder(tickets types.TicketsAccumulator) *types.ErrorCode {
	for i := 1; i < len(tickets); i++ {
		if bytes.Compare(tickets[i-1].Id[:], tickets[i].Id[:]) > 0 {
			err := SafroleErrorCode.BadTicketOrder
			return &err
		}
	}

	return nil
}

// (6.32) The extrinsic tickets must not contain any duplicate tickets
// (6.33) The new ticket accumulator must not contain any duplicate tickets
// (Validators should not submit the same ticket)
func VerifyTicketsDuplicate(tickets types.TicketsAccumulator) *types.ErrorCode {
	for i := 1; i < len(tickets); i++ {
		if bytes.Equal(tickets[i-1].Id[:], tickets[i].Id[:]) {
			err := SafroleErrorCode.DuplicateTicket
			return &err
		}
	}

	return nil
}

// Tickets Attempt must be less than or equal to TicketsPerValidator
func VerifyTicketsAttempt(tickets types.TicketsExtrinsic) *types.ErrorCode {
	for _, ticket := range tickets {
		// ticket.Attempt is an entry index (0-based)
		if ticket.Attempt >= types.TicketAttempt(types.TicketsPerValidator) {
			err := SafroleErrorCode.BadTicketAttempt
			return &err
		}
	}

	return nil
}

// createSignatureContext creates the context for the VRF signature
func createSignatureContext(_X_T string, _posteriorEta2 types.Entropy, _r types.TicketAttempt) []byte {
	X_T := []byte(_X_T)
	posteriorEta2 := _posteriorEta2[:]
	r := []byte{byte(_r)}

	context := []byte{}
	context = append(context, X_T...)
	context = append(context, posteriorEta2...)
	context = append(context, r...)

	return context
}

// (6.32)
// This function is not used in the current implementation, becasue we throw an
// error if we find a duplicate ticket in the new ticket (from extrinsic
// tickets).
func RemoveAndSortDuplicateTickets(tickets types.TicketsAccumulator) types.TicketsAccumulator {
	if len(tickets) == 0 {
		return tickets
	}

	sort.Slice(tickets, func(i, j int) bool {
		return bytes.Compare(tickets[i].Id[:], tickets[j].Id[:]) < 0
	})

	j := 0
	for i := 1; i < len(tickets); i++ {
		if !bytes.Equal(tickets[i].Id[:], tickets[j].Id[:]) {
			j++
			tickets[j] = tickets[i]
		}
	}

	return tickets[:j+1]
}

func Contains(tickets types.TicketsAccumulator, ticketId types.TicketId) bool {
	for _, ticket := range tickets {
		if bytes.Equal(ticket.Id[:], ticketId[:]) {
			return true
		}
	}

	return false
}

// (6.33)
// This function is not used in the current implementation, becasue we throw an
// error if we find a duplicate ticket in the new ticket accumulator.
func RemoveTicketsInGammaA(tickets, gammaA types.TicketsAccumulator) types.TicketsAccumulator {
	result := types.TicketsAccumulator{}
	for _, ticket := range tickets {
		if !Contains(gammaA, ticket.Id) {
			result = append(result, ticket)
		}
	}

	return result
}

// (6.34)
func GetPreviousTicketsAccumulator() types.TicketsAccumulator {
	s := store.GetInstance()

	// Get previous time slot index
	tau := s.GetPriorStates().GetTau()

	// Get current time slot index
	tauPrime := s.GetPosteriorStates().GetTau()

	e := GetEpochIndex(tau)
	ePrime := GetEpochIndex(tauPrime)

	if ePrime > e {
		return types.TicketsAccumulator{}
	} else {
		gammaA := s.GetPriorStates().GetGammaA()
		return gammaA
	}
}

// (6.34)
// create gamma_a'(New ticket accumulator)
func CreateNewTicketAccumulator(ringVerifier *vrf.Verifier) *types.ErrorCode {
	// 1. Verify the epoch tail
	// 2. Verify the attempt of the tickets
	// 3. Verify the tickets proof (return the new tickets)
	// 4. Verify the new tickets order
	// 5. Verify the new tickets duplicate
	// 6. Get the previous ticket accumulator
	// 7. Concatenate the new tickets and the previous ticket accumulator
	// 8. Sort the tickets by ticket id
	// 9. Select E tickets from the sorted tickets for the new ticket accumulator
	// 10. Set the new ticket accumulator to the posterior state

	// Get extrinsic tickets
	s := store.GetInstance()
	extrinsicTickets := s.GetLatestBlock().Extrinsic.Tickets

	// (6.30) Verify the epoch tail
	err := VerifyEpochTail(extrinsicTickets)
	if err != nil {
		return err
	}

	// Verify the attempt of the tickets
	err = VerifyTicketsAttempt(extrinsicTickets)
	if err != nil {
		// Extrinsic tickets attempt is invalid
		return err
	}

	// (6.31) Verify the tickets proof
	newTickets, err := VerifyTicketsProof(ringVerifier, extrinsicTickets)
	if err != nil {
		// Extrinsic tickets proof is invalid
		return err
	}

	// (6.32) Verify the new tickets order
	err = VerifyTicketsOrder(newTickets)
	if err != nil {
		// Extrinsic tickets order is invalid
		return err
	}

	// (6.32) Verify the new tickets duplicate
	err = VerifyTicketsDuplicate(newTickets)
	if err != nil {
		// Extrinsic tickets duplicate is invalid
		return err
	}

	// (6.34) Get previous ticket accumulator
	previousTicketsAccumulator := GetPreviousTicketsAccumulator()

	// (6.34) Concatenate the new tickets and the previous ticket accumulator
	newTicketsAccumulator := append(newTickets, previousTicketsAccumulator...)

	// (6.34) sort the tickets by ticket id
	// We already verified the duplicate tickets, so the newTicketsAccumulator
	// should not contain any duplicate tickets
	sort.Slice(newTicketsAccumulator, func(i, j int) bool {
		return bytes.Compare(newTicketsAccumulator[i].Id[:], newTicketsAccumulator[j].Id[:]) < 0
	})

	// (6.33) Verify the new tickets accmuulator
	err = VerifyTicketsDuplicate(newTicketsAccumulator)
	if err != nil {
		// Found a ticket duplicate (Someone submitted the same ticket)
		return err
	}

	// (6.34) select E tickets from the sorted tickets for the new ticket accumulator
	maxTicketsAccumulatorSize := types.EpochLength
	if len(newTicketsAccumulator) > maxTicketsAccumulatorSize {
		newTicketsAccumulator = newTicketsAccumulator[:maxTicketsAccumulatorSize]
	}

	// (6.34) set the new ticket accumulator to the posterior state
	s.GetPosteriorStates().SetGammaA(newTicketsAccumulator)

	return nil
}
