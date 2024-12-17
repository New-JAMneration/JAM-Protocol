package scale

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type Bytes struct {
	data   []byte
	offset int
}

func (s *Bytes) GetNextBytes(length int) ([]byte, error) {
	if s.offset+length > len(s.data) {
		return nil, errors.New("not enough bytes remaining")
	}
	data := s.data[s.offset : s.offset+length]
	s.offset += length
	return data, nil
}

func (s *Bytes) GetNextU8() (uint8, error) {
	data, err := s.GetNextBytes(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

func (s *Bytes) GetNextBool() (bool, error) {
	byteValue, err := s.GetNextU8()
	if err != nil {
		return false, fmt.Errorf("failed to read byte for boolean: %w", err)
	}

	if byteValue == 0 {
		return false, nil
	} else if byteValue == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid value for boolean: %d", byteValue)
	}
}

func (s *Bytes) GetRemainingBytes() []byte {
	data := s.data[s.offset:]
	s.offset = len(s.data)
	return data
}

func (s *Bytes) GetRemainingLength() int {
	return len(s.data) - s.offset
}

func (s *Bytes) Reset() {
	s.offset = 0
}

func (s *Bytes) ToHex() string {
	return "0x" + hex.EncodeToString(s.data)
}

func NewBytes(data interface{}) (*Bytes, error) {
	switch v := data.(type) {
	case []byte:
		return &Bytes{data: v, offset: 0}, nil
	case string:
		if len(v) >= 2 && v[:2] == "0x" {
			decoded, err := hex.DecodeString(v[2:])
			if err != nil {
				return nil, fmt.Errorf("invalid hex string: %w", err)
			}
			return &Bytes{data: decoded, offset: 0}, nil
		}
		return nil, errors.New("string data must start with '0x'")
	default:
		return nil, errors.New("unsupported data type")
	}
}
