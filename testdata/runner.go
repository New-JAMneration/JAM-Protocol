package testdata

type TestRunner interface {
	Run(data interface{}, runSTF bool) error
	Verify(data Testable) error
}
