package types

import (
	"fmt"
)

// U16
func (u *U16) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding U16")
	encoded, err := e.EncodeUintWithLength(uint64(*u), 2)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("U16: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// U32
func (u *U32) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding U32")
	encoded, err := e.EncodeUintWithLength(uint64(*u), 4)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("U32: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// HeaderHash
func (h *HeaderHash) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding HeaderHash")
	if _, err := e.buf.Write(h[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("HeaderHash: %v", h[:]))

	return nil
}

// StateRoot
func (s *StateRoot) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding StateRoot")
	if _, err := e.buf.Write(s[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("StateRoot: %v", s[:]))

	return nil
}

// OpaqueHash
func (o *OpaqueHash) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding OpaqueHash")
	if _, err := e.buf.Write(o[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("OpaqueHash: %v", o[:]))

	return nil
}

// TimeSlot
func (t *TimeSlot) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TimeSlot")
	encoded, err := e.EncodeUintWithLength(uint64(*t), 4)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("TimeSlot: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// Entropy
func (e *Entropy) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding Entropy")
	if _, err := enc.buf.Write(e[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Entropy: %v", e[:]))

	return nil
}

// BandersnatchPublic
func (bp *BandersnatchPublic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BandersnatchPublic")
	if _, err := e.buf.Write(bp[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BandersnatchPublic: %v", bp[:]))

	return nil
}

// EpochMark
func (em *EpochMark) Encode(e *Encoder) error {
	// Entropy
	if err := em.Entropy.Encode(e); err != nil {
		return err
	}

	// TicketsEntropy
	if err := em.TicketsEntropy.Encode(e); err != nil {
		return err
	}

	// Validators
	if len(em.Validators) != int(ValidatorsCount) {
		return fmt.Errorf("Validators length is not equal to ValidatorCount")
	}

	for _, validator := range em.Validators {
		if err := validator.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// TicketId
func (t *TicketId) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketId")
	if _, err := e.buf.Write(t[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("TicketId: %v", t[:]))

	return nil
}

// TicketAttempt
func (t *TicketAttempt) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketAttempt")
	bytes := []byte{byte(*t)}
	if _, err := e.buf.Write(bytes); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("TicketAttempt: %v", bytes))

	return nil
}

// TicketBody
func (tb *TicketBody) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketBody")
	if err := tb.Id.Encode(e); err != nil {
		return err
	}

	if err := tb.Attempt.Encode(e); err != nil {
		return err
	}

	return nil
}

// TicketsMark
func (tm *TicketsMark) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketsMark")

	if len(*tm) != int(EpochLength) {
		return fmt.Errorf("TicketsMark length is not equal to EpochLength")
	}

	for _, ticketBody := range *tm {
		if err := ticketBody.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Ed25519Public
func (ep *Ed25519Public) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Ed25519Public")
	if _, err := e.buf.Write(ep[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Ed25519Public: %v", ep[:]))

	return nil
}

// OffendersMark
func (om *OffendersMark) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding OffendersMark")

	err := e.EncodeLength(uint64(len(*om)))
	if err != nil {
		return err
	}

	for _, offender := range *om {
		if err := offender.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ValidatorIndex
func (v *ValidatorIndex) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ValidatorIndex")
	encoded, err := e.EncodeUintWithLength(uint64(*v), 2)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ValidatorIndex: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// BandersnatchVrfSignature
func (bvs *BandersnatchVrfSignature) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BandersnatchVrfSignature")
	if _, err := e.buf.Write(bvs[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BandersnatchVrfSignature: %v", bvs[:]))

	return nil
}

// Header
func (h *Header) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Header")
	// Parent
	if err := h.Parent.Encode(e); err != nil {
		return err
	}

	// ParentStateRoot
	if err := h.ParentStateRoot.Encode(e); err != nil {
		return err
	}

	// ExtrinsicHash
	if err := h.ExtrinsicHash.Encode(e); err != nil {
		return err
	}

	// Slot
	if err := h.Slot.Encode(e); err != nil {
		return err
	}

	// EpochMark
	// If epoch mark is nil, append 0 to the buffer, else append 1
	epochMarkIsNil := h.EpochMark == nil
	if epochMarkIsNil {
		if _, err := e.buf.Write([]byte{0}); err != nil {
			return err
		}
	} else {
		if _, err := e.buf.Write([]byte{1}); err != nil {
			return err
		}

		// Encode EpochMark
		if err := h.EpochMark.Encode(e); err != nil {
			return err
		}
	}

	// TicketsMark
	// If tickets mark is nil, append 0 to the buffer, else append 1
	ticketsMarkIsNil := h.TicketsMark == nil
	if ticketsMarkIsNil {
		if _, err := e.buf.Write([]byte{0}); err != nil {
			return err
		}
	} else {
		if _, err := e.buf.Write([]byte{1}); err != nil {
			return err
		}

		// Encode TicketsMark
		if err := h.TicketsMark.Encode(e); err != nil {
			return err
		}
	}

	// OffenderMark
	if err := h.OffendersMark.Encode(e); err != nil {
		return err
	}

	// AuthorIndex
	if err := h.AuthorIndex.Encode(e); err != nil {
		return err
	}

	// EntropySource
	if err := h.EntropySource.Encode(e); err != nil {
		return err
	}

	// Seal
	if err := h.Seal.Encode(e); err != nil {
		return err
	}

	return nil
}

// BandersnatchRingVrfSignature
func (b *BandersnatchRingVrfSignature) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BandersnatchRingVrfSignature")
	if _, err := e.buf.Write(b[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BandersnatchRingVrfSignature: %v", b[:]))

	return nil
}

// TicketEnvelop
func (t *TicketEnvelope) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketEnvelope")

	// Attempt
	if err := t.Attempt.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := t.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// TicketsExtrinsic
func (t *TicketsExtrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketsExtrinsic")

	if err := e.EncodeLength(uint64(len(*t))); err != nil {
		return err
	}

	for _, ticketEnvelope := range *t {
		if err := ticketEnvelope.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// SerivceId
func (s *ServiceId) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServiceId")
	encoded, err := e.EncodeUintWithLength(uint64(*s), 4)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ServiceId: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// ByteSequence
func (b *ByteSequence) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ByteSequence")

	if err := e.EncodeLength(uint64(len(*b))); err != nil {
		return err
	}

	if _, err := e.buf.Write(*b); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ByteSequence: %v", *b))

	return nil
}

// Preimage
func (p *Preimage) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Preimage")

	// Requester
	if err := p.Requester.Encode(e); err != nil {
		return err
	}

	// Blob
	if err := p.Blob.Encode(e); err != nil {
		return err
	}

	return nil
}

// PreimagesExtrinsic
func (p *PreimagesExtrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding PreimagesExtrinsic")

	if err := e.EncodeLength(uint64(len(*p))); err != nil {
		return err
	}

	for _, preimage := range *p {
		if err := preimage.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// WorkPackageHash
func (w *WorkPackageHash) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkPackageHash")
	if _, err := e.buf.Write(w[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("WorkPackageHash: %v", w[:]))

	return nil
}

// ErasureRoot
func (e *ErasureRoot) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding ErasureRoot")
	if _, err := enc.buf.Write(e[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ErasureRoot: %v", e[:]))

	return nil
}

// ExportsRoot
func (e *ExportsRoot) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding ExportsRoot")
	if _, err := enc.buf.Write(e[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ExportsRoot: %v", e[:]))

	return nil
}

// WorkPackageSpec
func (w *WorkPackageSpec) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkPackageSpec")

	// Hash
	if err := w.Hash.Encode(e); err != nil {
		return err
	}

	// Length
	if err := w.Length.Encode(e); err != nil {
		return err
	}

	// ErasureRoot
	if err := w.ErasureRoot.Encode(e); err != nil {
		return err
	}

	// ExportsRoot
	if err := w.ExportsRoot.Encode(e); err != nil {
		return err
	}

	// ExportsCount
	if err := w.ExportsCount.Encode(e); err != nil {
		return err
	}

	return nil
}

// BeefyRoot
func (b *BeefyRoot) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BeefyRoot")
	if _, err := e.buf.Write(b[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BeefyRoot: %v", b[:]))

	return nil
}

// RefineContext
func (r *RefineContext) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding RefineContext")

	// Anchor
	if err := r.Anchor.Encode(e); err != nil {
		return err
	}

	// StateRoot
	if err := r.StateRoot.Encode(e); err != nil {
		return err
	}

	// BeefyRoot
	if err := r.BeefyRoot.Encode(e); err != nil {
		return err
	}

	// LookupAnchor
	if err := r.LookupAnchor.Encode(e); err != nil {
		return err
	}

	// LookupAnchorSlot
	if err := r.LookupAnchorSlot.Encode(e); err != nil {
		return err
	}

	// Prerequisites
	if err := e.EncodeLength(uint64(len(r.Prerequisites))); err != nil {
		return err
	}

	for _, prerequisite := range r.Prerequisites {
		if err := prerequisite.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// SegmentRootLookupItem
func (s *SegmentRootLookupItem) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding SegmentRootLookupItem")

	// WorkPackageHash
	if err := s.WorkPackageHash.Encode(e); err != nil {
		return err
	}

	// SegmentTreeRoot
	if err := s.SegmentTreeRoot.Encode(e); err != nil {
		return err
	}

	return nil
}

// SegmentRootLookup
func (s *SegmentRootLookup) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding SegmentRootLookup")

	if err := e.EncodeLength(uint64(len(*s))); err != nil {
		return err
	}

	for _, segmentRootLookupItem := range *s {
		if err := segmentRootLookupItem.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// CoreIndex
func (c *CoreIndex) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding CoreIndex")
	encoded, err := e.EncodeUintWithLength(uint64(*c), 2)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("CoreIndex: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// Gas
func (g *Gas) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Gas")
	encoded, err := e.EncodeUintWithLength(uint64(*g), 8)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Gas: %v", encoded))

	if _, err := e.buf.Write(encoded); err != nil {
		return err
	}

	return nil
}

// WorkExecResult
func (w *WorkExecResult) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkExecResult")

	// Check the size of map
	if len(*w) != 1 {
		return fmt.Errorf("WorkExecResult size is not equal to 1")
	}

	var key WorkExecResultType
	for k := range *w {
		key = k
		break // Get the first key and exit loop
	}

	cLog(Yellow, fmt.Sprintf("WorkExecResultType: %v", key))

	switch key {
	case "ok":
		// Encode the first byte
		if _, err := e.buf.Write([]byte{0}); err != nil {
			return err
		}

		// Get the value and encode it.
		byteSequence := (*w)["ok"]

		// Encode the length of byte sequence
		if err := e.EncodeLength(uint64(len(byteSequence))); err != nil {
			return err
		}

		if _, err := e.buf.Write(byteSequence); err != nil {
			return err
		}

		cLog(Yellow, fmt.Sprintf("WorkExecResult: %v", byteSequence))

		return nil
	case "out-of-gas":
		if _, err := e.buf.Write([]byte{1}); err != nil {
			return err
		}
		return nil
	case "panic":
		if _, err := e.buf.Write([]byte{2}); err != nil {
			return err
		}
		return nil
	case "bad-exports":
		if _, err := e.buf.Write([]byte{3}); err != nil {
			return err
		}
		return nil
	case "bad-code":
		if _, err := e.buf.Write([]byte{4}); err != nil {
			return err
		}
		return nil
	case "code-oversize":
		if _, err := e.buf.Write([]byte{5}); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("WorkExecResultType is not valid")
	}
}

// WorkResult
func (w *WorkResult) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkResult")

	// ServiceId
	if err := w.ServiceId.Encode(e); err != nil {
		return err
	}

	// CodeHash
	if err := w.CodeHash.Encode(e); err != nil {
		return err
	}

	// PayloadHash
	if err := w.PayloadHash.Encode(e); err != nil {
		return err
	}

	// AccumulateGas
	if err := w.AccumulateGas.Encode(e); err != nil {
		return err
	}

	// Result
	if err := w.Result.Encode(e); err != nil {
		return err
	}

	return nil
}

// WorkReport
func (w *WorkReport) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkReport")

	// PackageSpec
	if err := w.PackageSpec.Encode(e); err != nil {
		return err
	}

	// Context
	if err := w.Context.Encode(e); err != nil {
		return err
	}

	// CoreIndex
	if err := w.CoreIndex.Encode(e); err != nil {
		return err
	}

	// AuthorizerHash
	if err := w.AuthorizerHash.Encode(e); err != nil {
		return err
	}

	// AuthOutput
	if err := w.AuthOutput.Encode(e); err != nil {
		return err
	}

	// SegmentRootLookup
	if err := w.SegmentRootLookup.Encode(e); err != nil {
		return err
	}

	// Results
	if err := e.EncodeLength(uint64(len(w.Results))); err != nil {
		return err
	}

	for _, result := range w.Results {
		if err := result.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Ed25519Signature
func (e *Ed25519Signature) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding Ed25519Signature")
	if _, err := enc.buf.Write(e[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Ed25519Signature: %v", e[:]))

	return nil
}

// ValidatorSignature
func (v *ValidatorSignature) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ValidatorSignature")

	// ValidatorIndex
	if err := v.ValidatorIndex.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := v.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// ReportGuarantee
func (r *ReportGuarantee) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ReportGuarantee")

	// Report
	if err := r.Report.Encode(e); err != nil {
		return err
	}

	// Slot
	if err := r.Slot.Encode(e); err != nil {
		return err
	}

	// Signatures
	if err := e.EncodeLength(uint64(len(r.Signatures))); err != nil {
		return err
	}

	for _, signature := range r.Signatures {
		if err := signature.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// GuaranteesExtrinsic
func (g *GuaranteesExtrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding GuaranteesExtrinsic")

	if err := e.EncodeLength(uint64(len(*g))); err != nil {
		return err
	}

	for _, guarantee := range *g {
		if err := guarantee.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AvailAssurance
func (a *AvailAssurance) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AvailAssurance")

	// Anchor
	if err := a.Anchor.Encode(e); err != nil {
		return err
	}

	if len(a.Bitfield) != int(AvailBitfieldBytes) {
		return fmt.Errorf("Bitfield length is not equal to AvailBitfieldBytes")
	}

	// Bitfield
	if _, err := e.buf.Write(a.Bitfield); err != nil {
		return err
	}

	// ValidatorIndex
	if err := a.ValidatorIndex.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := a.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// AssurancesExtrinsic
func (a *AssurancesExtrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AssurancesExtrinsic")

	if err := e.EncodeLength(uint64(len(*a))); err != nil {
		return err
	}

	for _, assurance := range *a {
		if err := assurance.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Judgement
func (j *Judgement) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Judgement")

	// Vote
	if j.Vote {
		if _, err := e.buf.Write([]byte{1}); err != nil {
			return err
		}
	} else {
		if _, err := e.buf.Write([]byte{0}); err != nil {
			return err
		}
	}

	// ValidatorIndex
	if err := j.Index.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := j.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// Verdict
func (v *Verdict) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Verdict")

	// Target
	if err := v.Target.Encode(e); err != nil {
		return err
	}

	// Age
	if err := v.Age.Encode(e); err != nil {
		return err
	}

	// Votes
	if len(v.Votes) != int(ValidatorsSuperMajority) {
		return fmt.Errorf("Votes length is not equal to ValidatorsSuperMajority")
	}

	for _, vote := range v.Votes {
		if err := vote.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// WorkReportHash
func (w *WorkReportHash) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkReportHash")
	if _, err := e.buf.Write(w[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("WorkReportHash: %v", w[:]))

	return nil
}

// Culprit
func (c *Culprit) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Culprit")

	// Target
	if err := c.Target.Encode(e); err != nil {
		return err
	}

	// Key
	if err := c.Key.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := c.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// Fault
func (f *Fault) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Fault")

	// Target
	if err := f.Target.Encode(e); err != nil {
		return err
	}

	// Vote
	if f.Vote {
		if _, err := e.buf.Write([]byte{1}); err != nil {
			return err
		}
		cLog(Yellow, fmt.Sprintf("Fault Vote: %v", 1))
	} else {
		if _, err := e.buf.Write([]byte{0}); err != nil {
			return err
		}
		cLog(Yellow, fmt.Sprintf("Fault Vote: %v", 0))
	}

	// Key
	if err := f.Key.Encode(e); err != nil {
		return err
	}

	// Signature
	if err := f.Signature.Encode(e); err != nil {
		return err
	}

	return nil
}

// DisputesExtrinsic
func (d *DisputesExtrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding DisputesExtrinsic")

	if err := e.EncodeLength(uint64(len(d.Verdicts))); err != nil {
		return err
	}

	for _, verdict := range d.Verdicts {
		if err := verdict.Encode(e); err != nil {
			return err
		}
	}

	if err := e.EncodeLength(uint64(len(d.Culprits))); err != nil {
		return err
	}

	for _, culprit := range d.Culprits {
		if err := culprit.Encode(e); err != nil {
			return err
		}
	}

	if err := e.EncodeLength(uint64(len(d.Faults))); err != nil {
		return err
	}

	for _, fault := range d.Faults {
		if err := fault.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Extrinsic
func (ex *Extrinsic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Extrinsic")

	// Tickets
	if err := ex.Tickets.Encode(e); err != nil {
		return err
	}

	// Preimages
	if err := ex.Preimages.Encode(e); err != nil {
		return err
	}

	// Guarantees
	if err := ex.Guarantees.Encode(e); err != nil {
		return err
	}

	// Assurances
	if err := ex.Assurances.Encode(e); err != nil {
		return err
	}

	// Disputes
	if err := ex.Disputes.Encode(e); err != nil {
		return err
	}

	return nil
}

// Block
func (b *Block) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Block")

	// Header
	if err := b.Header.Encode(e); err != nil {
		return err
	}

	// Extrinsic
	if err := b.Extrinsic.Encode(e); err != nil {
		return err
	}

	return nil
}

// Authorizer
func (a *Authorizer) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Authorizer")

	// CodeHash
	if err := a.CodeHash.Encode(e); err != nil {
		return err
	}

	// Params
	if err := a.Params.Encode(e); err != nil {
		return err
	}

	return nil
}

func (i *ImportSpec) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ImportSpec")

	// TreeRoot
	if err := i.TreeRoot.Encode(e); err != nil {
		return err
	}

	// Index
	if err := i.Index.Encode(e); err != nil {
		return nil
	}

	return nil
}

func (e *ExtrinsicSpec) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding ExtrinsicSpec")

	// Hash
	if err := e.Hash.Encode(enc); err != nil {
		return err
	}

	// Len
	if err := e.Len.Encode(enc); err != nil {
		return err
	}

	return nil
}

func (w *WorkItem) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkItem")

	// Service
	if err := w.Service.Encode(e); err != nil {
		return err
	}

	// CodeHash
	if err := w.CodeHash.Encode(e); err != nil {
		return err
	}

	// Payload
	if err := w.Payload.Encode(e); err != nil {
		return err
	}

	// RefineGasLimit
	if err := w.RefineGasLimit.Encode(e); err != nil {
		return err
	}

	// AccumulateGasLimit
	if err := w.AccumulateGasLimit.Encode(e); err != nil {
		return err
	}

	// ImportSegments
	if err := e.EncodeLength(uint64(len(w.ImportSegments))); err != nil {
		return err
	}

	for _, importSegment := range w.ImportSegments {
		if err := importSegment.Encode(e); err != nil {
			return err
		}
	}

	// Extrinsic
	if err := e.EncodeLength(uint64(len(w.Extrinsic))); err != nil {
		return err
	}

	for _, extrinsic := range w.Extrinsic {
		if err := extrinsic.Encode(e); err != nil {
			return err
		}
	}

	// ExportCount
	if err := w.ExportCount.Encode(e); err != nil {
		return err
	}

	return nil
}

// WorkPackage
func (w *WorkPackage) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding WorkPackage")

	// Authorization
	if err := w.Authorization.Encode(e); err != nil {
		return err
	}

	// AuthCodeHost
	if err := w.AuthCodeHost.Encode(e); err != nil {
		return err
	}

	// Authorizer
	if err := w.Authorizer.Encode(e); err != nil {
		return err
	}

	// Context
	if err := w.Context.Encode(e); err != nil {
		return err
	}

	// Items
	if err := e.EncodeLength(uint64(len(w.Items))); err != nil {
		return err
	}

	for _, item := range w.Items {
		if err := item.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ActivityRecord
func (a *ActivityRecord) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ActivityRecord")

	// Blocks
	if err := a.Blocks.Encode(e); err != nil {
		return err
	}

	// Tickets
	if err := a.Tickets.Encode(e); err != nil {
		return err
	}

	// PreImages
	if err := a.PreImages.Encode(e); err != nil {
		return err
	}

	// PreImagesSize
	if err := a.PreImagesSize.Encode(e); err != nil {
		return err
	}

	// Guarantees
	if err := a.Guarantees.Encode(e); err != nil {
		return err
	}

	// Assurances
	if err := a.Assurances.Encode(e); err != nil {
		return err
	}

	return nil
}

// ActivityRecords
func (a *ActivityRecords) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ActivityRecords")

	if len(*a) != int(ValidatorsCount) {
		return fmt.Errorf("ActivityRecords length is not equal to ValidatorsCount")
	}

	for _, record := range *a {
		if err := record.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Statistics
func (s *Statistics) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Statistics")

	// Current
	if err := s.Current.Encode(e); err != nil {
		return err
	}

	// Last
	if err := s.Last.Encode(e); err != nil {
		return err
	}

	return nil
}

// ValidatorMetadata
func (v *ValidatorMetadata) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ValidatorMetadata")
	if _, err := e.buf.Write(v[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("ValidatorMetadata: %v", v[:]))

	return nil
}

// BlsPublic
func (b *BlsPublic) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BlsPublic")
	if _, err := e.buf.Write(b[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BlsPublic: %v", b[:]))

	return nil
}

// Validator
func (v *Validator) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Validator")

	// Bandersnatch
	if err := v.Bandersnatch.Encode(e); err != nil {
		return err
	}

	// Ed25519
	if err := v.Ed25519.Encode(e); err != nil {
		return err
	}

	// Bls
	if err := v.Bls.Encode(e); err != nil {
		return err
	}

	// Metadata
	if err := v.Metadata.Encode(e); err != nil {
		return err
	}

	return nil
}

// ValidatorsData
func (v *ValidatorsData) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ValidatorsData")

	if len(*v) != int(ValidatorsCount) {
		return fmt.Errorf("ValidatorsData length is not equal to ValidatorsCount")
	}

	for _, validator := range *v {
		if err := validator.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// EntropyBuffer
func (e *EntropyBuffer) Encode(enc *Encoder) error {
	cLog(Cyan, "Encoding EntropyBuffer")

	if len(*e) != int(4) {
		return fmt.Errorf("EntropyBuffer length is not equal to 4")
	}

	for _, entropy := range *e {
		if err := entropy.Encode(enc); err != nil {
			return err
		}
	}

	return nil
}

// TicketsAccumulator
func (t *TicketsAccumulator) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketsAccumulator")

	if err := e.EncodeLength(uint64(len(*t))); err != nil {
		return err
	}

	for _, ticketBody := range *t {
		if err := ticketBody.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// BandersnatchRingCommitment
func (b *BandersnatchRingCommitment) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BandersnatchRingCommitment")

	if _, err := e.buf.Write(b[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("BandersnatchRingCommitment: %v", b[:]))

	return nil
}

// TicketsOrKeys
func (t *TicketsOrKeys) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding TicketsOrKeys")

	// if tickets is nil, append 0 to the buffer, else append 1

	if t.Tickets == nil && t.Keys == nil {
		return fmt.Errorf("Tickets and Keys are both nil")
	}

	if t.Tickets != nil && t.Keys != nil {
		return fmt.Errorf("Tickets and Keys are both not nil")
	}

	// Tickets
	if t.Tickets != nil {
		// prefix
		e.buf.Write([]byte{0})

		// Encode Tickets
		if len(t.Tickets) != EpochLength {
			return fmt.Errorf("Tickets length is not equal to EpochLength")
		}

		for _, ticketBody := range t.Tickets {
			if err := ticketBody.Encode(e); err != nil {
				return err
			}
		}

	}

	// Keys
	if t.Keys != nil {
		// prefix
		e.buf.Write([]byte{1})

		// Encode Keys
		if len(t.Keys) != EpochLength {
			return fmt.Errorf("Keys length is not equal to EpochLength")
		}

		for _, key := range t.Keys {
			if err := key.Encode(e); err != nil {
				return err
			}
		}
	}

	return nil
}
