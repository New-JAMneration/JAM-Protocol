package auditing

import (
	"crypto/ed25519"
	"fmt"
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/header"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// (17.1) Q ∈ ⟦W?⟧C(17.1)
// (17.2) Q ≡ ρ[c]w if ρ[c]w ∈ W
//
//	∅ otherwise 		c <− NC
//
// Construct Q: A list of length NC, one report per core if assigned
func GetQ() []*types.WorkReport {
	store := store.GetInstance()
	rho := store.GetPriorStates().GetState().Rho
	availableReports := store.GetAvailableWorkReportsPointer().GetAvailableWorkReports()

	// Create a lookup table of hashes for fast check
	available := make(map[types.WorkPackageHash]bool)
	for _, report := range availableReports {
		available[report.PackageSpec.Hash] = true
	}

	Q := make([]*types.WorkReport, types.CoresCount)
	for index, assignment := range rho {
		if assignment != nil {
			report := assignment.Report
			if available[report.PackageSpec.Hash] {
				Q[index] = &report
			}
		}
	}
	return Q
}

// (17.3) s0 ∈ F[]  ⟨XU ⌢ Y(Hv)⟩
//
//	κ[v]b
//
// (17.4) U = $jam_audit
// Generate s₀ from jam_audit and block author entropy using this validator's key
func GetS0(validatorIndex int) (types.BandersnatchVrfSignature, error) {
	store_instance := store.GetInstance()
	posteriorState := store_instance.GetPriorStates()
	header := store_instance.GetIntermediateHeader()

	// Get block author's entropy hash Y(Hᵥ)
	authorKey := posteriorState.GetKappa()[header.AuthorIndex].Bandersnatch
	authorHandler, err := safrole.CreateVRFHandler(authorKey)
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to create VRF handler for authorKey: %w", err)
	}
	yHv, err := authorHandler.VRFIetfOutput(header.EntropySource[:])
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to compute VRF output for author: %w", err)
	}

	// Combine context: Xᵁ = $jam_audit ⌢ Y(Hᵥ)
	context := append(types.ByteSequence(types.JamAudit[:]), yHv...)

	// s₀ = VRF⟨context⟩ using own key
	validatorKey := posteriorState.GetKappa()[validatorIndex].Bandersnatch
	validatorHandler, err := safrole.CreateVRFHandler(validatorKey)
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to create VRF handler for validatorKey: %w", err)
	}
	signature, err := validatorHandler.RingSign(context, []byte(""))
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to sign context: %w", err)
	}
	s0, err := validatorHandler.VRFIetfOutput(signature[:])
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to compute VRF output for validator: %w", err)
	}

	return types.BandersnatchVrfSignature(s0), nil
}

// (17.5) a0 = {(c, w) | (c, w) ∈ p⋅⋅⋅+10, w ≠ ∅}
// (17.6) where p = F([(c, Qc) | c c <− NC], r)
// (17.7) and r = Y(s0)
//
//	Generate initial audit assignment set a0 for validator v
func ComputeA0ForValidator(Q []*types.WorkReport, validatorIndex int) (output []types.AuditReport) { // TODO: get s0 from local data?
	// (17.3) get s0
	store_instance := store.GetInstance()
	s0, err := GetS0(validatorIndex)
	if err != nil {
		log.Fatalf("GetS0 failed: %v", err)
	}

	// (17.7) compute r = 𝒴(s₀)
	validatorKey := store_instance.GetPosteriorStates().GetKappa()[validatorIndex].Bandersnatch
	handler, err := safrole.CreateVRFHandler(validatorKey)
	if err != nil {
		log.Fatalf("failed to create VRF handler for validatorKey: %v", err)
	}
	r, err := handler.VRFIetfOutput(s0[:]) // (17.7)
	if err != nil {
		log.Fatalf("failed to compute VRF output for validator: %v", err)
	}

	// (17.6) p = F([...], r): shuffle all cores
	shuffle_array := make([]types.U32, types.CoresCount)
	for i := types.U32(0); i < types.U32(types.CoresCount); i++ {
		shuffle_array[i] = i
	}
	p := shuffle.Shuffle(shuffle_array, types.OpaqueHash(r))
	// (17.5) a0 = top 10 shuffled (c, w) where w ≠ ∅
	for _, idx := range p {
		if Q[idx] != nil {
			output = append(output, types.AuditReport{
				CoreID: types.CoreIndex(idx),
				Report: *Q[idx],
			})
			if len(output) == 10 {
				break
			}
		}
	}
	return output
}

