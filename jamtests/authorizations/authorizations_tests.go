package jamtests

import (
	"encoding/json"
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

type AuthorizationTestCase struct {
	Input     AuthorizationInput  `json:"input"`
	PreState  AuthorizationState  `json:"pre_state"`
	Output    AuthorizationOutput `json:"output"`
	PostState AuthorizationState  `json:"post_state"`
}

/*
This sequence is out of GP spec and derived from the Guarantees Extrinsic (E_G)

For the sake of construction simplicity, we prefer not to include the complete

		extrinsic here but rather focus only on the components relevant to advancing this subsystem's STF.

		CoreAuthorizers[i] = CoreAuthorizer {
	    	core: E_G[i].w.core,
	    	auth-hash: E_G[i].w.auth-hash
		}
*/
type CoreAuthorizer struct {
	CoreIndex      types.CoreIndex  `json:"core"`
	AuthorizerHash types.OpaqueHash `json:"auth_hash"`
}

type CoreAuthorizers []CoreAuthorizer

type AuthorizationInput struct {
	Slot  types.TimeSlot  `json:"slot"`
	Auths CoreAuthorizers `json:"auths"`
}

type AuthorizationOutput struct { // null
}

type AuthorizationState struct {
	Alpha  types.AuthPools  `json:"auth_pools"`
	Varphi types.AuthQueues `json:"auth_queues"`
}

// Unmarshal JSON CoreAuthorizer
func (c *CoreAuthorizer) UnmarshalJSON(data []byte) error {
	var err error

	var input struct {
		CoreIndex      types.CoreIndex  `json:"core"`
		AuthorizerHash types.OpaqueHash `json:"auth_hash"`
	}

	if err = json.Unmarshal(data, &input); err != nil {
		return err
	}

	c.CoreIndex = input.CoreIndex
	c.AuthorizerHash = input.AuthorizerHash

	return nil
}

// Unmarshal JSON CoreAuthorizers
func (c *CoreAuthorizers) UnmarshalJSON(data []byte) error {
	var err error

	var input []CoreAuthorizer
	if err = json.Unmarshal(data, &input); err != nil {
		return err
	}

	if len(input) == 0 {
		return nil
	}

	*c = input

	return nil
}

// Unmarshal JSON AuthorizationInput
func (i *AuthorizationInput) UnmarshalJSON(data []byte) error {
	var err error

	var input struct {
		Slot  types.TimeSlot  `json:"slot"`
		Auths CoreAuthorizers `json:"auths"`
	}

	if err = json.Unmarshal(data, &input); err != nil {
		return err
	}

	i.Slot = input.Slot
	i.Auths = input.Auths

	return nil
}

// CoreAuthorizer
func (c *CoreAuthorizer) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding CoreAuthorizer")
	var err error

	if err = c.CoreIndex.Decode(d); err != nil {
		return err
	}

	if err = c.AuthorizerHash.Decode(d); err != nil {
		return err
	}

	return nil
}

// AuthorizationInput
func (i *AuthorizationInput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AuthorizationInput")
	var err error

	if err = i.Slot.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	i.Auths = make([]CoreAuthorizer, length)
	for j := range i.Auths {
		if err = i.Auths[j].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// AuthorizationState
func (s *AuthorizationState) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AuthorizationState")
	var err error

	if err = s.Alpha.Decode(d); err != nil {
		return err
	}

	if err = s.Varphi.Decode(d); err != nil {
		return err
	}

	return nil
}

// AuthorizationOutput
func (o *AuthorizationOutput) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AuthorizationOutput")
	return nil
}

// AuthorizationTestCase
func (a *AuthorizationTestCase) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding AuthorizationTestCase")
	var err error

	if err = a.Input.Decode(d); err != nil {
		return err
	}

	if err = a.PreState.Decode(d); err != nil {
		return err
	}

	if err = a.Output.Decode(d); err != nil {
		return err
	}

	if err = a.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// CoreAuthorizer
func (c *CoreAuthorizer) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding CoreAuthorizer")
	var err error

	if err = c.CoreIndex.Encode(e); err != nil {
		return err
	}

	if err = c.AuthorizerHash.Encode(e); err != nil {
		return err
	}

	return nil
}

// AuthorizationInput
func (i *AuthorizationInput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AuthorizationInput")
	var err error

	if err = i.Slot.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(i.Auths))); err != nil {
		return err
	}

	for j := range i.Auths {
		if err = i.Auths[j].Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AuthorizationState
func (s *AuthorizationState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AuthorizationState")
	var err error

	if err = s.Alpha.Encode(e); err != nil {
		return err
	}

	if err = s.Varphi.Encode(e); err != nil {
		return err
	}

	return nil
}

// AuthorizationOutput
func (o *AuthorizationOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AuthorizationOutput")
	return nil
}

// AuthorizationTestCase
func (a *AuthorizationTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AuthorizationTestCase")
	var err error

	if err = a.Input.Encode(e); err != nil {
		return err
	}

	if err = a.PreState.Encode(e); err != nil {
		return err
	}

	if err = a.Output.Encode(e); err != nil {
		return err
	}

	if err = a.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

// TODO: Implement Dump method
func (a *AuthorizationTestCase) Dump() error {
	return nil
}

func (a *AuthorizationTestCase) GetPostState() interface{} {
	return a.PostState
}

func (a *AuthorizationTestCase) GetOutput() interface{} {
	return a.Output
}

func (a *AuthorizationTestCase) ExpectError() error {
	return nil
}

func (a *AuthorizationTestCase) Validate() error {
	return nil
}
