package utilities

import (
	"fmt"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func SerializeByteSequence(input []byte) (output jamTypes.ByteSequence) {
	return WrapByteSequence(jamTypes.ByteSequence(input[:])).Serialize()
}

func SerializeOpaqueHash(input jamTypes.OpaqueHash) (output jamTypes.ByteSequence) {
	return WrapOpaqueHash(input).Serialize()
}

func BlockSerialization(block jamTypes.Block) (output jamTypes.ByteSequence) {
	/*
		(C.13) E (H, ET (ET ), EP (EP ), EG(EG), EA(EA), ED(ED))
	*/
	output = append(output, HeaderSerialization(block.Header)...)
	output = append(output, ExtrinsicTicketSerialization(block.Extrinsic.Tickets)...)
	output = append(output, ExtrinsicPreimageSerialization(block.Extrinsic.Preimages)...)
	output = append(output, ExtrinsicGuaranteeSerialization(block.Extrinsic.Guarantees)...)
	output = append(output, ExtrinsicAssuranceSerialization(block.Extrinsic.Assurances)...)
	output = append(output, ExtrinsicDisputeSerialization(block.Extrinsic.Disputes)...)
	return output
}

func ExtrinsicTicketSerialization(tickets jamTypes.TicketsExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.14) E(↕ET)
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(tickets)))...)
	for _, ticket := range tickets {
		// ticket.Attempt
		output = append(output, SerializeU64(jamTypes.U64(ticket.Attempt))...)
		// ticket.Signature
		output = append(output, SerializeByteSequence(ticket.Signature[:])...)
	}
	return output
}

func ExtrinsicPreimageSerialization(preimages jamTypes.PreimagesExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.15) E(↕[s, ↕p])

		Requester ServiceId    `json:"requester,omitempty"`
		Blob      ByteSequence `json:"blob,omitempty"`
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(preimages)))...)
	for _, preimage := range preimages {
		// Preimage.Requester
		output = append(output, SerializeU64(jamTypes.U64(preimage.Requester))...)
		// Preimagt.Blob
		output = append(output, SerializeU64(jamTypes.U64(len(preimage.Blob)))...)
		output = append(output, SerializeByteSequence(preimage.Blob[:])...)
	}
	return output
}

// INFO: This is different between Appendix C (C.16) and (5.4), (5.5), (5.6).
func ExtrinsicGuaranteeSerialization(guarantees jamTypes.GuaranteesExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.16) E(↕[w, E4(t), ↕a])

		Report     WorkReport           `json:"report"`
		Slot       TimeSlot             `json:"slot,omitempty"`
		Signatures []ValidatorSignature `json:"signatures,omitempty"`
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(guarantees)))...)
	for _, guarantee := range guarantees {
		// WorkReport
		output = append(output, WorkReportSerialization(guarantee.Report)...)
		// Slot
		output = append(output, SerializeFixedLength(jamTypes.U32(guarantee.Slot), 4)...)
		// Signature
		output = append(output, SerializeU64(jamTypes.U64(len(guarantee.Signatures)))...)
		for _, signature := range guarantee.Signatures {
			output = append(output, SerializeU64(jamTypes.U64(signature.ValidatorIndex))...)
			output = append(output, SerializeByteSequence(signature.Signature[:])...)
		}
	}
	return output
}

func ExtrinsicAssuranceSerialization(assurances jamTypes.AssurancesExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.17) ↕[a, f, E2(v), s]

		Anchor         OpaqueHash       `json:"anchor,omitempty"`
		Bitfield       []byte           `json:"bitfield,omitempty"`
		ValidatorIndex ValidatorIndex   `json:"validator_index,omitempty"`
		Signature      Ed25519Signature `json:"signature,omitempty"`
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(assurances)))...)
	for _, assurance := range assurances {
		output = append(output, SerializeByteSequence(assurance.Anchor[:])...)
		output = append(output, SerializeByteSequence(assurance.Bitfield[:])...)
		output = append(output, SerializeFixedLength(jamTypes.U32(assurance.ValidatorIndex), 2)...)
		output = append(output, SerializeByteSequence(assurance.Signature[:])...)
	}
	return output
}

