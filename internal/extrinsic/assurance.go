package extrinsic

// Assurance equation 11.8
type Assurance struct {
	Anchor         string `json:"anchor"`
	Bitstring      string `json:"bitfield"`
	ValidatorIndex uint16 `json:"validator_index"`
	Signature      string `json:"signature"`
}
