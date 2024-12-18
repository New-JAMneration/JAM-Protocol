package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
)

type U64 struct {
}

func (u *U64) Process(s *scale_bytes.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(8)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (u *U64) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := u.getUint64(value)
	if !ok {
		return nil, errors.New("value is not uint64")
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf, nil
}

func (u *U64) getUint64(val interface{}) (uint64, bool) {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return 0, false
	}

	if v.Kind() == reflect.Uint64 {
		return v.Uint(), true
	}

	return 0, false
}

func NewU64() IType {
	return &U64{}
}
