// Package safemath provides overflow-protected integer arithmetic helpers.
//
// Used by ServiceAccount counter (items/octets) maintenance in
// Insert/Delete methods and by ThresholdBalance-style computations where
// silent wrap-around would corrupt state.
package safemath

import "errors"

// ErrOverflow indicates that an arithmetic operation overflowed/underflowed.
var ErrOverflow = errors.New("integer overflow")

// Integer covers every signed and unsigned integer kind (including types
// with a custom underlying integer type, via ~).
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Add returns a + b and true on success, or (0, false) on overflow.
func Add[T Integer](a, b T) (T, bool) {
	sum := a + b

	// (^T(0)) < 0 is true when T is a signed integer kind.
	if (^T(0)) < 0 {
		// Signed: overflow iff both operands have the same sign but sum has
		// the opposite sign.
		if (a^sum) < 0 && (b^sum) < 0 {
			return 0, false
		}
	} else {
		// Unsigned: overflow iff the sum wrapped around below either operand.
		if sum < a {
			return 0, false
		}
	}

	return sum, true
}

// Sub returns a - b and true on success, or (0, false) on overflow / underflow.
func Sub[T Integer](a, b T) (T, bool) {
	diff := a - b

	if (^T(0)) < 0 {
		// Signed: overflow iff operands differ in sign and the result differs
		// in sign from a.
		if (a^b) < 0 && (a^diff) < 0 {
			return 0, false
		}
	} else {
		// Unsigned: underflow iff a < b.
		if a < b {
			return 0, false
		}
	}

	return diff, true
}

// Mul returns a * b and true on success, or (0, false) on overflow.
func Mul[T Integer](a, b T) (T, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}

	result := a * b

	if (^T(0)) < 0 {
		// Signed: reverse-divide check.
		if result/b != a {
			return 0, false
		}

		// Special case: MinInt * -1.
		minusOne := ^T(0)
		if (a == minusOne && b == -b) || (b == minusOne && a == -a) {
			return 0, false
		}
	} else {
		// Unsigned: reverse-divide check.
		if result/b != a {
			return 0, false
		}
	}

	return result, true
}
