package jamtests

import (
	"encoding/json"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type SafroleTestCase struct {
	Input     SafroleInput  `json:"input"`
	PreState  SafroleState  `json:"pre_state"`
	Output    SafroleOutput `json:"output"`
	PostState SafroleState  `json:"post_state"`
}

type SafroleInput struct {
	Slot      types.TimeSlot         `json:"slot"`      // Current slot
	Entropy   types.Entropy          `json:"entropy"`   // Per block entropy (originated from block entropy source VRF)
	Extrinsic types.TicketsExtrinsic `json:"extrinsic"` // Safrole extrinsic
}

type SafroleOutputData struct {
	EpochMark   *types.EpochMark   `json:"epoch_mark,omitempty"`   // New epoch marker (optional).
	TicketsMark *types.TicketsMark `json:"tickets_mark,omitempty"` // Winning tickets marker (optional).
}

type SafroleErrorCode types.ErrorCode

type SafroleOutput struct {
	Ok  *SafroleOutputData `json:"ok,omitempty"`
	Err *SafroleErrorCode  `json:"err,omitempty"`
}

type SafroleState struct {
	Tau           types.TimeSlot                   `json:"tau"`            // Most recent block's timeslot
	Eta           types.EntropyBuffer              `json:"eta"`            // Entropy accumulator and epochal randomness
	Lambda        types.ValidatorsData             `json:"lambda"`         // Validator keys and metadata which were active in the prior epoch
	Kappa         types.ValidatorsData             `json:"kappa"`          // Validator keys and metadata currently active
	GammaK        types.ValidatorsData             `json:"gamma_k"`        // Validator keys for the following epoch
	Iota          types.ValidatorsData             `json:"iota"`           // Validator keys and metadata to be drawn from next
	GammaA        types.TicketsAccumulator         `json:"gamma_a"`        // Sealing-key contest ticket accumulator
	GammaS        types.TicketsOrKeys              `json:"gamma_s"`        // Sealing-key series of the current epoch
	GammaZ        types.BandersnatchRingCommitment `json:"gamma_z"`        // Bandersnatch ring commitment
	PostOffenders []types.Ed25519Public            `json:"post_offenders"` // Posterior offenders sequence
}

const (
	BadSlot          SafroleErrorCode = iota // 0 Timeslot value must be strictly monotonic
	UnexpectedTicket                         // 1 Received a ticket while in epoch's tail
	BadTicketOrder                           // 2 Tickets must be sorted
	BadTicketProof                           // 3 Invalid ticket ring proof
	BadTicketAttempt                         // 4 Invalid ticket attempt value
	Reserved                                 // 5 Reserved
	DuplicateTicket                          // 6 Found a ticket duplicate
)

var safroleErrorMap = map[string]SafroleErrorCode{
	"bad_slot":           BadSlot,
	"unexpected_ticket":  UnexpectedTicket,
	"bad_ticket_order":   BadTicketOrder,
	"bad_ticket_proof":   BadTicketProof,
	"bad_ticket_attempt": BadTicketAttempt,
	"reserved":           Reserved,
	"duplicate_ticket":   DuplicateTicket,
}

func (e *SafroleErrorCode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := safroleErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}
