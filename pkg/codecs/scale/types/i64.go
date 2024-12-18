package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type I64 struct {
}

func (i *I64) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(8)
	if err != nil {
		return 0, err
	}

	uintVal := binary.LittleEndian.Uint32(data)
	intVal := int16(uintVal)

	return intVal, nil
}

func (i *I64) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := i.getI64(value)
	if !ok {
		return nil, errors.New("value is not int64")
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(v))
	return buf, nil
}

func (i *I64) getI64(val interface{}) (int64, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Int64 {
		return v.Int(), true
	}

	return 0, false
}

func NewI64() IType {
	return &I64{}
}
