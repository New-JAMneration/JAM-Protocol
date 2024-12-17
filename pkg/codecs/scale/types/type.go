package types

import (
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type TypeMap struct {
	Name string
	Type string
}

type IType interface {
	Process(s *scale.Bytes) (interface{}, error)
	ProcessEncode(value interface{}) ([]byte, error)
}
