package types

import (
	"fmt"
	"strings"
)

var Registry = map[string]func() IType{
	"i8":           NewI8,
	"i16":          NewI16,
	"i32":          NewI32,
	"i64":          NewI64,
	"u8":           NewU8,
	"u16":          NewU16,
	"u32":          NewU32,
	"u64":          NewU64,
	"bool":         NewBool,
	"compact":      NewCompact,
	"compact<u32>": NewCompactU32,
	"hexbytes":     NewHexBytes,
	"null":         NewNull,
}

func GetType(typeString string) (IType, error) {

	info, err := extractFirstLayer(typeString)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(info.Wrapper) {
	case "vec":
		return NewVec(info.InnerType), nil
	case "option":
		return NewOption(info.InnerType), nil
	}

	if f, ok := Registry[strings.ToLower(typeString)]; ok {
		return f(), nil
	}

	return nil, fmt.Errorf("type not found, typeString: %s", typeString)
}

func RegisterType(m map[string]func() IType) {
	for key, f := range m {
		Registry[strings.ToLower(key)] = f
	}
}

type typeInfo struct {
	Wrapper   string
	InnerType string
}

func extractFirstLayer(typeStr string) (typeInfo, error) {
	var info typeInfo

	typeStr = strings.TrimSpace(typeStr)

	start := strings.Index(typeStr, "<")
	if start == -1 {
		info.Wrapper = ""
		info.InnerType = typeStr
		return info, nil
	}

	info.Wrapper = strings.TrimSpace(typeStr[:start])

	count := 0
	end := -1
	for i := start; i < len(typeStr); i++ {
		if typeStr[i] == '<' {
			count++
		} else if typeStr[i] == '>' {
			count--
			if count == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return info, fmt.Errorf("未找到匹配的 '>' 字符")
	}

	info.InnerType = strings.TrimSpace(typeStr[start+1 : end])

	return info, nil
}
