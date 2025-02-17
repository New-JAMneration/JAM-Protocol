package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

type Decoder struct {
	buf *bytes.Reader
}

func NewDecoder() *Decoder {
	return &Decoder{}
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
	if len(data) == 0 {
		return 0, errors.New("no data to deserialize U64")
	}
	prefix := data[0]
	data = data[1:]

	// If x = 0: E(x) = [0]
	if prefix == 0 {
		return 0, nil
	}

	// If prefix = 0xFF: E(x) = [255] || E_8(x)
	if prefix == 0xFF {
		if len(data) < 8 {
			return 0, errors.New("not enough data for 8-byte U64")
		}
		var x uint64
		for i := 0; i < 8; i++ {
			x |= uint64(data[i]) << (8 * i)
		}
		return x, nil
	}

	// Otherwise, attempt to find the correct l by checking each candidate l
	// and verifying that decoded x fits the expected range.
	var bestX uint64
	found := false
	for tryL := 0; tryL <= 7; tryL++ {
		base := 256 - (1 << (8 - tryL))
		if int(prefix) >= base {
			// floorVal = prefix - base
			floorVal := uint64(int(prefix) - base)

			if len(data) < tryL {
				// Not enough data for remainder, try next l
				continue
			}

			remainderData := data[:tryL]
			var remainder uint64
			for i := 0; i < tryL; i++ {
				remainder |= uint64(remainderData[i]) << (8 * i)
			}

			// x = floorVal*(2^(8*l)) + remainder
			power8l := uint64(1) << (8 * tryL)
			x := floorVal*power8l + remainder

			// Check range to confirm l:
			// 2^(7*l) ≤ x < 2^(7*(l+1))
			lowerBound := uint64(1) << (7 * tryL)
			upperBound := uint64(1) << (7 * (tryL + 1))

			if x >= lowerBound && x < upperBound {
				bestX = x
				data = data[tryL:] // consume remainder bytes
				found = true
				break
			}
		}
	}

	if !found {
		return 0, errors.New("invalid U64 encoding")
	}

	return bestX, nil
}

func (d *Decoder) IdentifyLength(byteValue byte) (uint8, error) {
	cLog(Cyan, "Identifying length")
	cLog(Yellow, fmt.Sprintf("Byte Value: %v", byteValue))
	// l = 0：0 -> 數值範圍將會是 0 <= x < 128
	// l = 1：128 -> 10000000 -> 代表要再讀取 1 個 byte
	// l = 2：192 -> 11000000 -> 代表要再讀取 2 個 byte
	// l = 3：224 -> 11100000 -> 代表要再讀取 3 個 byte
	// l = 4：240 -> 11110000 -> 代表要再讀取 4 個 byte
	// l = 5：248 -> 11111000 -> 代表要再讀取 5 個 byte
	// l = 6：252 -> 11111100 -> 代表要再讀取 6 個 byte
	// l = 7：254 -> 11111110 -> 代表要再讀取 7 個 byte
	// byte = 255 -> x < 2^64

	if byteValue < 128 {
		return 0, nil
	}

	if byteValue >= 128 && byteValue < 192 {
		return 1, nil
	}

	if byteValue >= 192 && byteValue < 224 {
		return 2, nil
	}

	if byteValue >= 224 && byteValue < 240 {
		return 3, nil
	}

	if byteValue >= 240 && byteValue < 248 {
		return 4, nil
	}

	if byteValue >= 248 && byteValue < 252 {
		return 5, nil
	}

	if byteValue >= 252 && byteValue < 254 {
		return 6, nil
	}

	if byteValue >= 254 && byteValue < 255 {
		return 7, nil
	}

	if byteValue == 255 {
		return 8, nil
	}

	return 0, fmt.Errorf("Failed to identify length: %v", byteValue)
}

func (d *Decoder) DecodeLength() (uint64, error) {
	// Read the first byte to know the lengthFlag of the slice
	lengthFlag, err := d.ReadLegnthFlag()
	if err != nil {
		return 0, err
	}
	cLog(Yellow, fmt.Sprintf("Length Flag: %v", lengthFlag))

	// IdentifyLength will return how many bytes you have to read after the
	// lengthFlag
	remaningByteSize, err := d.IdentifyLength(lengthFlag)
	if err != nil {
		return 0, err
	}
	cLog(Yellow, fmt.Sprintf("Remaining Byte Size: %v", remaningByteSize))

	// Read lengthBytes bytes to know the length of the slice
	// first byte + remaining bytes
	remainingLengthBytes := make([]byte, remaningByteSize)

	// Avoid getting an error if the cursor is at the end of the buffer
	if remaningByteSize != 0 {
		_, err = d.buf.Read(remainingLengthBytes)
		if err != nil {
			return 0, err
		}
	}

	// Concatenate the lengthFlag and the remaining bytes
	lengthBytesValue := append([]byte{lengthFlag}, remainingLengthBytes...)

	// Decode length
	length, err := d.DecodeUint(lengthBytesValue)
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
