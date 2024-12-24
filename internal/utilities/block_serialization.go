package utilities

import (
	"fmt"

	jamTypes "github.com/New-JAMneration/JAM-Protocol/internal/jam_types"
)

func SerializeByteArray(input []byte) (output jamTypes.ByteSequence) {
	return WrapByteSequence(jamTypes.ByteSequence(input[:])).Serialize()
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
		output = append(output, SerializeByteArray(ticket.Signature[:])...)
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
		output = append(output, SerializeByteArray(preimage.Blob[:])...)
	}
	return output
}

func ExtrinsicGuaranteeSerialization(guarantees jamTypes.GuaranteesExtrinsic) (output jamTypes.ByteSequence) {
	/*
		(C.16) E(↕[w, E4(t), ↕a])

		Report     WorkReport           `json:"report"`
		Slot       TimeSlot             `json:"slot,omitempty"`
		Signatures []ValidatorSignature `json:"signatures,omitempty"`
	*/
	output = append(output, SerializeU64(jamTypes.U64(len(guarantees.Guarantees)))...)
	for _, guarantee := range guarantees.Guarantees {
		// WorkReport
		output = append(output, WorkReportSerialization(guarantee.Report)...)
		// Slot
		output = append(output, SerializeFixedLength(jamTypes.U32(guarantee.Slot), 4)...)
		// Signature
		output = append(output, SerializeU64(jamTypes.U64(len(guarantee.Signatures)))...)
		for _, signature := range guarantee.Signatures {
			output = append(output, SerializeU64(jamTypes.U64(signature.ValidatorIndex))...)
			output = append(output, SerializeByteArray(signature.Signature[:])...)
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
		output = append(output, SerializeByteArray(assurance.Anchor[:])...)
		output = append(output, SerializeByteArray(assurance.Bitfield[:])...)
		output = append(output, SerializeFixedLength(jamTypes.U32(assurance.ValidatorIndex), 2)...)
		output = append(output, SerializeByteArray(assurance.Signature[:])...)
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
		output = append(output, SerializeByteArray(verdict.Target[:])...)
		output = append(output, SerializeFixedLength(jamTypes.U32(verdict.Age), 4)...)
		for _, judgement := range verdict.Votes {
			if !judgement.Vote {
				output = append(output, SerializeU64(jamTypes.U64(0))...)
			} else {
				output = append(output, SerializeU64(jamTypes.U64(1))...)
			}
			output = append(output, SerializeFixedLength(jamTypes.U32(judgement.Index), 2)...)
			output = append(output, SerializeByteArray(judgement.Signature[:])...)
		}
	}
	// culprit
	output = append(output, SerializeU64(jamTypes.U64(len(disputes.Culprits)))...)
	for _, culprit := range disputes.Culprits {
		output = append(output, SerializeByteArray(culprit.Target[:])...)
		output = append(output, SerializeByteArray(culprit.Key[:])...)
		output = append(output, SerializeByteArray(culprit.Signature[:])...)
	}
	// fault
	output = append(output, SerializeU64(jamTypes.U64(len(disputes.Faults)))...)
	for _, fault := range disputes.Faults {
		output = append(output, SerializeByteArray(fault.Target[:])...)
		if !fault.Vote {
			output = append(output, SerializeU64(jamTypes.U64(0))...)
		} else {
			output = append(output, SerializeU64(jamTypes.U64(1))...)
		}
		output = append(output, SerializeByteArray(fault.Key[:])...)
		output = append(output, SerializeByteArray(fault.Signature[:])...)
	}
	return output
}

func HeaderSerialization(header jamTypes.Header) (output jamTypes.ByteSequence) {
	// (C.19) E(H) = EU (H) ⌢ E(Hs)
	output = append(output, HeaderUSerialization(header)...)
	output = append(output, SerializeByteArray(header.Seal[:])...)
	return output
}

func HeaderUSerialization(header jamTypes.Header) (output jamTypes.ByteSequence) {
	// (C.20) EU (H) = E(Hp,Hr,Hx) ⌢ E4(Ht) ⌢ E(¿He, ¿Hw, ↕Ho, E2(Hi),Hv)
	// header.Parent, header.ParentStateRoot, header.ExtrinsicHash
	output = append(output, SerializeByteArray(header.Parent[:])...)
	output = append(output, SerializeByteArray(header.ParentStateRoot[:])...)
	output = append(output, SerializeByteArray(header.ExtrinsicHash[:])...)
	// header.Slot
	output = append(output, SerializeFixedLength(jamTypes.U32(header.Slot), 4)...)
	// ?He header.EpochMark
	{
		num, _ := EmptyOrPair(header.EpochMark)
		output = append(output, SerializeU64(jamTypes.U64(num))...)
		if header.EpochMark != nil {
			output = append(output, SerializeByteArray(header.EpochMark.Entropy[:])...)
			output = append(output, SerializeByteArray(header.EpochMark.TicketsEntropy[:])...)
			for _, validator := range header.EpochMark.Validators {
				output = append(output, SerializeByteArray(validator[:])...)
			}
		}
	}
	// ?Hw header.TicketsMark
	{
		num, _ := EmptyOrPair(header.TicketsMark)
		output = append(output, SerializeU64(jamTypes.U64(num))...)
		if header.TicketsMark != nil {
			for _, ticket := range *header.TicketsMark {
				output = append(output, SerializeByteArray(ticket.Id[:])...)
				output = append(output, SerializeU64(jamTypes.U64(ticket.Attempt))...)
			}
		}
	}
	// ↕Ho header.OffendersMark
	{
		output = append(output, SerializeU64(jamTypes.U64(len(header.OffendersMark)))...)
		for _, offender := range header.OffendersMark {
			output = append(output, SerializeByteArray(offender[:])...)
		}
	}
	// E2(Hi) header.AuthorIndex
	{
		output = append(output, SerializeFixedLength(jamTypes.U32(header.AuthorIndex), 2)...)
	}
	// Hv header.EntropySource
	{
		output = append(output, SerializeByteArray(header.EntropySource[:])...)
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
	output = append(output, SerializeByteArray(refine_context.Anchor[:])...)
	output = append(output, SerializeByteArray(refine_context.StateRoot[:])...)
	output = append(output, SerializeByteArray(refine_context.BeefyRoot[:])...)
	output = append(output, SerializeByteArray(refine_context.LookupAnchor[:])...)
	output = append(output, SerializeFixedLength(jamTypes.U32(refine_context.LookupAnchorSlot), 4)...)
	output = append(output, SerializeU64(jamTypes.U64(len(refine_context.Prerequisites)))...)
	for _, prerequest := range refine_context.Prerequisites {
		output = append(output, SerializeByteArray(prerequest[:])...)
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
	output = append(output, SerializeByteArray(work_package_spec.Hash[:])...)
	output = append(output, SerializeFixedLength(work_package_spec.Length, 4)...)
	output = append(output, SerializeByteArray(work_package_spec.ErasureRoot[:])...)
	output = append(output, SerializeByteArray(work_package_spec.ExportsRoot[:])...)
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
	output = append(output, SerializeByteArray(result.CodeHash[:])...)
	output = append(output, SerializeByteArray(result.PayloadHash[:])...)
	output = append(output, SerializeFixedLength(jamTypes.U64(result.AccumulateGas), 8)...)
	output = append(output, SerializeU64(jamTypes.U64(result.AccumulateGas))...)
	output = append(output, SerializeWorkExecResult(result.Result)...)
	// TODO map wrapper usage
	// m := MapWarpper{Value: result.Result}
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
	output = append(output, SerializeByteArray(work_report.AuthorizerHash[:])...)     // xa
	// xo
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.AuthOutput)))...)
	output = append(output, SerializeByteArray(work_report.AuthOutput[:])...)
	// xl
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.SegmentRootLookup)))...)
	for _, item := range work_report.SegmentRootLookup {
		output = append(output, SerializeByteArray(item.WorkPackageHash[:])...)
		output = append(output, SerializeByteArray(item.SegmentTreeRoot[:])...)
	}
	// xr
	output = append(output, SerializeU64(jamTypes.U64(len(work_report.Results)))...)
	for _, result := range work_report.Results {
		output = append(output, WorkResultSerialization(result)...)
	}
	return output
}

func SerializeWorkExecResult(result jamTypes.WorkExecResult) (output jamTypes.ByteSequence) {
	// (C.28)
	if len(result) == 1 {
		for key, value := range result {
			// TODO gray paper type update
			if key == "ok" {
				output = append(output, SerializeU64(jamTypes.U64(0))...)
				output = append(output, SerializeByteArray(value)...) // (0, ↕o)
			} else if key == "out-of-gas" {
				output = append(output, SerializeU64(jamTypes.U64(1))...)
			} else if key == "panic" {
				output = append(output, SerializeU64(jamTypes.U64(2))...)
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
