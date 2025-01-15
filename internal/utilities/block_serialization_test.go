package utilities

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type heder struct {
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
}

type ticketsExtrinsic []ticketEnvelope

type ticketEnvelope struct {
	Attempt   int    `json:"attempt,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type preimagesExtrinsic []preimage

type preimage struct {
	Requester int    `json:"requester,omitempty"`
	Blob      string `json:"blob,omitempty"`
}

type guaranteesExtrinsic []guarantee

type guarantee struct {
	Report     workReport `json:"report"`
	Slot       int        `json:"slot,omitempty"`
	Signatures []struct {
		ValidatorIndex int    `json:"validator_index,omitempty"`
		Signature      string `json:"signature,omitempty"`
	} `json:"signatures,omitempty"`
}

type workResult struct {
	ServiceId     int               `json:"service_id,omitempty"`
	CodeHash      string            `json:"code_hash,omitempty"`
	PayloadHash   string            `json:"payload_hash,omitempty"`
	AccumulateGas int               `json:"accumulate_gas,omitempty"`
	Result        map[string]string `json:"result,omitempty"`
}

type refineContext struct {
	Anchor           string   `json:"anchor,omitempty"`
	StateRoot        string   `json:"state_root,omitempty"`
	BeefyRoot        string   `json:"beefy_root,omitempty"`
	LookupAnchor     string   `json:"lookup_anchor,omitempty"`
	LookupAnchorSlot int      `json:"lookup_anchor_slot,omitempty"`
	Prerequisites    []string `json:"prerequisites,omitempty"`
}

type workReport struct {
	PackageSpec struct {
		Hash         string `json:"hash,omitempty"`
		Length       int    `json:"length,omitempty"`
		ErasureRoot  string `json:"erasure_root,omitempty"`
		ExportsRoot  string `json:"exports_root,omitempty"`
		ExportsCount int    `json:"exports_count,omitempty"`
	} `json:"package_spec"`
	Context           refineContext `json:"context"`
	CoreIndex         int           `json:"core_index,omitempty"`
	AuthorizerHash    string        `json:"authorizer_hash,omitempty"`
	AuthOutput        string        `json:"auth_output,omitempty"`
	SegmentRootLookup []struct {
		WorkPackageHash string `json:"work_package_hash,omitempty"`
		SegmentTreeRoot string `json:"segment_tree_root,omitempty"`
	} `json:"segment_root_lookup,omitempty"`
	Results []workResult `json:"results,omitempty"`
}

type assurancesExtrinsic []availAssurance

type availAssurance struct {
	Anchor         string `json:"anchor,omitempty"`
	Bitfield       string `json:"bitfield,omitempty"`
	ValidatorIndex int    `json:"validator_index,omitempty"`
	Signature      string `json:"signature,omitempty"`
}

type disputesExtrinsic struct {
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
}

type workItem struct {
	Service            int    `json:"service,omitempty"`
	CodeHash           string `json:"code_hash,omitempty"`
	Payload            string `json:"payload,omitempty"`
	RefineGasLimit     int    `json:"refine_gas_limit,omitempty"`
	AccumulateGasLimit int    `json:"accumulate_gas_limit,omitempty"`
	ImportSegments     []struct {
		TreeRoot string `json:"tree_root,omitempty"`
		Index    int    `json:"index,omitempty"`
	} `json:"import_segments,omitempty"`
	Extrinsic []struct {
		Hash string `json:"hash,omitempty"`
		Len  int    `json:"len,omitempty"`
	} `json:"extrinsic,omitempty"`
	ExportCount int `json:"export_count,omitempty"`
}

type workPackage struct {
	Authorization string `json:"authorization,omitempty"`
	AuthCodeHost  int    `json:"auth_code_host,omitempty"`
	Authorizer    struct {
		CodeHash string `json:"code_hash,omitempty"`
		Params   string `json:"params,omitempty"`
	} `json:"authorizer"`
	Context refineContext `json:"context"`
	Items   []workItem    `json:"items,omitempty"`
}

func readDataFromFile(filename string) []byte {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	// Read the file content
	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil
	}
	return byteValue
}

func unmarshalDataFromJson(filename string, data interface{}) {
	// Unmarshal the JSON data
	byteValue := readDataFromFile(filename)

	err := json.Unmarshal(byteValue, &data)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
	}
}

func readHeaderFromJson(filename string, data heder) (outputData types.Header) {
	unmarshalDataFromJson(filename, &data)

	outputData = types.Header{
		Parent:          types.HeaderHash(stringToHex(data.Parent)),
		ParentStateRoot: types.StateRoot(stringToHex(data.ParentStateRoot)),
		ExtrinsicHash:   types.OpaqueHash(stringToHex(data.ExtrinsicHash)),
		Slot:            types.TimeSlot(data.Slot),
		AuthorIndex:     types.ValidatorIndex(data.AuthorIndex),
		EntropySource:   types.BandersnatchVrfSignature(stringToHex(data.EntropySource)),
		Seal:            types.BandersnatchVrfSignature(stringToHex(data.Seal)),
	}

	if data.EpochMark != nil {
		outputData.EpochMark = &types.EpochMark{
			Entropy:        types.Entropy(stringToHex(data.EpochMark.Entropy)),
			TicketsEntropy: types.Entropy(stringToHex(data.EpochMark.TicketsEntropy)),
		}

		for _, data := range data.EpochMark.Validators {
			validator := stringToHex(data)
			outputData.EpochMark.Validators = append(outputData.EpochMark.Validators, types.BandersnatchPublic(validator))
		}
	}

	if data.TicketsMark != nil {
		outputData.TicketsMark = &types.TicketsMark{}

		for _, data := range *data.TicketsMark {
			*outputData.TicketsMark = append(*outputData.TicketsMark, types.TicketBody{
				Id:      types.TicketId(stringToHex(data.Id)),
				Attempt: types.TicketAttempt(data.Attempt),
			})
		}
	}

	if data.OffendersMark != nil {
		for _, data := range data.OffendersMark {
			offender := stringToHex(data)
			outputData.OffendersMark = append(outputData.OffendersMark, types.Ed25519Public(offender))
		}
	}

	return outputData
}

func readExtrinsicTicketFromJson(filename string, data ticketsExtrinsic) (outputData types.TicketsExtrinsic) {
	unmarshalDataFromJson(filename, &data)

	for _, data := range data {
		signature := stringToHex(data.Signature)
		outputData = append(outputData, types.TicketEnvelope{
			Attempt:   types.TicketAttempt(data.Attempt),
			Signature: types.BandersnatchRingVrfSignature(signature),
		})
	}

	return outputData
}

func readExtrinsicPreimageFromJson(filename string, data preimagesExtrinsic) (outputData types.PreimagesExtrinsic) {
	unmarshalDataFromJson(filename, &data)

	for _, data := range data {
		preimage := stringToHex(data.Blob)
		outputData = append(outputData, types.Preimage{
			Requester: types.ServiceId(data.Requester),
			Blob:      types.ByteSequence(preimage),
		})
	}

	return outputData
}

func readExtrinsicWorkResultFromJson(filename string, data workResult) (outputData types.WorkResult) {
	unmarshalDataFromJson(filename, &data)

	myMap := make(map[types.WorkExecResultType][]byte)
	for key, value := range data.Result {
		if value == "" {
			myMap[types.WorkExecResultType(key)] = []byte{}
		} else {
			myMap[types.WorkExecResultType(key)] = stringToHex(value)
		}
	}
	outputData = types.WorkResult{
		ServiceId:     types.ServiceId(data.ServiceId),
		CodeHash:      types.OpaqueHash(stringToHex(data.CodeHash)),
		PayloadHash:   types.OpaqueHash(stringToHex(data.PayloadHash)),
		AccumulateGas: types.Gas(data.AccumulateGas),
		Result:        myMap,
	}

	return outputData
}

func readExtrinsicWorkReportFromJson(filename string, data workReport) (outputData types.WorkReport) {
	unmarshalDataFromJson(filename, &data)

	outputData = types.WorkReport{
		PackageSpec: types.WorkPackageSpec{
			Hash:         types.WorkPackageHash(stringToHex(data.PackageSpec.Hash)),
			Length:       types.U32(data.PackageSpec.Length),
			ErasureRoot:  types.ErasureRoot(stringToHex(data.PackageSpec.ErasureRoot)),
			ExportsRoot:  types.ExportsRoot(stringToHex(data.PackageSpec.ExportsRoot)),
			ExportsCount: types.U16(data.PackageSpec.ExportsCount),
		},
		CoreIndex:      types.CoreIndex(data.CoreIndex),
		AuthorizerHash: types.OpaqueHash(stringToHex(data.AuthorizerHash)),
		AuthOutput:     types.ByteSequence(stringToHex(data.AuthOutput)),
	}
	outputData.Context = readExtrinsicRefineContextFromJson(filename, data.Context)

	for _, data := range data.SegmentRootLookup {
		outputData.SegmentRootLookup = append(outputData.SegmentRootLookup, types.SegmentRootLookupItem{
			WorkPackageHash: types.WorkPackageHash(stringToHex(data.WorkPackageHash)),
			SegmentTreeRoot: types.OpaqueHash(stringToHex(data.SegmentTreeRoot)),
		})
	}

	for _, data := range data.Results {
		outputData.Results = append(outputData.Results, readExtrinsicWorkResultFromJson(filename, data))
	}
	return outputData
}

func readExtrinsicRefineContextFromJson(filename string, data refineContext) (outputData types.RefineContext) {
	unmarshalDataFromJson(filename, &data)

	outputData = types.RefineContext{
		Anchor:           types.HeaderHash(stringToHex(data.Anchor)),
		StateRoot:        types.StateRoot(stringToHex(data.StateRoot)),
		BeefyRoot:        types.BeefyRoot(stringToHex(data.BeefyRoot)),
		LookupAnchor:     types.HeaderHash(stringToHex(data.LookupAnchor)),
		LookupAnchorSlot: types.TimeSlot(data.LookupAnchorSlot),
	}

	for _, data := range data.Prerequisites {
		outputData.Prerequisites = append(outputData.Prerequisites, types.OpaqueHash(stringToHex(data)))
	}

	return outputData
}

func readExtrinsicGuaranteeFromJson(filename string, data guaranteesExtrinsic) (outputData types.GuaranteesExtrinsic) {
	unmarshalDataFromJson(filename, &data)

	for num, data := range data {
		outputData = append(outputData, types.ReportGuarantee{
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
			outputData[num].Report.Context.Prerequisites = append(outputData[num].Report.Context.Prerequisites, types.OpaqueHash(stringToHex(data)))
		}

		for _, data := range data.Report.SegmentRootLookup {
			outputData[num].Report.SegmentRootLookup = append(outputData[num].Report.SegmentRootLookup, types.SegmentRootLookupItem{
				WorkPackageHash: types.WorkPackageHash(stringToHex(data.WorkPackageHash)),
				SegmentTreeRoot: types.OpaqueHash(stringToHex(data.SegmentTreeRoot)),
			})
		}

		for _, data := range data.Report.Results {
			myMap := make(map[types.WorkExecResultType][]byte)
			for key, value := range data.Result {
				if value == "" {
					myMap[types.WorkExecResultType(key)] = []byte{}
				} else {
					myMap[types.WorkExecResultType(key)] = stringToHex(value)
				}
			}
			outputData[num].Report.Results = append(outputData[num].Report.Results, types.WorkResult{
				ServiceId:     types.ServiceId(data.ServiceId),
				CodeHash:      types.OpaqueHash(stringToHex(data.CodeHash)),
				PayloadHash:   types.OpaqueHash(stringToHex(data.PayloadHash)),
				AccumulateGas: types.Gas(data.AccumulateGas),
				Result:        myMap,
			})
		}

		for _, data := range data.Signatures {
			signature := stringToHex(data.Signature)
			outputData[num].Signatures = append(outputData[num].Signatures, types.ValidatorSignature{
				ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
				Signature:      types.Ed25519Signature(signature),
			})
		}
	}
	return outputData
}

func readExtrinsicAssurancesFromJson(filename string, data assurancesExtrinsic) (outputData types.AssurancesExtrinsic) {
	unmarshalDataFromJson(filename, &data)

	for _, data := range data {
		signature := stringToHex(data.Signature)
		outputData = append(outputData, types.AvailAssurance{
			Anchor:         types.OpaqueHash(stringToHex(data.Anchor)),
			Bitfield:       []byte(stringToHex(data.Bitfield)[:]),
			ValidatorIndex: types.ValidatorIndex(data.ValidatorIndex),
			Signature:      types.Ed25519Signature(signature),
		})
	}

	return outputData
}

func readExtrinsicDisputesFromJson(filename string, data disputesExtrinsic) (outputData types.DisputesExtrinsic) {
	unmarshalDataFromJson(filename, &data)

	for num, data := range data.Verdicts {
		outputData.Verdicts = append(outputData.Verdicts, types.Verdict{
			Target: types.OpaqueHash(stringToHex(data.Target)),
			Age:    types.U32(data.Age),
		})
		for _, vote := range data.Votes {
			outputData.Verdicts[num].Votes = append(outputData.Verdicts[num].Votes, types.Judgement{
				Vote:      vote.Vote,
				Index:     types.ValidatorIndex(vote.Index),
				Signature: types.Ed25519Signature(stringToHex(vote.Signature)),
			})
		}
	}

	for _, data := range data.Culprits {
		outputData.Culprits = append(outputData.Culprits, types.Culprit{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	for _, data := range data.Faults {
		outputData.Faults = append(outputData.Faults, types.Fault{
			Target:    types.WorkReportHash(stringToHex(data.Target)),
			Vote:      data.Vote,
			Key:       types.Ed25519Public(stringToHex(data.Key)),
			Signature: types.Ed25519Signature(stringToHex(data.Signature)),
		})
	}

	return outputData
}

func readWorkItemFromJson(filename string, data workItem) (outputData types.WorkItem) {
	unmarshalDataFromJson(filename, &data)

	outputData = types.WorkItem{
		Service:            types.ServiceId(data.Service),
		CodeHash:           types.OpaqueHash(stringToHex(data.CodeHash)),
		Payload:            types.ByteSequence(stringToHex(data.Payload)),
		RefineGasLimit:     types.Gas(data.RefineGasLimit),
		AccumulateGasLimit: types.Gas(data.AccumulateGasLimit),
		ExportCount:        types.U16(data.ExportCount),
	}

	for _, data := range data.ImportSegments {
		outputData.ImportSegments = append(outputData.ImportSegments, types.ImportSpec{
			TreeRoot: types.OpaqueHash(stringToHex(data.TreeRoot)),
			Index:    types.U16(data.Index),
		})
	}

	for _, data := range data.Extrinsic {
		outputData.Extrinsic = append(outputData.Extrinsic, types.ExtrinsicSpec{
			Hash: types.OpaqueHash(stringToHex(data.Hash)),
			Len:  types.U32(data.Len),
		})
	}

	return outputData
}

func readWorkPackageFromJson(filename string, data workPackage) (outputData types.WorkPackage) {
	unmarshalDataFromJson(filename, &data)

	outputData = types.WorkPackage{
		Authorization: types.ByteSequence(stringToHex(data.Authorization)),
		AuthCodeHost:  types.ServiceId(data.AuthCodeHost),
		Authorizer: types.Authorizer{
			CodeHash: types.OpaqueHash(stringToHex(data.Authorizer.CodeHash)),
			Params:   types.ByteSequence(stringToHex(data.Authorizer.Params)),
		},
	}
	outputData.Context = readExtrinsicRefineContextFromJson(filename, data.Context)
	for _, data := range data.Items {
		outputData.Items = append(outputData.Items, readWorkItemFromJson(filename, data))
	}

	return outputData
}

func stringToHex(str string) []byte {
	bytes, err := hex.DecodeString(str[2:])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bytes
}

func TestHeaderSerialization(t *testing.T) {
	testCases := []struct {
		name           string
		header         types.Header
		expectedResult []byte
	}{
		{
			name:           "header_0",
			header:         readHeaderFromJson("data/header_0.json", heder{}),
			expectedResult: readDataFromFile("data/header_0.bin"),
		},
		{
			name:           "header_1",
			header:         readHeaderFromJson("data/header_1.json", heder{}),
			expectedResult: readDataFromFile("data/header_1.bin"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serializationResult := HeaderSerialization(tc.header)
			if !bytes.Equal(tc.expectedResult, serializationResult) {
				t.Errorf("\nExpected Serialization result: %v, \nGot: %v", tc.expectedResult, serializationResult)
			}
		})
	}
}

func TestExtrinsicTicketSerialization(t *testing.T) {
	extrinsicTicketData := readExtrinsicTicketFromJson("data/tickets_extrinsic.json", ticketsExtrinsic{})
	expectedResult := readDataFromFile("data/tickets_extrinsic.bin")
	serializationResult := ExtrinsicTicketSerialization(extrinsicTicketData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestExtrinsicPreimageSerialization(t *testing.T) {
	extrinsicPreimageData := readExtrinsicPreimageFromJson("data/preimages_extrinsic.json", preimagesExtrinsic{})
	expectedResult := readDataFromFile("data/preimages_extrinsic.bin")
	serializationResult := ExtrinsicPreimageSerialization(extrinsicPreimageData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestWorkResultSerialization(t *testing.T) {
	testCases := []struct {
		name           string
		workResult     types.WorkResult
		expectedResult []byte
	}{
		{
			name:           "work_result_0",
			workResult:     readExtrinsicWorkResultFromJson("data/work_result_0.json", workResult{}),
			expectedResult: readDataFromFile("data/work_result_0.bin"),
		},
		{
			name:           "work_result_1",
			workResult:     readExtrinsicWorkResultFromJson("data/work_result_1.json", workResult{}),
			expectedResult: readDataFromFile("data/work_result_1.bin"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serializationResult := WorkResultSerialization(tc.workResult)
			if !bytes.Equal(tc.expectedResult, serializationResult) {
				t.Errorf("\nExpected Serialization result: %v, \nGot: %v", tc.expectedResult, serializationResult)
			}
		})
	}
}

func TestWorkReportSerialization(t *testing.T) {
	extrinsicWorkReportData := readExtrinsicWorkReportFromJson("data/work_report.json", workReport{})
	expectedResult := readDataFromFile("data/work_report.bin")
	serializationResult := WorkReportSerialization(extrinsicWorkReportData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestRefineContextSerialization(t *testing.T) {
	extrinsicRefineContextData := readExtrinsicRefineContextFromJson("data/refine_context.json", refineContext{})
	expectedResult := readDataFromFile("data/refine_context.bin")
	serializationResult := RefineContextSerialization(extrinsicRefineContextData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestExtrinsicGuaranteeSerialization(t *testing.T) {
	extrinsicGuaranteeData := readExtrinsicGuaranteeFromJson("data/guarantees_extrinsic.json", guaranteesExtrinsic{})
	expectedResult := readDataFromFile("data/guarantees_extrinsic.bin")
	serializationResult := ExtrinsicGuaranteeSerialization(extrinsicGuaranteeData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestExtrinsicAssuranceSerialization(t *testing.T) {
	extrinsicAssuranceData := readExtrinsicAssurancesFromJson("data/assurances_extrinsic.json", assurancesExtrinsic{})
	expectedResult := readDataFromFile("data/assurances_extrinsic.bin")
	serializationResult := ExtrinsicAssuranceSerialization(extrinsicAssuranceData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestExtrinsicDisputeSerialization(t *testing.T) {
	extrinsicDisputeData := readExtrinsicDisputesFromJson("data/disputes_extrinsic.json", disputesExtrinsic{})
	expectedResult := readDataFromFile("data/disputes_extrinsic.bin")
	serializationResult := ExtrinsicDisputeSerialization(extrinsicDisputeData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestWorkItemSerialization(t *testing.T) {
	extrinsicWorkItemData := readWorkItemFromJson("data/work_item.json", workItem{})
	expectedResult := readDataFromFile("data/work_item.bin")
	serializationResult := WorkItemSerialization(extrinsicWorkItemData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestSerializeWorkPackage(t *testing.T) {
	extrinsicWorkPackageData := readWorkPackageFromJson("data/work_package.json", workPackage{})
	expectedResult := readDataFromFile("data/work_package.bin")
	serializationResult := SerializeWorkPackage(extrinsicWorkPackageData)

	if !bytes.Equal(expectedResult, serializationResult) {
		t.Errorf("\nExpected Serialization result: %v, \nGot: %v", expectedResult, serializationResult)
	}
}

func TestDeferredTransferSerialization(t *testing.T) {
	// Create a default RefineContext with sample values
	defaultDeferredTransfer := types.DeferredTransfer{}

	// Expected output (adjust based on SerializeByteArray and SerializeFixedLength behavior)
	expectedOutput := make([]byte, 152)

	result := DeferredTransferSerialization(defaultDeferredTransfer)

	if !bytes.Equal(result, expectedOutput) {
		t.Errorf("RefineContextSerialization() = %v, want %v", result, expectedOutput)
	}
}