// (17.8) let n = (T − P ⋅ Ht) / A
// - T  = current wall-clock time (seconds)
// - P  = SlotPeriod (e.g. 6s, constant)
// - Ht = slot number from block header
// - A  = TranchePeriod (e.g. 8s, constant)
// GetTranchIndex computes tranche index from wall-clock time and block slot
func GetTranchIndex() types.U64 {
	T := header.GetCurrentTimeInSecond()
	H_t := store.GetInstance().GetIntermediateHeader().Slot
	n := (types.U64(T) - types.U64(types.SlotPeriod)*types.U64(H_t)) / types.TranchePeriod
	return n
}

// (17.9) S ≡ Eκ[v]e ⟨XI + n ⌢ xn ⌢ H(H)⟩
// (17.10) where xn = E([E2(c) ⌢ H(w) S(c, w) ∈ an])
// (17.11) XI = $jam_announce

// BuildAnnouncement generates signature S over assigned audit reports in tranche n
func BuildAnnouncement(
	n types.U32,
	reports []types.AuditReport,
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validatorIndex int,
) types.Ed25519Signature {

	var xnPayload types.ByteSequence

	// (17.10) Build xn = hash(coreID || H(w)) for each (c, w) ∈ aₙ
	for _, value := range reports {
		xnPayload = append(xnPayload, utilities.SerializeFixedLength(types.U64(value.CoreID), 2)...) // E2(c)
		H_w := hashFunc(utilities.WorkReportSerialization(value.Report))                             // H(w)
		xnPayload = append(xnPayload, H_w[:]...)                                                     // concat
	}

	xn := utilities.SerializeByteSequence(xnPayload)

	// (17.11) XI = $jam_announce
	// Context = ⟨XI ⌢ n ⌢ xn ⌢ H(H)⟩
	H := store.GetInstance().GetIntermediateHeader()
	context := types.ByteSequence(types.JamAnnounce[:])                // XI
	context = append(context, utilities.SerializeFixedLength(n, 4)...) // n
	context = append(context, xn...)                                   // xn
	headerBytes := utilities.HeaderSerialization(H)
	headerHash := hashFunc(headerBytes)
	context = append(context, headerHash[:]...) // H(H)

	// (17.9) S = Sign⟨context⟩ using Ed25519
	ed25519Pub := store.GetInstance().GetPriorStates().GetKappa()[validatorIndex].Ed25519
	signature := ed25519.Sign(ed25519Pub[:], context)
	return types.Ed25519Signature(signature)
}

// GetAssignedValidators returns the list of validator indices assigned to audit a given work-report.
// (17.12) An ∶ W → ℘⟨NV⟩
func GetAssignedValidators(w types.WorkReport, assignmentMap types.AssignmentMap) []types.ValidatorIndex {
	if validators, ok := assignmentMap[w.PackageSpec.Hash]; ok {
		return validators
	}
	// key doesn't exist return ∅
	return []types.ValidatorIndex{}
}

// ClassifyJudgments categorizes the validators who gave positive (J⊤) or negative (J⊥) judgments for a given work report.
// It uses the work report's hash to determine matching judgments.
func ClassifyJudgments(
	report types.WorkReport,
	judgments []types.AuditReport, // All validators' audit judgments
) (positives map[types.ValidatorIndex]bool, negatives map[types.ValidatorIndex]bool) {
	positives = make(map[types.ValidatorIndex]bool)
	negatives = make(map[types.ValidatorIndex]bool)
	reportHash := report.PackageSpec.Hash

	for _, j := range judgments {
		if j.Report.PackageSpec.Hash == reportHash {
			if j.AuditResult {
				positives[j.ValidatorID] = true // J⊤: Positive judgments
			} else {
				negatives[j.ValidatorID] = true // J⊥: Negative judgments
			}
		}
	}
	return
}

// ∀n > 0 ∶
// (17.15) sn(w) ∈ F[] κ[v]b ⟨XU ⌢ Y(Hv ) ⌢ H(w) n⟩
// (17.16) an ≡ { V/256F Y(sn(w))0 < mn | w ∈ Q, w ≠ ∅}
// where mn = SAn−1(w) ∖ J⊺(w)S
/*
Start
  │
  │
  ├──► [Loop each WorkReport `w` in Q]
  │       │
  │       ├──► Is w == ∅ ? ──► Yes → Skip
  │       │                        ↓
  │       └──► Compute mₙ = |Aₙ₋₁(w) ∖ J⊤(w)|
  │               │
  │               ├──► Build VRF seed: <Xᵁ, 𝒴(Hᵥ), ℋ(w), n>
  │               ├──► sₙ(w) ← VRF(seed)
  │               ├──► y ← 𝒴(sₙ(w))₀
  │               │
  │               └──► If y < mₙ ? ──► Yes → Add (c, w) to aₙ
  │                                        ↓
  │                                     No → skip
  │
  ▼
End (aₙ ready)
*/

