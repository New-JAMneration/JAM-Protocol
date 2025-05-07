package testdata

type Testable interface {
	// Dump the test data to store
	Dump() error

	// Get State
	GetPostState() interface{}

	// Get Output
	GetOutput() interface{}

	// Expect Error
	ExpectError() error

	// Valide
	Validate() error
}
