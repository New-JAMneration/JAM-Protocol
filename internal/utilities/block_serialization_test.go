package utilities

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type blockJSON struct {
	Header struct {
		Parent          string `json:"parent,omitempty"`
		ParentStateRoot string `json:"parent_state_root,omitempty"`
		ExtrinsicHash   string `json:"extrinsic_hash,omitempty"`
		Slot            int    `json:"slot,omitempty"`
		EpochMark       *struct {
			Entropy        string   `json:"entropy,omitempty"`
			TicketsEntropy string   `json:"tickets_entropy,omitempty"`
			Validators     []string `json:"validators,omitempty"`
		} `json:"epoch_mark,omitempty"`
		TicketsMark *[]struct {
			Id      string `json:"id,omitempty"`
			Attempt int    `json:"attempt,omitempty"`
		} `json:"tickets_mark,omitempty"`
		OffendersMark []string `json:"offenders_mark,omitempty"`
		AuthorIndex   int      `json:"author_index,omitempty"`
		EntropySource string   `json:"entropy_source,omitempty"`
		Seal          string   `json:"seal,omitempty"`
	} `json:"header"`

	Extrinsic struct {
		Tickets []struct {
			Attempt   int    `json:"attempt,omitempty"`
			Signature string `json:"signature,omitempty"`
		} `json:"tickets,omitempty"`
		Preimages []struct {
			Requester int    `json:"requester,omitempty"`
			Blob      string `json:"blob,omitempty"`
		} `json:"preimages"`
		Guarantees []struct {
			Report struct {
				PackageSpec struct {
					Hash         string `json:"hash,omitempty"`
					Length       int    `json:"length,omitempty"`
					ErasureRoot  string `json:"erasure_root,omitempty"`
					ExportsRoot  string `json:"exports_root,omitempty"`
					ExportsCount int    `json:"exports_count,omitempty"`
				} `json:"package_spec"`
				Context struct {
					Anchor           string   `json:"anchor,omitempty"`
					StateRoot        string   `json:"state_root,omitempty"`
					BeefyRoot        string   `json:"beefy_root,omitempty"`
					LookupAnchor     string   `json:"lookup_anchor,omitempty"`
					LookupAnchorSlot int      `json:"lookup_anchor_slot,omitempty"`
					Prerequisites    []string `json:"prerequisites,omitempty"`
				} `json:"context"`
				CoreIndex         int    `json:"core_index,omitempty"`
				AuthorizerHash    string `json:"authorizer_hash,omitempty"`
				AuthOutput        string `json:"auth_output,omitempty"`
				SegmentRootLookup []struct {
					WorkPackageHash string `json:"work_package_hash,omitempty"`
					SegmentTreeRoot string `json:"segment_tree_root,omitempty"`
				} `json:"segment_root_lookup,omitempty"`
				Results []struct {
					ServiceId     int               `json:"service_id,omitempty"`
					CodeHash      string            `json:"code_hash,omitempty"`
					PayloadHash   string            `json:"payload_hash,omitempty"`
					AccumulateGas int               `json:"accumulate_gas,omitempty"`
					Result        map[string]string `json:"result,omitempty"`
				} `json:"results,omitempty"`
			} `json:"report"`
			Slot       int `json:"slot,omitempty"`
			Signatures []struct {
				ValidatorIndex int    `json:"validator_index,omitempty"`
				Signature      string `json:"signature,omitempty"`
			} `json:"signatures,omitempty"`
		} `json:"guarantees"`
		Assurances []struct {
			Anchor         string `json:"anchor,omitempty"`
			Bitfield       string `json:"bitfield,omitempty"`
			ValidatorIndex int    `json:"validator_index,omitempty"`
			Signature      string `json:"signature,omitempty"`
		} `json:"assurances"`
		Disputes struct {
			Verdicts []struct {
				Target string `json:"target,omitempty"`
				Age    int    `json:"age,omitempty"`
				Votes  []struct {
					Vote      bool   `json:"vote,omitempty"`
					Index     int    `json:"index,omitempty"`
					Signature string `json:"signature,omitempty"`
				} `json:"votes,omitempty"`
			} `json:"verdicts,omitempty"`
			Culprits []struct {
				Target    string `json:"target,omitempty"`
				Key       string `json:"key,omitempty"`
				Signature string `json:"signature,omitempty"`
			} `json:"culprits,omitempty"`
			Faults []struct {
				Target    string `json:"target,omitempty"`
				Vote      bool   `json:"vote,omitempty"`
				Key       string `json:"key,omitempty"`
				Signature string `json:"signature,omitempty"`
			} `json:"faults,omitempty"`
		} `json:"disputes"`
	} `json:"extrinsic"`
}