func ComputeAnForValidator(
	trancheIndex types.U64,
	workReportPool []*types.WorkReport,
	priorAssignments map[types.WorkPackageHash][]types.ValidatorIndex, // Aₙ₋₁(w)
	positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool, // J⊤(w)
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validatorIndex int,
) []types.AuditReport {
	var auditAssignments []types.AuditReport
	storeInstance := store.GetInstance()
	posteriorState := storeInstance.GetPosteriorStates()
	header := storeInstance.GetIntermediateHeader()

	// Author info for Y(Hᵥ)
	authorPublicKey := posteriorState.GetKappa()[header.AuthorIndex].Bandersnatch
	authorVRFHandler, err := safrole.CreateVRFHandler(authorPublicKey)
	if err != nil {
		log.Fatalf("failed to create VRF handler for author key: %v", err)
	}
	authorVRFOutput, err := authorVRFHandler.VRFIetfOutput(header.EntropySource[:])
	if err != nil {
		log.Fatalf("failed to compute VRF output for author: %v", err)
	}

	for _, workPtr := range workReportPool {
		if workPtr == nil {
			continue
		}
		report := *workPtr
		reportHash := report.PackageSpec.Hash
		assignedValidators := priorAssignments[reportHash]
		positiveJudgedMap := positiveJudgers[reportHash]

		// Count validators who have not submitted judgment
		noShowCount := 0
		for _, v := range assignedValidators {
			if !positiveJudgedMap[v] {
				noShowCount++
			}
		}
		if noShowCount == 0 {
			continue
		}

		// Build VRF context: $jam_audit ⌢ Y(Hv) ⌢ H(w) ⌢ n
		context := append(types.ByteSequence(types.JamAudit[:]), authorVRFOutput...)
		reportHashBytes := hashFunc(utilities.WorkReportSerialization(report))
		context = append(context, reportHashBytes[:]...)
		context = append(context, utilities.SerializeFixedLength(trancheIndex, 4)...)

		// Generate sₙ(w)
		validatorKey := posteriorState.GetKappa()[validatorIndex].Bandersnatch
		validatorVRFHandler, err := safrole.CreateVRFHandler(validatorKey)
		if err != nil {
			log.Fatalf("failed to create VRF handler for validator key: %v", err)
		}
		signature, err := validatorVRFHandler.RingSign(context, []byte(""))
		if err != nil {
			log.Fatalf("failed to sign context: %v", err)
		}
		snOutput, err := validatorVRFHandler.VRFIetfOutput(signature[:])
		if err != nil {
			log.Fatalf("failed to compute sn(w) VRF output: %v", err)
		}
		signature, err = validatorVRFHandler.RingSign(snOutput, []byte(""))
		if err != nil {
			log.Fatalf("failed to sign context: %v", err)
		}
		// Compute scaled guess: V / (256F)
		finalOutput, err := authorVRFHandler.VRFIetfOutput(signature[:])
		if err != nil {
			log.Fatalf("failed to compute final VRF output: %v", err)
		}
		vrfScaledGuess := int(finalOutput[0]) * types.ValidatorsCount / (256 * types.BiasFactor)

		if vrfScaledGuess < noShowCount {
			auditAssignments = append(auditAssignments, types.AuditReport{
				CoreID:      report.CoreIndex,
				Report:      report,
				ValidatorID: types.ValidatorIndex(validatorIndex),
				AuditResult: false,
			})
		}
	}
	return auditAssignments
}

/*
// e_n(w): 對應於公式 17.17 的實作
func EvaluateReport(
	report types.WorkReport,
	coreID types.CoreIndex,
	proposals []types.WorkPackageProposal, // proposal from peers
	decodeFunc func(types.WorkPackageProposal, types.CoreIndex) (types.WorkReport, bool),
	encodeFunc func(types.WorkReport) types.WorkPackageEncoding,
) *types.WorkReport {
	targetEncoding := encodeFunc(report)

	for _, p := range proposals {
		// check if the decoded report matches the given one
		if decoded, ok := decodeFunc(p, coreID); ok {
			if encodeFunc(decoded) == targetEncoding {
				return &decoded // match found: return the decoded report
			}
		}
	}
	return nil // ⊥ — evaluation failed
}*/

