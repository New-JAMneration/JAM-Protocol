package types

import "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"

type TypeMap struct {
	Name string
	Type string
}

type IType interface {
	Process(s *scale_bytes.Bytes) (interface{}, error)
	ProcessEncode(value interface{}) ([]byte, error)
}
