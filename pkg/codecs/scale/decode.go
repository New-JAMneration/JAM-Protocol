package scale

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/scale_bytes"
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/types"
	"runtime/debug"
)

func Decode(typeStr string, source interface{}, data interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			fmt.Printf("Error decode: %v\nStack: %s\n", r, stack)
			err = errors.New("Error decode " + fmt.Sprint(r))
		}
	}()

	t, err := types.GetType(typeStr)
	if err != nil {
		return nil, err
	}

	b, err := scale_bytes.NewBytes(source)
	if err != nil {
		return nil, err
	}

	result, err = t.Process(b)
	if err != nil {
		return nil, err
	}

	if data != nil {
		jsonData, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonData, &data)
		if err != nil {
			return nil, err
		}
	}

	return result, err
}
