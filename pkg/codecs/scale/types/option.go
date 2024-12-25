package types

import "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"

type Option struct {
	subType string
}

func (o *Option) Process(s *scale_bytes.Bytes) (interface{}, error) {
	optionByte := s.GetNextBytes(1)

	if o.subType != "" && optionByte[0] != 0x00 {
		t, gErr := GetType(o.subType)
		if gErr != nil {
			return nil, gErr
		}
		value, pErr := t.Process(s)
		if pErr != nil {
			return nil, pErr
		}
		return value, nil
	}

	return nil, nil
}

func (o *Option) ProcessEncode(value interface{}) ([]byte, error) {
	if value != nil && o.subType != "" {
		t, gErr := GetType(o.subType)
		if gErr != nil {
			return nil, gErr
		}
		encodedValue, err := t.ProcessEncode(value)
		if err != nil {
			return nil, err
		}
		return append([]byte{0x01}, encodedValue...), nil
	}

	return []byte{0x00}, nil
}

func NewOption(subType string) IType {
	return &Option{subType: subType}
}
