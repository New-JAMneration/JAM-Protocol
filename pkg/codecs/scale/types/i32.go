package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type I32 struct {
}

func (i *I32) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(4)
	if err != nil {
		return 0, err
	}

	uintVal := binary.LittleEndian.Uint32(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I32) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := i.getI32(value)
	if !ok {
		return nil, errors.New("value is not int32")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return buf, nil
}

func (i *I32) getI32(val interface{}) (int32, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Int32 {
		return int32(v.Int()), true
	}

	return 0, false
}

func NewI32() IType {
	return &I32{}
}
