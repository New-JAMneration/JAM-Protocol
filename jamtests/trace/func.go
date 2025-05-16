package jamtests

import "github.com/New-JAMneration/JAM-Protocol/internal/store"

func (s *TraceTestCase) Dump() error {
	return nil
}

func (s *TraceTestCase) GetPostState() interface{} {
	return s.PostState
}

func (s *TraceTestCase) GetOutput() interface{} {
	return nil
}

func (s *TraceTestCase) ExpectError() error {
	return nil
}

func (s *TraceTestCase) Validate() error {
	instance := store.GetInstance()
	if instance.GetLatestBlock().Header.ParentStateRoot != s.PostState.StateRoot {
		return s.CmpKeyVal()
	}
	return nil
}

func (s *TraceTestCase) CmpKeyVal() error {
	// TODO: Compare the KeyVal
	return nil
}
