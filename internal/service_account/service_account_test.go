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
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// Custom input struct for json
type Input_json struct {
	Preimages []preimage_json `json:"preimages,omitempty"`
	Slot      int             `json:"slot,omitempty"`
}

type preimage_json struct {
	Requester string `json:"hash,omitempty"`
	Blob      string `json:"blob,omitempty"`
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
	PreimageLookup []preimagelookup_json `json:"preimages,omitempty"`
	LookupDict     []history_json        `json:"history,omitempty"`
	// CodeHash       types.OpaqueHash `json:"code_hash,omitempty"`
	// Balance        types.U64        `json:"balance,omitempty"`
	// MinItemGas     types.Gas        `json:"min_item_gas,omitempty"`
	// MinMemoGas     types.Gas        `json:"min_memo_gas,omitempty"`
	// Items          types.U32        `json:"items,omitempty"`
	// Bytes          types.U64        `json:"bytes,omitempty"`
	// Minbalance     types.U64        `json:"minbalance,omitempty"`
}

type preimagelookup_json struct {
	Hash string `json:"hash,omitempty"`
	Blob string `json:"blob,omitempty"`
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
	Preimages types.PreimagesExtrinsic
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

	// initialize a empty myTestVector
	my = myTestVector{
		Input: myInput{},
		PreState: types.State{
			Delta: make(map[types.ServiceId]types.ServiceAccount),
		},
		Output: vector_json.Output,
		PostState: types.State{
			Delta: make(map[types.ServiceId]types.ServiceAccount),
		},
	}
	// convert json to mytypes

	// myTestVector.Input: Input_json -> myInput
	/*
		type Input_json struct {		type myInput struct {
			Preimages []preimage_json -> 	Preimages types.PreimagesExtrinsic (no need)
			Slot      int             -> 	types.TimeSlot
		}								}
	*/
	var (
		slot       = vector_json.Input.Slot
		slot_types = types.TimeSlot(slot)
	)
	my.Input.Slot = slot_types

	// myTestVector.PreState: state_json -> types.State
	/*
		type state_json struct {			  type types.State struct {
			Accounts []serviceAccount_json ->     Delta map[types.ServiceId]types.ServiceAccount
		}									  }
	*/
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
	for _, account := range vector_json.PostState.Accounts {
		var (
			id_types             = types.ServiceId(account.Id)
			info                 = account.Info
			preimageLookup_types = make(map[types.OpaqueHash]types.ByteSequence)
			lookupDict_types     = make(map[types.DictionaryKey]types.TimeSlotSet)
		)
		// convert preimageLookup
		for _, preimagelookup := range info.PreimageLookup {
			var (
				pHash, _ = hexToOpaqueHash(preimagelookup.Hash)
				pHash1   = HexToByteArray32(preimagelookup.Hash)

				pHash_types = types.OpaqueHash(pHash1)

				pBlob_types = types.ByteSequence(preimagelookup.Blob)
				blobHash    = hash.Blake2bHash(utilities.WrapByteSequence(pBlob_types).Serialize())
			)
			preimageLookup_types[pHash_types] = pBlob_types
			fmt.Println("\npreimage.Hash:", preimagelookup.Hash)
			fmt.Println("\npHash:", pHash)
			fmt.Println("\npHash1:", pHash1)
			fmt.Println("\npreimage.Blob:", preimagelookup.Blob)
			fmt.Println("\npBlob_types:", pBlob_types)
			fmt.Println("\nblobHash:", blobHash)
			fmt.Println("\npreimageLookup_types:", preimageLookup_types)
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
		Input:     my.Input,
		PreState:  my.PreState,
		Output:    vector_json.Output,
		PostState: my.PostState,
	}
	// fmt.Println("\nmy:", my)
	return my, err
}

func TestServiceAccount(t *testing.T) {
	// Load test vectors from JSON
	vectors := []string{
		"../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-1.json",
		// "../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-2.json",
	}

	for i, vec := range vectors {
		vectorIdx, vector := i, vec

		t.Run(fmt.Sprintf("Vector_%d", vectorIdx), func(t *testing.T) {
			my, err := loadInputFromJSON(vector)
			if err != nil {
				t.Fatalf("Failed to load input from JSON[%d]: %v", vectorIdx, err)
			}

			// Store initialization
			store.NewPriorStates()
			store.GetInstance().GetPriorStates().SetDelta(my.PreState.Delta)
			// t.Logf("my.PreState.Delta: %+v", my.PreState.Delta)
			priorDelta := store.GetInstance().GetPriorStates().GetState().Delta
			slot := my.Input.Slot

			// Perform operations on the ServiceAccount
			// 这里可以添加对 account 的操作，例如验证、更新等
			// 例如：
			// t.Logf("Load Prior Delta: %+v", priorDelta)

			t.Run("AccountTests", func(t *testing.T) {

				t.Logf("Prior Delta: %+v", priorDelta)
				for id, account := range priorDelta {
					Id, Account := id, account

					t.Run(fmt.Sprintf("Account_%d", Id), func(t *testing.T) {
						// t.Parallel()

						t.Run("FetchCodeByHash", func(t *testing.T) {
							// t.Parallel()
							code := FetchCodeByHash(Id, Account.CodeHash)

							expectedCode := priorDelta[Id].PreimageLookup[priorDelta[Id].CodeHash]

							if !reflect.DeepEqual(code, expectedCode) {
								t.Errorf("service %d fetches code \nwant: %v \nbut got %v", Id, expectedCode, code)
							}
						})

						t.Log("PreimageLookup:", Account.PreimageLookup)
						t.Run("ValidateAccount", func(t *testing.T) {
							err = ValidateAccount(Account)
							if err != nil {
								t.Errorf("Validation failed for id: %d Account: %v", Id, err)
							}
						})

						t.Run("HistoricalLookup", func(t *testing.T) {
							// t.Parallel()
							for hash := range account.PreimageLookup {
								result := HistoricalLookupFunction(Account, slot, hash)
								expected := account.PreimageLookup[hash]

								if !reflect.DeepEqual(result, expected) {
									t.Errorf("HistoricalLookupFunction for Account %v does not match expected result for vector[%d]", Account, vectorIdx)
								}
							}
						})

						t.Run("GetSerivecAccountDerivatives", func(t *testing.T) {
							// t.Parallel()
							accountDer := GetSerivecAccountDerivatives(Id)
							t.Log("accountDer:", accountDer)
							t.Log("a_i=2*|a_l|+|a_p|\n", accountDer.Items, 2*len(Account.LookupDict)+len(Account.PreimageLookup))
							t.Log("a_o=[ ∑_{(h,z)∈Key(a_l)}  81 + z  ] + [ ∑_{x∈Value(a_s)}	32 + |x| ]\n", accountDer.Bytes, 81*len(Account.LookupDict)+len(Account.LookupDict), 32*len(Account.PreimageLookup)+len(Account.PreimageLookup))
							t.Log("a_t = B_S + B_I*a_i + B_L*a_o\n", accountDer.Minbalance, types.U64(types.BasicMinBalance)+types.U64(types.U32(types.AdditionalMinBalancePerItem)*deltaDer[Id].Items)+types.U64(types.AdditionalMinBalancePerOctet)*deltaDer[Id].Bytes)
						})
					})
				}
				// Check output against expected output
				// 这里可以添加对输出的检查，例如：
				// if !reflect.DeepEqual(account, my.Output) {
				//     t.Errorf("Output does not match expected output for vector[%d]", i)
				// }

				// Check post state
				// postState := store.GetInstance().GetPosteriorStates().GetState()
				// if !reflect.DeepEqual(postState, my.PostState) {
				// 	t.Errorf("PostState does not match expected post state for vector[%d]", i)
				// }
			})
		})
	}
}
