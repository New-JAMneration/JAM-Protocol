package jamtests

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func (e *SafroleErrorCode) Error() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *e)
}

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
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}

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

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// SafroleInput
func (i *SafroleInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleInput")
	var err error

	if err = i.Slot.Encode(e); err != nil {
		return err
	}

	if err = i.Entropy.Encode(e); err != nil {
		return err
	}

	if err = i.Extrinsic.Encode(e); err != nil {
		return err
	}

	return nil
}

// SafroleOutputData
func (o *SafroleOutputData) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleOutputData")

	if o.EpochMark == nil {
		if err := e.WriteByte(0); err != nil {
			return err
		}
	} else {
		if err := e.WriteByte(1); err != nil {
			return err
		}

		if err := o.EpochMark.Encode(e); err != nil {
			return err
		}
	}

	if o.TicketsMark == nil {
		if err := e.WriteByte(0); err != nil {
			return err
		}
	} else {
		if err := e.WriteByte(1); err != nil {
			return err
		}
		if err := o.TicketsMark.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// SafroleErrorCode
func (e *SafroleErrorCode) Encode(enc *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleErrorCode")

	return nil
}

// SafroleOutput
func (o *SafroleOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleOutput")
	// Write a pointer flag
	// Ok = 0, Err = 1
	if o.Ok != nil {
		cLog(Yellow, "SafroleOutput is ok")
		if err := e.WriteByte(0); err != nil {
			return err
		}

		// Encode SafroleOutputData
		if err := o.Ok.Encode(e); err != nil {
			return err
		}

		return nil
	}

	if o.Err != nil {
		cLog(Yellow, "SafroleOutput is err")
		if err := e.WriteByte(1); err != nil {
			return err
		}

		// Encode SafroleErrorCode
		if err := e.WriteByte(byte(*o.Err)); err != nil {
			return err
		}

		cLog(Yellow, fmt.Sprintf("SafroleErrorCode: %v", *o.Err))
	}

	return nil
}

// SafroleState
func (s *SafroleState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleState")
	var err error

	if err = s.Tau.Encode(e); err != nil {
		return err
	}

	if err = s.Eta.Encode(e); err != nil {
		return err
	}

	if err = s.Lambda.Encode(e); err != nil {
		return err
	}

	if err = s.Kappa.Encode(e); err != nil {
		return err
	}

	if err = s.GammaK.Encode(e); err != nil {
		return err
	}

	if err = s.Iota.Encode(e); err != nil {
		return err
	}

	if err = s.GammaA.Encode(e); err != nil {
		return err
	}

	if err = s.GammaS.Encode(e); err != nil {
		return err
	}

	if err = s.GammaZ.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(s.PostOffenders))); err != nil {
		return err
	}

	for i := range s.PostOffenders {
		if err = s.PostOffenders[i].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// SafroleTestCase
func (t *SafroleTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding SafroleTestCase")
	var err error

	if err = t.Input.Encode(e); err != nil {
		return err
	}

	if err = t.PreState.Encode(e); err != nil {
		return err
	}

	if err = t.Output.Encode(e); err != nil {
		return err
	}

	if err = t.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

func (s *SafroleTestCase) Dump() error {
	store.ResetInstance()
	storeInstance := store.GetInstance()

	storeInstance.GetPriorStates().SetTau(s.PreState.Tau)
	storeInstance.GetProcessingBlockPointer().SetSlot(s.Input.Slot)
	storeInstance.GetPosteriorStates().SetTau(s.Input.Slot)

	storeInstance.GetPriorStates().SetEta(s.PreState.Eta)
	// Set eta^prime_0 here
	hash_input := append(s.PreState.Eta[0][:], s.Input.Entropy[:]...)
	storeInstance.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hash_input)))

	storeInstance.GetPriorStates().SetLambda(s.PreState.Lambda)
	storeInstance.GetPriorStates().SetKappa(s.PreState.Kappa)
	storeInstance.GetPriorStates().SetGammaK(s.PreState.GammaK)
	storeInstance.GetPriorStates().SetIota(s.PreState.Iota)
	storeInstance.GetPriorStates().SetGammaA(s.PreState.GammaA)
	storeInstance.GetPriorStates().SetGammaS(s.PreState.GammaS)
	storeInstance.GetPriorStates().SetGammaZ(s.PreState.GammaZ)

	storeInstance.GetPosteriorStates().SetPsiO(s.PreState.PostOffenders)

	// Add block with TicketsExtrinsic
	block := types.Block{
		Extrinsic: types.Extrinsic{
			Tickets: s.Input.Extrinsic,
		},
	}
	storeInstance.AddBlock(block)

	return nil
}

