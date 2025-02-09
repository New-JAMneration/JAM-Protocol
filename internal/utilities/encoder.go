package utilities

import (
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	jamtests_reports "github.com/New-JAMneration/JAM-Protocol/jamtests/reports"
	jamtests_safrole "github.com/New-JAMneration/JAM-Protocol/jamtests/safrole"
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

var limitSizeArrayTypeList = []reflect.Type{
	reflect.TypeOf([]types.BandersnatchPublic{}),
	reflect.TypeOf([]types.Judgement{}),
	reflect.TypeOf(types.TicketsMark{}),
	reflect.TypeOf(types.ActivityRecords{}),
	reflect.TypeOf(types.ValidatorsData{}),
	reflect.TypeOf([]types.TicketBody{}),
	reflect.TypeOf(types.AvailabilityAssignments{}),
	reflect.TypeOf(types.AuthPools{}),
	reflect.TypeOf(types.AuthQueue{}),
	reflect.TypeOf(types.AuthQueues{}),
}

func typeInList(t reflect.Type, typeList []reflect.Type) bool {
	for _, v := range typeList {
		if t == v {
			return true
		}
	}
	return false
}

type Encoder struct {
	output []byte
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(input interface{}) ([]byte, error) {
	v := reflect.ValueOf(input)

	if err := e.encodeValue(v); err != nil {
		return nil, err
	}

	return e.output, nil
}

func (e *Encoder) encodeValue(v reflect.Value) error {
	cLog(Cyan, fmt.Sprintf("Value type: %v", v.Type()))

	// ------------------------------
	// Struct
	// ------------------------------
	if e.isStruct(v) {
		cLog(Magenta, "Struct")

		if e.isTicketsOrKeys(v) {
			cLog(Magenta, "TicketsOrKeys")

			if err := e.encodeTicketsOrKeys(v.Interface()); err != nil {
				return err
			}

			return nil
		}

		if e.isSafroleOutput(v) {
			cLog(Magenta, "SafroleOutput")

			if err := e.encodeSafroleOutput(v.Interface()); err != nil {
				return err
			}

			return nil
		}

		if e.isReportsOutput(v) {
			cLog(Magenta, "ReportsOutput")

			if err := e.encodeReportsOutput(v.Interface()); err != nil {
				return err
			}

			return nil
		}

		// If the type is struct, encode each field
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if err := e.encodeValue(field); err != nil {
				return err
			}
		}

		return nil
	}

	// ------------------------------
	// Nil
	// ------------------------------
	if e.isNil(v) {
		cLog(Magenta, "Nil")
		return nil
	}

	// ------------------------------
	// Bool
	// ------------------------------
	if e.isBool(v) {
		cLog(Magenta, "Bool")
		e.encodeBool(v.Interface())
		return nil
	}

	// ------------------------------
	// Uint
	// ------------------------------
	if e.isU8(v) {
		cLog(Magenta, "Uint8")
		e.encodeInt(v.Interface())
		return nil
	}

	if e.isU16(v) {
		cLog(Magenta, "Uint16")
		e.encodeInt(v.Interface())
		return nil
	}

	if e.isU32(v) {
		cLog(Magenta, "Uint32")
		e.encodeInt(v.Interface())
		return nil
	}

	if e.isU64(v) {
		cLog(Magenta, "Uint64")
		e.encodeInt(v.Interface())
		return nil
	}

	// ------------------------------
	// String
	// ------------------------------
	if v.Kind() == reflect.String {
		cLog(Magenta, "String")

		return nil
	}

	// ------------------------------
	// Slice (variable length)
	// ------------------------------
	if e.isU8Slice(v) {
		if e.isBitField(v) {
			cLog(Magenta, "BitField")
			e.encodeBitField(v.Interface())
			return nil
		}

		cLog(Magenta, "U8 Slice")
		e.encodeU8Slice(v.Interface())
		return nil
	}

	if e.isSlice(v) {
		cLog(Magenta, "Slice")

		// If the slice is in the limitSizeArrayTypeList, don't encode the length of
		// the slice.
		if !typeInList(v.Type(), limitSizeArrayTypeList) {
			cLog(Yellow, "Not in the limitSizeArrayTypeList")

			// Calculate length of the slice
			e.encodeLength(uint64(v.Len()))
		}

		// Get element in the slice and encode it
		for i := 0; i < v.Len(); i++ {
			element := v.Index(i)
			if err := e.encodeValue(element); err != nil {
				return err
			}
		}

		return nil
	}

	// ------------------------------
	// Uint8 Array (Golang byte array)
	// OpaqueHash, PublicKey, Signature, Metadata
	// Fixed length array
	// ------------------------------
	if e.isU8Array(v) {
		cLog(Magenta, "Uint8 Array")
		e.encodeU8Array(v.Interface())
		return nil
	}

	// ------------------------------
	// Array
	// ------------------------------
	if e.isArray(v) {
		cLog(Magenta, "Array")
		e.encodeArray(v)
		return nil
	}

	// ------------------------------
	// Pointer
	// ------------------------------
	if e.isPointer(v) {
		cLog(Magenta, "Pointer")

		// Empty discriminator
		if v.IsNil() {
			prefix := []byte{0}
			e.output = append(e.output, prefix...)
			cLog(Yellow, fmt.Sprintf("Pointer is nil: %v", prefix))
		} else {
			prefix := []byte{1}
			e.output = append(e.output, prefix...)
			cLog(Yellow, fmt.Sprintf("Pointer is not nil: %v", prefix))

			value := v.Elem()
			if err := e.encodeValue(value); err != nil {
				return err
			}
		}

		return nil
	}

	// ------------------------------
	// Map
	// ------------------------------
	if e.isMap(v) {
		cLog(Magenta, "Map")

		// if the map is types.WorkExecResult
		if v.Type() == reflect.TypeOf(types.WorkExecResult{}) {
			if err := e.encodeWorkExecResult(v.Interface()); err != nil {
				return err
			}
		}

		return nil
	}

	return fmt.Errorf("Unsupported type: %v", v.Type())
}

func (e *Encoder) encodeArray(v reflect.Value) {
	// Get element in the array
	for i := 0; i < v.Len(); i++ {
		element := v.Index(i)
		e.encodeValue(element)
	}
}

func (e *Encoder) isNil(v reflect.Value) bool {
	return v.Kind() == reflect.Invalid
}

func (e *Encoder) isBool(v reflect.Value) bool {
	return v.Kind() == reflect.Bool
}

func (e *Encoder) isPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr
}

