package types

import (
	"errors"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
)

type Enum struct {
	typeMapping map[int]TypeMap
}

func (e *Enum) Process(s *scale_bytes.Bytes) (interface{}, error) {
	b, err := s.GetNextBytes(1)
	if err != nil {
		return nil, err
	}

	i := int(b[0])

	if len(e.typeMapping) == 0 {
		return nil, errors.New("no enum type")
	}

	if m, ok := e.typeMapping[i]; ok {
		t, err := GetType(m.Type)
		if err != nil {
			return nil, err
		}

		data, pErr := t.Process(s)
		if pErr != nil {
			return nil, pErr
		}

		return map[string]interface{}{m.Name: data}, nil
	}

	return nil, fmt.Errorf("index %d not present in enum value list", i)
}

func (e *Enum) ProcessEncode(value interface{}) ([]byte, error) {
	switch data := value.(type) {
	case map[string]interface{}:
		for enumKey, enumValue := range data {
			for idx, m := range e.typeMapping {
				if m.Name == enumKey {
					t, err := GetType(m.Type)
					if err != nil {
						return nil, err
					}
					b, err := t.ProcessEncode(enumValue)
					if err != nil {
						return nil, err
					}

					result := []byte{byte(idx)}
					result = append(result, b...)
					return result, nil
				}
			}
		}
	case string:
		for idx, m := range e.typeMapping {
			if m.Name == data {
				result := []byte{byte(idx)}
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("value %+v not present in value list of this enum", value)
}

func NewEnum(typeMapping map[int]TypeMap) IType {
	return &Enum{typeMapping: typeMapping}
}
