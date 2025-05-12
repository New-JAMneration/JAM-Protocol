package types

import (
	"encoding/binary"
	"fmt"
)

// HeaderHash
func (h *HeaderHash) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding HeaderHash")

	var val HeaderHash
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("HeaderHash: %x", val))

	*h = val
	return nil
}

// StateRoot
func (s *StateRoot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding StateRoot")

	var val StateRoot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("StateRoot: %x", val))

	*s = val
	return nil
}

// OpaqueHash
func (o *OpaqueHash) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding OpaqueHash")

	var val OpaqueHash
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("OpaqueHash: %x", val))

	*o = val
	return nil
}

// TimeSlot
func (t *TimeSlot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TimeSlot")

	var val TimeSlot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("TimeSlot: %v", val))

	*t = val
	return nil
}

func (e *EpochMarkValidatorKeys) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding EpochMarkValidatorKeys")

	var err error

	if err = e.Bandersnatch.Decode(d); err != nil {
		return err
	}

	if err = e.Ed25519.Decode(d); err != nil {
		return err
	}

	return nil
}

// EpochMark
func (e *EpochMark) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding EpochMark")

	var err error
	if err = e.Entropy.Decode(d); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("Entropy: %x", e.Entropy))

	if err = e.TicketsEntropy.Decode(d); err != nil {
		return err
	}

	cLog(Yellow, fmt.Sprintf("TicketsEntropy: %x", e.TicketsEntropy))

	// make the slice with validators count
	e.Validators = make([]EpochMarkValidatorKeys, ValidatorsCount)
	for i := 0; i < ValidatorsCount; i++ {
		if err = e.Validators[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// Entropy
func (e *Entropy) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Entropy")

	var val Entropy
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Entropy: %x", val))

	*e = val
	return nil
}

// BandersnatchPublic
func (b *BandersnatchPublic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BandersnatchPublic")

	var val BandersnatchPublic
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BandersnatchPublic: %x", val))

	*b = val
	return nil
}

// TicketId
func (t *TicketId) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketId")

	var val TicketId
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("TicketId: %x", val))

	*t = val
	return nil
}

// TicketAttempt
func (t *TicketAttempt) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketAttempt")

	var val TicketAttempt
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("TicketAttempt: %v", val))

	*t = val
	return nil
}

func (t *TicketBody) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketBody")

	var err error
	if err = t.Id.Decode(d); err != nil {
		return err
	}

	if err = t.Attempt.Decode(d); err != nil {
		return err
	}

	return nil
}

// TicketsMark
func (t *TicketsMark) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketsMark")

	var err error

	// make the slice with epoch length
	tickets := make([]TicketBody, EpochLength)
	for i := 0; i < EpochLength; i++ {
		if err = tickets[i].Decode(d); err != nil {
			return err
		}
	}

	*t = tickets

	return nil
}

// Ed25519Public
func (e *Ed25519Public) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Ed25519Public")

	var val Ed25519Public
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Ed25519Public: %x", val))

	*e = val
	return nil
}

// OffendersMark
// type OffendersMark []Ed25519Public
func (o *OffendersMark) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding OffendersMark")

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	offenders := make([]Ed25519Public, length)
	for i := uint64(0); i < length; i++ {
		if err = offenders[i].Decode(d); err != nil {
			return err
		}
	}

	*o = offenders

	return nil
}

// ValidatorIndex
func (v *ValidatorIndex) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ValidatorIndex")

	var val ValidatorIndex
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ValidatorIndex: %v", val))

	*v = val
	return nil
}

// BandersnatchVrfSignature
func (b *BandersnatchVrfSignature) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BandersnatchVrfSignature")

	var val BandersnatchVrfSignature
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BandersnatchVrfSignature: %x", val))

	*b = val
	return nil
}

// Header
func (t *Header) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TestHeader")

	var err error

	if err = t.Parent.Decode(d); err != nil {
		return err
	}

	if err = t.ParentStateRoot.Decode(d); err != nil {
		return err
	}

	if err = t.ExtrinsicHash.Decode(d); err != nil {
		return err
	}

	if err = t.Slot.Decode(d); err != nil {
		return err
	}

	epochMarkPointerFlag, err := d.ReadPointerFlag()
	epochMarkPointerIsNil := epochMarkPointerFlag == 0
	if epochMarkPointerIsNil {
		cLog(Yellow, "EpochMark is nil")
	} else {
		cLog(Yellow, "EpochMark is not nil")
		if t.EpochMark == nil {
			t.EpochMark = &EpochMark{}
		}

		if err = t.EpochMark.Decode(d); err != nil {
			return err
		}
	}

	ticketsMarkPointerFlag, err := d.ReadPointerFlag()
	ticketsMarkPointerIsNil := ticketsMarkPointerFlag == 0
	if ticketsMarkPointerIsNil {
		cLog(Yellow, "TicketsMark is nil")
	} else {
		cLog(Yellow, "TicketsMark is not nil")
		if t.TicketsMark == nil {
			t.TicketsMark = &TicketsMark{}
		}

		if err = t.TicketsMark.Decode(d); err != nil {
			return err
		}
	}

	if err = t.OffendersMark.Decode(d); err != nil {
		return err
	}

	if err = t.AuthorIndex.Decode(d); err != nil {
		return err
	}

	if err = t.EntropySource.Decode(d); err != nil {
		return err
	}

	if err = t.Seal.Decode(d); err != nil {
		return err
	}

	return nil
}

// BandersnatchRingVrfSignature
func (b *BandersnatchRingVrfSignature) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BandersnatchRingVrfSignature")

	var val BandersnatchRingVrfSignature
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BandersnatchRingVrfSignature: %x", val))
	cLog(Yellow, fmt.Sprintf("BandersnatchRingVrfSignature Length: %v", len(val)))

	*b = val
	return nil
}

func (t *TicketEnvelope) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketEnvelope")

	var err error

	if err = t.Attempt.Decode(d); err != nil {
		return err
	}

	if err = t.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

func (t *TicketsExtrinsic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketsExtrinsic")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	tickets := make([]TicketEnvelope, length)
	for i := uint64(0); i < length; i++ {
		if err = tickets[i].Decode(d); err != nil {
			return err
		}
	}

	*t = tickets

	return nil
}

// ServiceId
func (s *ServiceId) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ServiceId")

	var val ServiceId
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ServiceId: %v", val))

	*s = val
	return nil
}

