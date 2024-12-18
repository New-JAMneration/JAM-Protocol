package jam_types

import (
	"github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale"
	"io"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	InitScaleRegistry()

	filePath := "./test_data/work_result_0.bin"

	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		t.Error(err)
		return
	}

	var p WorkResult

	result, err := scale.Decode("workresult", raw, &p)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
	t.Log(p)

	encode, err := scale.Encode("workresult", p)
	if err != nil {
		t.Error(err)
		return
	}

	//t.Log(encode)
	//if !bytes.Equal(encode, raw) {
	//	t.Error("Encode failed")
	//}

	for i, b := range encode {
		if b != raw[i] {
			t.Error("Encode failed", i)
		}
	}
}
