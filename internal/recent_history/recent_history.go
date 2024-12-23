package recent_history

import (
	"bytes"
	"sort"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
	u "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	merkle "github.com/New-JAMneration/JAM-Protocol/internal/utilities/merkle_tree"
)

type State struct {
	Beta       []jamTypes.BlockInfo // prior state
	BetaDagger []jamTypes.BlockInfo // intermediate state
	BetaPrime  []jamTypes.BlockInfo // posterior state

}

// \mathbf{C} in GP from type B (12.15)
type BeefyCommitmentOutput []AccumulationOutput

// instant-used struct
type AccumulationOutput struct {
	jamTypes.ServiceId
	jamTypes.OpaqueHash
}

var maxBlocksHistory = jamTypes.MaxBlocksHistory

// remove duplicated blocks by BlockHash
func (s *State) RemoveDuplicate(headerhash jamTypes.HeaderHash) bool {
	for _, block := range s.Beta {
		if bytes.Equal(block.HeaderHash, headerhash) {
			return true
		}
	}
	return false
}

// beta^dagger (7.2)
func (s *State) AddToBetaDagger(beta State) {
	if len(beta.Beta) != maxBlocksHistory {
		for i := 0; i < len(beta.Beta)-1; i++ {
			block := beta.Beta[i]
			s.BetaDagger = append(s.BetaDagger, block)
		}
	} else {
		// duplicate beta into beta^dagger
		s.BetaDagger = append(s.BetaDagger, beta.Beta...)
	}

	// 确保 BetaDagger 不超过 maxBlocksHistory
	if len(s.BetaDagger) > maxBlocksHistory {
		// 删除最旧的元素，保持长度为 maxBlocksHistory
		s.BetaDagger = s.BetaDagger[len(s.BetaDagger)-maxBlocksHistory:]
	}
}

func (s *State) BetaDaggerToBetaPrime(betaDagger jamTypes.BlockInfo, n jamTypes.BlockInfo) {
	merkle.M()
	u.SerializeFixedLength()
	u.SerializeU64()
	hash.KeccakHash()
	hash.Blake2bHash()

}

// accumulation-result tree root $r$ (7.3)
func r(c BeefyCommitmentOutput) (output jamTypes.OpaqueHash) {
	serviceid := make([]comparable, 0, len(c.Value))
	for h := range c.Value {
		serviceid = append(serviceid, h)
	}
	sort.Slice(serviceid, func(i, j int) bool {
		return serviceid[i].Less(serviceid[j])
	})
	for s, h in c {
		ss := s.serialized()
		s4 := u.SerializeU64(s)
		s4.append(u.SerializeFixedLength(h))
	}
	merkle.Mb()
	return output
}

// Merkle Mountain Range $\mathbf{b}$
func b() {
	// MMR append func A
}

func p(Eg jamTypes.GuaranteesExtrinsic) {

}

func n(p, h jamTypes.OpaqueHash, b, s jamTypes.OpaqueHash) {

}
