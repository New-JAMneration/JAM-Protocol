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

type block struct {
	Header    header    `json:"header"`
	Extrinsic extrinsic `json:"extrinsic"`
}

type header struct {
	Parent          string       `json:"parent,omitempty"`
	ParentStateRoot string       `json:"parent_state_root,omitempty"`
	ExtrinsicHash   string       `json:"extrinsic_hash,omitempty"`
	Slot            types.U32    `json:"slot,omitempty"`
	EpochMark       *epochMark   `json:"epoch_mark,omitempty"`
	TicketsMark     *ticketsMark `json:"tickets_mark,omitempty"`
	OffendersMark   []string     `json:"offenders_mark,omitempty"`
	AuthorIndex     types.U16    `json:"author_index,omitempty"`
	EntropySource   string       `json:"entropy_source,omitempty"`
	Seal            string       `json:"seal,omitempty"`
}

type epochMark struct {
	Entropy        string   `json:"entropy,omitempty"`
	TicketsEntropy string   `json:"tickets_entropy,omitempty"`
	Validators     []string `json:"validators,omitempty"`
}

type ticketsMark []ticketBody

type ticketBody struct {
	Id      string   `json:"id,omitempty"`
	Attempt types.U8 `json:"attempt,omitempty"`
}

type extrinsic struct {
	Tickets    []ticketsExtrinsic   `json:"tickets,omitempty"`
	Preimages  []preimagesExtrinsic `json:"preimages"`
	Guarantees []reportGuarantee    `json:"guarantees"`
	Assurances []availAssurance     `json:"assurances,omitempty"`
	Disputes   disputesExtrinsic    `json:"disputes"`
}

type ticketsExtrinsic struct {
	Attempt   types.U8 `json:"attempt,omitempty"`
	Signature string   `json:"signature,omitempty"`
}

type preimagesExtrinsic struct {
	Requester types.U32 `json:"requester,omitempty"`
	Blob      string    `json:"blob,omitempty"`
}

type reportGuarantee struct {
	Report     workReport           `json:"report"`
	Slot       types.U32            `json:"slot,omitempty"`
	Signatures []validatorSignature `json:"signatures,omitempty"`
}

type validatorSignature struct {
	ValidatorIndex types.U16 `json:"validator_index,omitempty"`
	Signature      string    `json:"signature,omitempty"`
}

type workReport struct {
	PackageSpec       workPackageSpec         `json:"package_spec"`
	Context           refineContext           `json:"context"`
	CoreIndex         types.U16               `json:"core_index,omitempty"`
	AuthorizerHash    string                  `json:"authorizer_hash,omitempty"`
	AuthOutput        string                  `json:"auth_output,omitempty"`
	SegmentRootLookup []segmentRootLookupItem `json:"segment_root_lookup,omitempty"`
	Results           []workResult            `json:"results,omitempty"`
}

type workPackageSpec struct {
	Hash         string    `json:"hash,omitempty"`
	Length       types.U32 `json:"length,omitempty"`
	ErasureRoot  string    `json:"erasure_root,omitempty"`
	ExportsRoot  string    `json:"exports_root,omitempty"`
	ExportsCount types.U16 `json:"exports_count,omitempty"`
}

type refineContext struct {
	Anchor           string    `json:"anchor,omitempty"`
	StateRoot        string    `json:"state_root,omitempty"`
	BeefyRoot        string    `json:"beefy_root,omitempty"`
	LookupAnchor     string    `json:"lookup_anchor,omitempty"`
	LookupAnchorSlot types.U32 `json:"lookup_anchor_slot,omitempty"`
	Prerequisites    []string  `json:"prerequisites,omitempty"`
}

type segmentRootLookupItem struct {
	WorkPackageHash string `json:"work_package_hash,omitempty"`
	SegmentTreeRoot string `json:"segment_tree_root,omitempty"`
}

type workResult struct {
	ServiceId     types.U32                     `json:"service_id,omitempty"`
	CodeHash      string                        `json:"code_hash,omitempty"`
	PayloadHash   string                        `json:"payload_hash,omitempty"`
	AccumulateGas types.U64                     `json:"accumulate_gas,omitempty"`
	Result        map[workExecResultType][]byte `json:"result,omitempty"`
}

type workExecResultType string

type availAssurance struct {
	Anchor         string    `json:"anchor,omitempty"`
	Bitfield       string    `json:"bitfield,omitempty"`
	ValidatorIndex types.U16 `json:"validator_index,omitempty"`
	Signature      string    `json:"signature,omitempty"`
}