func readDataFromJson(filename string) []byte {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v", err)
		return nil
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return nil
	}
	return byteValue
}

func readBlockDataFromJson(filename string) types.Block {
	byteValue := readDataFromJson(filename)
	var testDataFromJson blockJSON
	var outputBlockData types.Block

	// Unmarshal the JSON data
	err := json.Unmarshal(byteValue, &testDataFromJson)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v", err)
		return types.Block{}
	}

	// Read Block Header data from JSON
	outputBlockData.Header = types.Header{
		Parent:          types.HeaderHash(stringToHex(testDataFromJson.Header.Parent)),
		ParentStateRoot: types.StateRoot(stringToHex(testDataFromJson.Header.ParentStateRoot)),
		ExtrinsicHash:   types.OpaqueHash(stringToHex(testDataFromJson.Header.ExtrinsicHash)),
		Slot:            types.TimeSlot(testDataFromJson.Header.Slot),
		AuthorIndex:     types.ValidatorIndex(testDataFromJson.Header.AuthorIndex),
		EntropySource:   types.BandersnatchVrfSignature(stringToHex(testDataFromJson.Header.EntropySource)),
		Seal:            types.BandersnatchVrfSignature(stringToHex(testDataFromJson.Header.Seal)),
	}

	if testDataFromJson.Header.EpochMark != nil {
		outputBlockData.Header.EpochMark = &types.EpochMark{
			Entropy:        types.Entropy(stringToHex(testDataFromJson.Header.EpochMark.Entropy)),
			TicketsEntropy: types.Entropy(stringToHex(testDataFromJson.Header.EpochMark.TicketsEntropy)),
		}

		for _, data := range testDataFromJson.Header.EpochMark.Validators {
			validator := stringToHex(data)
			outputBlockData.Header.EpochMark.Validators = append(outputBlockData.Header.EpochMark.Validators, types.BandersnatchPublic(validator))
		}
	}

	if testDataFromJson.Header.TicketsMark != nil {
		outputBlockData.Header.TicketsMark = &types.TicketsMark{}

		for _, data := range *testDataFromJson.Header.TicketsMark {
			*outputBlockData.Header.TicketsMark = append(*outputBlockData.Header.TicketsMark, types.TicketBody{
				Id:      types.TicketId(stringToHex(data.Id)),
				Attempt: types.TicketAttempt(data.Attempt),
			})
		}
	}

	for _, data := range testDataFromJson.Header.OffendersMark {
		offender := stringToHex(data)
		outputBlockData.Header.OffendersMark = append(outputBlockData.Header.OffendersMark, types.Ed25519Public(offender))
	}

	// Read Block Extrinsic data from JSON
	for _, data := range testDataFromJson.Extrinsic.Tickets {
		signature := stringToHex(data.Signature)
		outputBlockData.Extrinsic.Tickets = append(outputBlockData.Extrinsic.Tickets, types.TicketEnvelope{
			Attempt:   types.TicketAttempt(data.Attempt),
			Signature: types.BandersnatchRingVrfSignature(signature),
		})
	}

	for _, data := range testDataFromJson.Extrinsic.Preimages {
		preimage := stringToHex(data.Blob)
		outputBlockData.Extrinsic.Preimages = append(outputBlockData.Extrinsic.Preimages, types.Preimage{
			Requester: types.ServiceId(data.Requester),
			Blob:      types.ByteSequence(preimage),
		})
	}

	for num, data := range testDataFromJson.Extrinsic.Guarantees {
		outputBlockData.Extrinsic.Guarantees = append(outputBlockData.Extrinsic.Guarantees, types.ReportGuarantee{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:         types.WorkPackageHash(stringToHex(data.Report.PackageSpec.Hash)),
					Length:       types.U32(data.Report.PackageSpec.Length),
					ErasureRoot:  types.ErasureRoot(stringToHex(data.Report.PackageSpec.ErasureRoot)),
					ExportsRoot:  types.ExportsRoot(stringToHex(data.Report.PackageSpec.ExportsRoot)),
					ExportsCount: types.U16(data.Report.PackageSpec.ExportsCount),
				},
				Context: types.RefineContext{
					Anchor:           types.HeaderHash(stringToHex(data.Report.Context.Anchor)),
					StateRoot:        types.StateRoot(stringToHex(data.Report.Context.StateRoot)),
					BeefyRoot:        types.BeefyRoot(stringToHex(data.Report.Context.BeefyRoot)),
					LookupAnchor:     types.HeaderHash(stringToHex(data.Report.Context.LookupAnchor)),
					LookupAnchorSlot: types.TimeSlot(data.Report.Context.LookupAnchorSlot),
				},
				CoreIndex:      types.CoreIndex(data.Report.CoreIndex),
				AuthorizerHash: types.OpaqueHash(stringToHex(data.Report.AuthorizerHash)),
				AuthOutput:     types.ByteSequence(stringToHex(data.Report.AuthOutput)),
				Results:        nil,
			},
			Slot: types.TimeSlot(data.Slot),
		})

		for _, data := range data.Report.Context.Prerequisites {
			outputBlockData.Extrinsic.Guarantees[num].Report.Context.Prerequisites = append(outputBlockData.Extrinsic.Guarantees[num].Report.Context.Prerequisites, types.OpaqueHash(stringToHex(data)))
		}

		for _, data := range data.Report.SegmentRootLookup {
			outputBlockData.Extrinsic.Guarantees[num].Report.SegmentRootLookup = append(outputBlockData.Extrinsic.Guarantees[num].Report.SegmentRootLookup, types.SegmentRootLookupItem{
				WorkPackageHash: types.WorkPackageHash(stringToHex(data.WorkPackageHash)),
				SegmentTreeRoot: types.OpaqueHash(stringToHex(data.SegmentTreeRoot)),
			})
		}

		for _, data := range data.Report.Results {
			myMap := make(map[types.WorkExecResultType][]byte)
			for key, value := range data.Result {
				myMap[types.WorkExecResultType(key)] = stringToHex(value)
			}
			outputBlockData.Extrinsic.Guarantees[num].Report.Results = append(outputBlockData.Extrinsic.Guarantees[num].Report.Results, types.WorkResult{
				ServiceId:     types.ServiceId(data.ServiceId),
				CodeHash:      types.OpaqueHash(stringToHex(data.CodeHash)),
				PayloadHash:   types.OpaqueHash(stringToHex(data.PayloadHash)),
				AccumulateGas: types.Gas(data.AccumulateGas),
				Result:        myMap,
			})
		}

		for _, data := range data.Signatures {
			signature := stringToHex(data.Signature)
			outputBlockData.Extrinsic.Guarantees[num].Signatures = append(outputBlockData.Extrinsic.Guarantees[num].Signatures, types.ValidatorSignature{
				ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
				Signature:      types.Ed25519Signature(signature),
			})
		}

	}

	for _, data := range testDataFromJson.Extrinsic.Assurances {
		signature := stringToHex(data.Signature)
		outputBlockData.Extrinsic.Assurances = append(outputBlockData.Extrinsic.Assurances, types.AvailAssurance{
			Anchor:         types.OpaqueHash(stringToHex(data.Anchor)),
			Bitfield:       []byte(stringToHex(data.Bitfield)[:]),
			ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
			Signature:      types.Ed25519Signature(signature),
		})
	}

	for num, data := range testDataFromJson.Extrinsic.Disputes.Verdicts {
		outputBlockData.Extrinsic.Disputes.Verdicts = append(outputBlockData.Extrinsic.Disputes.Verdicts, types.Verdict{
			Target: types.OpaqueHash(stringToHex(data.Target)),
			Age:    types.U32(data.Age),
		})
		for _, vote := range data.Votes {
			outputBlockData.Extrinsic.Disputes.Verdicts[num].Votes = append(outputBlockData.Extrinsic.Disputes.Verdicts[num].Votes, types.Judgement{
				Vote:      vote.Vote,
				Index:     types.ValidatorIndex(vote.Index),
				Signature: types.Ed25519Signature(stringToHex(vote.Signature)),
			})
		}
	}

	for _, data := range testDataFromJson.Extrinsic.Disputes.Culprits {
		outputBlockData.Extrinsic.Disputes.Culprits = append(outputBlockData.Extrinsic.Disputes.Culprits, types.Culprit{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	for _, data := range testDataFromJson.Extrinsic.Disputes.Faults {
		outputBlockData.Extrinsic.Disputes.Faults = append(outputBlockData.Extrinsic.Disputes.Faults, types.Fault{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Vote:      data.Vote,
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	return outputBlockData
}

func stringToHex(str string) []byte {
	bytes, err := hex.DecodeString(str[2:])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bytes
}

func extrinsicSerialization(extrinsic types.Extrinsic) (output types.ByteSequence) {
	output = append(output, ExtrinsicTicketSerialization(extrinsic.Tickets)...)
	output = append(output, ExtrinsicPreimageSerialization(extrinsic.Preimages)...)
	output = append(output, ExtrinsicGuaranteeSerialization(extrinsic.Guarantees)...)
	output = append(output, ExtrinsicAssuranceSerialization(extrinsic.Assurances)...)
	output = append(output, ExtrinsicDisputeSerialization(extrinsic.Disputes)...)
	return output
}

// extrinsic hash formula (5.4)
func pocExtrinsicHash(extrinsic types.Extrinsic) (output types.OpaqueHash) {
	ticketSerializedHash := hash.Blake2bHash(ExtrinsicTicketSerialization(extrinsic.Tickets))
	preimageSerializedHash := hash.Blake2bHash(ExtrinsicPreimageSerialization(extrinsic.Preimages))
	assureanceSerializedHash := hash.Blake2bHash(ExtrinsicAssuranceSerialization(extrinsic.Assurances))
	disputeSerializedHash := hash.Blake2bHash(ExtrinsicDisputeSerialization(extrinsic.Disputes))

	// g (5.6)
	g := types.ByteSequence{}
	g = append(g, SerializeU64(types.U64(len(extrinsic.Guarantees)))...)
	for _, guarantee := range extrinsic.Guarantees {
		// w, WorkReport
		w := guarantee.Report
		wHash := hash.Blake2bHash(WorkReportSerialization(w))

		// t, Slot
		t := guarantee.Slot
		tSerialized := SerializeFixedLength(types.U32(t), 4)

		// a, Signatures (credential)
		signaturesLength, Signatures := LensElementPair(guarantee.Signatures)

		elementSerialized := types.ByteSequence{}
		elementSerialized = append(elementSerialized, SerializeByteArray(wHash[:])...)
		elementSerialized = append(elementSerialized, SerializeByteArray(tSerialized)...)
		elementSerialized = append(elementSerialized, SerializeU64(types.U64(signaturesLength))...)
		for _, signature := range Signatures {
			elementSerialized = append(elementSerialized, SerializeU64(types.U64(signature.ValidatorIndex))...)
			elementSerialized = append(elementSerialized, SerializeByteArray(signature.Signature[:])...)
		}

		// If the input type of serialization is octet sequence, we can directly
		// append it because it is already serialized.
		g = append(g, elementSerialized...)
	}

	gHash := hash.Blake2bHash(g)

	// Serialize the hash of the extrinsic elements
	serializedElements := types.ByteSequence{}
	serializedElements = append(serializedElements, WrapOpaqueHash(ticketSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, WrapOpaqueHash(preimageSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, WrapOpaqueHash(gHash).Serialize()...)
	serializedElements = append(serializedElements, WrapOpaqueHash(assureanceSerializedHash).Serialize()...)
	serializedElements = append(serializedElements, WrapOpaqueHash(disputeSerializedHash).Serialize()...)

	// Hash the serialized elements
	output = hash.Blake2bHash(serializedElements)

	return output
}

func TestBlockSerialization(t *testing.T) {
	testCases := []struct {
		name              string
		header            types.Header
		extrinsic         types.Extrinsic
		expectedHeader    types.HeaderHash
		expectedExtrinsic types.OpaqueHash
	}{}

	for i := 0; i <= 11; i++ {
		name := fmt.Sprintf("425530_%03d", i) // 格式化為 425530_000, 425530_001, ...
		header := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header
		extrinsic := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Extrinsic
		var expectedHeaderHash types.HeaderHash
		expectedExtrinsicHash := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header.ExtrinsicHash

		if i == 11 {
			expectedHeaderHash = readBlockDataFromJson("data/425531_000.json").Header.Parent // 對於 011，使用 425531_000 的 parent
		} else {
			expectedHeaderHash = readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i+1)).Header.Parent
		}

		testCases = append(testCases, struct {
			name              string
			header            types.Header
			extrinsic         types.Extrinsic
			expectedHeader    types.HeaderHash
			expectedExtrinsic types.OpaqueHash
		}{
			name:              name,
			header:            header,
			extrinsic:         extrinsic,
			expectedHeader:    expectedHeaderHash,
			expectedExtrinsic: expectedExtrinsicHash,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headerResult := HeaderSerialization(tc.header)
			headerHash := hash.Blake2bHash(headerResult)
			extrinsicResult := extrinsicSerialization(tc.extrinsic)
			extrinsicHash := hash.Blake2bHash(extrinsicResult)
			// pocExtrinsicHash := pocExtrinsicHash(tc.extrinsic)

			if headerHash != types.OpaqueHash(tc.expectedHeader) {
				t.Errorf("\nExpected Header Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedHeader[:]), hex.EncodeToString(headerHash[:]))
			}

			// if pocExtrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
			// 	t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(pocExtrinsicHash[:]))
			// }

			if extrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
				t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(extrinsicHash[:]))
			}
		})
	}
}

func TestExtrinsicSerialization(t *testing.T) {
	testCases := []struct {
		name              string
		extrinsic         types.Extrinsic
		expectedExtrinsic types.OpaqueHash
	}{
		{
			name:              "Assurances",
			extrinsic:         readBlockDataFromJson("data/425536_009.json").Extrinsic,
			expectedExtrinsic: readBlockDataFromJson("data/425536_009.json").Header.ExtrinsicHash,
		},
		{
			name:              "Guarantees",
			extrinsic:         readBlockDataFromJson("data/425538_002.json").Extrinsic,
			expectedExtrinsic: readBlockDataFromJson("data/425538_002.json").Header.ExtrinsicHash,
		},
		{
			name:              "Preimages",
			extrinsic:         readBlockDataFromJson("data/425539_002.json").Extrinsic,
			expectedExtrinsic: readBlockDataFromJson("data/425539_002.json").Header.ExtrinsicHash,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extrinsicResult := extrinsicSerialization(tc.extrinsic)
			extrinsicHash := hash.Blake2bHash(extrinsicResult)

			if extrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
				t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(extrinsicHash[:]))
			}
		})
	}
}

// func TestPocExtrinsicSerialization(t *testing.T) {
// 	testCases := []struct {
// 		name              string
// 		extrinsic         types.Extrinsic
// 		expectedExtrinsic types.OpaqueHash
// 	}{}

// 	for i := 0; i <= 11; i++ {
// 		name := fmt.Sprintf("425530_%03d", i) // 格式化為 425530_000, 425530_001, ...
// 		extrinsic := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Extrinsic
// 		expectedExtrinsicHash := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header.ExtrinsicHash

// 		testCases = append(testCases, struct {
// 			name              string
// 			extrinsic         types.Extrinsic
// 			expectedExtrinsic types.OpaqueHash
// 		}{
// 			name:              name,
// 			extrinsic:         extrinsic,
// 			expectedExtrinsic: expectedExtrinsicHash,
// 		})
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			pocExtrinsicHash := pocExtrinsicHash(tc.extrinsic)

// 			if pocExtrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
// 				t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(pocExtrinsicHash[:]))
// 			}
// 		})
// 	}
// }
