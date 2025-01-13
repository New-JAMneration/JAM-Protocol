package service_account

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// Custom input struct for json
type Input_json struct {
	Preimages []preimage_json `json:"preimages,omitempty"`
	Slot      int             `json:"slot,omitempty"`
}

type preimage_json struct {
	Hash string `json:"hash,omitempty"`
	Blob string `json:"blob,omitempty"`
}
type state_json struct {
	Accounts []serviceAccount_json `json:"accounts,omitempty"`
}

type serviceAccount_json struct {
	Id   int                     `json:"id,omitempty"`
	Info serviceAccountInfo_json `json:"info,omitempty"`
}

type serviceAccountInfo_json struct {
	// StorageDict    map[types.OpaqueHash]types.ByteSequence   `json:"storage_dict,omitempty"`
	PreimageLookup []preimage_json `json:"preimages,omitempty"`
	LookupDict     []history_json  `json:"history,omitempty"`
	// CodeHash       types.OpaqueHash `json:"code_hash,omitempty"`
	// Balance        types.U64        `json:"balance,omitempty"`
	// MinItemGas     types.Gas        `json:"min_item_gas,omitempty"`
	// MinMemoGas     types.Gas        `json:"min_memo_gas,omitempty"`
	// Items          types.U32        `json:"items,omitempty"`
	// Bytes          types.U64        `json:"bytes,omitempty"`
	// Minbalance     types.U64        `json:"minbalance,omitempty"`
}

type history_json struct {
	Key       dictionaryKey_json `json:"key,omitempty"`
	Timeslots []int              `json:"value,omitempty"`
}

type dictionaryKey_json struct {
	Hash   string `json:"hash,omitempty"`
	Length int    `json:"length,omitempty"`
}
type testVector_json struct {
	Input     Input_json  `json:"input,omitempty"`
	PreState  state_json  `json:"pre_state,omitempty"`
	Output    interface{} `json:"output,omitempty"`
	PostState state_json  `json:"post_state,omitempty"`
}

// mytypes
type myInput struct {
	Preimages map[types.OpaqueHash]types.ByteSequence
	Slot      types.TimeSlot
}

type myTestVector struct {
	Input     myInput
	PreState  types.State
	Output    interface{}
	PostState types.State
}

func hexToOpaqueHash(hexStr string) ([32]byte, error) {
	// remove "0x" prefix
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// decode hex string
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to decode hex: %v", err)
	}

	if len(bytes) != 32 {
		return [32]byte{}, fmt.Errorf("invalid length: expected 32 bytes, got %d", len(bytes))
	}

	var result [32]byte
	copy(result[:], bytes)
	return result, nil
}

