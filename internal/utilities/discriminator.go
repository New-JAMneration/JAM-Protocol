package utilities

import "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"

// LensElementPair will return the length of the slice and the input data itself
// input will be a slice
func LensElementPair(input []any) (int, []any) {
	// Equation C.8
	return len(input), input
}

// EmptyOrPair will return 0 if the input is empty(length == 0) and (1, input) otherwise
// 注意 : 傳進來要使用 pointer
func EmptyOrPair(input any) (int, any) {
	// Equation C.9
	// 只需判斷有 len() 的 input，如 slice , map
	// 無 len() 的 datatype 會被初始花為 0
	switch input := input.(type) {
	case []any:
		if len(input) == 0 {
			return 0, nil
		}
	case jam_types.ByteSequence:
		if len(input) == 0 {
			return 0, nil
		}
	case string:
		if len(input) == 0 {
			return 0, nil
		}
	default:
		return 1, input
	}
	return 1, input

}
