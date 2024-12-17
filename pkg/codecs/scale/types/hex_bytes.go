package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type HexBytes struct {
}

func (i *HexBytes) Process(s *scale.Bytes) (interface{}, error) {
	compactU32 := NewCompactU32()
	elementCount, err := compactU32.Process(s)
	if err != nil {
		return nil, err
	}

	nextBytes, gErr := s.GetNextBytes(elementCount.(int))
	if gErr != nil {
		return nil, gErr
	}

	return fmt.Sprintf("0x%s", hex.EncodeToString(nextBytes)), nil
}

func (i *HexBytes) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not string")
	}

	if len(v) < 2 || v[:2] != "0x" {
		return nil, errors.New(`hexBytes value should start with "0x"`)
	}

	rawValue, err := hex.DecodeString(v[2:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex value: %v", err)
	}

	compactU32 := NewCompactU32()
	lengthEncoded, err := compactU32.ProcessEncode(len(rawValue))
	if err != nil {
		return nil, err
	}

	return append(lengthEncoded, rawValue...), nil
}

func NewHexBytes() IType {
	return &HexBytes{}
}