// ByteSequence
func (b *ByteSequence) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ByteSequence")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	byteSequence := make([]byte, length)
	_, err = d.buf.Read(byteSequence)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ByteSequence: %x", byteSequence))

	*b = byteSequence

	return nil
}

// Preimage
func (p *Preimage) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Preimage")

	var err error

	if err = p.Requester.Decode(d); err != nil {
		return err
	}

	if err = p.Blob.Decode(d); err != nil {
		return err
	}

	return nil
}

// PreimagesExtrinsic
func (p *PreimagesExtrinsic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding PreimagesExtrinsic")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	preimages := make([]Preimage, length)
	for i := uint64(0); i < length; i++ {
		if err = preimages[i].Decode(d); err != nil {
			return err
		}
	}

	*p = preimages

	return nil
}

// Ed25519Signature
func (e *Ed25519Signature) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Ed25519Signature")

	var val Ed25519Signature
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Ed25519Signature: %x", val))

	*e = val
	return nil
}

// ValidatorSignature
func (v *ValidatorSignature) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ValidatorSignature")

	var err error

	if err = v.ValidatorIndex.Decode(d); err != nil {
		return err
	}

	if err = v.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

// ErausreRoot
func (e *ErasureRoot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ErasureRoot")

	var val ErasureRoot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ErasureRoot: %x", val))

	*e = val
	return nil
}

// ExportsRoot
func (e *ExportsRoot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ExportsRoot")

	var val ExportsRoot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExportsRoot: %x", val))

	*e = val
	return nil
}

// WorkPackageSpec
func (w *WorkPackageSpec) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkPackageSpec")

	var err error

	if err = w.Hash.Decode(d); err != nil {
		return err
	}

	if err = w.Length.Decode(d); err != nil {
		return err
	}

	if err = w.ErasureRoot.Decode(d); err != nil {
		return err
	}

	if err = w.ExportsRoot.Decode(d); err != nil {
		return err
	}

	if err = w.ExportsCount.Decode(d); err != nil {
		return err
	}

	return nil
}

// BeefyRoot
func (b *BeefyRoot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BeefyRoot")

	var val BeefyRoot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BeefyRoot: %x", val))

	*b = val
	return nil
}

func (r *RefineContext) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding RefineContext")

	var err error

	if err = r.Anchor.Decode(d); err != nil {
		return err
	}

	if err = r.StateRoot.Decode(d); err != nil {
		return err
	}

	if err = r.BeefyRoot.Decode(d); err != nil {
		return err
	}

	if err = r.LookupAnchor.Decode(d); err != nil {
		return err
	}

	if err = r.LookupAnchorSlot.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	// Default slice is nil, you can't make a slice with length 0
	if length == 0 {
		return nil
	}

	// Make the slice with length
	prerequisites := make([]OpaqueHash, length)
	for i := uint64(0); i < length; i++ {
		if err = prerequisites[i].Decode(d); err != nil {
			return err
		}
	}

	r.Prerequisites = prerequisites

	return nil
}

// CoreIndex
func (c *CoreIndex) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding CoreIndex")

	var val CoreIndex
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("CoreIndex: %v", val))

	*c = val
	return nil
}

// WorkPackageHash
func (w *WorkPackageHash) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkPackageHash")

	var val WorkPackageHash
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("WorkPackageHash: %x", val))

	*w = val
	return nil
}

// SegementRootLookupItem
func (s *SegmentRootLookupItem) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding SegmentRootLookupItem")

	var err error

	if err = s.WorkPackageHash.Decode(d); err != nil {
		return err
	}

	if err = s.SegmentTreeRoot.Decode(d); err != nil {
		return err
	}

	return nil
}

// SegmentRootLookup
func (s *SegmentRootLookup) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding SegmentRootLookup")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	segmentRootLookup := make([]SegmentRootLookupItem, length)
	for i := uint64(0); i < length; i++ {
		if err = segmentRootLookup[i].Decode(d); err != nil {
			return err
		}
	}

	*s = segmentRootLookup

	return nil
}

// WorkExecResult
func (w *WorkExecResult) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkExecResult")

	var err error

	// Get the first byte
	firstByte, err := d.buf.ReadByte()
	if err != nil {
		return err
	}

	// make the map
	*w = WorkExecResult{}

	switch firstByte {
	case 0:
		cLog(Yellow, "WorkExecResultOk")

		// decode byteseuqnce
		var byteSequence ByteSequence
		if err = byteSequence.Decode(d); err != nil {
			return err
		}

		// set the map
		(*w)["ok"] = byteSequence
	case 1:
		cLog(Yellow, "WorkExecResultOutOfGas")
		(*w)["out-of-gas"] = nil
	case 2:
		cLog(Yellow, "WorkExecResultPanic")
		(*w)["panic"] = nil
	case 3:
		cLog(Yellow, "WorkExecResultBadExports")
		(*w)["bad-exports"] = nil
	case 4:
		cLog(Yellow, "WorkExecResultReportOversize")
		(*w)["report-oversize"] = nil
	case 5:
		cLog(Yellow, "WorkExecResultBadCode")
		(*w)["bad-code"] = nil
	case 6:
		cLog(Yellow, "WorkExecResultCodeOversize")
		(*w)["code-oversize"] = nil
	}

	return nil
}

// Gas
func (g *Gas) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Gas")

	var val Gas
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Gas: %v", val))

	*g = val
	return nil
}

// RefineLoad
func (r *RefineLoad) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding RefineLoad")

	var err error

	gasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	r.GasUsed = U64(gasUsed)

	imports, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	r.Imports = U16(imports)

	extrinsicCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	r.ExtrinsicCount = U16(extrinsicCount)

	extrinsicSize, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	r.ExtrinsicSize = U32(extrinsicSize)

	exports, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	r.Exports = U16(exports)

	return nil
}

func (w *WorkResult) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkResult")

	var err error

	if err = w.ServiceId.Decode(d); err != nil {
		return err
	}

	if err = w.CodeHash.Decode(d); err != nil {
		return err
	}

	if err = w.PayloadHash.Decode(d); err != nil {
		return err
	}

	if err = w.AccumulateGas.Decode(d); err != nil {
		return err
	}

	if err = w.Result.Decode(d); err != nil {
		return err
	}

	if err = w.RefineLoad.Decode(d); err != nil {
		return err
	}

	return nil
}

