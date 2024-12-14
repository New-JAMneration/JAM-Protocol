package utilities

// LengthDiscriminator will return the length of the slice and the input data itself
func LengthDiscriminator(input []any) (int, []any) {
	// Equation C.8
	return len(input), input
}

// ConvenientDiscriminator will
func ConvenientDiscriminator(input *any) (int, any) {
	if input == nil {
		return 0, nil
	}
	return 1, input
}
