package work

// work package (14.2)
type WorkPackage struct {
	Authorization []byte        `json:"authorization"`  // authorization token
	AuthCodeHost  uint32        `json:"auth_code_host"` // host service index
	Authorizer    Authorizer    `json:"authorizer"`
	Context       RefineContext `json:"context"`
	WorkItems     []WorkItem    `json:"items"`
}

type Authorizer struct {
	CodeHash [32]byte `json:"code_hash"` // authorization code hash
	Params   []byte   `json:"params"`    // parameterization blob
}
