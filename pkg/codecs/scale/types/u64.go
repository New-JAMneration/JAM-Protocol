package types

import (
	"encoding/binary"
	"errors"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
)

type U64 struct {
}

func (u *U64) Process(s *scale.Bytes) (interface{}, error) {
	data, err := s.GetNextBytes(8)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (u *U64) ProcessEncode(value interface{}) ([]byte, error) {
	v, ok := value.(uint64)
	if !ok {
		return nil, errors.New("value is not uint64")
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return buf, nil
}

func NewU64() IType {
	return &U64{}
}