func (e *Encoder) isU8(v reflect.Value) bool {
	return v.Kind() == reflect.Uint8
}

func (e *Encoder) isU16(v reflect.Value) bool {
	return v.Kind() == reflect.Uint16
}

func (e *Encoder) isU32(v reflect.Value) bool {
	return v.Kind() == reflect.Uint32
}

func (e *Encoder) isU64(v reflect.Value) bool {
	return v.Kind() == reflect.Uint64
}

func (e *Encoder) isStruct(v reflect.Value) bool {
	return v.Kind() == reflect.Struct
}

func (e *Encoder) isSlice(v reflect.Value) bool {
	return v.Kind() == reflect.Slice
}

func (e *Encoder) isArray(v reflect.Value) bool {
	return v.Kind() == reflect.Array
}

func (e *Encoder) isU8Array(v reflect.Value) bool {
	return v.Kind() == reflect.Array && v.Type().Elem().Kind() == reflect.Uint8
}

func (e *Encoder) isU8Slice(v reflect.Value) bool {
	return v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8
}

func (e *Encoder) isBitField(v reflect.Value) bool {
	if e.isU8Slice(v) {
		return v.Type() == reflect.TypeOf([]byte{})
	}
	return false
}

func (e *Encoder) isTicketsOrKeys(v reflect.Value) bool {
	return v.Type() == reflect.TypeOf(types.TicketsOrKeys{})
}

func (e *Encoder) isSafroleOutput(v reflect.Value) bool {
	return v.Type() == reflect.TypeOf(jamtests_safrole.SafroleOutput{})
}

func (e *Encoder) isReportsOutput(v reflect.Value) bool {
	return v.Type() == reflect.TypeOf(jamtests_reports.ReportsOutput{})
}

func (e *Encoder) isMap(v reflect.Value) bool {
	return v.Kind() == reflect.Map
}

