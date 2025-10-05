package merklization

import (
	"bytes"
	"errors"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

const NODE_SIZE = 512

type LeafPath struct {
	Pair types.StateKeyVal
	Path []types.OpaqueHash // hashes from root to leaf
}

// CE129Request: client → server
type CE129Request struct {
	HeaderHash types.OpaqueHash
	StartKey   types.StateKey
	EndKey     types.StateKey
	MaxSize    int
}

// CE129Response: server → client
type CE129Response struct {
	StartPath []types.OpaqueHash
	EndPath   []types.OpaqueHash
	Pairs     []types.StateKeyVal
}

// TODO: this is just a temp example for DB root → []LeafPath
var GlobalMerklePathMap = make(map[types.StateRoot][]LeafPath)

// TODO: review the exact implement of the CE129 handler after the DB design is done
func CE129Handler(request CE129Request) (CE129Response, error) {
	LeafPaths, ok := GlobalMerklePathMap[types.StateRoot(request.HeaderHash)]
	if !ok {
		return CE129Response{}, errors.New("header hash not found")
	}

	result := CE129Response{}
	var startPath, endPath []types.OpaqueHash
	foundStart := false

	for _, leafPath := range LeafPaths {
		key := leafPath.Pair.Key
		// TODO: check the actual logic of the range compare
		if bytes.Compare(key[:], request.StartKey[:]) >= 0 &&
			bytes.Compare(key[:], request.EndKey[:]) < 0 {
			result.Pairs = append(result.Pairs, leafPath.Pair)
			if !foundStart {
				startPath = leafPath.Path
				foundStart = true
			}
			endPath = leafPath.Path
		}
	}
	result.StartPath = startPath
	result.EndPath = endPath
	// TODO: Check the actual logic of handling the MaxSize
	if len(result.Pairs) > request.MaxSize {
		result.Pairs = result.Pairs[:request.MaxSize]
	}
	// TODO: check the actual logic of empty result case
	return result, nil
}

// bytesToBits converts a byte sequence to a bit sequence.
// The function defined in graypaper 3.7.3. Boolean Values
func bytesToBits(s types.ByteSequence) types.BitSequence {
	// Initialize the bit sequence.
	bitSequence := types.BitSequence{}

	// Iterate over the byte sequence.
	for _, byte := range s {
		// Iterate over the bits in the byte.
		for i := 0; i < 8; i++ {
			// Calculate the bit.
			bit := ((byte >> uint(7-i)) & 1)

			// Append the bit (bool) to the bit sequence.
			bitSequence = append(bitSequence, bit == 1)
		}
	}

	// Return the bit sequence.
	return bitSequence
}

// bitsToBytes converts a BitSequence to a ByteSequence
func bitsToBytes(bits types.BitSequence) (types.ByteSequence, error) {
	if len(bits)%8 != 0 {
		return nil, errors.New("bit sequence length must be a multiple of 8")
	}

	byteLength := len(bits) / 8
	bytes := make(types.ByteSequence, byteLength)

	for i, bit := range bits {
		if bit {
			bytes[i/8] |= 1 << uint(7-i%8)
		}
	}

	return bytes, nil
}

// BranchEncoding encodes a branch node.
func BranchEncoding(left, right types.OpaqueHash) types.BitSequence {
	branchPrefixBit := types.BitSequence{false}
	leftBits := bytesToBits(left[:])[1:]
	rightBits := bytesToBits(right[:])

	encoding := types.BitSequence{}
	encoding = append(encoding, branchPrefixBit...) // 1 bit
	encoding = append(encoding, leftBits...)        // 255 bits
	encoding = append(encoding, rightBits...)       // 256 bits

	return encoding // 512 bits
}

func embeddedValueLeaf(key types.StateKey, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                       // 1 bit
	embeddedValueLeafPrefixBit := types.BitSequence{false}         // 1 bit
	prefix := append(leftPrefixBit, embeddedValueLeafPrefixBit...) // 2 bits

	valueSize := types.U32(len(value))
	serializedValueSize := utilities.SerializeFixedLength(valueSize, 1)
	valueSizeBits := bytesToBits(serializedValueSize)

	encoding := types.BitSequence{}
	encoding = append(encoding, prefix...)
	encoding = append(encoding, valueSizeBits[2:]...)
	encoding = append(encoding, bytesToBits(key[:])...)
	encoding = append(encoding, bytesToBits(value)...)

	// Calculate the size, if it has space left, fill it with 0
	if len(encoding) < NODE_SIZE {
		encoding = append(encoding, make(types.BitSequence, NODE_SIZE-len(encoding))...)
	}

	return encoding
}

func regularLeaf(key types.StateKey, value types.ByteSequence) types.BitSequence {
	leftPrefixBit := types.BitSequence{true}                                    // 1 bit
	regularLeafPrefixBit := types.BitSequence{true}                             // 1 bit
	fillZeroBits := types.BitSequence{false, false, false, false, false, false} // 6 bits

	encoding := types.BitSequence{}
	encoding = append(encoding, leftPrefixBit...)
	encoding = append(encoding, regularLeafPrefixBit...)
	encoding = append(encoding, fillZeroBits...)
	encoding = append(encoding, bytesToBits(key[:])...)
	valueHash := hash.Blake2bHash(value)
	encoding = append(encoding, bytesToBits(valueHash[:])...)

	return encoding
}

func LeafEncoding(key types.StateKey, value types.ByteSequence) types.BitSequence {
	if len(value) <= 32 {
		return embeddedValueLeaf(key, value)
	} else {
		return regularLeaf(key, value)
	}
}

// Convert a types.BitSequence to a string
func bitSequenceToString(bitSequence types.BitSequence) string {
	str := ""
	for _, bit := range bitSequence {
		if bit {
			str += "1"
		} else {
			str += "0"
		}
	}
	return str
}

// INFO: Convert the BitSequence to a bitstrings, because we cannot use []bool as a
// key in a map
type MerklizationInput map[string]types.StateKeyVal

func Merklization(d MerklizationInput) (types.OpaqueHash, []LeafPath) {
	if len(d) == 0 {
		// zero hash
		return types.OpaqueHash{}, nil
	}

	// FIXME: 為什麼 graypaper 要寫 {(k, v)}, 而不是判斷長度？
	if len(d) == 1 {
		for _, stateKeyVal := range d {
			leftEncoding := LeafEncoding(stateKeyVal.Key, stateKeyVal.Value)
			bytes, _ := bitsToBytes(leftEncoding)
			leafHash := hash.Blake2bHash(bytes)
			return leafHash, []LeafPath{{
				Pair: stateKeyVal,
				Path: []types.OpaqueHash{leafHash},
			}}
		}
	}

	l := make(MerklizationInput)
	r := make(MerklizationInput)
	for key, value := range d {
		isLeft := key[0] == '0'
		if isLeft {
			l[key[1:]] = value
		}

		isRight := key[0] == '1'
		if isRight {
			r[key[1:]] = value
		}
	}
	leftHash, leftPaths := Merklization(l)
	rightHash, rightPaths := Merklization(r)
	branchEncoding := BranchEncoding(leftHash, rightHash)
	bytes, _ := bitsToBytes(branchEncoding)
	nodeHash := hash.Blake2bHash(bytes)
	for i := range leftPaths {
		leftPaths[i].Path = append(leftPaths[i].Path, nodeHash)
	}
	for i := range rightPaths {
		rightPaths[i].Path = append(rightPaths[i].Path, nodeHash)
	}

	allPaths := append(leftPaths, rightPaths...)

	return nodeHash, allPaths
}

// basic Merklization function
// $M_{\sigma}(\sigma)$
// Input: $T(\sigma)$ a dictionary, serialized states
func MerklizationSerializedState(serializedState types.StateKeyVals) types.StateRoot {
	merklizationInput := make(MerklizationInput)

	// Convert the StateKeyVals to merklization input
	for _, stateKeyVal := range serializedState {
		key := bitSequenceToString(bytesToBits(stateKeyVal.Key[:]))
		value := types.StateKeyVal{
			Key:   stateKeyVal.Key,
			Value: stateKeyVal.Value,
		}

		merklizationInput[key] = value
	}
	root, paths := Merklization(merklizationInput)

	// TODO: refactor after the DB design is done
	GlobalMerklePathMap[types.StateRoot(root)] = paths
	return types.StateRoot(root)
}

// MerklizationState is a function that takes a state and returns the
// Merklization of the state.
// (D.5)
func MerklizationState(state types.State) types.StateRoot {
	// serializedState, err := StateSerialize(state)
	serializedState, _ := StateEncoder(state)

	return MerklizationSerializedState(serializedState)
}
