package safrole

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// basic types
type U8 uint8

type U32 uint32

// fixed-length byte arrays
type ByteArray32 [32]U8

type OpaqueHash ByteArray32
type Ed25519Key ByteArray32
type BlsKey [144]U8
type BandersnatchKey ByteArray32

type EpochKeys []BandersnatchKey

func (e EpochKeys) Validate() error {
	if len(e) != jam_types.EpochLength {
		return fmt.Errorf("EpochKeys must have exactly %d entries, got %d", jam_types.EpochLength, len(e))
	}
	return nil
}

type TicketsBodies []TicketBody

func (t TicketsBodies) Validate() error {
	if len(t) != jam_types.EpochLength {
		return fmt.Errorf("TicketsBodies must have exactly %d entries, got %d", jam_types.EpochLength, len(t))
	}
	return nil
}

// define enumerations
type TicketsOrKeys struct {
	Tickets *TicketsBodies
	Keys    *EpochKeys
}

// define structures
type TicketBody struct {
	ID      OpaqueHash `json:"id"`
	Attempt U8         `json:"attempt"`
}

type ValidatorData struct {
	Bandersnatch BandersnatchKey `json:"bandersnatch"`
	Ed25519      Ed25519Key      `json:"ed25519"`
	Bls          BlsKey          `json:"bls"`
	Metadata     [128]U8         `json:"metadata"`
}

type ValidatorsData []ValidatorData

func (v ValidatorsData) Validate() error {
	if len(v) != jam_types.ValidatorsCount {
		return fmt.Errorf("ValidatorsData must have exactly %d entries, got %d", jam_types.ValidatorsCount, len(v))
	}
	return nil
}

type TicketEnvelope struct {
	Attempt   U8      `json:"attempt"`
	Signature [784]U8 `json:"signature"`
}

type EpochMark struct {
	Entropy    OpaqueHash        `json:"entropy"`
	Validators []BandersnatchKey `json:"validators"`
}

func (e *EpochMark) Validate() error {
	if len(e.Validators) != jam_types.ValidatorsCount {
		return fmt.Errorf("EpochMark must have exactly %d validators, got %d", jam_types.ValidatorsCount, len(e.Validators))
	}
	return nil
}

type TicketsMark []TicketBody

func (t TicketsMark) Validate() error {
	if len(t) != jam_types.EpochLength {
		return fmt.Errorf("TicketsMark must have exactly %d TicketBody entries, got %d", jam_types.EpochLength, len(t))
	}
	return nil
}

// output markers
type OutputMarks struct {
	EpochMark   *EpochMark   `json:"epoch_mark,omitempty"`   // New epoch signal
	TicketsMark *TicketsMark `json:"tickets_mark,omitempty"` // Tickets signal
}

// state relevant to Safrole protocol
type State struct {
	Tau    U32            `json:"tau"`     // Most recent block's timeslot
	Eta    [4]OpaqueHash  `json:"eta"`     // Entropy accumulator and epochal randomness
	Lambda ValidatorsData `json:"lambda"`  // Validator keys and metadata which were active in the prior epoch
	Kappa  ValidatorsData `json:"kappa"`   // Validator keys and metadata currently active
	GammaK ValidatorsData `json:"gamma_k"` // Validator keys for the following epoch
	Iota   ValidatorsData `json:"iota"`    // Validator keys and metadata to be drawn from next
	GammaA []TicketBody   `json:"gamma_a"` // Sealing-key contest ticket accumulator
	GammaS TicketsOrKeys  `json:"gamma_s"` //	Sealing-key series of the current epoch
	GammaZ [144]U8        `json:"gamma_z"` // Bandersnatch ring commitment
}

// input for Safrole protocol
type Input struct {
	Slot      U32              `json:"slot"`      // Current slot
	Entropy   OpaqueHash       `json:"entropy"`   // Per block entropy (originated from block entropy source VRF)
	Extrinsic []TicketEnvelope `json:"extrinsic"` // Safrole extrinsic
}

// output from Safrole protocol
type Output struct {
	Ok  *OutputMarks     `json:"ok,omitempty"`  // Markers
	Err *CustomErrorCode `json:"err,omitempty"` // Error code
}

// Safrole state transition function execution dump
type Testcase struct {
	Input     Input  `json:"input"`      // Input
	PreState  State  `json:"pre_state"`  // Pre-execution state
	Output    Output `json:"output"`     // Output
	PostState State  `json:"post_state"` // Post-execution state
}
