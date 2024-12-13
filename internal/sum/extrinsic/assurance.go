package extrinsic

// Assurance formula is in GP section 11.2.1   Equation 11.10
type Assurance struct {
	Anchor         string // json:"anchor"
	Bitstring      string // json:"bitfield"
	ValidatorIndex uint16 // json:"validator_index"
	Signature      string // json:"signature"
}
