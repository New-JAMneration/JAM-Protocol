package extrinsic

// Tickets (6.6)(6.29)
type Tickets struct {
	EntryIndex    uint16    `json:"attempt,omitempty"`   // naturenumber
	ValidityProof [784]byte `json:"signature,omitempty"` // bandersnatch ringvrf
}
