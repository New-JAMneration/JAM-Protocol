package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

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

// unmarshal safrole input
func (i *SafroleInput) UnmarshalJSON(data []byte) error {
	var err error

	var input struct {
		Slot      types.TimeSlot         `json:"slot"`
		Entropy   types.Entropy          `json:"entropy"`
		Extrinsic types.TicketsExtrinsic `json:"extrinsic"`
	}

	if err = json.Unmarshal(data, &input); err != nil {
		return err
	}

	i.Slot = input.Slot
	i.Entropy = input.Entropy

	if len(input.Extrinsic) == 0 {
		return nil
	}

	i.Extrinsic = input.Extrinsic

	return nil
}

// unmarshal safrole state
func (s *SafroleState) UnmarshalJSON(data []byte) error {
	var err error

	var state struct {
		Tau           types.TimeSlot                   `json:"tau"`
		Eta           types.EntropyBuffer              `json:"eta"`
		Lambda        types.ValidatorsData             `json:"lambda"`
		Kappa         types.ValidatorsData             `json:"kappa"`
		GammaK        types.ValidatorsData             `json:"gamma_k"`
		Iota          types.ValidatorsData             `json:"iota"`
		GammaA        types.TicketsAccumulator         `json:"gamma_a"`
		GammaS        types.TicketsOrKeys              `json:"gamma_s"`
		GammaZ        types.BandersnatchRingCommitment `json:"gamma_z"`
		PostOffenders []types.Ed25519Public            `json:"post_offenders"`
	}

	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}

	s.Tau = state.Tau
	s.Eta = state.Eta
	s.Lambda = state.Lambda
	s.Kappa = state.Kappa
	s.GammaK = state.GammaK
	s.Iota = state.Iota
	s.GammaA = state.GammaA
	s.GammaS = state.GammaS
	s.GammaZ = state.GammaZ

	if len(state.PostOffenders) == 0 {
		return nil
	}

	s.PostOffenders = state.PostOffenders

	return nil
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

// SafroleTestCase
func (t *SafroleTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = t.Input.Decode(d); err != nil {
		return err
	}

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Output.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// SafroleInput
func (i *SafroleInput) Decode(d *types.Decoder) error {
	var err error

	if err = i.Slot.Decode(d); err != nil {
		return err
	}

	if err = i.Entropy.Decode(d); err != nil {
		return err
	}

	if err = i.Extrinsic.Decode(d); err != nil {
		return err
	}

	return nil
}

// SafroleOutputData
func (o *SafroleOutputData) Decode(d *types.Decoder) error {
	var err error

	epochMarkPointerFlag, err := d.ReadPointerFlag()
	epochMarkPointerIsNil := epochMarkPointerFlag == 0
	if epochMarkPointerIsNil {
		cLog(Yellow, "EpochMark is nil")
	} else {
		cLog(Yellow, "..EpochMark is not nil")
		if o.EpochMark == nil {
			o.EpochMark = &types.EpochMark{}
		}

		if err = o.EpochMark.Decode(d); err != nil {
			return err
		}
	}

	ticketsMarkPointerFlag, err := d.ReadPointerFlag()
	ticketsMarkPointerIsNil := ticketsMarkPointerFlag == 0
	if ticketsMarkPointerIsNil {
		cLog(Yellow, "TicketsMark is nil")
	} else {
		cLog(Yellow, "TicketsMark is not nil")
		if o.TicketsMark == nil {
			o.TicketsMark = &types.TicketsMark{}
		}

		if err = o.TicketsMark.Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// SafroleOutput
func (o *SafroleOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding SafroleOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "SafroleOutput is ok")

		if o.Ok == nil {
			o.Ok = &SafroleOutputData{}
		}

		if err = o.Ok.Decode(d); err != nil {
			return err
		}
		return nil
	} else {
		cLog(Yellow, "SafroleOutput is err")
		cLog(Yellow, "Decoding SafroleErrorCode")

		// Read a byte as error code
		errByte, err := d.ReadErrorByte()
		if err != nil {
			return err
		}

		o.Err = (*SafroleErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("SafroleErrorCode: %v", *o.Err))
	}

	return nil
}

// SafroleState
func (s *SafroleState) Decode(d *types.Decoder) error {
	var err error

	if err = s.Tau.Decode(d); err != nil {
		return err
	}

	if err = s.Eta.Decode(d); err != nil {
		return err
	}

	if err = s.Lambda.Decode(d); err != nil {
		return err
	}

	if err = s.Kappa.Decode(d); err != nil {
		return err
	}

	if err = s.GammaK.Decode(d); err != nil {
		return err
	}

	if err = s.Iota.Decode(d); err != nil {
		return err
	}

	if err = s.GammaA.Decode(d); err != nil {
		return err
	}

	if err = s.GammaS.Decode(d); err != nil {
		return err
	}

	if err = s.GammaZ.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	s.PostOffenders = make([]types.Ed25519Public, length)
	for i := range s.PostOffenders {
		if err = s.PostOffenders[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}