func ExtrinsicDisputeSerialization(disputes jamTypes.DisputesExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.18) E(↕[(r, E4(a), [(v, E2(i), s)]] , ↕c, ↕f)

		Verdicts []Verdict `json:"verdicts,omitempty"`
		Culprits []Culprit `json:"culprits,omitempty"`
		Faults   []Fault   `json:"faults,omitempty"`
	*/
	// verdict
	output = append(output, SerializeU64(jamTypes.U64(len(disputes.Verdicts)))...)
	for _, verdict := range disputes.Verdicts {
		output = append(output, SerializeByteSequence(verdict.Target[:])...)
		output = append(output, SerializeFixedLength(jamTypes.U32(verdict.Age), 4)...)
		for _, judgement := range verdict.Votes {
			if !judgement.Vote {
				output = append(output, SerializeU64(jamTypes.U64(0))...)
			} else {
				output = append(output, SerializeU64(jamTypes.U64(1))...)
			}
			output = append(output, SerializeFixedLength(jamTypes.U32(judgement.Index), 2)...)
			output = append(output, SerializeByteSequence(judgement.Signature[:])...)
		}
	}
	// culprit
	output = append(output, SerializeU64(jamTypes.U64(len(disputes.Culprits)))...)
	for _, culprit := range disputes.Culprits {
		output = append(output, SerializeByteSequence(culprit.Target[:])...)
		output = append(output, SerializeByteSequence(culprit.Key[:])...)
		output = append(output, SerializeByteSequence(culprit.Signature[:])...)
	}
	// fault
	output = append(output, SerializeU64(jamTypes.U64(len(disputes.Faults)))...)
	for _, fault := range disputes.Faults {
		output = append(output, SerializeByteSequence(fault.Target[:])...)
		if !fault.Vote {
			output = append(output, SerializeU64(jamTypes.U64(0))...)
		} else {
			output = append(output, SerializeU64(jamTypes.U64(1))...)
		}
		output = append(output, SerializeByteSequence(fault.Key[:])...)
		output = append(output, SerializeByteSequence(fault.Signature[:])...)
	}
	return output
}

func HeaderSerialization(header jamTypes.Header) (output jamTypes.ByteSequence) {
	// (C.19) E(H) = EU (H) ⌢ E(Hs)
	output = append(output, HeaderUSerialization(header)...)
	output = append(output, SerializeByteSequence(header.Seal[:])...)
	return output
}

func HeaderUSerialization(header jamTypes.Header) (output jamTypes.ByteSequence) {
	// (C.20) EU (H) = E(Hp,Hr,Hx) ⌢ E4(Ht) ⌢ E(¿He, ¿Hw, ↕Ho, E2(Hi),Hv)
	// header.Parent, header.ParentStateRoot, header.ExtrinsicHash
	output = append(output, SerializeByteSequence(header.Parent[:])...)
	output = append(output, SerializeByteSequence(header.ParentStateRoot[:])...)
	output = append(output, SerializeOpaqueHash(header.ExtrinsicHash)...)
	// header.Slot
	output = append(output, SerializeFixedLength(jamTypes.U32(header.Slot), 4)...)
	// ?He header.EpochMark
	{
		num, _ := EmptyOrPair(header.EpochMark)
		output = append(output, SerializeU64(jamTypes.U64(num))...)
		if header.EpochMark != nil {
			output = append(output, SerializeByteSequence(header.EpochMark.Entropy[:])...)
			output = append(output, SerializeByteSequence(header.EpochMark.TicketsEntropy[:])...)
			for _, validator := range header.EpochMark.Validators {
				output = append(output, SerializeByteSequence(validator[:])...)
			}
		}
	}
	// ?Hw header.TicketsMark
	{
		num, _ := EmptyOrPair(header.TicketsMark)
		output = append(output, SerializeU64(jamTypes.U64(num))...)
		if header.TicketsMark != nil {
			for _, ticket := range *header.TicketsMark {
				output = append(output, SerializeByteSequence(ticket.Id[:])...)
				output = append(output, SerializeU64(jamTypes.U64(ticket.Attempt))...)
			}
		}
	}
	// ↕Ho header.OffendersMark
	{
		output = append(output, SerializeU64(jamTypes.U64(len(header.OffendersMark)))...)
		for _, offender := range header.OffendersMark {
			output = append(output, SerializeByteSequence(offender[:])...)
		}
	}
	// E2(Hi) header.AuthorIndex
	{
		output = append(output, SerializeFixedLength(jamTypes.U32(header.AuthorIndex), 2)...)
	}
	// Hv header.EntropySource
	{
		output = append(output, SerializeByteSequence(header.EntropySource[:])...)
	}
	return output
}