// WorkReport
func (w *WorkReport) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkReport")

	var err error

	if err = w.PackageSpec.Decode(d); err != nil {
		return err
	}

	if err = w.Context.Decode(d); err != nil {
		return err
	}

	if err = w.CoreIndex.Decode(d); err != nil {
		return err
	}

	if err = w.AuthorizerHash.Decode(d); err != nil {
		return err
	}

	if err = w.AuthOutput.Decode(d); err != nil {
		return err
	}

	if err = w.SegmentRootLookup.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	results := make([]WorkResult, length)
	for i := uint64(0); i < length; i++ {
		if err = results[i].Decode(d); err != nil {
			return err
		}
	}

	w.Results = results

	authGasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}

	w.AuthGasUsed = U64(authGasUsed)

	return nil
}

// ReportGuarantee
func (r *ReportGuarantee) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ReportGuarantee")

	var err error

	// Report
	if err = r.Report.Decode(d); err != nil {
		return err
	}

	// Slot
	if err = r.Slot.Decode(d); err != nil {
		return err
	}

	// Signatures
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	signatures := make([]ValidatorSignature, length)
	for i := uint64(0); i < length; i++ {
		if err = signatures[i].Decode(d); err != nil {
			return err
		}
	}

	r.Signatures = signatures

	return nil
}

// GuaranteesExtrinsic
func (g *GuaranteesExtrinsic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding GuaranteesExtrinsic")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	guarantees := make([]ReportGuarantee, length)
	for i := uint64(0); i < length; i++ {
		if err = guarantees[i].Decode(d); err != nil {
			return err
		}
	}
	*g = guarantees

	return nil
}

// AvailAssurance
func (a *AvailAssurance) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AvailAssurance")

	var err error

	if err = a.Anchor.Decode(d); err != nil {
		return err
	}

	if err = a.Bitfield.Decode(d); err != nil {
		return err
	}

	if err = a.ValidatorIndex.Decode(d); err != nil {
		return err
	}

	if err = a.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

func (bf *Bitfield) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Bitfield")

	bytes := make([]byte, AvailBitfieldBytes)
	_, err := d.buf.Read(bytes)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BitField: %x", bytes))

	bitfield, err := MakeBitfieldFromByteSlice(bytes)
	if err != nil {
		return err
	}

	*bf = bitfield
	return nil
}

// AssurancesExtrinsic
func (a *AssurancesExtrinsic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AssurancesExtrinsic")

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	assurances := make([]AvailAssurance, length)
	for i := uint64(0); i < length; i++ {
		if err = assurances[i].Decode(d); err != nil {
			return err
		}
	}

	*a = assurances

	return nil
}

// WorkReportHash
func (w *WorkReportHash) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkReportHash")

	var val WorkReportHash
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("WorkReportHash: %x", val))

	*w = val
	return nil
}

// Culprit
func (c *Culprit) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Culprit")

	var err error

	if err = c.Target.Decode(d); err != nil {
		return err
	}

	if err = c.Key.Decode(d); err != nil {
		return err
	}

	if err = c.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

// Fault
func (f *Fault) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Fault")

	var err error

	if err = f.Target.Decode(d); err != nil {
		return err
	}

	// read a byte for bool
	vote, err := d.buf.ReadByte()
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Vote: %v", vote))

	f.Vote = vote == 1

	if err = f.Key.Decode(d); err != nil {
		return err
	}

	if err = f.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

// U8
func (u *U8) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding U8")

	var val U8
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("U8: %v", val))

	*u = val
	return nil
}

// U16
func (u *U16) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding U16")

	var val U16
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("U16: %v", val))

	*u = val
	return nil
}

// U32
func (u *U32) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding U32")

	var val U32
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("U32: %v", val))

	*u = val
	return nil
}

// U64
func (u *U64) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding U64")

	var val U64
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("U64: %v", val))

	*u = val
	return nil
}

// Judgement
func (j *Judgement) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Judgement")

	var err error

	// read a byte for bool
	vote, err := d.buf.ReadByte()
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("Vote: %v", vote))

	j.Vote = vote == 1

	if err = j.Index.Decode(d); err != nil {
		return err
	}

	if err = j.Signature.Decode(d); err != nil {
		return err
	}

	return nil
}

func (v *Verdict) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Verdict")

	var err error

	if err = v.Target.Decode(d); err != nil {
		return err
	}

	if err = v.Age.Decode(d); err != nil {
		return err
	}

	length := ValidatorsSuperMajority

	// make the slice with length
	votes := make([]Judgement, length)
	for i := 0; i < length; i++ {
		if err = votes[i].Decode(d); err != nil {
			return err
		}
	}

	v.Votes = votes

	return nil
}

// DisputesExtrinsic
func (d *DisputesExtrinsic) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding DisputesExtrinsic")

	var err error

	length, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if length != 0 {
		// make the slice with length
		verdicts := make([]Verdict, length)
		for i := uint64(0); i < length; i++ {
			if err = verdicts[i].Decode(decoder); err != nil {
				return err
			}
		}
		d.Verdicts = verdicts
	}

	length, err = decoder.DecodeLength()
	if err != nil {
		return err
	}

	if length != 0 {
		// make the slice with length
		culprits := make([]Culprit, length)
		for i := uint64(0); i < length; i++ {
			if err = culprits[i].Decode(decoder); err != nil {
				return err
			}
		}

		d.Culprits = culprits
	}

	length, err = decoder.DecodeLength()
	if err != nil {
		return err
	}

	if length != 0 {
		// make the slice with length
		faults := make([]Fault, length)
		for i := uint64(0); i < length; i++ {
			if err = faults[i].Decode(decoder); err != nil {
				return err
			}
		}

		d.Faults = faults
	}

	return nil
}

// Extrinsic
func (e *Extrinsic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Extrinsic")

	var err error

	if err = e.Tickets.Decode(d); err != nil {
		return err
	}

	if err = e.Preimages.Decode(d); err != nil {
		return err
	}

	if err = e.Guarantees.Decode(d); err != nil {
		return err
	}

	if err = e.Assurances.Decode(d); err != nil {
		return err
	}

	if err = e.Disputes.Decode(d); err != nil {
		return err
	}

	return nil
}

// Block
func (b *Block) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Block")

	var err error

	if err = b.Header.Decode(d); err != nil {
		return err
	}

	if err = b.Extrinsic.Decode(d); err != nil {
		return err
	}

	return nil
}

// ImportSpec
func (i *ImportSpec) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ImportSpec")

	var err error

	if err = i.TreeRoot.Decode(d); err != nil {
		return err
	}

	if err = i.Index.Decode(d); err != nil {
		return err
	}

	return nil
}

