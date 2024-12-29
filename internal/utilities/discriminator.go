package utilities

import (
	"fmt"
	"reflect"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type Discriminator struct {
	Value []Serializable
}

// C.1.4. Discriminator Encoding.
func (d Discriminator) Serialize() types.ByteSequence {
	length := types.U64(len(d.Value))
	return append(WrapU64(length).Serialize(), SerializableSequence(d.Value).Serialize()...)

}

// LensElementPair will return the length of the slice and the input data itself
// input will be a slice
func LensElementPair[T any](input []T) (int, []T) {
	// Equation C.8
	return len(input), input
}

// EmptyOrPair will return the length of the slice and the input data itself
// if return 2, it means the input type is not supported
func EmptyOrPair(input interface{}) (int, any) {
	// Equation C.9
	if input == nil {
		return 0, nil
	}

	value := reflect.ValueOf(input)

	switch value.Kind() {
	case reflect.Struct:
		if value.NumField() == 0 {
			return 0, nil
		}
		// check all the fields in struct
		for i := 0; i < value.NumField(); i++ {
			fmt.Println(value.Field(i).Kind())
			f := value.Field(i)
			if f.Kind() == reflect.Slice && f.Len() != 0 {
				return 1, input
			} else if f.Kind() == reflect.Array {
				return 1, input
			} else if f.Kind() == reflect.Ptr {
				if value.Field(i).IsNil() {
					return 0, nil
				}
				elem := value.Field(i).Elem()
				if elem.Kind() == reflect.Slice && elem.Len() != 0 {
					return 1, input
				}
			} else {
				err := fmt.Errorf("input type is not supported currently")
				return 2, err
			}
		}
		return 0, nil
	case reflect.Slice:
		if value.Len() == 0 {

			return 0, nil
		}
	case reflect.Pointer:
		if value.IsNil() {
			fmt.Println("input is nil")
			return 0, nil
		}
		elem := value.Elem()
		if elem.Kind() == reflect.Slice && elem.Len() == 0 {
			return 0, nil
		}
	default:
		err := fmt.Errorf("input type is not supported ")
		return 2, err
	}
	return 1, input
}
