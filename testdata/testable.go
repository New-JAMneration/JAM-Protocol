package testdata

type Testable interface {
	// Dump the test data to store
	Dump() error

	// Get State
	GetPostState() interface{}

	// Valide
	Validate() error
}
