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
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// Custom input struct for json
type Input_json struct {
	PreimageExtrinsic []preimage_json `json:"preimages,omitempty"`
	Slot              int             `json:"slot,omitempty"`
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
	LookupDict     []lookupdict_json     `json:"history,omitempty"`
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

type lookupdict_json struct {
	Key       dictKey_json `json:"key,omitempty"`
	Timeslots []int        `json:"value,omitempty"`
}

type dictKey_json struct {
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
	PreimageExtrinsic types.PreimagesExtrinsic
	Slot              types.TimeSlot
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
		type Input_json struct {		         type myInput struct {
			PreimageExtrinsic []preimage_json -> 	PreimageExtrinsic types.PreimagesExtrinsic (no need)
			Slot              int             -> 	Slot              types.TimeSlot
		}								         }
	*/
	var (
		slot       = vector_json.Input.Slot
		slot_types = types.TimeSlot(slot)
	)
	my.Input.Slot = slot_types

	// myTestVector.PreState: state_json -> types.State
	/*
		type state_json struct {			                              		type types.State struct {
			Accounts []serviceAccount_json {                       ->     			Delta map[types.ServiceId]types.ServiceAccount {
				Id   int                                   				 				PreimageLookup map[types.OpaqueHash]types.ByteSequence   // a_p
				Info serviceAccountInfo_json { 											LookupDict     map[types.DictionaryKey]types.TimeSlotSet // a_l
					PreimageLookup []preimagelookup_json { 							}
						Hash string												}
						Blob string
					}
					LookupDict     []lookupdict_json {
						Key       dictKey_json {
							Hash   string
							Length int
						}
						Timeslots []int
					}
				}
			}
		}
	*/
	for _, account := range vector_json.PreState.Accounts {
		var (
			id_types             = types.ServiceId(account.Id)
			preimageLookup_types = make(map[types.OpaqueHash]types.ByteSequence)
			lookupDict_types     = make(map[types.DictionaryKey]types.TimeSlotSet)
		)
		// convert preimageLookup
		for _, preimageLookup := range account.Info.PreimageLookup {
			var (
				pHash, _    = hexToOpaqueHash(preimageLookup.Hash)
				pHash_types = types.OpaqueHash(pHash)

				pBlob_types = types.ByteSequence(preimageLookup.Blob)
			)
			preimageLookup_types[pHash_types] = pBlob_types
		}
		// convert lookupDict
		for _, lookupDict := range account.Info.LookupDict {
			var (
				lKeyHash, _    = hexToOpaqueHash(lookupDict.Key.Hash)
				lKeyHash_types = types.OpaqueHash(lKeyHash)

				lKeylength       = lookupDict.Key.Length
				lKeylength_types = types.U32(lKeylength)

				timeslots       = lookupDict.Timeslots
				timeslots_types = types.TimeSlotSet{}
			)
			for _, timeslot := range timeslots {
				timeslots_types = append(timeslots_types, types.TimeSlot(timeslot))
			}
			lookupDict_types[types.DictionaryKey{Hash: lKeyHash_types, Length: lKeylength_types}] = timeslots_types
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
		type state_json struct {			                              		type types.State struct {
			Accounts []serviceAccount_json {                       ->     			Delta map[types.ServiceId]types.ServiceAccount {
				Id   int                                   				 				PreimageLookup map[types.OpaqueHash]types.ByteSequence   // a_p
				Info serviceAccountInfo_json { 											LookupDict     map[types.DictionaryKey]types.TimeSlotSet // a_l
					PreimageLookup []preimagelookup_json { 							}
						Hash string												}
						Blob string
					}
					LookupDict     []lookupdict_json {
						Key       dictKey_json {
							Hash   string
							Length int
						}
						Timeslots []int
					}
				}
			}
		}
	*/
	for _, account := range vector_json.PostState.Accounts {
		var (
			id_types             = types.ServiceId(account.Id)
			preimageLookup_types = make(map[types.OpaqueHash]types.ByteSequence)
			lookupDict_types     = make(map[types.DictionaryKey]types.TimeSlotSet)
		)
		// convert preimageLookup
		for _, preimageLookup := range account.Info.PreimageLookup {
			var (
				pHash, _    = hexToOpaqueHash(preimageLookup.Hash)
				pHash_types = types.OpaqueHash(pHash)

				pBlob_types = types.ByteSequence(preimageLookup.Blob)
			)
			preimageLookup_types[pHash_types] = pBlob_types
		}
		// convert lookupDict
		for _, lookupDict := range account.Info.LookupDict {
			var (
				lKeyHash, _    = hexToOpaqueHash(lookupDict.Key.Hash)
				lKeyHash_types = types.OpaqueHash(lKeyHash)

				lKeylength       = lookupDict.Key.Length
				lKeylength_types = types.U32(lKeylength)

				timeslots       = lookupDict.Timeslots
				timeslots_types = make([]types.TimeSlot, 3)
			)
			for i, timeslot := range timeslots {
				timeslots_types[i] = types.TimeSlot(timeslot)
			}
			lookupDict_types[types.DictionaryKey{Hash: lKeyHash_types, Length: lKeylength_types}] = timeslots_types
		}

		serviceAccount_types := types.ServiceAccount{
			PreimageLookup: preimageLookup_types,
			LookupDict:     lookupDict_types,
		}
		// fmt.Println("serviceAccount_types:", serviceAccount_types)
		// iterate over accounts
		my.PostState.Delta[id_types] = serviceAccount_types
		// fmt.Println("\nmy.PostState.Delta:", my.PostState.Delta)
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

func TestFetchCodeByHash(t *testing.T) {
	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode     = types.ByteSequence("0x123456789")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: mockCode,
			},
		}
	)

	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// fetch code by hash
	code := FetchCodeByHash(mockId, mockCodeHash)

	// check if code is equal to mockCode
	if code == nil {
		t.Errorf("FetchCodeByHash failed: non exist code for codeHash %v", mockCodeHash)
	} else if !reflect.DeepEqual(code, mockCode) {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockCode, code)
	}
}

func TestValidateAccount(t *testing.T) {
	// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)

	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode     = types.ByteSequence("0x123456789")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		preimage     = mockCode

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			// h = H(p)
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			// (h, |p|) ∈ K(a_l)
			LookupDict: map[types.DictionaryKey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(preimage))}: {},
			},
		}
	)
	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// test ValidateAccount
	err := ValidatePreimageLookupDict(mockId)
	if err != nil {
		t.Errorf("ValidateAccount failed: %v", err)
	}
}

