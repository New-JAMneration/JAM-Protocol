// 6.7. The Extrinsic and Tickets (graypaper 0.5.4)
package safrole

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	vrf "github.com/New-JAMneration/JAM-Protocol/pkg/Rust-VRF/vrf-func-ffi/src"
)

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
func removeAndSortDuplicateTickets(tickets types.TicketsAccumulator) types.TicketsAccumulator {
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

func contains(tickets types.TicketsAccumulator, ticketId types.TicketId) bool {
	for _, ticket := range tickets {
		if bytes.Equal(ticket.Id[:], ticketId[:]) {
			return true
		}
	}

	return false
}

// (6.33)
func removeTicketsInGammaA(tickets, gammaA types.TicketsAccumulator) types.TicketsAccumulator {
	result := types.TicketsAccumulator{}
	for _, ticket := range tickets {
		if !contains(gammaA, ticket.Id) {
			result = append(result, ticket)
		}
	}

	return result
}

// (6.31) (6.32) (6.33)
// n: the set of new tickets
func createNewTickets(extrinsicTickets types.TicketsExtrinsic) types.TicketsAccumulator {
	s := store.GetInstance()
	posteriorState := s.GetPosteriorState()
	gammaK := posteriorState.Gamma.GammaK
	ring := []byte{}
	for _, validator := range gammaK {
		ring = append(ring, []byte(validator.Bandersnatch[:])...)
	}

	skBytes := []byte{}
	ringSize := uint(len(gammaK))
	proverIdx := uint(0)
	ringVRFHandler, err := vrf.NewHandler(ring, skBytes, ringSize, proverIdx)
	if err != nil {
		fmt.Printf("Failed to create RingVRF Handler: %v\n", err)
	}

	n := types.TicketsAccumulator{}
	for _, ticket := range extrinsicTickets {
		context := createSignatureContext(types.JamTicketSeal, posteriorState.Eta[2], ticket.Attempt)
		message := []byte{}
		signature := ticket.Signature[:]
		output, err := ringVRFHandler.RingVerify(context, message, signature)
		if err != nil {
			fmt.Printf("Failed to verify signature: %v\n", err)
		}

		ticket := types.TicketBody{
			Id:      types.TicketId(output), // y
			Attempt: ticket.Attempt,         // r
		}

		n = append(n, ticket)
	}

	// (6.32)
	n = removeAndSortDuplicateTickets(n)

	// (6.33)
	gammaA := posteriorState.Gamma.GammaA
	n = removeTicketsInGammaA(n, gammaA)

	return n
}

// (6.34)
// gamma_a: Previous ticket accumulator
func getExistentTickets() types.TicketsAccumulator {
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

	if ePrime > e {
		return types.TicketsAccumulator{}
	} else {
		gammaA := priorState.Gamma.GammaA
		return gammaA
	}
}

// (6.34)
// gamma_a': New ticket accumulator
func CreateNewTicketAccumulator(extrinsicTickets types.TicketsExtrinsic) {
	n := createNewTickets(extrinsicTickets)
	existentTickets := getExistentTickets()

	tickets := append(existentTickets, n...)

	// sort the tickets by ticket id
	sort.Slice(tickets, func(i, j int) bool {
		return bytes.Compare(tickets[i].Id[:], tickets[j].Id[:]) < 0
	})

	// select E tickets from the sorted tickets for the new ticket accumulator
	newTicketAccumulator := tickets[:types.EpochLength]

	// set the new ticket accumulator to the posterior state
	s := store.GetInstance()
	s.GetPosteriorStates().SetGammaA(newTicketAccumulator)
}
