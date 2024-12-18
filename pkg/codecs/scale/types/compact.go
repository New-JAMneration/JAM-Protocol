package types

import (
	"encoding/binary"
	"errors"
	bytes2 "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"math"
)

type Compact struct {
	CompactLength int
	CompactBytes  []byte
}

func (c *Compact) ProcessCompactBytes(s *bytes2.Bytes) (int, error) {
	data, err := s.GetNextBytes(1)
	if err != nil {
		return 0, err
	}

	b := data[0]

	var v int

	switch {
	case b == 0:
		v = 0
	case b == 0xff:
		v = int(binary.LittleEndian.Uint64(data))
	default:
		// Find the first zero bit from the left
		length := 0
		for i := 0; i < 8; i++ {
			if (b & (0b10000000 >> i)) == 0 {
				length = i
				break
			}
		}

		// Get subsequent scale_bytes
		buf, nErr := s.GetNextBytes(length)
		if nErr != nil {
			return 0, errors.New("failed to read length of compact_bytes")
		}

		// Calculate remaining part (`rem`) and combine to get final value
		rem := int(b & ((1 << (7 - length)) - 1))
		if len(buf) == 0 {
			v = rem << (8 * length)
		} else {
			v = int(binary.LittleEndian.Uint64(buf)) + (rem << (8 * length))
		}
	}

	return v, nil
}

func (c *Compact) Process(s *bytes2.Bytes) (interface{}, error) {
	return c.ProcessCompactBytes(s)
}

func (c *Compact) ProcessEncode(value interface{}) ([]byte, error) {
	data, ok := value.(int)
	if !ok {
		return nil, errors.New("value is not int")
	}

	if data <= 0b00111111 {
		return []byte{byte(data << 2)}, nil
	} else if data <= 0b0011111111111111 {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16((data<<2)|0b01))
		return b, nil
	} else if data <= 0b00111111111111111111111111111111 {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32((data<<2)|0b10))
		return b, nil
	} else {
		for bytesLength := 4; bytesLength <= 68; bytesLength++ {
			if math.Pow(2, float64(8*(bytesLength-1))) <= float64(data) &&
				float64(data) < math.Pow(2, float64(8*bytesLength)) {

				headerByte := byte(((bytesLength - 4) << 2) | 0b11)
				valueBytes := make([]byte, bytesLength)
				binary.LittleEndian.PutUint64(valueBytes, uint64(data))

				return append([]byte{headerByte}, valueBytes...), nil
			}
		}
		return nil, errors.New("value out of range")
	}
}

func NewCompact() IType {
	return &Compact{}
}
