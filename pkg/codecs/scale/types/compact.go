package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	bytes2 "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
)

type Compact struct {
	CompactLength int
	CompactBytes  []byte
}

func (c *Compact) ProcessCompactBytes(s *bytes2.Bytes) (uint64, error) {
	data := s.GetNextBytes(1)

	b := data[0]

	var v uint64

	switch {
	case b == 0:
		v = 0
	case b == 0xff:
		v = binary.LittleEndian.Uint64(data)
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
		buf := s.GetNextBytes(length)

		// Calculate remaining part (`rem`) and combine to get final value
		rem := int(b & ((1 << (7 - length)) - 1))
		if len(buf) == 0 {
			v = uint64(rem << (8 * length))
		} else {
			v = binary.LittleEndian.Uint64(buf) + uint64(rem<<(8*length))
		}
	}

	return v, nil
}

func (c *Compact) Process(s *bytes2.Bytes) (interface{}, error) {
	return c.ProcessCompactBytes(s)
}

// encodeTo
// https://github.com/davxy/parity-scale-codec/blob/98b8a44133eb26b2f7fc8d867a928dbf5b64e897/src/compact.rs#L397
func (c *Compact) encodeTo(x uint64) ([]byte, error) {
	var buf bytes.Buffer

	if x == 0 {
		// 寫入單一字節 0
		_, err := buf.Write([]byte{0})
		return buf.Bytes(), err
	}

	// 尋找最小的 l 在 0..8 中，使得 2^(7*l) <= x < 2^(7*(l+1))
	var l int
	found := false
	for i := 0; i < 8; i++ {
		lower := uint64(1) << (7 * i)
		upper := uint64(1) << (7 * (i + 1))
		if lower <= x && x < upper {
			l = i
			found = true
			break
		}
	}

	if found {
		// 計算第一個字節: (2^8 - 2^(8-l)) + (x / 2^(8*l))
		prefix := (uint64(1) << 8) - (uint64(1) << (8 - l))
		prefix += x / (uint64(1) << (8 * l))
		firstByte := byte(prefix)

		// 寫入第一個字節
		_, err := buf.Write([]byte{firstByte})
		if err != nil {
			return nil, err
		}

		// 計算剩餘的字節: x % 2^(8*l)
		rem := x % (uint64(1) << (8 * l))
		remBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(remBytes, rem)
		// 只寫入最低的 'l' 個字節
		_, err = buf.Write(remBytes[:l])
		return buf.Bytes(), err
	}

	// 如果沒有找到 l，則寫入 0xFF 後跟 x 的 8 個小端字節
	_, err := buf.Write([]byte{0xFF})
	if err != nil {
		return nil, err
	}

	xBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(xBytes, x)
	_, err = buf.Write(xBytes)
	return buf.Bytes(), err
}

func (c *Compact) ProcessEncode(value interface{}) ([]byte, error) {
	data, ok := value.(int)
	if !ok {
		return nil, errors.New("value is not int")
	}

	to, err := c.encodeTo(uint64(data))
	return to, err
}

func NewCompact() IType {
	return &Compact{}
}