// encode types.WorkExecResult
func (e *Encoder) encodeWorkExecResult(value interface{}) error {
	v := reflect.ValueOf(value)
	mapKey := reflect.ValueOf(value).MapKeys()[0]
	resultIndex := 0

	switch mapKey.Interface().(types.WorkExecResultType) {
	case "ok":
		resultIndex = 0
	case "out-of-gas":
		resultIndex = 1
	case "panic":
		resultIndex = 2
	case "bad-exports":
		resultIndex = 3
	case "bad-code":
		resultIndex = 4
	case "code-oversize":
		resultIndex = 5
	default:
		return fmt.Errorf("Unsupported WorkExecResultType: %v", mapKey.Interface().(types.WorkExecResultType))
	}

	// encode the map key
	encodedKey := e.encodeUint(uint64(resultIndex))
	e.output = append(e.output, encodedKey...)
	cLog(Yellow, fmt.Sprintf("Map key: %v", encodedKey))

	if resultIndex == 0 { // ok
		// encode the value
		value := v.MapIndex(mapKey)
		e.encodeU8Slice(value.Interface())
	}

	return nil
}

func (e *Encoder) encodeSafroleOutput(value interface{}) error {
	v := reflect.ValueOf(value)

	// Check the output is ok or err
	isOk := !v.Field(0).IsNil()
	isErr := !v.Field(1).IsNil()

	if isOk && isErr {
		return fmt.Errorf("SafroleOutput should not contain both ok and err")
	}

	if !isOk && !isErr {
		return fmt.Errorf("SafroleOutput should contain either ok or err")
	}

	if isOk {
		// append prefix 0
		prefix := []byte{0}
		e.output = append(e.output, prefix...)
		cLog(Yellow, fmt.Sprintf("Output is ok: %v", prefix))

		// encode the ok value
		okValue := v.Field(0).Elem()
		if err := e.encodeValue(okValue); err != nil {
			return err
		}

		return nil
	}

	if isErr {
		// append prefix 1
		prefix := []byte{1}
		e.output = append(e.output, prefix...)
		cLog(Yellow, fmt.Sprintf("Output is err: %v", prefix))

		// encode the error code
		errValue := v.Field(1).Elem()
		if err := e.encodeValue(errValue); err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (e *Encoder) encodeReportsOutput(value interface{}) error {
	v := reflect.ValueOf(value)

	// Check the output is ok or err
	isOk := !v.Field(0).IsNil()
	isErr := !v.Field(1).IsNil()

	if isOk && isErr {
		return fmt.Errorf("ReportsOutput should not contain both ok and err")
	}

	if !isOk && !isErr {
		return fmt.Errorf("ReportsOutput should contain either ok or err")
	}

	if isOk {
		// append prefix 0
		prefix := []byte{0}
		e.output = append(e.output, prefix...)
		cLog(Yellow, fmt.Sprintf("Output is ok: %v", prefix))

		// encode the ok value
		okValue := v.Field(0).Elem()
		if err := e.encodeValue(okValue); err != nil {
			return err
		}

		return nil
	}

	if isErr {
		// append prefix 1
		prefix := []byte{1}
		e.output = append(e.output, prefix...)
		cLog(Yellow, fmt.Sprintf("Output is err: %v", prefix))

		// encode the error code
		errValue := v.Field(1).Elem()
		if err := e.encodeValue(errValue); err != nil {
			return err
		}

		return nil
	}

	return nil
}

// OpaqueHash, PublicKey, Signature, Metadata...
// Fixed length array
func (e *Encoder) encodeU8Array(value interface{}) {
	v := reflect.ValueOf(value)

	tmp := make([]byte, v.Len())
	reflect.Copy(reflect.ValueOf(tmp), v) // copy the value to the byte array
	cLog(Yellow, fmt.Sprintf("Slice: %v", tmp))
	e.output = append(e.output, tmp...)
}

// ByteSequence
// Variable length array
func (e *Encoder) encodeU8Slice(value interface{}) {
	v := reflect.ValueOf(value)

	// Calculate length of the slice
	e.encodeLength(uint64(v.Len()))

	tmp := make([]byte, v.Len())
	reflect.Copy(reflect.ValueOf(tmp), v) // copy the value to the byte array
	e.output = append(e.output, tmp...)
	cLog(Yellow, fmt.Sprintf("Slice: %v", tmp))
}

func (e *Encoder) encodeBitField(value interface{}) {
	v := reflect.ValueOf(value)

	tmp := make([]byte, v.Len())
	reflect.Copy(reflect.ValueOf(tmp), v) // copy the value to the byte array
	cLog(Yellow, fmt.Sprintf("Bit Field: %v", tmp))
	e.output = append(e.output, tmp...)
}

// Integer encoding
func (e *Encoder) encodeInt(value interface{}) {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Uint8:
		result := e.encodeUintWithLength(v.Uint(), 1)
		cLog(Yellow, fmt.Sprintf("Uint8: %v", result))
		e.output = append(e.output, result...)
	case reflect.Uint16:
		result := e.encodeUintWithLength(v.Uint(), 2)
		cLog(Yellow, fmt.Sprintf("Uint16: %v", result))
		e.output = append(e.output, result...)
	case reflect.Uint32:
		result := e.encodeUintWithLength(v.Uint(), 4)
		cLog(Yellow, fmt.Sprintf("Uint32: %v", result))
		e.output = append(e.output, result...)
	case reflect.Uint64:
		result := e.encodeUintWithLength(v.Uint(), 8)
		cLog(Yellow, fmt.Sprintf("Uint64: %v", result))
		e.output = append(e.output, result...)
	default:
		fmt.Println("Unsupported integer type")
	}
}

// TODO: error handling
func (e *Encoder) encodeUintWithLength(value uint64, l int) []byte {
	if l == 0 {
		return []byte{}
	}

	out := make([]byte, l)
	for i := 0; i < l; i++ {
		out[i] = byte(value & 0xFF)
		value >>= 8
	}

	return out
}

// TODO: error handling
func (e *Encoder) encodeUint(value uint64) []byte {
	// If x = 0: E(x) = [0]
	if value == 0 {
		return []byte{0}
	}

	// Attempt to find l in [1..8] such that 2^(7*l) ≤ x < 2^(7*(l+1))
	for l := 0; l <= 7; l++ {
		l64 := uint(l)
		lowerBound := uint64(1) << (7 * l64)       // 2^(7*l)
		upperBound := uint64(1) << (7 * (l64 + 1)) // 2^(7*(l+1))
		if value >= lowerBound && value < upperBound {
			// Found suitable l.
			power8l := uint64(1) << (8 * l64)
			remainder := value % power8l
			floor := value / power8l

			// prefix = 2^8 - 2^(8-l) + floor(x / 2^(8*l))
			prefix := byte((256 - (1 << (8 - l64))) + floor)

			return append([]byte{prefix}, e.encodeUintWithLength(remainder, l)...)
		}
	}

	fmt.Println("No suitable l found")

	// If no suitable l found:
	// E(x) = [2^8 - 1] || E_8(x) = [255] || SerializeFixedLength(x,8)
	return append([]byte{0xFF}, e.encodeUintWithLength(value, 8)...)
}

func (e *Encoder) encodeBool(value interface{}) {
	v := reflect.ValueOf(value)

	// bool to uint64
	var result uint64
	if v.Bool() {
		result = 1
	} else {
		result = 0
	}

	resultBytes := e.encodeUint(result)
	e.output = append(e.output, resultBytes...)

	cLog(Yellow, fmt.Sprintf("Bool: %v", resultBytes))
}

func (e *Encoder) encodeTicketsOrKeys(value interface{}) error {
	v := reflect.ValueOf(value)

	isTickets := v.Field(0).Len() != 0
	isKeys := v.Field(1).Len() != 0

	if isTickets && isKeys {
		return fmt.Errorf("TicketsOrKeys should not contain both tickets and keys")
	}

	if !isTickets && !isKeys {
		return fmt.Errorf("TicketsOrKeys should contain either tickets or keys")
	}

	if isTickets {
		// append prefix 0
		prefix := []byte{0}
		e.output = append(e.output, prefix...)

		cLog(Yellow, fmt.Sprintf("is tickets: %v", prefix))

		// encode the tickets
		tickets := v.Field(0)
		if err := e.encodeValue(tickets); err != nil {
			return err
		}
	}

	if isKeys {
		// append prefix 1
		prefix := []byte{1}
		e.output = append(e.output, prefix...)

		cLog(Yellow, fmt.Sprintf("is keys: %v", prefix))

		// encode the keys
		keys := v.Field(1)
		if err := e.encodeValue(keys); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeLength(length uint64) error {
	encodedLength := e.encodeUint(length)
	e.output = append(e.output, encodedLength...)

	cLog(Yellow, fmt.Sprintf("Length: %v", encodedLength))

	return nil
}
