package types

import (
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type I8 struct {
}

func (i *I8) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(1)
	if err != nil {
		return 0, err
	}

	return int8(data[0]), nil
}

func (i *I8) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(int8)
	if !ok {
		return nil, errors.New("value is not int8")
	}

	buf := []byte{byte(v)}
	return buf, nil
}

func NewI8() IType {
	return &I8{}
}