func (s *SafroleTestCase) GetPostState() interface{} {
	return s.PostState
}

func (s *SafroleTestCase) GetOutput() interface{} {
	return s.Output
}

func (s *SafroleTestCase) ExpectError() error {
	if s.Output.Err == nil {
		return nil
	}
	return s.Output.Err
}

func (s *SafroleTestCase) Validate() error {
	storeInstance := store.GetInstance()
	// Set eta^prime_0 here
	hash_input := append(s.PreState.Eta[0][:], s.Input.Entropy[:]...)
	storeInstance.GetPosteriorStates().SetEta0(types.Entropy(hash.Blake2bHash(hash_input)))
	/*
		Check EpochMark and TicketsMark
	*/
	ourEpochMarker := storeInstance.GetProcessingBlockPointer().GetEpochMark()
	ourTicketsMark := storeInstance.GetProcessingBlockPointer().GetTicketsMark()

	if s.Output.Ok.EpochMark != nil && ourEpochMarker != nil {
		if !reflect.DeepEqual(s.Output.Ok.EpochMark, ourEpochMarker) {
			diff := cmp.Diff(s.Output.Ok.EpochMark, ourEpochMarker)
			return fmt.Errorf("epoch marker mismatch:\n%v", diff)
		}
	} else if s.Output.Ok.TicketsMark != nil && ourTicketsMark != nil {
		if !reflect.DeepEqual(s.Output.Ok.TicketsMark, ourTicketsMark) {
			diff := cmp.Diff(s.Output.Ok.TicketsMark, ourTicketsMark)
			return fmt.Errorf("tickets mark mismatch:\n%v", diff)
		}
	}

	/*
		Check PosteriorStates
	*/
	if !reflect.DeepEqual(s.PostState.Tau, storeInstance.GetPosteriorStates().GetTau()) {
		diff := cmp.Diff(s.PostState.Tau, storeInstance.GetPosteriorStates().GetTau())
		return fmt.Errorf("tau mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.Eta, storeInstance.GetPosteriorStates().GetEta()) {
		diff := cmp.Diff(s.PostState.Eta, storeInstance.GetPosteriorStates().GetEta())
		return fmt.Errorf("eta mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.Lambda, storeInstance.GetPosteriorStates().GetLambda()) {
		diff := cmp.Diff(s.PostState.Lambda, storeInstance.GetPosteriorStates().GetLambda())
		return fmt.Errorf("lambda mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.Kappa, storeInstance.GetPosteriorStates().GetKappa()) {
		diff := cmp.Diff(s.PostState.Kappa, storeInstance.GetPosteriorStates().GetKappa())
		return fmt.Errorf("kappa mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.GammaK, storeInstance.GetPosteriorStates().GetGammaK()) {
		diff := cmp.Diff(s.PostState.GammaK, storeInstance.GetPosteriorStates().GetGammaK())
		return fmt.Errorf("gamma_k mismatch:\n%v", diff)
		/*
			We don't compare iota here, this state will be update in accumulation
		*/
		// } else if !reflect.DeepEqual(expectedState.Iota, storeInstance.GetPosteriorStates().GetIota()) {
		// diff := cmp.Diff(expectedState.Iota, storeInstance.GetPosteriorStates().GetIota())
		// t.Errorf("iota mismatch:\n%v", diff)
	} else if !cmp.Equal(s.PostState.GammaA, storeInstance.GetPosteriorStates().GetGammaA(), cmpopts.EquateEmpty()) {
		diff := cmp.Diff(s.PostState.GammaA, storeInstance.GetPosteriorStates().GetGammaA(), cmpopts.EquateEmpty())
		return fmt.Errorf("gamma_a mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.GammaS, storeInstance.GetPosteriorStates().GetGammaS()) {
		diff := cmp.Diff(s.PostState.GammaS, storeInstance.GetPosteriorStates().GetGammaS())
		return fmt.Errorf("gamma_s mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.GammaZ, storeInstance.GetPosteriorStates().GetGammaZ()) {
		diff := cmp.Diff(s.PostState.GammaZ, storeInstance.GetPosteriorStates().GetGammaZ())
		return fmt.Errorf("gamma_z mismatch:\n%v", diff)
	} else if !reflect.DeepEqual(s.PostState.PostOffenders, storeInstance.GetPosteriorStates().GetPsiO()) {
		diff := cmp.Diff(s.PostState.PostOffenders, storeInstance.GetPosteriorStates().GetPsiO())
		return fmt.Errorf("post_offenders mismatch:\n%v", diff)
	}

	return nil
}