// ExtrinsicSpec
func (e *ExtrinsicSpec) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ExtrinsicSpec")

	var err error

	if err = e.Hash.Decode(d); err != nil {
		return err
	}

	if err = e.Len.Decode(d); err != nil {
		return err
	}

	return nil
}

// WorkItem
func (w *WorkItem) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkItem")

	var err error

	if err = w.Service.Decode(d); err != nil {
		return err
	}

	if err = w.CodeHash.Decode(d); err != nil {
		return err
	}

	if err = w.Payload.Decode(d); err != nil {
		return err
	}

	if err = w.RefineGasLimit.Decode(d); err != nil {
		return err
	}

	if err = w.AccumulateGasLimit.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	importSegments := make([]ImportSpec, length)
	for i := uint64(0); i < length; i++ {
		if err = importSegments[i].Decode(d); err != nil {
			return err
		}
	}

	w.ImportSegments = importSegments

	length, err = d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	extrinsic := make([]ExtrinsicSpec, length)
	for i := uint64(0); i < length; i++ {
		if err = extrinsic[i].Decode(d); err != nil {
			return err
		}
	}

	w.Extrinsic = extrinsic

	if err = w.ExportCount.Decode(d); err != nil {
		return err
	}

	return nil
}

// Authorizer
func (a *Authorizer) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Authorizer")

	var err error

	if err = a.CodeHash.Decode(d); err != nil {
		return err
	}

	if err = a.Params.Decode(d); err != nil {
		return err
	}

	return nil
}

// WorkPackage
func (w *WorkPackage) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkPackage")

	var err error

	if err = w.Authorization.Decode(d); err != nil {
		return err
	}

	if err = w.AuthCodeHost.Decode(d); err != nil {
		return err
	}

	if err = w.Authorizer.Decode(d); err != nil {
		return err
	}

	if err = w.Context.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	items := make([]WorkItem, length)
	for i := uint64(0); i < length; i++ {
		if err = items[i].Decode(d); err != nil {
			return err
		}
	}

	w.Items = items

	return nil
}

// ActivityRecord
func (a *ActivityRecord) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ActivityRecord")

	var err error

	if err = a.Blocks.Decode(d); err != nil {
		return err
	}

	if err = a.Tickets.Decode(d); err != nil {
		return err
	}

	if err = a.PreImages.Decode(d); err != nil {
		return err
	}

	if err = a.PreImagesSize.Decode(d); err != nil {
		return err
	}

	if err = a.Guarantees.Decode(d); err != nil {
		return err
	}

	if err = a.Assurances.Decode(d); err != nil {
		return err
	}

	return nil
}

// ActivityRecords
func (a *ActivityRecords) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ActivityRecords")

	var err error

	// make the slice with length
	records := make([]ActivityRecord, ValidatorsCount)
	for i := 0; i < ValidatorsCount; i++ {
		if err = records[i].Decode(d); err != nil {
			return err
		}
	}

	*a = records

	return nil
}

// CoreActivityRecord
func (c *CoreActivityRecord) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding CoreActivityRecord")

	var err error

	cLog(Cyan, "Decoding DALoad")
	daLoad, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.DALoad = U32(daLoad)
	cLog(Yellow, fmt.Sprintf("DALoad: %v", c.DALoad))

	cLog(Cyan, "Decoding Popularity")
	popularity, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.Popularity = U16(popularity)
	cLog(Yellow, fmt.Sprintf("Popularity: %v", c.Popularity))

	cLog(Cyan, "Decoding Imports")
	imports, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.Imports = U16(imports)
	cLog(Yellow, fmt.Sprintf("Imports: %v", c.Imports))

	cLog(Cyan, "Decoding Exports")
	exports, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.Exports = U16(exports)
	cLog(Yellow, fmt.Sprintf("Exports: %v", c.Exports))

	cLog(Cyan, "Decoding ExtrinsicSize")
	extrinsicSize, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.ExtrinsicSize = U32(extrinsicSize)
	cLog(Yellow, fmt.Sprintf("ExtrinsicSize: %v", c.ExtrinsicSize))

	cLog(Cyan, "Decoding ExtrinsicCount")
	extrinsicCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.ExtrinsicCount = U16(extrinsicCount)
	cLog(Yellow, fmt.Sprintf("ExtrinsicCount: %v", c.ExtrinsicCount))

	cLog(Cyan, "Decoding AccumulateCount")
	bundleSize, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.BundleSize = U32(bundleSize)
	cLog(Yellow, fmt.Sprintf("BundleSize: %v", c.BundleSize))

	cLog(Cyan, "Decoding AccumulateCount")
	gasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	c.GasUsed = U64(gasUsed)
	cLog(Yellow, fmt.Sprintf("GasUsed: %v", c.GasUsed))

	return nil
}

// type CoresStatistics []CoreActivityRecord
func (c *CoresStatistics) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding CoresStatistics")

	var err error

	// make the slice with length
	cores := make([]CoreActivityRecord, CoresCount)
	for i := 0; i < CoresCount; i++ {
		if err = cores[i].Decode(d); err != nil {
			return err
		}
	}

	*c = cores

	return nil
}

