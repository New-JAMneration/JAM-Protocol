package types

import (
	"encoding/hex"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"reflect"
	"strings"
)

type Vec struct {
	subType string
}

func (v *Vec) Process(s *scale_bytes.Bytes) (interface{}, error) {
	compactU32 := NewCompactU32()
	elementCount, err := compactU32.Process(s)
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for i := 0; i < int(elementCount.(uint64)); i++ {
		t, gErr := GetType(v.subType)
		if gErr != nil {
			return nil, gErr
		}

		result, pErr := t.Process(s)
		if pErr != nil {
			return nil, pErr
		}

		results = append(results, result)
	}

	return results, nil
}

func (v *Vec) ProcessEncode(value interface{}) ([]byte, error) {
	var data []byte

	elementCount := CompactU32{}

	if v.subType == "u8" {
		var byteArray []byte

		switch val := value.(type) {
		case string:
			if strings.HasPrefix(val, "0x") {
				var err error
				byteArray, err = hex.DecodeString(val[2:])
				if err != nil {
					return nil, errors.New("failed to decode hex string")
				}
			} else {
				byteArray = []byte(val)
			}
		case []byte:
			byteArray = val
		case []interface{}:
			for _, b := range val {
				if byteVal, ok := b.(byte); ok {
					byteArray = append(byteArray, byteVal)
				}
			}
		default:
			return nil, errors.New("unsupported value type for scale_bytes conversion")
		}

		lengthEncoded, err := elementCount.ProcessEncode(len(byteArray))
		if err != nil {
			return nil, err
		}
		data = append(lengthEncoded, byteArray...)
		return data, nil
	}

	kind := reflect.TypeOf(value).Kind()

	if kind == reflect.Slice {
		s := reflect.ValueOf(value)
		lengthEncoded, err := elementCount.ProcessEncode(s.Len())
		if err != nil {
			return nil, err
		}

		data = append(data, lengthEncoded...)
		for i := 0; i < s.Len(); i++ {
			t, gErr := GetType(v.subType)
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

	return nil, errors.New("unsupported value type for Vec conversion")
}

func NewVec(subType string) IType {
	return &Vec{subType: subType}
}
