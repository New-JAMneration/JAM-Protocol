package testdata

type TestRunner interface {
	Run(data interface{}) error
	Verify(data Testable) error
}
