package types

import (
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
)

type Bool struct {
}

func (b *Bool) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBool()
	if err != nil {
		return 0, err
	}

	return &data, nil
}

func (b *Bool) ProcessEncode(value interface{}) ([]byte, error) {
	if data, ok := value.(bool); ok {
		if data {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	}

	return nil, fmt.Errorf("not bool")
}

func NewBool() IType {
	return &Bool{}
}