// ServiceActivityRecord
func (s *ServiceActivityRecord) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ServiceActivityRecord")

	var err error

	cLog(Cyan, "Decoding ProvidedCount")
	providedCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.ProvidedCount = U16(providedCount)
	cLog(Yellow, fmt.Sprintf("ProvidedCount: %v", s.ProvidedCount))

	cLog(Cyan, "Decoding ProvidedSize")
	providedSize, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.ProvidedSize = U32(providedSize)
	cLog(Yellow, fmt.Sprintf("ProvidedSize: %v", s.ProvidedSize))

	cLog(Cyan, "Decoding RefinementCount")
	refinementCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.RefinementCount = U32(refinementCount)
	cLog(Yellow, fmt.Sprintf("RefinementCount: %v", refinementCount))

	cLog(Cyan, "Decoding RefinementGasUsed")
	refinementGasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.RefinementGasUsed = U64(refinementGasUsed)
	cLog(Yellow, fmt.Sprintf("RefinementGasUsed: %v", refinementGasUsed))

	cLog(Cyan, "Decoding Imports")
	imports, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.Imports = U32(imports)
	cLog(Yellow, fmt.Sprintf("Imports: %v", imports))

	cLog(Cyan, "Decoding Exports")
	exports, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.Exports = U32(exports)
	cLog(Yellow, fmt.Sprintf("Exports: %v", exports))

	cLog(Cyan, "Decoding ExtrinsicSize")
	extrinsicSize, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.ExtrinsicSize = U32(extrinsicSize)
	cLog(Yellow, fmt.Sprintf("ExtrinsicSize: %v", extrinsicSize))

	cLog(Cyan, "Decoding ExtrinsicCount")
	extrinsicCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.ExtrinsicCount = U32(extrinsicCount)
	cLog(Yellow, fmt.Sprintf("ExtrinsicCount: %v", extrinsicCount))

	cLog(Cyan, "Decoding AccumulateCount")
	accumulateCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.AccumulateCount = U32(accumulateCount)
	cLog(Yellow, fmt.Sprintf("AccumulateCount: %v", accumulateCount))

	cLog(Cyan, "Decoding AccumulateGasUsed")
	accumulateGasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.AccumulateGasUsed = U64(accumulateGasUsed)
	cLog(Yellow, fmt.Sprintf("AccumulateGasUsed: %v", accumulateGasUsed))

	cLog(Cyan, "Decoding OnTransfersCount")
	onTransfersCount, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.OnTransfersCount = U32(onTransfersCount)
	cLog(Yellow, fmt.Sprintf("OnTransfersCount: %v", onTransfersCount))

	cLog(Cyan, "Decoding OnTransfersGasUsed")
	onTransfersGasUsed, err := d.DecodeInteger()
	if err != nil {
		return err
	}
	s.OnTransfersGasUsed = U64(onTransfersGasUsed)
	cLog(Yellow, fmt.Sprintf("OnTransfersGasUsed: %v", onTransfersGasUsed))

	return nil
}

// ServicesStatistics
func (s *ServicesStatistics) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ServicesStatistics")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the map
	services := make(ServicesStatistics)

	for i := uint64(0); i < length; i++ {
		serviceId, err := d.DecodeInteger()
		if err != nil {
			return err
		}

		var serviceActivityRecord ServiceActivityRecord
		if err = serviceActivityRecord.Decode(d); err != nil {
			return err
		}

		services[ServiceId(serviceId)] = serviceActivityRecord
	}

	*s = services

	return nil
}

// Statistics
func (s *Statistics) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Statistics")

	var err error

	if err = s.ValsCurrent.Decode(d); err != nil {
		return err
	}

	if err = s.ValsLast.Decode(d); err != nil {
		return err
	}

	if err = s.Cores.Decode(d); err != nil {
		return err
	}

	if err = s.Services.Decode(d); err != nil {
		return err
	}

	return nil
}

// BlsPublic
func (b *BlsPublic) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BlsPublic")

	var val BlsPublic
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BlsPublic: %x", val))

	*b = val
	return nil
}

// ValidatorMetadata
func (v *ValidatorMetadata) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ValidatorMetadata")

	var val ValidatorMetadata
	err := binary.Read(d.buf, binary.LittleEndian, &val)
	if err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ValidatorMetadata: %x", val))

	*v = val

	return nil
}

// Validator
func (v *Validator) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Validator")

	var err error

	if err = v.Bandersnatch.Decode(d); err != nil {
		return err
	}

	if err = v.Ed25519.Decode(d); err != nil {
		return err
	}

	if err = v.Bls.Decode(d); err != nil {
		return err
	}

	if err = v.Metadata.Decode(d); err != nil {
		return err
	}

	return nil
}

// ValidatorsData
func (v *ValidatorsData) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ValidatorsData")

	var err error

	cLog(Yellow, fmt.Sprintf("ValidatorsCount: %v", ValidatorsCount))

	// make the slice with length
	validators := make([]Validator, ValidatorsCount)
	for i := 0; i < ValidatorsCount; i++ {
		if err = validators[i].Decode(d); err != nil {
			return err
		}
	}

	*v = validators

	return nil
}

// EntropyBuffer
func (e *EntropyBuffer) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding EntropyBuffer")

	var err error

	// make the slice with length
	entropyBuffer := EntropyBuffer{}
	for i := 0; i < 4; i++ {
		if err = entropyBuffer[i].Decode(d); err != nil {
			return err
		}
	}
	cLog(Yellow, fmt.Sprintf("EntropyBuffer: %x", entropyBuffer))

	*e = entropyBuffer

	return nil
}

// TicketsAccumulator
func (t *TicketsAccumulator) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketsAccumulator")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with epoch length
	tickets := make([]TicketBody, length)
	for i := uint64(0); i < length; i++ {
		if err = tickets[i].Decode(d); err != nil {
			return err
		}
	}

	*t = tickets

	return nil
}

// TicketsOrKeys
func (t *TicketsOrKeys) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding TicketsOrKeys")

	var err error

	// if the first byte is 0, it means Tickets is nil
	// Otherwise, it means Tickets is not nil

	firstByte, err := d.ReadPointerFlag()
	isTickets := firstByte == 0
	isKeys := firstByte == 1

	if isTickets {
		cLog(Cyan, "TicketsOrKeys is Tickets")
		// make the slice with epoch length
		tickets := make([]TicketBody, EpochLength)
		for i := 0; i < EpochLength; i++ {
			if err = tickets[i].Decode(d); err != nil {
				return err
			}
		}

		t.Tickets = tickets
		return nil
	}

	if isKeys {
		cLog(Cyan, "TicketsOrKeys is Keys")
		// make the slice with epoch length
		keys := make([]BandersnatchPublic, EpochLength)
		for i := 0; i < EpochLength; i++ {
			if err = keys[i].Decode(d); err != nil {
				return err
			}
		}

		t.Keys = keys
		return nil
	}

	return nil
}

// BandersnatchRingCommitment
func (b *BandersnatchRingCommitment) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BandersnatchRingCommitment")

	var err error

	var val BandersnatchRingCommitment
	if err = binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("BandersnatchRingCommitment: %x", val))

	*b = val
	return nil
}

// AvailabilityAssignment
func (a *AvailabilityAssignment) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AvailabilityAssignment")

	var err error

	if err = a.Report.Decode(d); err != nil {
		return err
	}

	if err = a.Timeout.Decode(d); err != nil {
		return err
	}

	return nil
}

