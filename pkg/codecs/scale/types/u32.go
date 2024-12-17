package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type U32 struct {
}

func (u *U32) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data), nil
}

func (u *U32) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(uint32)
	if !ok {
		return nil, errors.New("value is not uint32")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf, nil
}

func NewU32() IType {
	return &U32{}
}
