package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	bytes2 "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type U8 struct {
}

func (u *U8) Process(s *bytes2.Bytes) (interface{}, error) {
	data, err := s.GetNextU8()
	if err != nil {
		return 0, err
	}

	return data, nil
}

func (u *U8) ProcessEncode(value interface{}) ([]byte, error) {
	u8, ok := u.getUint8(value)
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

func (u *U8) getUint8(val interface{}) (uint8, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Uint8 {
		return uint8(v.Uint()), true
	}

	return 0, false
}

func NewU8() IType {
	return &U8{}
}
