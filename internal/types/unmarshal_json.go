package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func (v *Validator) UnmarshalJSON(data []byte) error {
	var temp struct {
		Bandersnatch string `json:"bandersnatch,omitempty"`
		Ed25519      string `json:"ed25519,omitempty"`
		Bls          string `json:"bls,omitempty"`
		Metadata     string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	bandersnatchBytes, err := hex.DecodeString(temp.Bandersnatch[2:])
	if err != nil {
		return err
	}

	ed25519Bytes, err := hex.DecodeString(temp.Ed25519[2:])
	if err != nil {
		return err
	}

	blsBytes, err := hex.DecodeString(temp.Bls[2:])
	if err != nil {
		return err
	}

	metadataBytes, err := hex.DecodeString(temp.Metadata[2:])
	if err != nil {
		return err
	}

	v.Bandersnatch = BandersnatchPublic(bandersnatchBytes)
	v.Ed25519 = Ed25519Public(ed25519Bytes)
	v.Bls = BlsPublic(blsBytes)
	v.Metadata = ValidatorMetadata(metadataBytes)

	return nil
}

func (s *ServiceInfo) UnmarshalJSON(data []byte) error {
	var temp struct {
		CodeHash             string    `json:"code_hash,omitempty"`
		Balance              U64       `json:"balance,omitempty"`
		MinItemGas           Gas       `json:"min_item_gas,omitempty"`
		MinMemoGas           Gas       `json:"min_memo_gas,omitempty"`
		Bytes                U64       `json:"bytes,omitempty"`
		DepositOffset        U64       `json:"deposit_offset,omitempty"`
		Items                U32       `json:"items,omitempty"`
		CreationSlot         TimeSlot  `json:"creation_slot,omitempty"`
		LastAccumulationSlot TimeSlot  `json:"last_accumulation_slot,omitempty"`
		ParentService        ServiceId `json:"parent_service,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	s.DepositOffset = temp.DepositOffset

	codeHashBytes, err := hex.DecodeString(temp.CodeHash[2:])
	if err != nil {
		return err
	}
	s.CodeHash = OpaqueHash(codeHashBytes)

	s.Balance = temp.Balance
	s.MinItemGas = temp.MinItemGas
	s.MinMemoGas = temp.MinMemoGas
	s.CreationSlot = temp.CreationSlot
	s.LastAccumulationSlot = temp.LastAccumulationSlot
	s.ParentService = temp.ParentService
	s.Bytes = temp.Bytes
	s.Items = temp.Items

	return nil
}

func (r *RefineContext) UnmarshalJSON(data []byte) error {
	var temp struct {
		Anchor           string   `json:"anchor,omitempty"`
		StateRoot        string   `json:"state_root,omitempty"`
		BeefyRoot        string   `json:"beefy_root,omitempty"`
		LookupAnchor     string   `json:"lookup_anchor,omitempty"`
		LookupAnchorSlot U64      `json:"lookup_anchor_slot,omitempty"`
		Prerequisites    []string `json:"prerequisites,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	anchorBytes, err := hex.DecodeString(temp.Anchor[2:])
	if err != nil {
		return err
	}
	r.Anchor = HeaderHash(anchorBytes)

	stateRootBytes, err := hex.DecodeString(temp.StateRoot[2:])
	if err != nil {
		return err
	}
	r.StateRoot = StateRoot(stateRootBytes)

	beefyRootBytes, err := hex.DecodeString(temp.BeefyRoot[2:])
	if err != nil {
		return err
	}
	r.BeefyRoot = BeefyRoot(beefyRootBytes)

	lookupAnchorBytes, err := hex.DecodeString(temp.LookupAnchor[2:])
	if err != nil {
		return err
	}
	r.LookupAnchor = HeaderHash(lookupAnchorBytes)

	r.LookupAnchorSlot = TimeSlot(temp.LookupAnchorSlot)

	for _, prerequisite := range temp.Prerequisites {
		prerequisiteBytes, err := hex.DecodeString(prerequisite[2:])
		if err != nil {
			return err
		}
		r.Prerequisites = append(r.Prerequisites, OpaqueHash(prerequisiteBytes))
	}

	return nil
}

func (a *Authorizer) UnmarshalJSON(data []byte) error {
	var temp struct {
		CodeHash string `json:"code_hash,omitempty"`
		Params   string `json:"params,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	codeHashBytes, err := hex.DecodeString(temp.CodeHash[2:])
	if err != nil {
		return err
	}
	a.CodeHash = OpaqueHash(codeHashBytes)

	paramsBytes, err := hex.DecodeString(temp.Params[2:])
	if err != nil {
		return err
	}
	a.Params = ByteSequence(paramsBytes)

	return nil
}

func (i *ImportSpec) UnmarshalJSON(data []byte) error {
	var temp struct {
		TreeRoot string `json:"tree_root,omitempty"`
		Index    U16    `json:"index,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	treeRootBytes, err := hex.DecodeString(temp.TreeRoot[2:])
	if err != nil {
		return err
	}
	i.TreeRoot = OpaqueHash(treeRootBytes)

	i.Index = temp.Index

	return nil
}

func (e *ExtrinsicSpec) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash string `json:"hash,omitempty"`
		Len  U32    `json:"len,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}
	e.Hash = OpaqueHash(hashBytes)

	e.Len = temp.Len

	return nil
}

func (w *WorkItem) UnmarshalJSON(data []byte) error {
	var temp struct {
		Service            U32             `json:"service,omitempty"`
		CodeHash           string          `json:"code_hash,omitempty"`
		Payload            string          `json:"payload,omitempty"`
		RefineGasLimit     Gas             `json:"refine_gas_limit,omitempty"`
		AccumulateGasLimit Gas             `json:"accumulate_gas_limit,omitempty"`
		ImportSegments     []ImportSpec    `json:"import_segments,omitempty"`
		Extrinsic          []ExtrinsicSpec `json:"extrinsic,omitempty"`
		ExportCount        U16             `json:"export_count,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	w.Service = ServiceId(temp.Service)

	codeHashBytes, err := hex.DecodeString(temp.CodeHash[2:])
	if err != nil {
		return err
	}
	w.CodeHash = OpaqueHash(codeHashBytes)

	payloadBytes, err := hex.DecodeString(temp.Payload[2:])
	if err != nil {
		return err
	}
	w.Payload = ByteSequence(payloadBytes)

	w.RefineGasLimit = temp.RefineGasLimit
	w.AccumulateGasLimit = temp.AccumulateGasLimit
	w.ImportSegments = temp.ImportSegments
	w.Extrinsic = temp.Extrinsic
	w.ExportCount = temp.ExportCount

	return nil
}

func (w *WorkPackage) UnmarshalJSON(data []byte) error {
	var temp struct {
		Authorization    string        `json:"authorization,omitempty"`
		AuthCodeHost     U32           `json:"auth_code_host,omitempty"`
		AuthCodeHash     string        `json:"auth_code_hash,omitempty"`
		AuthorizerConfig string        `json:"authorizer_config,omitempty"`
		Context          RefineContext `json:"context"`
		Items            []WorkItem    `json:"items,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	authorizationBytes, err := hex.DecodeString(temp.Authorization[2:])
	if err != nil {
		return err
	}
	w.Authorization = ByteSequence(authorizationBytes)

	w.AuthCodeHost = ServiceId(temp.AuthCodeHost)

	codeHashBytes, err := hex.DecodeString(temp.AuthCodeHash[2:])
	if err != nil {
		return err
	}

	paramsBytes, err := hex.DecodeString(temp.AuthorizerConfig[2:])
	if err != nil {
		return err
	}

	w.AuthCodeHash = OpaqueHash(codeHashBytes)
	w.AuthorizerConfig = ByteSequence(paramsBytes)
	w.Context = temp.Context
	w.Items = temp.Items

	return nil
}

// RefineLoad
func (r *RefineLoad) UnmarshalJSON(data []byte) error {
	var temp struct {
		GasUsed        Gas `json:"gas_used,omitempty"`
		Imports        U16 `json:"imports,omitempty"`
		ExtrinsicCount U16 `json:"extrinsic_count,omitempty"`
		ExtrinsicSize  U32 `json:"extrinsic_size,omitempty"`
		Exports        U16 `json:"exports,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.GasUsed = temp.GasUsed
	r.Imports = temp.Imports
	r.ExtrinsicCount = temp.ExtrinsicCount
	r.ExtrinsicSize = temp.ExtrinsicSize
	r.Exports = temp.Exports

	return nil
}

func (w *WorkResult) UnmarshalJSON(data []byte) error {
	var temp struct {
		ServiceId     U32            `json:"service_id,omitempty"`
		CodeHash      string         `json:"code_hash,omitempty"`
		PayloadHash   string         `json:"payload_hash,omitempty"`
		AccumulateGas Gas            `json:"accumulate_gas,omitempty"`
		Result        WorkExecResult `json:"result,omitempty"`
		RefineLoad    RefineLoad     `json:"refine_load,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	w.ServiceId = ServiceId(temp.ServiceId)

	codeHashBytes, err := hex.DecodeString(temp.CodeHash[2:])
	if err != nil {
		return err
	}
	w.CodeHash = OpaqueHash(codeHashBytes)

	payloadHashBytes, err := hex.DecodeString(temp.PayloadHash[2:])
	if err != nil {
		return err
	}
	w.PayloadHash = OpaqueHash(payloadHashBytes)

	w.AccumulateGas = temp.AccumulateGas
	w.Result = temp.Result
	w.RefineLoad = temp.RefineLoad

	return nil
}

func (w *WorkPackageSpec) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash         string `json:"hash,omitempty"`
		Length       U32    `json:"length,omitempty"`
		ErasureRoot  string `json:"erasure_root,omitempty"`
		ExportsRoot  string `json:"exports_root,omitempty"`
		ExportsCount U16    `json:"exports_count,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}
	w.Hash = WorkPackageHash(hashBytes)

	w.Length = temp.Length

	erasureRootBytes, err := hex.DecodeString(temp.ErasureRoot[2:])
	if err != nil {
		return err
	}
	w.ErasureRoot = ErasureRoot(erasureRootBytes)

	exportsRootBytes, err := hex.DecodeString(temp.ExportsRoot[2:])
	if err != nil {
		return err
	}
	w.ExportsRoot = ExportsRoot(exportsRootBytes)

	w.ExportsCount = temp.ExportsCount

	return nil
}

func (s *SegmentRootLookupItem) UnmarshalJSON(data []byte) error {
	// jam-test-vectors
	var temp struct {
		WorkPackageHash string `json:"work_package_hash,omitempty"`
		SegmentTreeRoot string `json:"segment_tree_root,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// To read file from different test data source (davxy/jam-test-vectors, jam-duna/jamtestnet)
	if temp.WorkPackageHash == "" && temp.SegmentTreeRoot == "" {
		// jamtestnet
		var temp2 struct {
			WorkPackageHash string `json:"hash,omitempty"`
			SegmentTreeRoot string `json:"exports_root,omitempty"`
		}

		if err := json.Unmarshal(data, &temp2); err != nil {
			return err
		}

		temp.WorkPackageHash = temp2.WorkPackageHash
		temp.SegmentTreeRoot = temp2.SegmentTreeRoot
	}

	workPackageHashBytes, err := hex.DecodeString(temp.WorkPackageHash[2:])
	if err != nil {
		return err
	}
	s.WorkPackageHash = WorkPackageHash(workPackageHashBytes)

	segmentTreeRootBytes, err := hex.DecodeString(temp.SegmentTreeRoot[2:])
	if err != nil {
		return err
	}
	s.SegmentTreeRoot = OpaqueHash(segmentTreeRootBytes)

	return nil
}

func (w *WorkReport) UnmarshalJSON(data []byte) error {
	var temp struct {
		PackageSpec       WorkPackageSpec   `json:"package_spec"`
		Context           RefineContext     `json:"context"`
		CoreIndex         CoreIndex         `json:"core_index,omitempty"`
		AuthorizerHash    string            `json:"authorizer_hash,omitempty"`
		AuthOutput        string            `json:"auth_output,omitempty"`
		SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"`
		Results           []WorkResult      `json:"results,omitempty"`
		AuthGasUsed       Gas               `json:"auth_gas_used,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	w.PackageSpec = temp.PackageSpec
	w.Context = temp.Context
	w.CoreIndex = temp.CoreIndex

	authorizerHashBytes, err := hex.DecodeString(temp.AuthorizerHash[2:])
	if err != nil {
		return err
	}
	w.AuthorizerHash = OpaqueHash(authorizerHashBytes)

	authOutputBytes, err := hex.DecodeString(temp.AuthOutput[2:])
	if err != nil {
		return err
	}

	// if authOutputBytes is empty, set to nil
	if len(authOutputBytes) == 0 {
		w.AuthOutput = nil
	} else {
		w.AuthOutput = ByteSequence(authOutputBytes)
	}

	w.SegmentRootLookup = temp.SegmentRootLookup
	w.Results = temp.Results
	w.AuthGasUsed = temp.AuthGasUsed

	return nil
}

func (m *Mmr) UnmarshalJSON(data []byte) error {
	var temp struct {
		Peaks []string `json:"peaks,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	for _, peak := range temp.Peaks {
		if peak == "" {
			m.Peaks = append(m.Peaks, nil)
			continue
		}

		peakBytes, err := hex.DecodeString(peak[2:])
		if err != nil {
			return err
		}
		m.Peaks = append(m.Peaks, MmrPeak(peakBytes))
	}

	return nil
}

func (r *ReportedWorkPackage) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash        string `json:"hash,omitempty"`
		ExportsRoot string `json:"exports_root,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}
	r.Hash = WorkReportHash(hashBytes)

	exportsRootBytes, err := hex.DecodeString(temp.ExportsRoot[2:])
	if err != nil {
		return err
	}
	r.ExportsRoot = ExportsRoot(exportsRootBytes)

	return nil
}

func (b *BlockInfo) UnmarshalJSON(data []byte) error {
	var temp struct {
		HeaderHash string                `json:"header_hash,omitempty"`
		BeefyRoot  string                `json:"beefy_root,omitempty"`
		StateRoot  string                `json:"state_root,omitempty"`
		Reported   []ReportedWorkPackage `json:"reported,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	headerHashBytes, err := hex.DecodeString(temp.HeaderHash[2:])
	if err != nil {
		return err
	}
	b.HeaderHash = HeaderHash(headerHashBytes)

	beefyRootBytes, err := hex.DecodeString(temp.BeefyRoot[2:])
	if err != nil {
		return err
	}
	b.BeefyRoot = OpaqueHash(beefyRootBytes)

	stateRootBytes, err := hex.DecodeString(temp.StateRoot[2:])
	if err != nil {
		return err
	}
	b.StateRoot = StateRoot(stateRootBytes)

	if len(temp.Reported) == 0 {
		b.Reported = nil
	} else {
		b.Reported = temp.Reported
	}

	return nil
}

func (t *TicketEnvelope) UnmarshalJSON(data []byte) error {
	var temp struct {
		Attempt   U8     `json:"attempt,omitempty"`
		Signature string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}
	t.Signature = BandersnatchRingVrfSignature(signatureBytes)
	t.Attempt = TicketAttempt(temp.Attempt)

	return nil
}

func (t *TicketBody) UnmarshalJSON(data []byte) error {
	var temp struct {
		Id      string `json:"id,omitempty"`
		Attempt U8     `json:"attempt,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	idBytes, err := hex.DecodeString(temp.Id[2:])
	if err != nil {
		return err
	}
	t.Id = TicketId(idBytes)
	t.Attempt = TicketAttempt(temp.Attempt)

	return nil
}

func (t *TicketsOrKeys) UnmarshalJSON(data []byte) error {
	var temp struct {
		Tickets []TicketBody         `json:"tickets,omitempty"`
		Keys    []BandersnatchPublic `json:"keys,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Tickets) > 0 {
		t.Tickets = temp.Tickets
	}

	if len(temp.Keys) > 0 {
		t.Keys = temp.Keys
	}

	return nil
}

func (j *Judgement) UnmarshalJSON(data []byte) error {
	var temp struct {
		Vote      bool   `json:"vote,omitempty"`
		Index     U32    `json:"index,omitempty"`
		Signature string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	j.Vote = temp.Vote
	j.Index = ValidatorIndex(temp.Index)

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}
	j.Signature = Ed25519Signature(signatureBytes)

	return nil
}

func (v *Verdict) UnmarshalJSON(data []byte) error {
	var temp struct {
		Target string      `json:"target,omitempty"`
		Age    U32         `json:"age,omitempty"`
		Votes  []Judgement `json:"votes,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	targetBytes, err := hex.DecodeString(temp.Target[2:])
	if err != nil {
		return err
	}
	v.Target = OpaqueHash(targetBytes)

	v.Age = temp.Age
	v.Votes = temp.Votes

	return nil
}

func (c *Culprit) UnmarshalJSON(data []byte) error {
	var temp struct {
		Target    string `json:"target,omitempty"`
		Key       string `json:"key,omitempty"`
		Signature string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	targetBytes, err := hex.DecodeString(temp.Target[2:])
	if err != nil {
		return err
	}

	keyBytes, err := hex.DecodeString(temp.Key[2:])
	if err != nil {
		return err
	}

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}

	c.Target = WorkReportHash(targetBytes)
	c.Key = Ed25519Public(keyBytes)
	c.Signature = Ed25519Signature(signatureBytes)

	return nil
}

func (f *Fault) UnmarshalJSON(data []byte) error {
	var temp struct {
		Target    string `json:"target,omitempty"`
		Vote      bool   `json:"vote,omitempty"`
		Key       string `json:"key,omitempty"`
		Signature string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	targetBytes, err := hex.DecodeString(temp.Target[2:])
	if err != nil {
		return err
	}

	keyBytes, err := hex.DecodeString(temp.Key[2:])
	if err != nil {
		return err
	}

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}

	f.Target = WorkReportHash(targetBytes)
	f.Vote = temp.Vote
	f.Key = Ed25519Public(keyBytes)
	f.Signature = Ed25519Signature(signatureBytes)

	return nil
}

func (d *DisputesRecords) UnmarshalJSON(data []byte) error {
	var temp struct {
		Good      []string `json:"good,omitempty"`
		Bad       []string `json:"bad,omitempty"`
		Wonky     []string `json:"wonky,omitempty"`
		Offenders []string `json:"offenders,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	for _, good := range temp.Good {
		goodBytes, err := hex.DecodeString(good[2:])
		if err != nil {
			return err
		}
		d.Good = append(d.Good, WorkReportHash(goodBytes))
	}

	for _, bad := range temp.Bad {
		badBytes, err := hex.DecodeString(bad[2:])
		if err != nil {
			return err
		}
		d.Bad = append(d.Bad, WorkReportHash(badBytes))
	}

	for _, wonky := range temp.Wonky {
		wonkyBytes, err := hex.DecodeString(wonky[2:])
		if err != nil {
			return err
		}
		d.Wonky = append(d.Wonky, WorkReportHash(wonkyBytes))
	}

	for _, offender := range temp.Offenders {
		offenderBytes, err := hex.DecodeString(offender[2:])
		if err != nil {
			return err
		}
		d.Offenders = append(d.Offenders, Ed25519Public(offenderBytes))
	}

	return nil
}

func (p *Preimage) UnmarshalJSON(data []byte) error {
	var temp struct {
		Requester U32    `json:"requester,omitempty"`
		Blob      string `json:"blob,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	p.Requester = ServiceId(temp.Requester)

	blobBytes, err := hex.DecodeString(temp.Blob[2:])
	if err != nil {
		return err
	}

	if len(blobBytes) == 0 {
		p.Blob = nil
	} else {
		p.Blob = blobBytes
	}

	return nil
}

func (a *AvailAssurance) UnmarshalJSON(data []byte) error {
	var temp struct {
		Anchor         string `json:"anchor,omitempty"`
		Bitfield       string `json:"bitfield,omitempty"`
		ValidatorIndex U32    `json:"validator_index,omitempty"`
		Signature      string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	anchorBytes, err := hex.DecodeString(temp.Anchor[2:])
	if err != nil {
		return err
	}
	a.Anchor = OpaqueHash(anchorBytes)

	bitfield, err := MakeBitfieldFromHexString(temp.Bitfield)
	if err != nil {
		return err
	}
	a.Bitfield = bitfield

	a.ValidatorIndex = ValidatorIndex(temp.ValidatorIndex)

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}
	a.Signature = Ed25519Signature(signatureBytes)

	return nil
}

func (bf *Bitfield) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}

	bitfield, err := MakeBitfieldFromHexString(hexStr)
	if err != nil {
		return err
	}

	*bf = bitfield

	return nil
}

func (v *ValidatorSignature) UnmarshalJSON(data []byte) error {
	var temp struct {
		ValidatorIndex U32    `json:"validator_index,omitempty"`
		Signature      string `json:"signature,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	v.ValidatorIndex = ValidatorIndex(temp.ValidatorIndex)

	signatureBytes, err := hex.DecodeString(temp.Signature[2:])
	if err != nil {
		return err
	}
	v.Signature = Ed25519Signature(signatureBytes)

	return nil
}

//	type EpochMarkValidatorKeys struct {
//		Bandersnatch BandersnatchPublic
//		Ed25519      Ed25519Public
//	}

// EpochMarkValidatorKeys
func (e *EpochMarkValidatorKeys) UnmarshalJSON(data []byte) error {
	var temp struct {
		Bandersnatch string `json:"bandersnatch,omitempty"`
		Ed25519      string `json:"ed25519,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	bandersnatchBytes, err := hex.DecodeString(temp.Bandersnatch[2:])
	if err != nil {
		return err
	}

	ed25519Bytes, err := hex.DecodeString(temp.Ed25519[2:])
	if err != nil {
		return err
	}

	e.Bandersnatch = BandersnatchPublic(bandersnatchBytes)
	e.Ed25519 = Ed25519Public(ed25519Bytes)

	return nil
}

func (e *EpochMark) UnmarshalJSON(data []byte) error {
	var temp struct {
		Entropy        string                   `json:"entropy,omitempty"`
		TicketsEntropy string                   `json:"tickets_entropy,omitempty"`
		Validators     []EpochMarkValidatorKeys `json:"validators,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	entropyBytes, err := hex.DecodeString(temp.Entropy[2:])
	if err != nil {
		return err
	}
	e.Entropy = Entropy(entropyBytes)

	ticketsEntropyBytes, err := hex.DecodeString(temp.TicketsEntropy[2:])
	if err != nil {
		return err
	}
	e.TicketsEntropy = Entropy(ticketsEntropyBytes)

	e.Validators = temp.Validators

	return nil
}

func (h *Header) UnmarshalJSON(data []byte) error {
	var temp struct {
		Parent          string                   `json:"parent,omitempty"`
		ParentStateRoot string                   `json:"parent_state_root,omitempty"`
		ExtrinsicHash   string                   `json:"extrinsic_hash,omitempty"`
		Slot            U64                      `json:"slot,omitempty"`
		EpochMark       *EpochMark               `json:"epoch_mark,omitempty"`
		TicketsMark     *TicketsMark             `json:"tickets_mark,omitempty"`
		OffendersMark   OffendersMark            `json:"offenders_mark,omitempty"`
		AuthorIndex     U32                      `json:"author_index,omitempty"`
		EntropySource   BandersnatchVrfSignature `json:"entropy_source,omitempty"`
		Seal            BandersnatchVrfSignature `json:"seal,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	parentBytes, err := hex.DecodeString(temp.Parent[2:])
	if err != nil {
		return err
	}
	h.Parent = HeaderHash(parentBytes)

	parentStateRootBytes, err := hex.DecodeString(temp.ParentStateRoot[2:])
	if err != nil {
		return err
	}
	h.ParentStateRoot = StateRoot(parentStateRootBytes)

	extrinsicHashBytes, err := hex.DecodeString(temp.ExtrinsicHash[2:])
	if err != nil {
		return err
	}
	h.ExtrinsicHash = OpaqueHash(extrinsicHashBytes)

	h.Slot = TimeSlot(temp.Slot)
	h.EpochMark = temp.EpochMark
	h.TicketsMark = temp.TicketsMark
	h.OffendersMark = temp.OffendersMark
	h.AuthorIndex = ValidatorIndex(temp.AuthorIndex)
	h.EntropySource = temp.EntropySource
	h.Seal = temp.Seal

	return nil
}

// Entropy
func (e *Entropy) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(e[:], decoded)
	return nil
}

func (w *WorkExecResult) UnmarshalJSON(data []byte) error {
	raw := make(map[string]string)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*w = make(WorkExecResult)

	for key, value := range raw {
		// To read test file from jam-test-vectors traces
		// replace "_" with "-" in the key
		key = strings.ReplaceAll(key, "_", "-")
		resultType := WorkExecResultType(key)

		if key == "ok" {
			// "ok" is a byte sequence
			decoded, err := hex.DecodeString(value[2:])
			if err != nil {
				return fmt.Errorf("failed to decode hex for key %s: %w", key, err)
			}

			// if decoded is empty, set to nil
			if len(decoded) == 0 {
				(*w)[resultType] = nil
			} else {
				(*w)[resultType] = decoded
			}
		}

		if value == "" {
			(*w)[resultType] = nil
		}
	}

	return nil
}

func (a *AuthorizerHash) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(a[:], decoded)
	return nil
}

func (w *WorkPackageHash) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(w[:], decoded)
	return nil
}

func (s *StateRoot) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(s[:], decoded)
	return nil
}

type AccountInfoHistory map[LookupMetaMapkey]TimeSlotSet

type AccountInfo struct {
	Preimages PreimagesExtrinsic `json:"preimages"`
	History   AccountInfoHistory `json:"history"`
}

func (aih *AccountInfoHistory) UnmarshalJSON(data []byte) error {
	var raw []struct {
		Key   LookupMetaMapkey `json:"key"`
		Value TimeSlotSet      `json:"value"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*aih = make(AccountInfoHistory)
	for _, item := range raw {
		(*aih)[item.Key] = item.Value
	}

	return nil
}

// unmarshal AccumulateRoot
func (a *AccumulateRoot) UnmarshalJSON(data []byte) error {
	decoded, err := parseFixedByteArray(data, 32)
	if err != nil {
		return err
	}
	copy(a[:], decoded)
	return nil
}

// SegmentRootLookup
func (s *SegmentRootLookup) UnmarshalJSON(data []byte) error {
	var temp []SegmentRootLookupItem
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*s = temp
	return nil
}

// OffendersMark
func (o *OffendersMark) UnmarshalJSON(data []byte) error {
	var temp []string
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	for _, offender := range temp {
		offenderBytes, err := hex.DecodeString(offender[2:])
		if err != nil {
			return err
		}
		*o = append(*o, Ed25519Public(offenderBytes))
	}

	return nil
}

// DisputesExtrinsic
func (d *DisputesExtrinsic) UnmarshalJSON(data []byte) error {
	var temp struct {
		Verdicts []Verdict `json:"verdicts,omitempty"`
		Culprits []Culprit `json:"culprits,omitempty"`
		Faults   []Fault   `json:"faults,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Verdicts) == 0 {
		d.Verdicts = nil
	} else {
		d.Verdicts = temp.Verdicts
	}

	if len(temp.Culprits) == 0 {
		d.Culprits = nil
	} else {
		d.Culprits = temp.Culprits
	}

	if len(temp.Faults) == 0 {
		d.Faults = nil
	} else {
		d.Faults = temp.Faults
	}

	return nil
}

// Extrinsic
func (e *Extrinsic) UnmarshalJSON(data []byte) error {
	var temp struct {
		Tickets    TicketsExtrinsic    `json:"tickets,omitempty"`
		Preimages  PreimagesExtrinsic  `json:"preimages"`
		Guarantees GuaranteesExtrinsic `json:"guarantees"`
		Assurances AssurancesExtrinsic `json:"assurances,omitempty"`
		Disputes   DisputesExtrinsic   `json:"disputes"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Tickets) == 0 {
		e.Tickets = nil
	} else {
		e.Tickets = temp.Tickets
	}

	if len(temp.Preimages) == 0 {
		e.Preimages = nil
	} else {
		e.Preimages = temp.Preimages
	}

	if len(temp.Guarantees) == 0 {
		e.Guarantees = nil
	} else {
		e.Guarantees = temp.Guarantees
	}

	if len(temp.Assurances) == 0 {
		e.Assurances = nil
	} else {
		e.Assurances = temp.Assurances
	}

	e.Disputes = temp.Disputes

	return nil
}

// TicketsAccumulator
func (t *TicketsAccumulator) UnmarshalJSON(data []byte) error {
	var temp []TicketBody
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*t = temp
	return nil
}

// AuthPool
func (a *AuthPool) UnmarshalJSON(data []byte) error {
	var temp []AuthorizerHash
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*a = temp

	return nil
}

// AuthPools
func (a *AuthPools) UnmarshalJSON(data []byte) error {
	var temp []AuthPool
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*a = temp
	return nil
}

// AvailabilityAssignment
func (a *AvailabilityAssignment) UnmarshalJSON(data []byte) error {
	var temp struct {
		Report  WorkReport `json:"report"`
		Timeout TimeSlot   `json:"timeout,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Report = temp.Report
	a.Timeout = temp.Timeout

	return nil
}

// AvailabilityAssignments
func (a *AvailabilityAssignments) UnmarshalJSON(data []byte) error {
	var temp []*AvailabilityAssignment
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	for i := range temp {
		item := AvailabilityAssignmentsItem(temp[i])
		*a = append(*a, item)
	}

	return nil
}

// AuthQueues
func (a *AuthQueues) UnmarshalJSON(data []byte) error {
	var temp []AuthQueue
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*a = temp
	return nil
}

// ReadyRecord
func (r *ReadyRecord) UnmarshalJSON(data []byte) error {
	var temp struct {
		Report       WorkReport
		Dependencies []WorkPackageHash
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Dependencies) == 0 {
		r.Dependencies = nil
	} else {
		r.Dependencies = temp.Dependencies
	}

	r.Report = temp.Report

	return nil
}

// ReadyQueueItem
func (r *ReadyQueueItem) UnmarshalJSON(data []byte) error {
	var temp []ReadyRecord
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*r = temp

	return nil
}

// ReadyQueue
func (r *ReadyQueue) UnmarshalJSON(data []byte) error {
	var temp []ReadyQueueItem
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*r = temp

	return nil
}

// AccumulatedQueueItem
func (a *AccumulatedQueueItem) UnmarshalJSON(data []byte) error {
	var temp []WorkPackageHash
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*a = temp

	return nil
}

// AccumulatedQueue
func (a *AccumulatedQueue) UnmarshalJSON(data []byte) error {
	var temp []AccumulatedQueueItem
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*a = temp

	return nil
}

// BlocksHistory
func (b *BlocksHistory) UnmarshalJSON(data []byte) error {
	var temp []BlockInfo
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*b = temp

	return nil
}

// Beta
func (b *RecentBlocks) UnmarshalJSON(data []byte) error {
	var temp struct {
		History BlocksHistory `json:"history,omitempty"`
		Mmr     Mmr           `json:"mmr,omitempty"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.History) == 0 {
		b.History = nil
		b.Mmr = Mmr{}
	} else {
		b.History = temp.History
		b.Mmr = temp.Mmr
	}

	*b = temp

	return nil
}

// // AccumulatedHistory
// func (a *AccumulatedHistory) UnmarshalJSON(data []byte) error {
// 	var temp []WorkPackageHash
// 	if err := json.Unmarshal(data, &temp); err != nil {
// 		return err
// 	}

// 	if len(temp) == 0 {
// 		return nil
// 	}

// 	*a = temp

// 	return nil
// }

// // AccumulatedHistories
// func (a *AccumulatedHistories) UnmarshalJSON(data []byte) error {
// 	var temp []AccumulatedHistory
// 	if err := json.Unmarshal(data, &temp); err != nil {
// 		return err
// 	}

// 	if len(temp) == 0 {
// 		return nil
// 	}

// 	*a = temp

// 	return nil
// }

// ServiceAccountState
func (a *ServiceAccountState) UnmarshalJSON(data []byte) error {
	var temp []AccountDTO
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	// Init ServiceAccountState map
	*a = make(ServiceAccountState)

	// Assign the temp into the accounts map
	for _, account := range temp {
		var serviceAccount ServiceAccount

		var serviceInfo ServiceInfo
		serviceInfo.CodeHash = account.Data.Service.CodeHash
		serviceInfo.Balance = account.Data.Service.Balance
		serviceInfo.MinItemGas = account.Data.Service.MinItemGas
		serviceInfo.MinMemoGas = account.Data.Service.MinMemoGas
		serviceInfo.Bytes = account.Data.Service.Bytes
		serviceInfo.Items = account.Data.Service.Items

		serviceAccount.ServiceInfo = serviceInfo

		if len(account.Data.Preimages) != 0 {
			serviceAccount.PreimageLookup = make(PreimagesMapEntry)

			for _, preimage := range account.Data.Preimages {
				serviceAccount.PreimageLookup[preimage.Hash] = preimage.Blob
			}
		} else {
			serviceAccount.PreimageLookup = nil
		}

		if len(account.Data.LookupMeta) != 0 {
			serviceAccount.LookupDict = make(map[LookupMetaMapkey]TimeSlotSet)

			for _, lookup := range account.Data.LookupMeta {
				key := LookupMetaMapkey(lookup.Key)
				serviceAccount.LookupDict[key] = lookup.Val
			}
		} else {
			serviceAccount.LookupDict = nil
		}

		serviceAccount.StorageDict = account.Data.Storage

		(*a)[account.Id] = serviceAccount
	}

	return nil
}

// Account
func (a *AccountDTO) UnmarshalJSON(data []byte) error {
	var temp struct {
		Id   U32            `json:"id"`
		Data AccountDataDTO `json:"data"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Id = ServiceId(temp.Id)
	a.Data = temp.Data

	return nil
}

// AccountDataDTO
func (a *AccountDataDTO) UnmarshalJSON(data []byte) error {
	var temp struct {
		Service    ServiceInfo             `json:"service"`
		Preimages  []PreimagesMapEntryDTO  `json:"preimages"`
		LookupMeta []LookupMetaMapEntryDTO `json:"lookup_meta"`
		Storage    Storage                 `json:"storage"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	a.Service = temp.Service

	if len(temp.Preimages) == 0 {
		a.Preimages = nil
	} else {
		a.Preimages = temp.Preimages
	}

	if len(temp.LookupMeta) == 0 {
		a.LookupMeta = nil
	} else {
		a.LookupMeta = temp.LookupMeta
	}

	a.Storage = temp.Storage

	return nil
}

// Storage
func (s *Storage) UnmarshalJSON(data []byte) error {
	var temp map[string]string
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		return nil
	}

	*s = make(Storage)
	for key, value := range temp {
		keyBytes, err := hex.DecodeString(key[2:])
		if err != nil {
			return err
		}

		valueBytes, err := hex.DecodeString(value[2:])
		if err != nil {
			return err
		}

		if len(valueBytes) == 0 {
			(*s)[string(keyBytes)] = nil
		} else {
			(*s)[string(keyBytes)] = ByteSequence(valueBytes)
		}
	}

	return nil
}

// LookupMetaMapEntryDTO
func (l *LookupMetaMapEntryDTO) UnmarshalJSON(data []byte) error {
	var temp struct {
		Key LookupMetaMapkey `json:"key"`
		Val []TimeSlot       `json:"value"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	l.Key = temp.Key

	if len(temp.Val) == 0 {
		l.Val = nil
	} else {
		l.Val = temp.Val
	}

	return nil
}

// LookupMetaMapkey
func (l *LookupMetaMapkey) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash   OpaqueHash `json:"hash"`
		Length U32        `json:"length"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	l.Hash = temp.Hash
	l.Length = temp.Length

	return nil
}

// PreimagesMapEntryDTO
func (p *PreimagesMapEntryDTO) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash OpaqueHash `json:"hash"`
		Blob string     `json:"blob"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	p.Hash = temp.Hash

	blobBytes, err := hex.DecodeString(temp.Blob[2:])
	if err != nil {
		return err
	}
	p.Blob = blobBytes

	return nil
}

// Priviliges
func (p *Privileges) UnmarshalJSON(data []byte) error {
	var temp struct {
		Bless       U32                      `json:"bless"`      // Manager
		Assign      []U32                    `json:"assign"`     // AlterPhi
		Designate   U32                      `json:"designate"`  // AlterIota
		AlwaysAccum []AlwaysAccumulateMapDTO `json:"always_acc"` // AutoAccumulateGasLimits
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	p.Bless = ServiceId(temp.Bless)
	p.Assign = make(ServiceIdList, len(temp.Assign))
	for i, id := range temp.Assign {
		p.Assign[i] = ServiceId(id)
	}
	p.Designate = ServiceId(temp.Designate)

	if len(temp.AlwaysAccum) == 0 {
		p.AlwaysAccum = make(AlwaysAccumulateMap)
	}

	for _, entry := range temp.AlwaysAccum {
		p.AlwaysAccum[entry.ServiceId] = entry.Gas
	}

	return nil
}

// ServicesStatistics
func (s *ServicesStatistics) UnmarshalJSON(data []byte) error {
	var temp []struct {
		Id     U32                   `json:"id"`
		Record ServiceActivityRecord `json:"record"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp) == 0 {
		*s = nil
		return nil
	}

	*s = make(ServicesStatistics, len(temp))
	for _, item := range temp {
		(*s)[ServiceId(item.Id)] = item.Record
	}

	return nil
}

func (s *Statistics) UnmarshalJSON(data []byte) error {
	var temp struct {
		ValsCurr ValidatorsStatistics `json:"vals_curr,omitempty"`
		ValsLast ValidatorsStatistics `json:"vals_last,omitempty"`
		Cores    CoresStatistics      `json:"cores,omitempty"`
		Services ServicesStatistics   `json:"services,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	s.ValsCurr = temp.ValsCurr
	s.ValsLast = temp.ValsLast
	s.Cores = temp.Cores

	if len(temp.Services) == 0 {
		s.Services = nil
	} else {
		s.Services = temp.Services
	}

	return nil
}

func (s *StateKey) UnmarshalJSON(data []byte) error {
	decoded, err := hex.DecodeString(string(data[2 : len(data)-1]))
	if err != nil {
		return fmt.Errorf("failed to decode hex string: %w", err)
	}

	if len(decoded) != len(StateKey{}) {
		return fmt.Errorf("decoded length %d does not match expected length %d", len(decoded), len(StateKey{}))
	}

	copy(s[:], decoded)

	return nil
}

// TraceState UnmarshalJSON unmarshals a JSON-encoded StateKeyVals.
func (s *StateKeyVals) UnmarshalJSON(data []byte) error {
	var raw []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = make(StateKeyVals, len(raw))
	for i, kv := range raw {
		// key
		decodedKey, err := hex.DecodeString(kv.Key[2:])
		if err != nil {
			return fmt.Errorf("failed to decode hex string: %w", err)
		}

		(*s)[i].Key = StateKey(decodedKey)

		// value
		decodedValue, err := hex.DecodeString(kv.Value[2:])
		if err != nil {
			return fmt.Errorf("failed to decode hex string: %w", err)
		}

		if len(decodedValue) == 0 {
			(*s)[i].Value = nil
		} else {
			(*s)[i].Value = decodedValue
		}
	}

	return nil
}