// (17.18) n = {Sκ[v]e (Xe(w) ⌢ H(w)) S (c, w) ∈ an}
func BuildJudgements(
	tranche types.U64,
	reports []types.AuditReport, // (c, w) ∈ aₙ
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validator_index int,
) []types.AuditReport {
	for index, item := range reports {
		report := item.Report

		// Determine context string
		var context types.ByteSequence
		if item.AuditResult {
			context = []byte("$jam_valid")
		} else {
			context = []byte("$jam_invalid")
		}

		// Hash the report content
		hashW := hashFunc(utilities.WorkReportSerialization(report)) // 𝓗(w)
		context = append(context, hashW[:]...)                       // X_e(w) ⌢ 𝓗(w)

		// Sign the message
		ed25519_public := store.GetInstance().GetPosteriorStates().GetKappa()[validator_index].Ed25519
		signature := ed25519.Sign(ed25519_public[:], context)
		reports[index].Signature = types.Ed25519Signature(signature)
	}

	return reports
}

// (17.19) U(w)⇔∨J⊥(w)=∅∧∃n:An​(w)⊆J⊤(w)∣J⊤(w)∣>(2/3)V

func IsWorkReportAudited(
	report types.WorkReport,
	judgments []types.AuditReport,
	assignments []types.ValidatorIndex, // An(w)
) bool {
	// collect judgement for this report
	positives, negatives := ClassifyJudgments(report, judgments)

	// Rule 1: J⊥(w) == ∅ ∧ ∃n : Aₙ(w) ⊆ J⊤(w)
	if len(negatives) == 0 {
		allConfirmed := true
		for _, v := range assignments {
			if !positives[v] {
				allConfirmed = false
				break
			}
		}
		if allConfirmed {
			return true
		}
	}

	// Rule 2: |J⊤(w)| > 2/3 V
	if len(positives) >= types.ValidatorsSuperMajority {
		return true
	}

	// Otherwise: Not audited
	return false
}

func FilterJudgements(judgements []types.AuditReport, targetHash types.WorkPackageHash) []types.AuditReport {
	var out []types.AuditReport
	for _, j := range judgements {
		if j.Report.PackageSpec.Hash == targetHash {
			out = append(out, j)
		}
	}
	return out
}

// (17.20) U ⇔ ∀w ∈ W ∶ U (w)
func IsBlockAudited(
	reports []types.WorkReport,
	allJudgements []types.AuditReport,
	assignmentMap map[types.WorkPackageHash][]types.ValidatorIndex,
) bool {
	for _, report := range reports {
		hash := report.PackageSpec.Hash
		assignments := GetAssignedValidators(report, assignmentMap)
		judgements := FilterJudgements(allJudgements, hash)

		if !IsWorkReportAudited(report, judgements, assignments) {
			return false // All report should be audited
		}
	}
	return true
}

func ProcessAuditing(validatorIndex int) error {
	// (17.1)~(17.2): Get Q = one WorkReport per core (if exists in W)
	Q := GetQ()

	// Collect assignment map A_n-1(w)
	assignmentMap := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	judgementPool := []types.AuditReport{}

	// J⊤(w): positiveJudgers
	positiveJudgers := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)

	// (17.3)~(17.7): Compute initial assignment a₀
	a0 := ComputeA0ForValidator(Q, validatorIndex)

	// Store a₀ assignment for J⊤ use
	for _, item := range a0 {
		hash := item.Report.PackageSpec.Hash
		assignmentMap[hash] = append(assignmentMap[hash], types.ValidatorIndex(validatorIndex))
	}

	// (17.8): Compute tranche index n
	n := GetTranchIndex()

	// (17.15)~(17.16): Compute aₙ
	aN := ComputeAnForValidator(
		n,
		Q,
		assignmentMap,
		positiveJudgers,
		hash.Blake2bHash,
		validatorIndex,
	)

	// Merge aₙ into J⊤ tracking map (if result is positive, will be signed next)
	for _, a := range aN {
		if a.AuditResult {
			hash := a.Report.PackageSpec.Hash
			if _, ok := positiveJudgers[hash]; !ok {
				positiveJudgers[hash] = make(map[types.ValidatorIndex]bool)
			}
			positiveJudgers[hash][a.ValidatorID] = true
		}
	}

	// (17.18): Sign judgement and attach signature
	signed := BuildJudgements(n, aN, hash.Blake2bHash, validatorIndex)

	// Collect signed audit reports
	judgementPool = append(judgementPool, signed...)

	// Optional: validate per report (17.19) and per block (17.20)
	for _, report := range Q {
		if report != nil {
			hash := report.PackageSpec.Hash
			assignments := assignmentMap[hash]
			judgements := FilterJudgements(judgementPool, hash)

			if !IsWorkReportAudited(*report, judgements, assignments) {
				fmt.Printf("Report not audited: Core %d, Hash %x\n", report.CoreIndex, hash)
			}
		}
	}

	return nil
}