func RefineContextSerialization(refine_context jamTypes.RefineContext) (output jamTypes.ByteSequence) {
	/*
		(C.21) E(x ∈ X) ≡ E(xa, xs, xb, xl) ⌢ E4(xt) ⌢ E(↕xp)

		Anchor           HeaderHash   `json:"anchor,omitempty"`
		StateRoot        StateRoot    `json:"state_root,omitempty"`
		BeefyRoot        BeefyRoot    `json:"beefy_root,omitempty"`
		LookupAnchor     HeaderHash   `json:"lookup_anchor,omitempty"`
		LookupAnchorSlot TimeSlot     `json:"lookup_anchor_slot,omitempty"`
		Prerequisites    []OpaqueHash `json:"prerequisites,omitempty"`
	*/
	output = append(output, SerializeByteSequence(refine_context.Anchor[:])...)
	output = append(output, SerializeByteSequence(refine_context.StateRoot[:])...)
	output = append(output, SerializeByteSequence(refine_context.BeefyRoot[:])...)
	output = append(output, SerializeByteSequence(refine_context.LookupAnchor[:])...)
	output = append(output, SerializeFixedLength(jamTypes.U32(refine_context.LookupAnchorSlot), 4)...)
	output = append(output, SerializeU64(jamTypes.U64(len(refine_context.Prerequisites)))...)
	for _, prerequest := range refine_context.Prerequisites {
		output = append(output, SerializeByteSequence(prerequest[:])...)
	}
	return output
}

func WorkPackageSpecSerialization(work_package_spec jamTypes.WorkPackageSpec) (output jamTypes.ByteSequence) {
	/*
		(C.22) E(x ∈ S) ≡ E(xh) ⌢ E4(xl) ⌢ E(xu, xe) ⌢ E2(xn)

		Hash         WorkPackageHash `json:"hash,omitempty"`
		Length       U32             `json:"length,omitempty"`
		ErasureRoot  ErasureRoot     `json:"erasure_root,omitempty"`
		ExportsRoot  ExportsRoot     `json:"exports_root,omitempty"`
		ExportsCount U16             `json:"exports_count,omitempty"`
	*/
	output = append(output, SerializeByteSequence(work_package_spec.Hash[:])...)
	output = append(output, SerializeFixedLength(work_package_spec.Length, 4)...)
	output = append(output, SerializeByteSequence(work_package_spec.ErasureRoot[:])...)
	output = append(output, SerializeByteSequence(work_package_spec.ExportsRoot[:])...)
	output = append(output, SerializeFixedLength(jamTypes.U32(work_package_spec.ExportsCount), 2)...)
	return output
}

func WorkResultSerialization(result jamTypes.WorkResult) (output jamTypes.ByteSequence) {
	/*
		(C.23) E(x ∈ L) ≡ E4(xs) ⌢ E(xc, xl) ⌢ E8(xg) ⌢ E(O(xo))

		ServiceId     ServiceId      `json:"service_id,omitempty"`
		CodeHash      OpaqueHash     `json:"code_hash,omitempty"`
		PayloadHash   OpaqueHash     `json:"payload_hash,omitempty"`
		AccumulateGas Gas            `json:"accumulate_gas,omitempty"`
		Result        WorkExecResult `json:"result,omitempty"`
	*/
	output = append(output, SerializeFixedLength(jamTypes.U64(result.ServiceId), 4)...)
	output = append(output, SerializeOpaqueHash(result.CodeHash)...)
	output = append(output, SerializeOpaqueHash(result.PayloadHash)...)
	output = append(output, SerializeFixedLength(jamTypes.U64(result.AccumulateGas), 8)...)
	output = append(output, SerializeWorkExecResult(result.Result)...)
	return output
}

