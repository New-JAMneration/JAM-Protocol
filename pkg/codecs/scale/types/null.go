package types

import "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"

type Null struct {
}

func (n *Null) Process(s *scale_bytes.Bytes) (interface{}, error) {
	return nil, nil
}

func (n *Null) ProcessEncode(value interface{}) ([]byte, error) {
	return nil, nil
}

func NewNull() IType {
	return &Null{}
}
