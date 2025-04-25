package types

import (
	"bytes"
	"fmt"
	"sort"
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

// U64
func (u *U64) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding U64")
	encoded, err := e.EncodeUintWithLength(uint64(*u), 8)
	if err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("U64: %v", encoded))

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

// EpochMarkValidatorKeys
func (emvk *EpochMarkValidatorKeys) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding EpochMarkValidatorKeys")

	// Bandersnatch
	if err := emvk.Bandersnatch.Encode(e); err != nil {
		return err
	}

	// Ed25519
	if err := emvk.Ed25519.Encode(e); err != nil {
		return err
	}

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

// RefineLoad
// INFO: This struct use C.6 integer encoding
func (r *RefineLoad) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding RefineLoad")

	// GasUsed
	if err := e.EncodeInteger(uint64(r.GasUsed)); err != nil {
		return err
	}

	// Imports
	if err := e.EncodeInteger(uint64(r.Imports)); err != nil {
		return err
	}

	// ExtrinsicCount
	if err := e.EncodeInteger(uint64(r.ExtrinsicCount)); err != nil {
		return err
	}

	// ExtrinsicSize
	if err := e.EncodeInteger(uint64(r.ExtrinsicSize)); err != nil {
		return err
	}

	// Exports
	if err := e.EncodeInteger(uint64(r.Exports)); err != nil {
		return err
	}

	return nil
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

	// RefineLoad
	if err := w.RefineLoad.Encode(e); err != nil {
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

	// AuthGasUsed
	// INFO: This field is encoded as C.6 integer
	if err := e.EncodeInteger(uint64(w.AuthGasUsed)); err != nil {
		return err
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

	// TODO: not sure if we still need this check
	if len(a.Bitfield) != int(AvailBitfieldBytes) {
		return fmt.Errorf("Bitfield length is not equal to AvailBitfieldBytes")
	}

	// Bitfield
	if err := a.Bitfield.Encode(e); err != nil {
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

func (bf *Bitfield) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Bitfield")

	_, err := e.buf.Write(bf.ToOctetSlice())
	return err
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
		return err
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

// INFO: πC , πS 是使用 C.6 進行序列化
func (c *CoreActivityRecord) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding CoreActivityRecord")

	// DALoad
	cLog(Cyan, "Encoding DALoad")
	if err := e.EncodeInteger(uint64(c.DALoad)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("DALoad: %v", c.DALoad))

	// Popularity
	cLog(Cyan, "Encoding Popularity")
	if err := e.EncodeInteger(uint64(c.Popularity)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Popularity: %v", c.Popularity))

	// Imports
	cLog(Cyan, "Encoding Imports")
	if err := e.EncodeInteger(uint64(c.Imports)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Imports: %v", c.Imports))

	// Exports
	cLog(Cyan, "Encoding Exports")
	if err := e.EncodeInteger(uint64(c.Exports)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Exports: %v", c.Exports))

	// ExtrinsicSize
	cLog(Cyan, "Encoding ExtrinsicSize")
	if err := e.EncodeInteger(uint64(c.ExtrinsicSize)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExtrinsicSize: %v", c.ExtrinsicSize))

	// ExtrinsicCount
	cLog(Cyan, "Encoding ExtrinsicCount")
	if err := e.EncodeInteger(uint64(c.ExtrinsicCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExtrinsicCount: %v", c.ExtrinsicCount))

	// BundleSize
	cLog(Cyan, "Encoding BundleSize")
	if err := e.EncodeInteger(uint64(c.BundleSize)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BundleSize: %v", c.BundleSize))

	// GasUSed
	cLog(Cyan, "Encoding GasUSed")
	if err := e.EncodeInteger(uint64(c.GasUsed)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("GasUSed: %v", c.GasUsed))

	return nil
}

// INFO: πC , πS 是使用 C.6 進行序列化
func (s *ServiceActivityRecord) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServiceActivityRecord")

	// ProvidedCount
	cLog(Cyan, "Encoding ProvidedCount")
	if err := e.EncodeInteger(uint64(s.ProvidedCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ProvidedCount: %v", s.ProvidedCount))

	// ProvidedSize
	cLog(Cyan, "Encoding ProvidedSize")
	if err := e.EncodeInteger(uint64(s.ProvidedSize)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ProvidedSize: %v", s.ProvidedSize))

	// RefinementCount
	cLog(Cyan, "Encoding RefinementCount")
	if err := e.EncodeInteger(uint64(s.RefinementCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("RefinementCount: %v", s.RefinementCount))

	// RefinementGasUsed
	cLog(Cyan, "Encoding RefinementGasUsed")
	if err := e.EncodeInteger(uint64(s.RefinementGasUsed)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("RefinementGasUsed: %v", s.RefinementGasUsed))

	// Imports
	cLog(Cyan, "Encoding Imports")
	if err := e.EncodeInteger(uint64(s.Imports)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Imports: %v", s.Imports))

	// Exports
	cLog(Cyan, "Encoding Exports")
	if err := e.EncodeInteger(uint64(s.Exports)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Exports: %v", s.Exports))

	// ExtrinsicSize
	cLog(Cyan, "Encoding ExtrinsicSize")
	if err := e.EncodeInteger(uint64(s.ExtrinsicSize)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExtrinsicSize: %v", s.ExtrinsicSize))

	// ExtrinsicCount
	cLog(Cyan, "Encoding ExtrinsicCount")
	if err := e.EncodeInteger(uint64(s.ExtrinsicCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExtrinsicCount: %v", s.ExtrinsicCount))

	// AccumulateCount
	cLog(Cyan, "Encoding AccumulateCount")
	if err := e.EncodeInteger(uint64(s.AccumulateCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("AccumulateCount: %v", s.AccumulateCount))

	// AccumulateGasUsed
	cLog(Cyan, "Encoding AccumulateGasUsed")
	if err := e.EncodeInteger(uint64(s.AccumulateGasUsed)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("AccumulateGasUsed: %v", s.AccumulateGasUsed))

	// OnTransfersCount
	cLog(Cyan, "Encoding OnTransfersCount")
	if err := e.EncodeInteger(uint64(s.OnTransfersCount)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("OnTransfersCount: %v", s.OnTransfersCount))

	// OnTransfersGasUsed
	cLog(Cyan, "Encoding OnTransfersGasUsed")
	if err := e.EncodeInteger(uint64(s.OnTransfersGasUsed)); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("OnTransfersGasUsed: %v", s.OnTransfersGasUsed))

	return nil
}

// type ServicesStatistics map[ServiceId]ServiceActivityRecord
func (s *ServicesStatistics) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServicesStatistics")

	// Encode the size of the map
	if err := e.EncodeLength(uint64(len(*s))); err != nil {
		return err
	}

	// Before encoding the map, sort the keys
	keys := make([]ServiceId, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}

	// Sort the keys (ServiceId)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// Iterate over the map and encode the key and value
	for _, key := range keys {
		// Key (ServiceId) (C.6)
		if err := e.EncodeInteger(uint64(key)); err != nil {
			return err
		}

		// value (ServiceActivityRecord)
		value := (*s)[key]
		if err := value.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// CoresStatistics
// TODO: πC , πS 是使用 C.6 進行序列化
func (c *CoresStatistics) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding CoresStatistics")

	if len(*c) != CoresCount {
		return fmt.Errorf("CoresStatistics length is not equal to CoresCount")
	}

	for _, record := range *c {
		if err := record.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Statistics
func (s *Statistics) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Statistics")

	// ValsCurrent
	if err := s.ValsCurrent.Encode(e); err != nil {
		return err
	}

	// ValsLast
	if err := s.ValsLast.Encode(e); err != nil {
		return err
	}

	// Cores
	if err := s.Cores.Encode(e); err != nil {
		return err
	}

	// Services
	if err := s.Services.Encode(e); err != nil {
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

// AvailabilityAssignment
func (a *AvailabilityAssignment) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AvailabilityAssignment")

	// Report
	if err := a.Report.Encode(e); err != nil {
		return err
	}

	// Timeout
	if err := a.Timeout.Encode(e); err != nil {
		return err
	}

	return nil
}

// AvailabilityAssignments
func (a *AvailabilityAssignments) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AvailabilityAssignments")

	if len(*a) != int(CoresCount) {
		return fmt.Errorf("AvailabilityAssignments length is not equal to CoresCount")
	}

	for _, item := range *a {
		if item == nil {
			cLog(Yellow, "AvailabilityAssignments item is nil")
			cLog(Yellow, "Appending 0 to the buffer")
			if _, err := e.buf.Write([]byte{0}); err != nil {
				return err
			}
		} else {
			cLog(Yellow, "AvailabilityAssignments item is not nil")
			cLog(Yellow, "Appending 1 to the buffer")
			if _, err := e.buf.Write([]byte{1}); err != nil {
				return err
			}

			if err := (*item).Encode(e); err != nil {
				return err
			}
		}
	}

	return nil
}

// Mmr
func (m *Mmr) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Mmr")

	if err := e.EncodeLength(uint64(len(m.Peaks))); err != nil {
		return err
	}

	for _, peak := range m.Peaks {
		// Encode pointer
		if peak == nil {
			if err := e.EncodeLength(0); err != nil {
				return err
			}
		} else {
			if err := e.EncodeLength(1); err != nil {
				return err
			}

			// Encode the peak
			if err := (*peak).Encode(e); err != nil {
				return err
			}
		}
	}

	return nil
}

// ReportedWorkPackage
func (r *ReportedWorkPackage) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ReportedWorkPackage")

	// Hash
	if err := r.Hash.Encode(e); err != nil {
		return err
	}

	// ExportsRoot
	if err := r.ExportsRoot.Encode(e); err != nil {
		return err
	}

	return nil
}

// BlockInfo
func (bi *BlockInfo) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BlockInfo")

	// HeaderHash
	if err := bi.HeaderHash.Encode(e); err != nil {
		return err
	}

	// Mmr
	if err := bi.Mmr.Encode(e); err != nil {
		return err
	}

	// StateRoot
	if err := bi.StateRoot.Encode(e); err != nil {
		return err
	}

	// Reported
	if err := e.EncodeLength(uint64(len(bi.Reported))); err != nil {
		return err
	}

	for _, reportedWorkPackage := range bi.Reported {
		if err := reportedWorkPackage.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// BlocksHistory
func (bh *BlocksHistory) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding BlocksHistory")

	if len(*bh) > int(MaxBlocksHistory) {
		return fmt.Errorf("BlocksHistory length is greater than BlocksHistoryLength")
	}

	if err := e.EncodeLength(uint64(len(*bh))); err != nil {
		return err
	}

	for _, blockInfo := range *bh {
		if err := blockInfo.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AuthorizerHash
func (ah *AuthorizerHash) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AuthorizerHash")
	if _, err := e.buf.Write(ah[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("AuthorizerHash: %v", ah[:]))

	return nil
}

// AuthPool
func (ap *AuthPool) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AuthPool")

	if len(*ap) > int(AuthPoolMaxSize) {
		return fmt.Errorf("AuthPool length is greater than AuthPoolMaxSize")
	}

	if err := e.EncodeLength(uint64(len(*ap))); err != nil {
		return err
	}

	for _, authorizerHash := range *ap {
		if err := authorizerHash.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AuthPools
func (ap *AuthPools) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AuthPools")

	// CoresCount
	if len(*ap) != int(CoresCount) {
		return fmt.Errorf("AuthPools length is not equal to CoresCount")
	}

	for _, authPool := range *ap {
		if err := authPool.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ServiceInfo
func (s *ServiceInfo) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServiceInfo")

	// CodeHash
	if err := s.CodeHash.Encode(e); err != nil {
		return err
	}

	// Balance
	if err := s.Balance.Encode(e); err != nil {
		return err
	}

	// MinItemGas
	if err := s.MinItemGas.Encode(e); err != nil {
		return err
	}

	// MinMemoGas
	if err := s.MinMemoGas.Encode(e); err != nil {
		return err
	}

	// Bytes
	if err := s.Bytes.Encode(e); err != nil {
		return err
	}

	// Items
	if err := s.Items.Encode(e); err != nil {
		return err
	}

	return nil
}

// MetaCode
func (m *MetaCode) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding MetaCode")

	// Metadata (dynamic length)
	if err := e.EncodeLength(uint64(len(m.Metadata))); err != nil {
		return err
	}

	if _, err := e.buf.Write(m.Metadata); err != nil {
		return err
	}

	// Code
	if _, err := e.buf.Write(m.Code); err != nil {
		return err
	}

	return nil
}

// DisputesRecords
func (d *DisputesRecords) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding DisputesRecords")

	// Good
	if err := e.EncodeLength(uint64(len(d.Good))); err != nil {
		return err
	}

	for _, good := range d.Good {
		if err := good.Encode(e); err != nil {
			return err
		}
	}

	// Bad
	if err := e.EncodeLength(uint64(len(d.Bad))); err != nil {
		return err
	}

	for _, bad := range d.Bad {
		if err := bad.Encode(e); err != nil {
			return err
		}
	}

	// Wonky
	if err := e.EncodeLength(uint64(len(d.Wonky))); err != nil {
		return err
	}

	for _, wonky := range d.Wonky {
		if err := wonky.Encode(e); err != nil {
			return err
		}
	}

	// Offenders
	if err := e.EncodeLength(uint64(len(d.Offenders))); err != nil {
		return err
	}

	for _, offender := range d.Offenders {
		if err := offender.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AuthQueue
func (aq *AuthQueue) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AuthQueue")

	// AuthQueueSize
	if len(*aq) != int(AuthQueueSize) {
		return fmt.Errorf("AuthQueue length is not equal to AuthQueueSize")
	}

	for _, authorizerHash := range *aq {
		if err := authorizerHash.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AuthQueues
func (aq *AuthQueues) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AuthQueues")

	// CoresCount
	if len(*aq) != int(CoresCount) {
		return fmt.Errorf("AuthQueues length is not equal to CoresCount")
	}

	for _, authQueue := range *aq {
		if err := authQueue.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ReadyRecord
func (r *ReadyRecord) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ReadyRecord")

	// Report
	if err := r.Report.Encode(e); err != nil {
		return err
	}

	// Dependencies
	if err := e.EncodeLength(uint64(len(r.Dependencies))); err != nil {
		return err
	}

	for _, dependency := range r.Dependencies {
		if err := dependency.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ReadyQueueItem
func (r *ReadyQueueItem) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ReadyQueueItem")

	if err := e.EncodeLength(uint64(len(*r))); err != nil {
		return err
	}

	for _, readyRecord := range *r {
		if err := readyRecord.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// ReadyQueue
func (rq *ReadyQueue) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ReadyQueue")

	if len(*rq) != int(EpochLength) {
		return fmt.Errorf("ReadyQueue length is not equal to EpochLength")
	}

	for _, readyQueueItem := range *rq {
		if err := readyQueueItem.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AccumulatedQueueItem
func (a *AccumulatedQueueItem) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AccumulateQueueItem")

	if err := e.EncodeLength(uint64(len(*a))); err != nil {
		return err
	}

	for _, workPackageHash := range *a {
		if err := workPackageHash.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AccumulatedQueue
func (aq *AccumulatedQueue) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AccumulateQueue")

	if len(*aq) != int(EpochLength) {
		return fmt.Errorf("AccumulateQueue length is not equal to EpochLength")
	}

	for _, accumulatedQueueItem := range *aq {
		if err := accumulatedQueueItem.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AlwaysAccumulateMapItem
func (a *AlwaysAccumulateMap) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AlwaysAccumulateMapItem")

	// Encode the size of the map
	if err := e.EncodeLength(uint64(len(*a))); err != nil {
		return err
	}

	// Before encoding, sort the map by key
	key := make([]ServiceId, 0, len(*a))
	for k := range *a {
		key = append(key, k)
	}

	// Sort the keys (ServiceId)
	sort.Slice(key, func(i, j int) bool {
		return key[i] < key[j]
	})

	// Iterate over the map and encode the key and value
	for _, k := range key {
		// ServiceId
		if err := k.Encode(e); err != nil {
			return err
		}

		value := (*a)[k]

		// Gas
		if err := value.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Privileges
func (p *Privileges) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Privileges")

	// Bless
	if err := p.Bless.Encode(e); err != nil {
		return err
	}

	// Assign
	if err := p.Assign.Encode(e); err != nil {
		return err
	}

	// Designate
	if err := p.Designate.Encode(e); err != nil {
		return err
	}

	// AlwaysAccum (dictionary)
	if err := p.AlwaysAccum.Encode(e); err != nil {
		return err
	}

	return nil
}

// AccumulateRoot
func (a *AccumulateRoot) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding AccumulateRoot")
	if _, err := e.buf.Write(a[:]); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("AccumulateRoot: %v", a[:]))

	return nil
}

// Gamma
func (g *Gamma) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Gamma")

	// GammaK
	if err := g.GammaK.Encode(e); err != nil {
		return err
	}

	// GammaZ
	if err := g.GammaZ.Encode(e); err != nil {
		return err
	}

	// GammaS
	if err := g.GammaS.Encode(e); err != nil {
		return err
	}

	// GammaA
	if err := g.GammaA.Encode(e); err != nil {
		return err
	}

	return nil
}

// // AccumulatedHistory
// func (ah *AccumulatedHistory) Encode(e *Encoder) error {
// 	cLog(Cyan, "Encoding AccumulatedHistory")

// 	if err := e.EncodeLength(uint64(len(*ah))); err != nil {
// 		return err
// 	}

// 	for _, workPackageHash := range *ah {
// 		if err := workPackageHash.Encode(e); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // AccumulatedHistories
// func (ah *AccumulatedHistories) Encode(e *Encoder) error {
// 	cLog(Cyan, "Encoding AccumulatedHistories(Xi)")

// 	if len(*ah) != int(EpochLength) {
// 		return fmt.Errorf("AccumulatedHistories length is not equal to EpochLength")
// 	}

// 	for _, accumulatedHistory := range *ah {
// 		if err := accumulatedHistory.Encode(e); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func (s *Storage) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Storage")

	// Dictionary size
	if err := e.EncodeLength(uint64(len(*s))); err != nil {
		return err
	}

	// Before encoding, sort the map by key
	keys := make([]OpaqueHash, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}

	// Sort the keys (OpaqueHash)
	sort.Slice(keys, func(i, j int) bool {
		// OpaqueHash is a byte array, so we need to compare byte by byte
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})

	// Encode the dictionary
	for _, key := range keys {
		// Encode the length of the key
		if err := e.EncodeLength(uint64(len(key))); err != nil {
			return err
		}

		// OpaqueHash
		if err := key.Encode(e); err != nil {
			return err
		}

		// ByteSequence
		value := (*s)[key]

		if err := value.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// LookupMetaMapkey
func (l *LookupMetaMapkey) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding LookupMetaMapkey")

	// Hash
	if err := l.Hash.Encode(e); err != nil {
		return err
	}

	// Length
	if err := l.Length.Encode(e); err != nil {
		return err
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding PreimagesMapEntry")

	// Dictionary size
	if err := e.EncodeLength(uint64(len(*p))); err != nil {
		return err
	}

	// Before encoding, sort the map by key
	keys := make([]OpaqueHash, 0, len(*p))
	for k := range *p {
		keys = append(keys, k)
	}

	// Sort the keys (OpaqueHash)
	sort.Slice(keys, func(i, j int) bool {
		// OpaqueHash is a byte array, so we need to compare byte by byte
		return bytes.Compare(keys[i][:], keys[j][:]) < 0
	})

	// Encode the dictionary
	for _, key := range keys {
		// OpaqueHash
		if err := key.Encode(e); err != nil {
			return err
		}

		// ByteSequence
		value := (*p)[key]

		if err := value.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// LookupMetaMapEntry
func (l *LookupMetaMapEntry) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding LookupMetaMapEntry")

	// Dictionary size
	if err := e.EncodeLength(uint64(len(*l))); err != nil {
		return err
	}

	// Before encoding, sort the map by key
	keys := make([]LookupMetaMapkey, 0, len(*l))
	for k := range *l {
		keys = append(keys, k)
	}

	// Sort the keys (DictionaryKey)
	sort.Slice(keys, func(i, j int) bool {
		iHash := keys[i].Hash
		jHash := keys[j].Hash

		iLength := keys[i].Length
		jLength := keys[j].Length

		// Compare keys with hash (first) and length (second)
		if bytes.Equal(iHash[:], jHash[:]) {
			return iLength < jLength
		} else {
			return bytes.Compare(iHash[:], jHash[:]) < 0
		}
	})

	// Encode the dictionary
	for _, key := range keys {
		// LookupMetaMapkey
		if err := key.Encode(e); err != nil {
			return err
		}

		// TimeSlot
		value := (*l)[key]

		if err := e.EncodeLength(uint64(len(value))); err != nil {
			return err
		}

		for _, timeSlot := range value {
			if err := timeSlot.Encode(e); err != nil {
				return err
			}
		}
	}

	return nil
}

// ServiceAccount
func (s *ServiceAccount) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServiceAccount")

	if err := s.ServiceInfo.Encode(e); err != nil {
		return err
	}

	if err := s.PreimageLookup.Encode(e); err != nil {
		return err
	}

	if err := s.LookupDict.Encode(e); err != nil {
		return err
	}

	if err := s.StorageDict.Encode(e); err != nil {
		return err
	}

	return nil
}

// ServiceAccountState
// We don't need to convert to DTO, we can encode directly
func (s *ServiceAccountState) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding ServiceAccountState")

	// Encode the size of the map
	if err := e.EncodeLength(uint64(len(*s))); err != nil {
		return err
	}

	// Before encoding, sort the map by key
	keys := make([]ServiceId, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}

	// Sort the keys (ServiceId, U32)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		// Encode the key (ServiceId)
		if err := k.Encode(e); err != nil {
			return err
		}

		serviceAccount := (*s)[k]

		// Encode the value (ServiceAccount)
		if err := serviceAccount.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// State
func (s *State) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding State")

	// Alpha
	if err := s.Alpha.Encode(e); err != nil {
		return err
	}

	// Varphi
	if err := s.Varphi.Encode(e); err != nil {
		return err
	}

	// Beta
	if err := s.Beta.Encode(e); err != nil {
		return err
	}

	// Gamma
	if err := s.Gamma.Encode(e); err != nil {
		return err
	}

	// Psi
	if err := s.Psi.Encode(e); err != nil {
		return err
	}

	// Eta
	if err := s.Eta.Encode(e); err != nil {
		return err
	}

	// Iota
	if err := s.Iota.Encode(e); err != nil {
		return err
	}

	// Kappa
	if err := s.Kappa.Encode(e); err != nil {
		return err
	}

	// Lambda
	if err := s.Lambda.Encode(e); err != nil {
		return err
	}

	// Rho
	if err := s.Rho.Encode(e); err != nil {
		return err
	}

	// Tau
	if err := s.Tau.Encode(e); err != nil {
		return err
	}

	// Chi
	if err := s.Chi.Encode(e); err != nil {
		return err
	}

	// Pi
	if err := s.Pi.Encode(e); err != nil {
		return err
	}

	// Theta
	if err := s.Theta.Encode(e); err != nil {
		return err
	}

	// Xi
	if err := s.Xi.Encode(e); err != nil {
		return err
	}

	// Delta
	if err := s.Delta.Encode(e); err != nil {
		return err
	}

	return nil
}

// deferredTransfer
func (d *DeferredTransfer) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding DeferredTransfer")

	// SenderID
	if err := d.SenderID.Encode(e); err != nil {
		return err
	}

	// ReceiverID
	if err := d.ReceiverID.Encode(e); err != nil {
		return err
	}

	// Balance
	if err := d.Balance.Encode(e); err != nil {
		return err
	}

	// Memo
	if _, err := e.buf.Write(d.Memo[:]); err != nil {
		return err
	}

	// GasLimit
	if err := d.GasLimit.Encode(e); err != nil {
		return err
	}

	return nil
}

func (d *DeferredTransfers) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding DeferredTransfers")

	if err := e.EncodeLength(uint64(len(*d))); err != nil {
		return err
	}

	for _, deferredTransfer := range *d {
		if err := deferredTransfer.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

func (o *Operand) Encode(e *Encoder) error {
	cLog(Cyan, "Encoding Operand")

	// Hash
	if err := o.Hash.Encode(e); err != nil {
		return err
	}

	// ExportsRoot
	if err := o.ExportsRoot.Encode(e); err != nil {
		return err
	}

	// AuthorizerHash
	if err := o.AuthorizerHash.Encode(e); err != nil {
		return err
	}

	// AuthOutput
	if err := o.AuthOutput.Encode(e); err != nil {
		return err
	}

	// PayloadHash
	if err := o.PayloadHash.Encode(e); err != nil {
		return err
	}

	// Result
	if err := o.Result.Encode(e); err != nil {
		return err
	}

	return nil
}
