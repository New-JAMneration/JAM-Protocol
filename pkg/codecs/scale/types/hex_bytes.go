package types

import (
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
)

type HexBytes struct {
}

func (i *HexBytes) Process(s *scale_bytes.Bytes) (interface{}, error) {
	compactU32 := NewCompactU32()
	elementCount, err := compactU32.Process(s)
	if err != nil {
		return nil, err
	}

	nextBytes, gErr := s.GetNextBytes(elementCount.(int))
	if gErr != nil {
		return nil, gErr
	}

	return nextBytes, nil
}

func (i *HexBytes) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.([]uint8)
	if !ok {
		return nil, errors.New("value is not bytes")
	}

	compactU32 := NewCompactU32()
	lengthEncoded, err := compactU32.ProcessEncode(len(v))
	if err != nil {
		return nil, err
	}

	return append(lengthEncoded, v...), nil
}

func NewHexBytes() IType {
	return &HexBytes{}
}