func WorkReportSerialization(work_report jamTypes.WorkReport) (output jamTypes.ByteSequence) {
	/*
		(C.24) E(x ∈ W) ≡ E(xs, xx, xc, xa, ↕xo, ↕xl, ↕xr)

		PackageSpec       WorkPackageSpec   `json:"package_spec"`
		Context           RefineContext     `json:"context"`
		CoreIndex         CoreIndex         `json:"core_index,omitempty"`
		AuthorizerHash    OpaqueHash        `json:"authorizer_hash,omitempty"`
		AuthOutput        ByteSequence      `json:"auth_output,omitempty"`
		SegmentRootLookup SegmentRootLookup `json:"segment_root_lookup,omitempty"`
		Results           []WorkResult      `json:"results,omitempty"`
	*/
	output = append(output, WorkPackageSpecSerialization(work_report.PackageSpec)...) // xs
	output = append(output, RefineContextSerialization(work_report.Context)...)       // xx
	output = append(output, SerializeU64(jamTypes.U64(work_report.CoreIndex))...)     // xc
	output = append(output, SerializeByteSequence(work_report.AuthorizerHash[:])...)  // xa
	// xo
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.AuthOutput)))...)
	output = append(output, SerializeByteSequence(work_report.AuthOutput[:])...)
	// xl
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.SegmentRootLookup)))...)
	for _, item := range work_report.SegmentRootLookup {
		output = append(output, SerializeByteSequence(item.WorkPackageHash[:])...)
		output = append(output, SerializeByteSequence(item.SegmentTreeRoot[:])...)
	}
	// xr
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.Results)))...)
	for _, result := range work_report.Results {
		output = append(output, WorkResultSerialization(result)...)
	}
	return output
}

func SerializeWorkPackage(work_package jamTypes.WorkPackage) (output jamTypes.ByteSequence) {
	/*
		(C.25) E(x ∈ P) ≡ E(↕xj, E4(xh), xu, ↕xp, xx, ↕xw)
		type WorkPackage struct {
			Authorization ByteSequence  `json:"authorization,omitempty"`
			AuthCodeHost  ServiceId     `json:"auth_code_host,omitempty"`
			Authorizer    Authorizer    `json:"authorizer"`
			Context       RefineContext `json:"context"`
			Items         []WorkItem    `json:"items,omitempty"`
		}
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(work_package.Authorization)))...)
	output = append(output, SerializeByteSequence(work_package.Authorization)...)
	output = append(output, SerializeFixedLength(jamTypes.U64(work_package.AuthCodeHost), 4)...)
	output = append(output, SerializeOpaqueHash(work_package.Authorizer.CodeHash)...)
	output = append(output, SerializeU64(jamTypes.U64(len(work_package.Authorizer.Params)))...)
	output = append(output, SerializeByteSequence(work_package.Authorizer.Params)...)
	output = append(output, RefineContextSerialization(work_package.Context)...)
	output = append(output, SerializeU64(jamTypes.U64(len(work_package.Items)))...)
	for _, work_item := range work_package.Items {
		output = append(output, WorkItemSerialization(work_item)...)
	}
	return output
}

func WorkItemSerialization(work_item jamTypes.WorkItem) (output jamTypes.ByteSequence) {
	/*
		(C.26) E(x ∈ I) ≡ E(E4(xs), xc, ↕xy, E8(xg), ↕E#I(xi), ↕[(h, E4(i)) ∣ (h, i) <− xx], E2(xe))
		type WorkItem struct {
			Service            ServiceId       `json:"service,omitempty"`
			CodeHash           OpaqueHash      `json:"code_hash,omitempty"`
			Payload            ByteSequence    `json:"payload,omitempty"`
			RefineGasLimit     Gas             `json:"refine_gas_limit,omitempty"`
			AccumulateGasLimit Gas             `json:"accumulate_gas_limit,omitempty"`
			ImportSegments     []ImportSpec    `json:"import_segments,omitempty"`
			Extrinsic          []ExtrinsicSpec `json:"extrinsic,omitempty"`
			ExportCount        U16             `json:"export_count,omitempty"`
		}
	*/
	// ↕[(h, E4(i)) ∣ (h, i) <− xx] TODO: \se_4(i) should be just \se(i), but we should wait for this to be written.". check with haha
	output = append(output, SerializeFixedLength(jamTypes.U64(work_item.Service), 4)...)
	output = append(output, SerializeOpaqueHash(work_item.CodeHash)...)
	output = append(output, SerializeU64(jamTypes.U64(len(work_item.Payload)))...)
	output = append(output, SerializeByteSequence(work_item.Payload)...)

	output = append(output, SerializeFixedLength(jamTypes.U64(work_item.RefineGasLimit), 8)...)
	output = append(output, SerializeFixedLength(jamTypes.U64(work_item.AccumulateGasLimit), 8)...)

	output = append(output, SerializeU64(jamTypes.U64(len(work_item.ImportSegments)))...)
	for _, import_spec := range work_item.ImportSegments {
		output = append(output, SerializeImportSpec(import_spec)...)
	}

	output = append(output, SerializeU64(jamTypes.U64(len(work_item.Extrinsic)))...)
	for _, extrinsic_spec := range work_item.Extrinsic {
		output = append(output, SerializeOpaqueHash(extrinsic_spec.Hash)...)
		output = append(output, SerializeFixedLength(jamTypes.U64(extrinsic_spec.Len), 4)...)
	}
	output = append(output, SerializeFixedLength(jamTypes.U64(work_item.ExportCount), 2)...)

	return output
}