// type AvailabilityAssignments []AvailabilityAssignmentsItem
func (a *AvailabilityAssignments) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AvailabilityAssignments")

	for i := 0; i < CoresCount; i++ {
		pointerFlag, err := d.ReadPointerFlag()
		if err != nil {
			return err
		}

		pointerIsNil := pointerFlag == 0
		if pointerIsNil {
			cLog(Yellow, "AvailabilityAssignmentsItem is nil")
			item := (*AvailabilityAssignment)(nil)
			*a = append(*a, item)
			continue
		}

		cLog(Yellow, "AvailabilityAssignmentsItem is not nil")

		// Decode the AvailabilityAssignment
		var assignment AvailabilityAssignment
		if err = assignment.Decode(d); err != nil {
			return err
		}

		item := &assignment
		*a = append(*a, item)
	}

	return nil
}

// Mmr
func (m *Mmr) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Mmr")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	peaks := make([]MmrPeak, length)
	for i := uint64(0); i < length; i++ {
		// check pointer flag
		pointerFlag, err := d.ReadPointerFlag()
		if err != nil {
			return err
		}
		pointerIsNil := pointerFlag == 0
		if pointerIsNil {
			cLog(Yellow, "MmrPeak is nil")
		}

		if !pointerIsNil {
			cLog(Yellow, "MmrPeak is not nil")
			var val OpaqueHash
			if err = val.Decode(d); err != nil {
				return err
			}
			peaks[i] = MmrPeak(&val)
		}
	}

	m.Peaks = peaks

	return nil
}

// ReportedWorkPackage
func (r *ReportedWorkPackage) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ReportedWorkPackage")

	var err error

	if err = r.Hash.Decode(d); err != nil {
		return err
	}

	if err = r.ExportsRoot.Decode(d); err != nil {
		return err
	}

	return nil
}

// BlockInfo
func (b *BlockInfo) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BlockInfo")

	var err error

	if err = b.HeaderHash.Decode(d); err != nil {
		return err
	}

	if err = b.Mmr.Decode(d); err != nil {
		return err
	}

	if err = b.StateRoot.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	reported := make([]ReportedWorkPackage, length)
	for i := uint64(0); i < length; i++ {
		if err = reported[i].Decode(d); err != nil {
			return err
		}
	}

	b.Reported = reported

	return nil
}

// BlocksHistory
func (b *BlocksHistory) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding BlocksHistory")

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	history := make([]BlockInfo, length)
	for i := uint64(0); i < length; i++ {
		if err = history[i].Decode(d); err != nil {
			return err
		}
	}

	*b = history

	return nil
}

// AuthorizerHash
func (a *AuthorizerHash) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AuthorizerHash")

	var val AuthorizerHash
	if err := val.Decode(d); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("AuthorizerHash: %x", val))

	*a = val
	return nil
}

// AuthPool
func (a *AuthPool) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AuthPool")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	pool := make([]OpaqueHash, length)
	for i := uint64(0); i < length; i++ {
		if err = pool[i].Decode(d); err != nil {
			return err
		}
	}

	// convert to AuthPool
	for i := 0; i < len(pool); i++ {
		// convert to AuthorizerHash
		authorizerHash := AuthorizerHash(pool[i])
		// append value to AuthPool
		*a = append(*a, authorizerHash)
	}

	return nil
}

// AuthPools
func (a *AuthPools) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AuthPools")

	// make the slice with length
	pools := make([]AuthPool, CoresCount)
	for i := 0; i < CoresCount; i++ {
		if err := pools[i].Decode(d); err != nil {
			return err
		}
	}

	*a = pools

	return nil
}

// ServiceInfo
func (s *ServiceInfo) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ServiceInfo")

	var err error

	if err = s.CodeHash.Decode(d); err != nil {
		return err
	}

	if err = s.Balance.Decode(d); err != nil {
		return err
	}

	if err = s.MinItemGas.Decode(d); err != nil {
		return err
	}

	if err = s.MinMemoGas.Decode(d); err != nil {
		return err
	}

	if err = s.Bytes.Decode(d); err != nil {
		return err
	}

	if err = s.Items.Decode(d); err != nil {
		return err
	}

	return nil
}

// Decode MetaCode
func (m *MetaCode) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding MetaCode")

	var err error

	if d.buf.Len() == 0 {
		return nil
	}

	// Decode Length
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Decode the Metadata
	metadata := make([]byte, length)
	if _, err = d.buf.Read(metadata); err != nil {
		return err
	}

	m.Metadata = ByteSequence(metadata)

	// Decode the Code (remaining bytes)
	code := make([]byte, d.buf.Len())
	if _, err = d.buf.Read(code); err != nil {
		return err
	}

	m.Code = ByteSequence(code)

	return nil
}

// DisputesRecords
func (d *DisputesRecords) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding DisputesRecords")

	var err error

	// Good
	goodLength, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if goodLength != 0 {
		good := make([]WorkReportHash, goodLength)
		for i := uint64(0); i < goodLength; i++ {
			if err = good[i].Decode(decoder); err != nil {
				return err
			}
		}
		cLog(Yellow, fmt.Sprintf("Good: %v", good))

		d.Good = good
	}

	// Bad
	badLength, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if badLength != 0 {
		bad := make([]WorkReportHash, badLength)
		for i := uint64(0); i < badLength; i++ {
			if err = bad[i].Decode(decoder); err != nil {
				return err
			}
		}
		cLog(Yellow, fmt.Sprintf("Bad: %v", bad))

		d.Bad = bad
	}

	// Wonky
	wonkyLength, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if wonkyLength != 0 {
		wonky := make([]WorkReportHash, wonkyLength)
		for i := uint64(0); i < wonkyLength; i++ {
			if err = wonky[i].Decode(decoder); err != nil {
				return err
			}
		}
		cLog(Yellow, fmt.Sprintf("Wonky: %v", wonky))

		d.Wonky = wonky
	}

	// Offenders
	offendersLength, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if offendersLength != 0 {
		offenders := make([]Ed25519Public, offendersLength)
		for i := uint64(0); i < offendersLength; i++ {
			if err = offenders[i].Decode(decoder); err != nil {
				return err
			}
		}
		cLog(Yellow, fmt.Sprintf("Offenders: %v", offenders))

		d.Offenders = offenders
	}

	return nil
}

// AuthQueue
func (a *AuthQueue) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AuthQueue")

	// make the slice with length
	queue := make([]OpaqueHash, AuthQueueSize)
	for i := 0; i < AuthQueueSize; i++ {
		if err := queue[i].Decode(d); err != nil {
			return err
		}
	}

	// convert to AuthQueue
	for i := 0; i < len(queue); i++ {
		// convert to AuthorizerHash
		authorizerHash := AuthorizerHash(queue[i])
		// append value to AuthQueue
		*a = append(*a, authorizerHash)
	}

	return nil
}

