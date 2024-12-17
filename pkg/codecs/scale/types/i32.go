package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type I32 struct {
}

func (i *I32) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(4)
	if err != nil {
		return 0, err
	}

	uintVal := binary.LittleEndian.Uint32(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I32) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(int32)
	if !ok {
		return nil, errors.New("value is not int32")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return buf, nil
}

func NewI32() IType {
	return &I32{}
}
