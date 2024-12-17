package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type I16 struct {
}

func (i *I16) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(2)
	if err != nil {
		return 0, err
	}

	uintVal := binary.LittleEndian.Uint16(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I16) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(int16)
	if !ok {
		return nil, errors.New("value is not int16")
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(v))
	return buf, nil
}

func NewI16() IType {
	return &I16{}
}
