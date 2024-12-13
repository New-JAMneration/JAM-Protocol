package extrinsic

// Assurance formula is in GP section 11.2.1   Equation 11.10
type Assurance struct {
	Anchor         string `json:"anchor"`
	Bitstring      string `json:"bitfield"`
	ValidatorIndex uint16 `json:"validator_index"`
	Signature      string `json:"signature"`
}

// NewAssurance is a constructor, not sure if bitstring should be initialized all zero
func NewAssurance(anchor string, bitstring string, validatorIndex uint16, signature string) *Assurance {
	return &Assurance{
		Anchor:         anchor,
		Bitstring:      bitstring,
		ValidatorIndex: validatorIndex,
		Signature:      signature,
	}
}
