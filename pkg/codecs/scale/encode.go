package scale

import (
	"errors"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/types"
	"reflect"
)

func Encode(typeStr string, value interface{}) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Error encoding " + fmt.Sprint(r))
		}
	}()

	t, err := types.GetType(typeStr)
	if err != nil {
		return nil, err
	}

	m := structToMap(value)

	data, err = t.ProcessEncode(m)
	return data, err
}

func structToMap(obj interface{}) interface{} {
	val := reflect.ValueOf(obj)

	// 如果是指針，則取其指向的值
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		result := make(map[string]interface{})

		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			if !val.Field(i).CanInterface() {
				continue
			}

			key := field.Name
			if tag, ok := field.Tag.Lookup("json"); ok {
				tagParts := parseTag(tag)
				if tagParts[0] != "" {
					key = tagParts[0]
				}
				if tagParts[0] == "-" {
					continue
				}
			}

			value := val.Field(i)
			result[key] = convertValue(value)
		}
		return result
	} else if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		var result []interface{}

		for i := 0; i < val.Len(); i++ {
			element := val.Index(i)
			result = append(result, convertValue(element))
		}
		return result
	}

	result := make(map[string]interface{})
	result["value"] = convertValue(val)
	return result
}

func convertValue(val reflect.Value) interface{} {
	if !val.IsValid() {
		return nil
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return convertValue(val.Elem())
	case reflect.Struct:
		return structToMap(val.Interface())
	case reflect.Slice, reflect.Array:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return val.Bytes()
		}
		var slice []interface{}
		for i := 0; i < val.Len(); i++ {
			slice = append(slice, convertValue(val.Index(i)))
		}
		return slice
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return val.Interface()
		}
		m := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			m[key.String()] = convertValue(val.MapIndex(key))
		}
		return m
	default:
		return val.Interface()
	}
}

func parseTag(tag string) []string {
	var parts []string
	for i, char := range tag {
		if char == ',' {
			parts = append(parts, tag[:i])
			parts = append(parts, tag[i+1:])
			return parts
		}
	}
	parts = append(parts, tag)
	return parts
}
