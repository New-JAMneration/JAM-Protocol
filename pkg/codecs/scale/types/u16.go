package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type U16 struct {
}

func (u *U16) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data), nil
}

func (u *U16) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := u.getUint16(value)
	if !ok {
		return nil, errors.New("value is not uint16")
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return buf, nil
}

func (u *U16) getUint16(val interface{}) (uint16, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Uint16 {
		return uint16(v.Uint()), true
	}

	return 0, false
}

func NewU16() IType {
	return &U16{}
}
