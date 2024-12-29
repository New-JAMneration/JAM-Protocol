package types

import (
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
)

type Struct struct {
	typeMapping []TypeMap
}

func (st *Struct) Process(s *scale_bytes.Bytes) (interface{}, error) {
	result := make(map[string]interface{})

	for _, m := range st.typeMapping {
		t, gErr := GetType(m.Type)
		if gErr != nil {
			return nil, gErr
		}

		data, pErr := t.Process(s)
		if pErr != nil {
			return nil, pErr
		}

		result[m.Name] = data
	}

	return result, nil
}

func (st *Struct) ProcessEncode(value interface{}) ([]byte, error) {
	data, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New("value is not map")
	}

	var b []byte
	for _, m := range st.typeMapping {
		t, gErr := GetType(m.Type)
		if gErr != nil {
			return nil, gErr
		}

		v, ok := data[m.Name]
		if !ok {
			return nil, errors.New("key not found")
		}

		encode, pErr := t.ProcessEncode(v)
		if pErr != nil {
			return nil, pErr
		}

		b = append(b, encode...)
	}

	return b, nil
}

func NewStruct(typeMapping []TypeMap) IType {
	return &Struct{typeMapping: typeMapping}
}
