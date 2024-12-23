package recent_history

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	"testing"
)

func TestAddToBetaDagger(t *testing.T) {
	// 创建一个 State 实例 state1
	state1 := State{
		Beta: []jam_types.BlockInfo{
			{HeaderHash: jam_types.HeaderHash{0x01}}, // 初始化 HeaderHash
			{HeaderHash: jam_types.HeaderHash{0x02}},
		},
	}

	// 创建另一个 State 实例 state2
	state2 := State{
		Beta: []jam_types.BlockInfo{
			{HeaderHash: jam_types.HeaderHash{0x03}},
			{HeaderHash: jam_types.HeaderHash{0x04}},
		},
	}

	// 调用 AddToBetaDagger 方法
	state1.AddToBetaDagger(state2)

	// 验证 state1 的 BetaDagger 是否包含 state2 的 Beta
	if len(state1.BetaDagger) != len(state2.Beta) {
		t.Errorf("Expected BetaDagger length %d, got %d", len(state2.Beta), len(state1.BetaDagger))
	}

	// 验证每个 HeaderHash 是否正确
	for i, block := range state2.Beta {
		if state1.BetaDagger[i].HeaderHash != block.HeaderHash {
			t.Errorf("Expected HeaderHash %v, got %v", block.HeaderHash, state1.BetaDagger[i].HeaderHash)
		}
	}
}