func TicketBodySerialization(ticket_body jamTypes.TicketBody) (output jamTypes.ByteSequence) {
	/*
		(C.27) E(x ∈ C) ≡ E(xy, xr)

		type TicketBody struct {
			Id      TicketId      `json:"id,omitempty"`
			Attempt TicketAttempt `json:"attempt,omitempty"`
		}
	*/
	output = append(output, SerializeByteSequence(ticket_body.Id[:])...)
	output = append(output, SerializeU64(jamTypes.U64(ticket_body.Attempt))...)
	return output
}

func SerializeWorkExecResult(result jamTypes.WorkExecResult) (output jamTypes.ByteSequence) {
	// (C.28)
	/*
			const (
			WorkExecResultOk           WorkExecResultType = "ok"
			WorkExecResultOutOfGas                        = "out-of-gas"
			WorkExecResultPanic                           = "panic"
			WorkExecResultBadExports                      = "bad-exports"
			WorkExecResultBadCode                         = "bad-code"
			WorkExecResultCodeOversize                    = "code-oversize"
		)

	*/
	if len(result) == 1 {
		for key, value := range result {
			if key == "ok" {
				output = append(output, SerializeU64(jamTypes.U64(0))...)
				output = append(output, SerializeByteSequence(value)...) // (0, ↕o)
			} else if key == "out-of-gas" {
				output = append(output, SerializeU64(jamTypes.U64(1))...)
			} else if key == "panic" {
				output = append(output, SerializeU64(jamTypes.U64(2))...)
			} else if key == "bad-exports" {
				output = append(output, SerializeU64(jamTypes.U64(3))...)
			} else if key == "bad-code" {
				output = append(output, SerializeU64(jamTypes.U64(4))...)
			} else if key == "code-oversize" {
				output = append(output, SerializeU64(jamTypes.U64(5))...)
			}
		}
	} else {
		fmt.Println("Map size expected to be 1")
	}
	return output
}

func SerializeImportSpec(import_spec jamTypes.ImportSpec) (output jamTypes.ByteSequence) {
	// (C.29) case 1 (h, E2(i)) if h ∈ H
	// TODO check case2 use case
	h, i := import_spec.TreeRoot, import_spec.Index
	output = append(output, SerializeOpaqueHash(h)...)
	output = append(output, SerializeFixedLength(jamTypes.U64(i), 2)...)
	return output
}
