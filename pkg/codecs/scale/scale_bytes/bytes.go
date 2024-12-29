package scale_bytes

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type Bytes struct {
	data   []byte
	offset int
}

func (s *Bytes) GetNextBytes(length int) []byte {
	if s.offset+length > len(s.data) {
		data := s.data[s.offset:]
		s.offset = len(s.data)
		return data
	}
	data := s.data[s.offset : s.offset+length]
	s.offset += length
	return data
}

func (s *Bytes) GetNextU8() uint8 {
	data := s.GetNextBytes(1)
	return data[0]
}

func (s *Bytes) GetNextBool() (bool, error) {
	byteValue := s.GetNextU8()

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
