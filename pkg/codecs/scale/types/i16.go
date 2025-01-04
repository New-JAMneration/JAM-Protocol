package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type I16 struct {
}

func (i *I16) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data := s.GetNextBytes(2)

	uintVal := binary.LittleEndian.Uint16(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I16) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := i.getI16(value)
	if !ok {
		return nil, errors.New("value is not int16")
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(v))
	return buf, nil
}

func (i *I16) getI16(val interface{}) (int16, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Int16 {
		return int16(v.Int()), true
	}

	return 0, false
}

func NewI16() IType {
	return &I16{}
}
