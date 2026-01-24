package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		logger.Debugf("%s%s%s", color, string, Reset)
	}
}

type Decoder struct {
	buf            *bytes.Reader
	HashSegmentMap HashSegmentMap
}

func NewDecoder() *Decoder {
	return &Decoder{
		HashSegmentMap: nil,
	}
}

func (d *Decoder) Decode(data []byte, v interface{}) error {
	d.buf = bytes.NewReader(data)
	return d.decodeStruct(v)
}

func (d *Decoder) ReadPointerFlag() (byte, error) {
	cLog(Cyan, "Reading pointer flag")
	firstByte, err := d.buf.ReadByte()
	if err != nil {
		return 0, err
	}
	return firstByte, nil
}

func (d *Decoder) ReadLegnthFlag() (byte, error) {
	cLog(Cyan, "Reading length flag")
	firstByte, err := d.buf.ReadByte()
	if err != nil {
		return 0, err
	}
	return firstByte, nil
}

func (d *Decoder) ReadErrorByte() (byte, error) {
	cLog(Cyan, "Reading error byte")
	firstByte, err := d.buf.ReadByte()
	if err != nil {
		return 0, err
	}
	return firstByte, nil
}

func (d *Decoder) DecodeUint(data []byte) (uint64, error) {
	if len(data) < 1 {
		return 0, errors.New("no data to deserialize U64")
	}
	prefix := data[0]

	// If x < 0x80: E(x) = [x]
	if prefix < 0x80 {
		return uint64(prefix), nil
	}

	// If prefix = 0xFF: E(x) = [255] || E_8(x)
	if prefix == 0xFF {
		if len(data) < 9 {
			return 0, errors.New("not enough data for 8-byte U64")
		}
		return binary.LittleEndian.Uint64(data[1:9]), nil
	}

	l := bits.LeadingZeros8(^prefix)
	needed := l + 1
	if len(data) < needed {
		return 0, errors.New("not enough data for U64")
	}

	base := 0xFF - (uint8(1) << (8 - uint(l))) + 1
	floorVal := uint64(prefix - base)

	var x uint64
	switch l {
	case 1:
		x = (floorVal << 8) | uint64(data[1])
	case 2:
		x = (floorVal << 16) | uint64(binary.LittleEndian.Uint16(data[1:3]))
	case 3:
		if len(data) >= 5 {
			x = (floorVal << 24) | (uint64(binary.LittleEndian.Uint32(data[1:5])) & 0xFFFFFF)
		} else {
			x = (floorVal << 24) | uint64(data[1]) | uint64(data[2])<<8 | uint64(data[3])<<16
		}
	case 4:
		x = (floorVal << 32) | uint64(binary.LittleEndian.Uint32(data[1:5]))
	default:
		remainder := uint64(0)
		for i := range l {
			remainder |= uint64(data[i+1]) << (8 * uint(i))
		}
		x = (floorVal << (8 * uint(l))) | remainder
	}

	// Validate encoding: x >= 2^(7*l)
	if x < (uint64(1) << (7 * uint(l))) {
		return 0, errors.New("invalid U64 encoding")
	}

	return x, nil
}

func (d *Decoder) IdentifyLength(byteValue byte) uint8 {
	return uint8(bits.LeadingZeros8(^byteValue))
}

func (d *Decoder) decodeUintFromReader() (uint64, error) {
	lengthFlag, err := d.buf.ReadByte()
	if err != nil {
		return 0, err
	}

	remainingByteSize := d.IdentifyLength(lengthFlag)
	totalSize := 1 + remainingByteSize

	bytesValue := make([]byte, totalSize)
	bytesValue[0] = lengthFlag

	if remainingByteSize > 0 {
		n, err := d.buf.Read(bytesValue[1:])
		if err != nil {
			return 0, err
		}
		if n != int(remainingByteSize) {
			return 0, errors.New("not enough data to read remaining bytes")
		}
	}

	return d.DecodeUint(bytesValue)
}

func (d *Decoder) DecodeInteger() (uint64, error) {
	return d.decodeUintFromReader()
}

// C.6 Deserialization
func (d *Decoder) DecodeLength() (uint64, error) {
	cLog(Yellow, "Reading length flag")
	length, err := d.decodeUintFromReader()
	if err != nil {
		return 0, err
	}
	cLog(Yellow, fmt.Sprintf("Slice Length: %v", length))
	return length, nil
}

func (d *Decoder) decodeStruct(v interface{}) error {
	if decodable, ok := v.(Decodable); ok {
		return decodable.Decode(d)
	}
	return binary.Read(d.buf, binary.LittleEndian, v)
}

type Decodable interface {
	Decode(d *Decoder) error
}

func (d *Decoder) SetHashSegmentMap(hashSegmentMap HashSegmentMap) {
	d.HashSegmentMap = hashSegmentMap
}
