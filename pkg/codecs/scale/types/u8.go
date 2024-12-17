package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type U8 struct {
}

func (u *U8) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextU8()
	if err != nil {
		return 0, err
	}

	return data, nil
}

func (u *U8) ProcessEncode(value interface{}) ([]byte, error) {
	u8, ok := value.(uint8)
	if !ok {
		return nil, errors.New("value is not uint8")
	}

	if u8 < 0 || u8 > 1<<8-1 {
		return nil, errors.New("value out of range for uint8")
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewU8() IType {
	return &U8{}
}
