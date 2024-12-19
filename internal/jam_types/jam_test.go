package jam_types

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func TestWorkResult(t *testing.T) {
	InitScaleRegistry()

	t.Run("work_result_0.bin", func(t *testing.T) {
		workerResultScale("./test_data/work_result_0.bin", t)
	})

	t.Run("work_result_1.bin", func(t *testing.T) {
		workerResultScale("./test_data/work_result_1.bin", t)
	})
}

func workerResultScale(filePath string, t *testing.T) {
	data, err := readFile(filePath)
	if err != nil {
		t.Error(err)
		return
	}

	w := &WorkResult{}

	if err := w.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := w.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestHeader(t *testing.T) {
	InitScaleRegistry()

	t.Run("header_0.bin", func(t *testing.T) {
		headerScale("./test_data/header_0.bin", t)
	})

	t.Run("header_1.bin", func(t *testing.T) {
		headerScale("./test_data/header_1.bin", t)
	})
}

func headerScale(filePath string, t *testing.T) {
	data, err := readFile(filePath)
	if err != nil {
		t.Error(err)
		return
	}

	h := &Header{}

	if err := h.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := h.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestAssurancesExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/assurances_extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	a := &AssurancesExtrinsic{}

	if err := a.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := a.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestDisputesExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/disputes_extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	a := &DisputesExtrinsic{}

	if err := a.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := a.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	e := &Extrinsic{}

	if err := e.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := e.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestTicketsExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/tickets_extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	te := &TicketsExtrinsic{}

	if err := te.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := te.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestPreimagesExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/preimages_extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &PreimagesExtrinsic{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestGuaranteesExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/guarantees_extrinsic.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &GuaranteesExtrinsic{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestRefineContextExtrinsic(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/refine_context.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &RefineContext{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestWorkItem(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/work_item.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &WorkItem{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestWorkPackage(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/work_package.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &WorkPackage{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}

func TestWorkReport(t *testing.T) {
	InitScaleRegistry()

	data, err := readFile("./test_data/work_report.bin")
	if err != nil {
		t.Error(err)
		return
	}

	p := &WorkReport{}

	if err := p.ScaleDecode(data); err != nil {
		t.Error(err)
		return
	}

	encode, err := p.ScaleEncode()
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(encode, data) {
		t.Error("Encode failed")
	}
}