func loadInputFromJSON(filePath string) (my myTestVector, err error) {
	var vector_json testVector_json
	data, err := os.ReadFile(filePath)
	if err != nil {
		return my, err
	}
	err = json.Unmarshal(data, &vector_json)

	// convert json to mytypes

	// myTestVector.Input: Input_json -> myInput
	/*
		type Input_json struct {		type myInput struct {
			Preimages []preimage_json -> 	map[types.OpaqueHash]types.ByteSequence
			Slot      int             -> 	types.TimeSlot
		}								}
	*/
	my.Input.Preimages = make(map[types.OpaqueHash]types.ByteSequence)
	var (
		slot       = vector_json.Input.Slot
		slot_types = types.TimeSlot(slot)
	)
	for _, preimage := range vector_json.Input.Preimages {
		var (
			inputHash, _    = hexToOpaqueHash(preimage.Hash)
			inputHash_types = types.OpaqueHash(inputHash)
			length_types    = types.ByteSequence(preimage.Blob)
		)
		my.Input.Preimages[inputHash_types] = length_types
	}

	// myTestVector.PreState: state_json -> types.State
	/*
		type state_json struct {			  type types.State struct {
			Accounts []serviceAccount_json ->     Delta map[types.ServiceId]types.ServiceAccount
		}									  }
	*/
	my.PreState.Delta = make(map[types.ServiceId]types.ServiceAccount)
	for _, account := range vector_json.PreState.Accounts {
		var (
			id_types             = types.ServiceId(account.Id)
			info                 = account.Info
			preimageLookup_types = make(map[types.OpaqueHash]types.ByteSequence)
			lookupDict_types     = make(map[types.DictionaryKey]types.TimeSlotSet)
		)
		// convert preimageLookup
		for _, preimage := range info.PreimageLookup {
			var (
				pHash, _    = hexToOpaqueHash(preimage.Hash)
				pHash_types = types.OpaqueHash(pHash)

				pBlob_types = types.ByteSequence(preimage.Blob)
			)
			preimageLookup_types[pHash_types] = pBlob_types
		}
		// convert lookupDict
		for _, history := range info.LookupDict {
			var (
				hKeyHash, _    = hexToOpaqueHash(history.Key.Hash)
				hKeyHash_types = types.OpaqueHash(hKeyHash)

				hKeylength       = history.Key.Length
				hKeylength_types = types.U32(hKeylength)

				timeslots       = history.Timeslots
				timeslots_types = make([]types.TimeSlot, len(timeslots))
			)
			for i, timeslot := range timeslots {
				timeslots_types[i] = types.TimeSlot(timeslot)
			}
			lookupDict_types[types.DictionaryKey{Hash: hKeyHash_types, Length: hKeylength_types}] = timeslots_types
		}

		serviceAccount_types := types.ServiceAccount{
			PreimageLookup: preimageLookup_types,
			LookupDict:     lookupDict_types,
		}
		// fmt.Println("serviceAccount_types:", serviceAccount_types)
		// iterate over accounts
		my.PreState.Delta[id_types] = serviceAccount_types
		// fmt.Println("\nmy.PreState.Delta:", my.PreState.Delta)
	}

	// myTestVector.PostState: state_json -> types.State
	/*
		type state_json struct {			  type types.State struct {
			Accounts []serviceAccount_json ->     Delta map[types.ServiceId]types.ServiceAccount
		}									  }
	*/
	my.PostState.Delta = make(map[types.ServiceId]types.ServiceAccount)
	for _, account := range vector_json.PostState.Accounts {
		var (
			id_types             = types.ServiceId(account.Id)
			info                 = account.Info
			preimageLookup_types = make(map[types.OpaqueHash]types.ByteSequence)
			lookupDict_types     = make(map[types.DictionaryKey]types.TimeSlotSet)
		)
		// convert preimageLookup
		for _, preimage := range info.PreimageLookup {
			var (
				pHash, _    = hexToOpaqueHash(preimage.Hash)
				pHash_types = types.OpaqueHash(pHash)

				pBlob_types = types.ByteSequence(preimage.Blob)
			)
			preimageLookup_types[pHash_types] = pBlob_types
		}
		// convert lookupDict
		for _, history := range info.LookupDict {
			var (
				hKeyHash, _    = hexToOpaqueHash(history.Key.Hash)
				hKeyHash_types = types.OpaqueHash(hKeyHash)

				hKeylength       = history.Key.Length
				hKeylength_types = types.U32(hKeylength)

				timeslots       = history.Timeslots
				timeslots_types = make([]types.TimeSlot, len(timeslots))
			)
			for i, timeslot := range timeslots {
				timeslots_types[i] = types.TimeSlot(timeslot)
			}
			lookupDict_types[types.DictionaryKey{Hash: hKeyHash_types, Length: hKeylength_types}] = timeslots_types
		}

		serviceAccount_types := types.ServiceAccount{
			PreimageLookup: preimageLookup_types,
			LookupDict:     lookupDict_types,
		}
		// iterate over accounts
		my.PostState.Delta[id_types] = serviceAccount_types
	}

	// Finally, we construct myTestVector
	my = myTestVector{
		Input: myInput{
			// Preimages: make(map[types.OpaqueHash]types.ByteSequence),
			Slot: slot_types,
		},
		// PostState: types.State{
		// 	Delta: make(map[types.ServiceId]types.ServiceAccount),
		// },
		Output: vector_json.Output,
		// PostState: types.State{
		// 	Delta: make(map[types.ServiceId]types.ServiceAccount),
		// },
	}
	return my, err
}

func TestServiceAccount(t *testing.T) {
	// Load test vectors from JSON
	vectors := []string{
		"../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-1.json",
		// "../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-2.json",
	}

	for i, vector := range vectors {
		my, err := loadInputFromJSON(vector)
		if err != nil {
			t.Fatalf("Failed to load input from JSON[%d]: %v", i, err)
		}

		// Create a new ServiceAccount
		var (
			preimages = my.Input.Preimages
			slot      = my.Input.Slot
		)

		// Store initialization
		store.NewPriorStates()
		store.NewIntermediateStates()
		store.NewPosteriorStates()
		store.GetInstance().GetPriorStates().SetDelta(my.PreState.Delta)
		priorDelta := store.GetInstance().GetPriorStates().GetState().Delta

		// Perform operations on the ServiceAccount
		// 这里可以添加对 account 的操作，例如验证、更新等
		// 例如：
		t.Run("TestCheckAccountExistence", func(t *testing.T) {
			// Update account
			CheckAccountExistence()

			intermediateDelta := store.GetInstance().GetIntermediateStates().GetDelta()
			if !reflect.DeepEqual(intermediateDelta, priorDelta) {
				t.Errorf("intermediateDelta does not match expected priorDelta for vector[%d]", i)
			}
		})

		for id, account := range priorDelta {
			t.Run("TestValidaeAccount", func(t *testing.T) {
				// Validate account
				err = ValidateAccount(account)
				if err != nil {
					t.Errorf("Validation failed for id: %d account: %v", id, err)
				}
			})

			t.Run("TestHistoricalLookupFunction", func(t *testing.T) {
				for hash := range preimages {
					result := HistoricalLookupFunction(account, slot, hash)
					expected := preimages[hash]

					if !reflect.DeepEqual(result, expected) {
						t.Errorf("HistoricalLookupFunction for account %v does not match expected result for vector[%d]", account, i)
					}
				}
			})
			t.Run("TestUpdateSerivecAccount", func(t *testing.T) {
				UpdateSerivecAccount()
			})
		}

		// Check output against expected output
		// 这里可以添加对输出的检查，例如：
		// if !reflect.DeepEqual(account, my.Output) {
		//     t.Errorf("Output does not match expected output for vector[%d]", i)
		// }

		// Check post state
		postState := store.GetInstance().GetPosteriorStates().GetState()
		if !reflect.DeepEqual(postState, my.PostState) {
			t.Errorf("PostState does not match expected post state for vector[%d]", i)
		}
	}
}
