package extrinsic

// Disbutes 10.2

type Disbutes struct {
	Verdicts []Verdicts `json:"verdicts,omitempty"`
	Culprits []Culprits `json:"culprits,omitempty"`
	Faults   []Faults   `json:"faults,omitempty"`
}

type Vote struct {
	Vote           bool     `json:"vote,omitempty"`
	ValidatorIndex uint16   `json:"index,omitempty"`     // N_V : 1023
	Signature      [64]byte `json:"signature,omitempty"` // ed25519
}

type Verdicts struct {
	WRHash       [32]byte `json:"target,omitempty"` // work report hash
	ValidatorSet int      `json:"age"`              // 0, 1 : lambda, kappa
	Vote         `json:"votes,omitempty"`
}

type Culprits struct {
	WRHash    [32]byte `json:"target,omitempty"`    // work report hash
	Key       [32]byte `json:"key,omitempty"`       // validators ed25519 pk
	Signature [64]byte `json:"signature,omitempty"` // ed25519
}

type Faults struct {
	WRHash    [32]byte `json:"target,omitempty"` // work report hash
	Vote      bool     `json:"vote,omitempty"`
	Key       [32]byte `json:"key,omitempty"`       // validators ed25519 pk
	Signature [64]byte `json:"signature,omitempty"` // ed25519
}
