package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type U16 struct {
}

func (u *U16) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data), nil
}

func (u *U16) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(uint16)
	if !ok {
		return nil, errors.New("value is not uint16")
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return buf, nil
}

func NewU16() IType {
	return &U16{}
}
