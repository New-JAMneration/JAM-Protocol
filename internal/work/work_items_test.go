package work_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkItemJSONSerialization(t *testing.T) {
	jsonData := `{
		"service": 16909060,
		"code_hash": "0x70a50829851e8f6a8c80f92806ae0e95eb7c06ad064e311cc39107b3219e532e",
		"payload": "0x0102030405",
		"refine_gas_limit": 42,
		"accumulate_gas_limit": 42,
		"import_segments": [
			{
				"tree_root": "0x461236a7eb29dcffc1dd282ce1de0e0ed691fc80e91e02276fe8f778f088a1b8",
				"index": 0
			},
			{
				"tree_root": "0xe7cb536522c1c1b41fff8021055b774e929530941ea12c10f1213c56455f29ad",
				"index": 1
			},
			{
				"tree_root": "0xb0a487a4adf6a0eda5d69ddd2f8b241cf44204f0ff793e993e5e553b7862a1dc",
				"index": 2
			}
		],
		"extrinsic": [
			{
				"hash": "0x381a0e351c5593018bbc87dd6694695caa1c0c1ddb24e70995da878d89495bf1",
				"len": 16
			},
			{
				"hash": "0x6c437d85cd8327f42a35d427ede1b5871347d3aae7442f2df1ff80f834acf17a",
				"len": 17
			}
		],
		"export_count": 4
	}`

	// 定義 WorkItem 結構
	type WorkItem struct {
		Service            uint32 `json:"service"`
		CodeHash           string `json:"code_hash"`
		Payload            string `json:"payload"`
		RefineGasLimit     uint64 `json:"refine_gas_limit"`
		AccumulateGasLimit uint64 `json:"accumulate_gas_limit"`
		ImportSegments     []struct {
			TreeRoot string `json:"tree_root"`
			Index    uint   `json:"index"`
		} `json:"import_segments"`
		Extrinsic []struct {
			Hash string `json:"hash"`
			Len  uint   `json:"len"`
		} `json:"extrinsic"`
		ExportCount uint `json:"export_count"`
	}

	// 測試 JSON 解碼
	var item WorkItem
	err := json.Unmarshal([]byte(jsonData), &item)
	assert.NoError(t, err, "JSON Unmarshal should not fail")

	// 驗證字段
	assert.Equal(t, uint32(16909060), item.Service)
	assert.Equal(t, "0x70a50829851e8f6a8c80f92806ae0e95eb7c06ad064e311cc39107b3219e532e", item.CodeHash)
	assert.Equal(t, "0x0102030405", item.Payload)
	assert.Equal(t, uint64(42), item.RefineGasLimit)
	assert.Equal(t, uint64(42), item.AccumulateGasLimit)
	assert.Len(t, item.ImportSegments, 3)
	assert.Equal(t, "0x461236a7eb29dcffc1dd282ce1de0e0ed691fc80e91e02276fe8f778f088a1b8", item.ImportSegments[0].TreeRoot)
	assert.Equal(t, uint(0), item.ImportSegments[0].Index)
	assert.Len(t, item.Extrinsic, 2)
	assert.Equal(t, "0x381a0e351c5593018bbc87dd6694695caa1c0c1ddb24e70995da878d89495bf1", item.Extrinsic[0].Hash)
	assert.Equal(t, uint(16), item.Extrinsic[0].Len)
	assert.Equal(t, uint(4), item.ExportCount)

	// 測試 JSON 編碼
	encodedData, err := json.Marshal(item)
	assert.NoError(t, err, "JSON Marshal should not fail")

	// 測試重新解碼後是否一致
	var newItem WorkItem
	err = json.Unmarshal(encodedData, &newItem)
	assert.NoError(t, err, "JSON Unmarshal after Marshal should not fail")
	assert.Equal(t, item, newItem, "Original and re-decoded items should match")
}