func TestHistoricalLookupFunction(t *testing.T) {
	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode     = types.ByteSequence("0x123456789")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		preimage     = mockCode

		mockTimestamp = types.TimeSlot(42)

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			LookupDict: map[types.DictionaryKey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(preimage))}: {mockTimestamp},
			},
		}
	)
	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// test HistoricalLookupFunction
	result := HistoricalLookupFunction(mockAccount, mockTimestamp, mockCodeHash)
	if !reflect.DeepEqual(result, preimage) {
		t.Errorf("HistoricalLookupFunction failed: expected %v, got %v", preimage, result)
	}
}

func TestGetSerivecAccountDerivatives(t *testing.T) {
	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		mockCode        = types.ByteSequence("0x123456789")
		mockCodeHash    = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		preimage        = mockCode
		mockPreimageLen = types.U32(len(preimage))

		mockTimestamp = types.TimeSlot(42)

		// create mock id and ServiceAccount
		mockId      = types.ServiceId(3)
		mockAccount = types.ServiceAccount{
			StorageDict: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			LookupDict: map[types.DictionaryKey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: mockPreimageLen}: {mockTimestamp},
			},
		}
	)
	// set to prior states
	store.NewPriorStates()
	store.GetInstance().GetPriorStates().SetDelta(map[types.ServiceId]types.ServiceAccount{
		mockId: mockAccount,
	})

	// test GetSerivecAccountDerivatives
	accountDer := GetSerivecAccountDerivatives(mockId)
	t.Log("accountDer:", accountDer)
	t.Logf("a_i=2*|a_l|+|a_s|\n LHS: %v, RHS: %v", accountDer.Items, 2*len(mockAccount.LookupDict)+len(mockAccount.StorageDict))
	var totalZ types.U32
	for key := range mockAccount.LookupDict {
		totalZ += key.Length
	}
	var totalX int
	for _, value := range mockAccount.StorageDict {
		totalX += len(value)
	}
	t.Logf("a_o=[ ∑_{(h,z)∈Key(a_l)}  81 + z ] + [ ∑_{x∈Value(a_s)} 32 + |x| ]\n LHS: %v, RHS: %v + %v", accountDer.Bytes, 81+totalZ, 32+totalX)
	t.Logf("a_t = B_S + B_I*a_i + B_L*a_o\n LHS: %v, RHS: %v", accountDer.Minbalance, types.U64(types.BasicMinBalance)+types.U64(types.U32(types.AdditionalMinBalancePerItem)*accountDer.Items)+types.U64(types.AdditionalMinBalancePerOctet)*accountDer.Bytes)
}

func TestServiceAccount(t *testing.T) {
	// Load test vectors from JSON
	vectors := []string{
		"../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-1.json",
		// "../../pkg/test_data/jam-test-vectors/preimages/data/preimage_needed-2.json",
	}

	for i, vec := range vectors {
		vectorIdx, vector := i, vec

		t.Run(fmt.Sprintf("Vector_%d", vectorIdx+1), func(t *testing.T) {
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
							t.Log("Account.CodeHash:", Account.CodeHash)
							code := FetchCodeByHash(Id, Account.CodeHash)

							expectedCode := Account.PreimageLookup[Account.CodeHash]

							if !reflect.DeepEqual(code, expectedCode) {
								t.Errorf("service %d fetches code \nwant: %v \nbut got %v", Id, expectedCode, code)
							}
						})

						t.Log("PreimageLookup:", Account.PreimageLookup)
						t.Run("ValidateAccount", func(t *testing.T) {
							err = ValidatePreimageLookupDict(Id)
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
							t.Log("a_t = B_S + B_I*a_i + B_L*a_o\n", accountDer.Minbalance, types.U64(types.BasicMinBalance)+types.U64(types.U32(types.AdditionalMinBalancePerItem)*accountDer.Items)+types.U64(types.AdditionalMinBalancePerOctet)*accountDer.Bytes)
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