// AuthQueues
func (a *AuthQueues) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AuthQueues")

	queues := make([]AuthQueue, CoresCount)
	for i := 0; i < CoresCount; i++ {
		if err := queues[i].Decode(d); err != nil {
			return err
		}
	}

	*a = queues

	return nil
}

// AccumulateRoot
func (a *AccumulateRoot) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AccumulateRoot")

	var val AccumulateRoot
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("AccumulateRoot: %x", val))

	*a = val
	return nil
}

// ReadyRecord
func (r *ReadyRecord) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ReadyRecord")

	var err error

	if err = r.Report.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	dependencies := make([]WorkPackageHash, length)
	for i := uint64(0); i < length; i++ {
		if err = dependencies[i].Decode(d); err != nil {
			return err
		}
	}

	r.Dependencies = dependencies

	return nil
}

// 	ReadyQueueItem       []ReadyRecord

// ReadyQueueItem
func (r *ReadyQueueItem) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ReadyQueueItem")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	records := make([]ReadyRecord, length)
	for i := uint64(0); i < length; i++ {
		if err = records[i].Decode(d); err != nil {
			return err
		}
	}

	*r = records

	return nil
}

// ReadyQueue
func (r *ReadyQueue) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ReadyQueue")

	// make the slice with epoch length
	queue := make([]ReadyQueueItem, EpochLength)
	for i := 0; i < EpochLength; i++ {
		if err := queue[i].Decode(d); err != nil {
			return err
		}
	}

	*r = queue

	return nil
}

// AccumulatedQueueItem
func (a *AccumulatedQueueItem) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AccumulatedQueueItem")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	items := make([]WorkPackageHash, length)
	for i := uint64(0); i < length; i++ {
		if err = items[i].Decode(d); err != nil {
			return err
		}
	}

	*a = items

	return nil
}

// AccumulatedQueue
func (a *AccumulatedQueue) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AccumulatedQueue")

	// make the slice with epoch length
	queue := make([]AccumulatedQueueItem, EpochLength)
	for i := 0; i < EpochLength; i++ {
		if err := queue[i].Decode(d); err != nil {
			return err
		}
	}

	*a = queue

	return nil
}

// AlwaysAccumulateMapItem
func (a *AlwaysAccumulateMap) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding AlwaysAccumulateMapItem")

	var err error

	// Encode the size of the map
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	// make the map with length
	*a = make(AlwaysAccumulateMap, length)

	for i := uint64(0); i < length; i++ {
		var key ServiceId
		if err = key.Decode(d); err != nil {
			return err
		}

		var val Gas
		if err = val.Decode(d); err != nil {
			return err
		}

		(*a)[key] = val
	}

	return nil
}

// Privileges
func (p *Privileges) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Privileges")

	var err error

	if err = p.Bless.Decode(d); err != nil {
		return err
	}

	if err = p.Assign.Decode(d); err != nil {
		return err
	}

	if err = p.Designate.Decode(d); err != nil {
		return err
	}

	if err = p.AlwaysAccum.Decode(d); err != nil {
		return err
	}

	return nil
}

// Gamma
func (g *Gamma) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Gamma")

	var err error

	if err = g.GammaK.Decode(d); err != nil {
		return err
	}

	if err = g.GammaZ.Decode(d); err != nil {
		return err
	}

	if err = g.GammaS.Decode(d); err != nil {
		return err
	}

	if err = g.GammaA.Decode(d); err != nil {
		return err
	}

	return nil
}

// LookupMetaMapkey
func (d *LookupMetaMapkey) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding LookupMetaMapkey")

	var err error

	if err = d.Hash.Decode(decoder); err != nil {
		return err
	}

	if err = d.Length.Decode(decoder); err != nil {
		return err
	}

	return nil
}

// LookupMetaMapEntry
func (l *LookupMetaMapEntry) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding LookupMetaMapEntry")

	// Decode the size of the map
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Init the map
	*l = make(LookupMetaMapEntry, length)
	for i := uint64(0); i < length; i++ {
		var key LookupMetaMapkey
		if err = key.Decode(d); err != nil {
			return err
		}

		timeSlotSetSize, err := d.DecodeLength()
		if err != nil {
			return err
		}

		if timeSlotSetSize == 0 {
			(*l)[key] = nil
		} else {
			// make the slice with timeSlotSetSize
			val := make([]TimeSlot, timeSlotSetSize)
			for i := uint64(0); i < timeSlotSetSize; i++ {
				if err = val[i].Decode(d); err != nil {
					return err
				}
			}

			(*l)[key] = val
		}
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding PreimagesMapEntry")

	// Decode the size of the map
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Init the map
	*p = make(PreimagesMapEntry, length)

	for i := uint64(0); i < length; i++ {
		var key OpaqueHash
		if err = key.Decode(d); err != nil {
			return err
		}

		var val ByteSequence
		if err = val.Decode(d); err != nil {
			return err
		}

		(*p)[key] = val
	}

	return nil
}

// Storage
func (s *Storage) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Storage")

	// Decode the size of the map
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Init the map
	*s = make(Storage, length)
	for i := uint64(0); i < length; i++ {
		// Decode the of the key
		// INFO: we want to read the vectors from jamtestnet, so we follow the same
		// pattern as in the jamtestnet. They put the length of the key before the
		// key
		length, err := d.DecodeLength()
		if err != nil {
			return err
		}

		if length == 0 {
			return nil
		}

		var key OpaqueHash
		if err = key.Decode(d); err != nil {
			return err
		}

		var val ByteSequence
		if err = val.Decode(d); err != nil {
			return err
		}

		(*s)[key] = val
	}

	return nil
}

// ServiceAccount
func (s *ServiceAccount) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ServiceAccount")

	if err := s.ServiceInfo.Decode(d); err != nil {
		return err
	}

	if err := s.PreimageLookup.Decode(d); err != nil {
		return err
	}

	if err := s.LookupDict.Decode(d); err != nil {
		return err
	}

	if err := s.StorageDict.Decode(d); err != nil {
		return err
	}

	return nil
}

// Accounts (Delta)
func (a *ServiceAccountState) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding Accounts")

	// Encode the size of the map
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// Init the map
	*a = make(ServiceAccountState, length)

	for i := uint64(0); i < length; i++ {
		// Decode key (ServiceId)
		var key ServiceId
		if err = key.Decode(d); err != nil {
			return err
		}

		// Decode value (ServiceAccount)
		var value ServiceAccount
		if err = value.Decode(d); err != nil {
			return err
		}

		(*a)[key] = value
	}

	return nil
}

