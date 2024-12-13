package extrinsic

// Assurance is in graypaper section 11.2.1 equation 11.10
type Assurance struct {
	Anchor         [32]byte // json : "anchor"
	Bitstring      string   // json : "bitfield"
	ValidatorIndex uint16   // json : "validator_index"
	Signature      [64]byte // json : "signature"
}