type disputesExtrinsic struct {
	Verdicts []verdict `json:"verdicts,omitempty"`
	Culprits []culprit `json:"culprits,omitempty"`
	Faults   []fault   `json:"faults,omitempty"`
}

type verdict struct {
	Target string      `json:"target,omitempty"`
	Age    types.U32   `json:"age,omitempty"`
	Votes  []judgement `json:"votes,omitempty"`
}

type judgement struct {
	Vote      bool      `json:"vote,omitempty"`
	Index     types.U16 `json:"index,omitempty"`
	Signature string    `json:"signature,omitempty"`
}

type culprit struct {
	Target    string `json:"target,omitempty"`
	Key       string `json:"key,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type fault struct {
	Target    string `json:"target,omitempty"`
	Vote      bool   `json:"vote,omitempty"`
	Key       string `json:"key,omitempty"`
	Signature string `json:"signature,omitempty"`
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
	// var testExtrinsic types.Block
	byteValue := readDataFromJson(filename)
	// Unmarshal the JSON data
	var testDataFromJson block
	var blockData types.Block
	err := json.Unmarshal(byteValue, &testDataFromJson)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v", err)
		return types.Block{}
	}

	// Read Block Header data from JSON
	blockData.Header = types.Header{
		Parent:          types.HeaderHash(stringToHex(testDataFromJson.Header.Parent)),
		ParentStateRoot: types.StateRoot(stringToHex(testDataFromJson.Header.ParentStateRoot)),
		ExtrinsicHash:   types.OpaqueHash(stringToHex(testDataFromJson.Header.ExtrinsicHash)),
		Slot:            types.TimeSlot(testDataFromJson.Header.Slot),
		AuthorIndex:     types.ValidatorIndex(testDataFromJson.Header.AuthorIndex),
		EntropySource:   types.BandersnatchVrfSignature(stringToHex(testDataFromJson.Header.EntropySource)),
		Seal:            types.BandersnatchVrfSignature(stringToHex(testDataFromJson.Header.Seal)),
	}

	if testDataFromJson.Header.EpochMark != nil {
		blockData.Header.EpochMark = &types.EpochMark{
			Entropy:        types.Entropy(stringToHex(testDataFromJson.Header.EpochMark.Entropy)),
			TicketsEntropy: types.Entropy(stringToHex(testDataFromJson.Header.EpochMark.TicketsEntropy)),
		}

		for _, data := range testDataFromJson.Header.EpochMark.Validators {
			validator := stringToHex(data)
			blockData.Header.EpochMark.Validators = append(blockData.Header.EpochMark.Validators, types.BandersnatchPublic(validator))
		}
	}

	if testDataFromJson.Header.TicketsMark != nil {
		blockData.Header.TicketsMark = &types.TicketsMark{}

		for _, data := range *testDataFromJson.Header.TicketsMark {
			*blockData.Header.TicketsMark = append(*blockData.Header.TicketsMark, types.TicketBody{
				Id:      types.TicketId(stringToHex(data.Id)),
				Attempt: types.TicketAttempt(data.Attempt),
			})
		}
	}

	for _, data := range testDataFromJson.Header.OffendersMark {
		offender := stringToHex(data)
		blockData.Header.OffendersMark = append(blockData.Header.OffendersMark, types.Ed25519Public(offender))
	}

	// Read Block Extrinsic data from JSON
	blockData.Extrinsic = types.Extrinsic{}

	for _, data := range testDataFromJson.Extrinsic.Tickets {
		signature := stringToHex(data.Signature)
		blockData.Extrinsic.Tickets = append(blockData.Extrinsic.Tickets, types.TicketEnvelope{
			Attempt:   types.TicketAttempt(data.Attempt),
			Signature: types.BandersnatchRingVrfSignature(signature),
		})
	}

	for _, data := range testDataFromJson.Extrinsic.Preimages {
		preimage := stringToHex(data.Blob)
		blockData.Extrinsic.Preimages = append(blockData.Extrinsic.Preimages, types.Preimage{
			Requester: types.ServiceId(data.Requester),
			Blob:      types.ByteSequence(preimage),
		})
	}

	for num, data := range testDataFromJson.Extrinsic.Guarantees {
		blockData.Extrinsic.Guarantees = append(blockData.Extrinsic.Guarantees, types.ReportGuarantee{
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
			blockData.Extrinsic.Guarantees[num].Report.Context.Prerequisites = append(blockData.Extrinsic.Guarantees[num].Report.Context.Prerequisites, types.OpaqueHash(stringToHex(data)))
		}

		for _, data := range data.Report.SegmentRootLookup {
			blockData.Extrinsic.Guarantees[num].Report.SegmentRootLookup = append(blockData.Extrinsic.Guarantees[num].Report.SegmentRootLookup, types.SegmentRootLookupItem{
				WorkPackageHash: types.WorkPackageHash(stringToHex(data.WorkPackageHash)),
				SegmentTreeRoot: types.OpaqueHash(stringToHex(data.SegmentTreeRoot)),
			})
		}

		for _, data := range data.Report.Results {
			myMap := make(map[types.WorkExecResultType][]byte)
			for key, value := range data.Result {
				myMap[types.WorkExecResultType(key)] = value
			}
			blockData.Extrinsic.Guarantees[num].Report.Results = append(blockData.Extrinsic.Guarantees[num].Report.Results, types.WorkResult{
				ServiceId:     types.ServiceId(data.ServiceId),
				CodeHash:      types.OpaqueHash(stringToHex(data.CodeHash)),
				PayloadHash:   types.OpaqueHash(stringToHex(data.PayloadHash)),
				AccumulateGas: types.Gas(data.AccumulateGas),
				Result:        myMap,
			})
		}

		for _, data := range data.Signatures {
			signature := stringToHex(data.Signature)
			blockData.Extrinsic.Guarantees[num].Signatures = append(blockData.Extrinsic.Guarantees[num].Signatures, types.ValidatorSignature{
				ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
				Signature:      types.Ed25519Signature(signature),
			})
		}

	}

	for _, data := range testDataFromJson.Extrinsic.Assurances {
		signature := stringToHex(data.Signature)
		blockData.Extrinsic.Assurances = append(blockData.Extrinsic.Assurances, types.AvailAssurance{
			Anchor:         types.OpaqueHash(stringToHex(data.Anchor)),
			Bitfield:       []byte(stringToHex(data.Bitfield)[:]),
			ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
			Signature:      types.Ed25519Signature(signature),
		})
	}

	for num, data := range testDataFromJson.Extrinsic.Disputes.Verdicts {
		blockData.Extrinsic.Disputes.Verdicts = append(blockData.Extrinsic.Disputes.Verdicts, types.Verdict{
			Target: types.OpaqueHash(stringToHex(data.Target)),
			Age:    types.U32(data.Age),
		})
		for _, vote := range data.Votes {
			blockData.Extrinsic.Disputes.Verdicts[num].Votes = append(blockData.Extrinsic.Disputes.Verdicts[num].Votes, types.Judgement{
				Vote:      vote.Vote,
				Index:     types.ValidatorIndex(vote.Index),
				Signature: types.Ed25519Signature(stringToHex(vote.Signature)),
			})
		}
	}

	for _, data := range testDataFromJson.Extrinsic.Disputes.Culprits {
		blockData.Extrinsic.Disputes.Culprits = append(blockData.Extrinsic.Disputes.Culprits, types.Culprit{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	for _, data := range testDataFromJson.Extrinsic.Disputes.Faults {
		blockData.Extrinsic.Disputes.Faults = append(blockData.Extrinsic.Disputes.Faults, types.Fault{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Vote:      data.Vote,
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	return blockData
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
	serializedElements = append(serializedElements, WrapByteArray32(types.ByteArray32(ticketSerializedHash)).Serialize()...)
	serializedElements = append(serializedElements, WrapByteArray32(types.ByteArray32(preimageSerializedHash)).Serialize()...)
	serializedElements = append(serializedElements, WrapByteArray32(types.ByteArray32(gHash)).Serialize()...)
	serializedElements = append(serializedElements, WrapByteArray32(types.ByteArray32(assureanceSerializedHash)).Serialize()...)
	serializedElements = append(serializedElements, WrapByteArray32(types.ByteArray32(disputeSerializedHash)).Serialize()...)

	output = hash.Blake2bHash(serializedElements)

	return output
}

// func payloadHash(extrinsic types.Extrinsic) (output types.OpaqueHash) {
// 	serializedElements := types.ByteSequence{}
// 	serializedElements =  append(serializedElements, SerializeU64(types.U64(extrinsic.Guarantees.CoreIndex))...)

// 	return output
// }

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
			// extrinsicResult := extrinsicSerialization(tc.extrinsic)
			// extrinsicHash := hash.Blake2bHash(extrinsicResult)
			pocExtrinsicHash := pocExtrinsicHash(tc.extrinsic)

			if headerHash != types.OpaqueHash(tc.expectedHeader) {
				t.Errorf("\nExpected Header Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedHeader[:]), hex.EncodeToString(headerHash[:]))
			}

			if pocExtrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
				t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(pocExtrinsicHash[:]))
			}

			// if extrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
			// 	t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(extrinsicHash[:]))
			// }
		})
	}
}

func TestHeaderSerialization(t *testing.T) {
	testCases := []struct {
		name           string
		header         types.Header
		extrinsic      types.Extrinsic
		expectedHeader types.HeaderHash
	}{}

	for i := 0; i <= 11; i++ {
		name := fmt.Sprintf("425530_%03d", i) // 格式化為 425530_000, 425530_001, ...
		header := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header
		extrinsic := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Extrinsic
		var expectedHeaderHash types.HeaderHash

		if i == 11 {
			expectedHeaderHash = readBlockDataFromJson("data/425531_000.json").Header.Parent // 對於 011，使用 425531_000 的 parent
		} else {
			expectedHeaderHash = readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i+1)).Header.Parent
		}

		testCases = append(testCases, struct {
			name           string
			header         types.Header
			extrinsic      types.Extrinsic
			expectedHeader types.HeaderHash
		}{
			name:           name,
			header:         header,
			extrinsic:      extrinsic,
			expectedHeader: expectedHeaderHash,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headerResult := HeaderSerialization(tc.header)
			headerHash := hash.Blake2bHash(headerResult)
			// extrinsicResult := extrinsicSerialization(tc.extrinsic)
			// extrinsicHash := hash.Blake2bHash(extrinsicResult)

			if headerHash != types.OpaqueHash(tc.expectedHeader) {
				t.Errorf("\nExpected Header Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedHeader[:]), hex.EncodeToString(headerHash[:]))
			}
		})
	}
}

func TestExtrinsicSerialization(t *testing.T) {
	testCases := []struct {
		name              string
		header            types.Header
		extrinsic         types.Extrinsic
		expectedExtrinsic types.OpaqueHash
	}{}

	for i := 0; i <= 11; i++ {
		name := fmt.Sprintf("425530_%03d", i) // 格式化為 425530_000, 425530_001, ...
		header := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header
		extrinsic := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Extrinsic
		expectedExtrinsicHash := readBlockDataFromJson(fmt.Sprintf("data/425530_%03d.json", i)).Header.ExtrinsicHash

		testCases = append(testCases, struct {
			name              string
			header            types.Header
			extrinsic         types.Extrinsic
			expectedExtrinsic types.OpaqueHash
		}{
			name:              name,
			header:            header,
			extrinsic:         extrinsic,
			expectedExtrinsic: expectedExtrinsicHash,
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// extrinsicResult := extrinsicSerialization(tc.extrinsic)
			// extrinsicHash := hash.Blake2bHash(extrinsicResult)
			pocExtrinsicHash := pocExtrinsicHash(tc.extrinsic)
			if pocExtrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
				t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(pocExtrinsicHash[:]))
			}

			// if extrinsicHash != types.OpaqueHash(tc.expectedExtrinsic) {
			// 	t.Errorf("\nExpected Extrinsic Hash: %v, \nGot: %v", hex.EncodeToString(tc.expectedExtrinsic[:]), hex.EncodeToString(extrinsicHash[:]))
			// }
		})
	}
}

// func TestExtrinsicSerialization(t *testing.T) {
// 	testCases := []struct {
// 		name      string
// 		extrinsic types.Extrinsic
// 		expected  string
// 	}{
// 		{
// 			name:      "LoadExtrinsic",
// 			extrinsic: readBlockDataFromJson("data/425530_007.json").Extrinsic,
// 			expected:  "b90d1b7ccbc999f230283d67f2a207709e21875d2d2097b6ae24e97db748bde9",
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result := extrinsicSerialization(tc.extrinsic)
// 			resultHash := hash.Blake2bHash(result)

// 			if hex.EncodeToString(resultHash[:]) != tc.expected {
// 				poc := pocExtrinsicHash(tc.extrinsic)
// 				fmt.Printf("POC Hash: %v\n", hex.EncodeToString(poc[:]))
// 				fmt.Printf("oringinal hash: %v\n", hex.EncodeToString(resultHash[:]))
// 				fmt.Printf("expected hash: %v\n", tc.expected)
// 				t.Errorf("Expected: %v, got: %v", tc.expected, hex.EncodeToString(resultHash[:]))
// 			}
// 		})
// 	}
// }
