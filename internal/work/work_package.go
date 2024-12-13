package work

// work package (14.2)
type WorkPackage struct {
	Authorization []byte // authorization token
	AuthCodeHost  uint32 // host service index
	Authorizer    Authorizer
	Context       RefineContext
	WorkItems     []WorkItem
}

type Authorizer struct {
	CodeHash [32]byte // authorization code hash
	Params   []byte   // parameterization blob
}
