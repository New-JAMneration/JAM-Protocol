package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type I64 struct {
}

func (i *I64) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(8)
	if err != nil {
		return 0, err
	}

	uintVal := binary.LittleEndian.Uint32(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I64) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(int64)
	if !ok {
		return nil, errors.New("value is not int64")
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(v))
	return buf, nil
}

func NewI64() IType {
	return &I64{}
}
