package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/logger"
)

// DecodeJSONByte decodes a JSON-encoded byte array
func DecodeJSONByte(input []byte) []byte {
	toJSON, _ := json.Marshal(input)
	out := string(toJSON)[1 : len(string(toJSON))-1]
	return hexToBytes(out)
}

// hexToBytes converts a hex string (with 0x prefix) to bytes
func hexToBytes(hexString string) []byte {
	bytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		logger.Errorf("failed to decode hex string: %v", err)
	}
	return bytes
}

// parseFixedByteArray parses a JSON byte array with a fixed expected length.
// It supports both hex string format ("0x...") and array format ([1,2,3,...]).
func parseFixedByteArray(data []byte, expectedLen int) ([]byte, error) {
	// Peek at first byte to see if it starts with '[' or '"'
	if len(data) > 0 && data[0] == '[' {
		arr, err := parseNormalByteArray(data, expectedLen)
		if err != nil {
			return nil, err
		}
		return arr, nil
	}

	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return nil, err
	}

	if len(hexStr) < 2 || hexStr[:2] != "0x" {
		return nil, fmt.Errorf("invalid hex format: %s", hexStr)
	}

	decoded, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return nil, err
	}

	if len(decoded) != expectedLen {
		return nil, fmt.Errorf("invalid length: expected %v bytes, got %v", expectedLen, len(decoded))
	}

	return decoded, nil
}

// parseNormalByteArray parses a JSON array of bytes
func parseNormalByteArray(data []byte, size int) ([]byte, error) {
	// Peek at first byte to see if it starts with '[' or '"'
	if len(data) > 0 && data[0] == '[' {
		// Data is an array like [0,255,34,...]
		var arr []byte
		if err := json.Unmarshal(data, &arr); err != nil {
			return arr, err
		}
		if len(arr) != size {
			return nil, fmt.Errorf("invalid length: expected %d bytes, got %d", size, len(arr))
		}
		return arr, nil
	}

	return nil, fmt.Errorf("invalid format for parseNormalByteArray: expected %d bytes, got %d", size, len(data))
}
