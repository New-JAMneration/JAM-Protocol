package types

import (
	"encoding/hex"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
	"strings"
)

type FixedArray struct {
	elementCount int
	subType      string
}

func (f *FixedArray) Process(s *scale_bytes.Bytes) (interface{}, error) {
	if f.elementCount == 0 {
		return nil, nil
	}

	if strings.ToLower(f.subType) == "u8" {
		nextBytes := s.GetNextBytes(f.elementCount)
		return nextBytes, nil
	}

	var result []interface{}

	for i := 0; i < f.elementCount; i++ {
		t, err := GetType(f.subType)
		if err != nil {
			return nil, err
		}

		data, pErr := t.Process(s)
		if pErr != nil {
			return nil, pErr
		}
		result = append(result, data)
	}

	return result, nil
}

func (f *FixedArray) ProcessEncode(value interface{}) ([]byte, error) {
	var data []byte

	if f.subType == "u8" {

		switch val := value.(type) {
		case string:
			if strings.HasPrefix(val, "0x") {
				var err error
				data, err = hex.DecodeString(val[2:])
				if err != nil {
					return nil, errors.New("failed to decode hex string")
				}
			} else {
				data = []byte(val)
			}
		case []byte:
			data = val
		case []interface{}:
			for _, b := range val {
				if byteVal, ok := b.(byte); ok {
					data = append(data, byteVal)
				}
			}
		default:
			return nil, errors.New("unsupported value type for scale_bytes conversion")
		}

		return data, nil
	}

	kind := reflect.TypeOf(value).Kind()

	if kind == reflect.Slice {
		s := reflect.ValueOf(value)
		if s.Len() != f.elementCount {
			return nil, errors.New("invalid length for fixed array")
		}

		for i := 0; i < s.Len(); i++ {
			t, gErr := GetType(f.subType)
			if gErr != nil {
				return nil, gErr
			}
			encode, eErr := t.ProcessEncode(s.Index(i).Interface())
			if eErr != nil {
				return nil, eErr
			}
			data = append(data, encode...)
		}
		return data, nil
	}

	return nil, errors.New("unsupported value type for FixedArray conversion")
}

func NewFixedArray(elementCount int, subType string) IType {
	return &FixedArray{
		elementCount: elementCount,
		subType:      subType,
	}
}
