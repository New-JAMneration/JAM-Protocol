package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type U32 struct {
}

func (u *U32) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data), nil
}

func (u *U32) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := u.getUint32(value)
	if !ok {
		return nil, errors.New("value is not uint32")
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf, nil
}

func (u *U32) getUint32(val interface{}) (uint32, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Uint32 {
		return uint32(v.Uint()), true
	}

	return 0, false
}

func NewU32() IType {
	return &U32{}
}
