package scale

import (
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/types"
	"reflect"
)

func Encode(typeStr string, value interface{}) ([]byte, error) {
	t, err := types.GetType(typeStr)
	if err != nil {
		return nil, err
	}

	m := structToMap(value)

	return t.ProcessEncode(m)
}

func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
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

		value := val.Field(i).Interface()
		result[key] = convertValue(reflect.ValueOf(value))
	}

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
