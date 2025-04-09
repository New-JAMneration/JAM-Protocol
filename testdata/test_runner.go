package testdata

type TestRunner interface {
	Run(data interface{}) error
}