// // AccumulatedHistory
// func (a *AccumulatedHistory) Decode(d *Decoder) error {
// 	cLog(Cyan, "Decoding AccumulatedHistory")

// 	length, err := d.DecodeLength()
// 	if err != nil {
// 		return err
// 	}

// 	if length == 0 {
// 		return nil
// 	}

// 	// make the slice with length
// 	history := make([]WorkPackageHash, length)
// 	for i := uint64(0); i < length; i++ {
// 		if err = history[i].Decode(d); err != nil {
// 			return err
// 		}
// 	}

// 	*a = history

// 	return nil
// }

// // AccumulatedHistories (Xi)
// func (a *AccumulatedHistories) Decode(d *Decoder) error {
// 	cLog(Cyan, "Decoding AccumulatedHistories")

// 	// make the slice with epoch length
// 	histories := make([]AccumulatedHistory, EpochLength)
// 	for i := 0; i < EpochLength; i++ {
// 		if err := histories[i].Decode(d); err != nil {
// 			return err
// 		}
// 	}

// 	*a = histories

// 	return nil
// }

// State
func (s *State) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding State")

	var err error

	if err = s.Alpha.Decode(d); err != nil {
		return err
	}

	if err = s.Varphi.Decode(d); err != nil {
		return err
	}

	if err = s.Beta.Decode(d); err != nil {
		return err
	}

	if err = s.Gamma.Decode(d); err != nil {
		return err
	}

	if err = s.Psi.Decode(d); err != nil {
		return err
	}

	if err = s.Eta.Decode(d); err != nil {
		return err
	}

	if err = s.Iota.Decode(d); err != nil {
		return err
	}

	if err = s.Kappa.Decode(d); err != nil {
		return err
	}

	if err = s.Lambda.Decode(d); err != nil {
		return err
	}

	if err = s.Rho.Decode(d); err != nil {
		return err
	}

	if err = s.Tau.Decode(d); err != nil {
		return err
	}

	if err = s.Chi.Decode(d); err != nil {
		return err
	}

	if err = s.Pi.Decode(d); err != nil {
		return err
	}

	if err = s.Theta.Decode(d); err != nil {
		return err
	}

	if err = s.Xi.Decode(d); err != nil {
		return err
	}

	if err = s.Delta.Decode(d); err != nil {
		return err
	}

	return nil
}

// deferredTransfer
func (d *DeferredTransfer) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding DeferredTransfer")

	var err error

	if err = d.SenderID.Decode(decoder); err != nil {
		return err
	}

	if err = d.ReceiverID.Decode(decoder); err != nil {
		return err
	}

	if err = d.Balance.Decode(decoder); err != nil {
		return err
	}

	// Read 128 bytes
	if err = binary.Read(decoder.buf, binary.LittleEndian, &d.Memo); err != nil {
		return err
	}

	if err = d.GasLimit.Decode(decoder); err != nil {
		return err
	}

	return nil
}

func (d *DeferredTransfers) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding DeferredTransfers")

	length, err := decoder.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make the slice with length
	transfers := make([]DeferredTransfer, length)
	for i := uint64(0); i < length; i++ {
		if err = transfers[i].Decode(decoder); err != nil {
			return err
		}
	}

	*d = transfers

	return nil
}

func (o *Operand) Decode(decoder *Decoder) error {
	cLog(Cyan, "Decoding Operand")

	var err error

	if err = o.Hash.Decode(decoder); err != nil {
		return err
	}

	if err = o.ExportsRoot.Decode(decoder); err != nil {
		return err
	}

	if err = o.AuthorizerHash.Decode(decoder); err != nil {
		return err
	}

	if err = o.AuthOutput.Decode(decoder); err != nil {
		return err
	}

	if err = o.PayloadHash.Decode(decoder); err != nil {
		return err
	}

	if err = o.GasLimit.Decode(decoder); err != nil {
		return err
	}

	if err = o.Result.Decode(decoder); err != nil {
		return err
	}

	return nil
}

func (e *ExtrinsicData) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ExtrinsicData")
	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	data := make([]byte, length)
	if _, err := d.buf.Read(data); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("ExtrinsicData: %x", data))
	*e = data
	return nil
}

func (l *ExtrinsicDataList) Decode(d *Decoder) error {
	length, err := d.DecodeLength()
	if err != nil {
		return err
	}
	result := make([]ExtrinsicData, length)
	for i := range result {
		var ed ExtrinsicData
		if err := ed.Decode(d); err != nil {
			return err
		}
		result[i] = ed
	}
	*l = result
	return nil
}

func (s *ExportSegment) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding ExportSegment")
	var val ExportSegment
	if err := binary.Read(d.buf, binary.LittleEndian, &val); err != nil {
		return err
	}
	cLog(Yellow, fmt.Sprintf("OpaqueHash: %x", val))

	*s = val
	return nil
}

func (m *ExportSegmentMatrix) Decode(d *Decoder) error {
	outerLen, err := d.DecodeLength()
	if err != nil {
		return err
	}
	result := make(ExportSegmentMatrix, outerLen)
	for i := range result {
		innerLen, err := d.DecodeLength()
		if err != nil {
			return err
		}
		row := make([]ExportSegment, innerLen)
		for j := range row {
			var seg ExportSegment
			if err := seg.Decode(d); err != nil {
				return err
			}
			row[j] = seg
		}
		result[i] = row
	}
	*m = result
	return nil
}

func (m *OpaqueHashMatrix) Decode(d *Decoder) error {
	outerLen, err := d.DecodeLength()
	if err != nil {
		return err
	}
	result := make(OpaqueHashMatrix, outerLen)
	for i := range result {
		innerLen, err := d.DecodeLength()
		if err != nil {
			return err
		}
		row := make([]OpaqueHash, innerLen)
		for j := range row {
			var op OpaqueHash
			if err := op.Decode(d); err != nil {
				return err
			}
			row[j] = op
		}
		result[i] = row
	}
	*m = result
	return nil
}

func (b *WorkPackageBundle) Decode(d *Decoder) error {
	cLog(Cyan, "Decoding WorkPackageBundle")
	var err error

	if err = b.Package.Decode(d); err != nil {
		return err
	}
	if err = b.Extrinsics.Decode(d); err != nil {
		return err
	}
	if err = b.ImportSegments.Decode(d); err != nil {
		return err
	}
	if err = b.ImportProofs.Decode(d); err != nil {
		return err
	}

	return nil
}
