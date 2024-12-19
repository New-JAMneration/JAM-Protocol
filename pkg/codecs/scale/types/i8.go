package types

import (
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type I8 struct {
}

func (i *I8) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data := s.GetNextBytes(1)

	return int8(data[0]), nil
}

func (i *I8) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := i.getI8(value)
	if !ok {
		return nil, errors.New("value is not int8")
	}

	buf := []byte{byte(v)}
	return buf, nil
}

func (i *I8) getI8(val interface{}) (int8, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Int16 {
		return int8(v.Int()), true
	}

	return 0, false
}

func NewI8() IType {
	return &I8{}
}
