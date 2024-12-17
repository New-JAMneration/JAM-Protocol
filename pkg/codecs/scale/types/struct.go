package types

import (
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type Struct struct {
	typeMapping []TypeMap
}

func (st *Struct) Process(s *scale.Bytes) (interface{}, error) {
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
	var data []byte
	for _, m := range st.typeMapping {
		t, gErr := GetType(m.Type)
		if gErr != nil {
			return nil, gErr
		}
		encode, pErr := t.ProcessEncode(value)
		if pErr != nil {
			return nil, pErr
		}

		data = append(data, encode...)
	}

	return data, nil
}

func NewStruct(typeMapping []TypeMap) IType {
	return &Struct{typeMapping: typeMapping}
}
