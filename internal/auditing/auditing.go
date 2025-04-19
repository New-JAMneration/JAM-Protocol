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
	header := store.GetInstance().GetIntermediateHeader()              // H
	context := types.ByteSequence(types.JamAnnounce[:])                // XI
	context = append(context, utilities.SerializeFixedLength(n, 4)...) // n
	context = append(context, xn...)                                   // xn
	headerBytes := utilities.HeaderSerialization(header)
	headerHash := hashFunc(headerBytes) // H(H)
	context = append(context, headerHash[:]...)

	// (17.9) S = Sign⟨context⟩ using Ed25519
	validator_key := store.GetInstance().GetPriorStates().GetKappa()[validatorIndex].Ed25519
	signature := ed25519.Sign(validator_key[:], context)
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

// (17.15) sn(w) ∈ F[] κ[v]b ⟨XU ⌢ Y(Hv ) ⌢ H(w) n⟩
// (17.16) an ≡ { V/256F Y(sn(w))0 < mn | w ∈ Q, w ≠ ∅}
// where mn = SAn−1(w) ∖ J⊺(w)S

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
// e_n(w): 17.17 產生 audit 結果
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
	auditReports []types.AuditReport, // (c, w) ∈ aₙ
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validator_index int,
) []types.AuditReport {
	for index, audit := range auditReports {
		report := audit.Report
		// Xe
		var context types.ByteSequence
		if audit.AuditResult {
			context = []byte("$jam_valid")
		} else {
			context = []byte("$jam_invalid")
		}

		// Hash the report content
		hashW := hashFunc(utilities.WorkReportSerialization(report)) // H(w)
		context = append(context, hashW[:]...)                       // Xe(w) ⌢ H(w)

		// Sign the message
		validator_key := store.GetInstance().GetPosteriorStates().GetKappa()[validator_index].Ed25519
		signature := ed25519.Sign(validator_key[:], context)
		auditReports[index].Signature = types.Ed25519Signature(signature)
	}

	return auditReports
}

// (17.19) Determines if a single work report is considered audited.
func IsWorkReportAudited(
	report types.WorkReport,
	judgments []types.AuditReport,
	assignedValidators []types.ValidatorIndex, // Aₙ(w)
) bool {
	positiveJudges, negativeJudges := ClassifyJudgments(report, judgments)

	// Rule 1: No negatives AND all assigned validators gave a positive judgment
	if len(negativeJudges) == 0 {
		allAssignedConfirmed := true
		for _, validator := range assignedValidators {
			if !positiveJudges[validator] {
				allAssignedConfirmed = false
				break
			}
		}
		if allAssignedConfirmed {
			return true
		}
	}

	// Rule 2: Supermajority of positive judgments
	if len(positiveJudges) >= types.ValidatorsSuperMajority {
		return true
	}

	return false
}

// Filters audit judgments that match the target report hash.
func FilterJudgments(judgments []types.AuditReport, targetHash types.WorkPackageHash) []types.AuditReport {
	var filtered []types.AuditReport
	for _, j := range judgments {
		if j.Report.PackageSpec.Hash == targetHash {
			filtered = append(filtered, j)
		}
	}
	return filtered
}

// (17.20) Checks if ALL work reports are fully audited.
func IsBlockAudited(
	workReports []types.WorkReport,
	allJudgments []types.AuditReport,
	assignmentMap map[types.WorkPackageHash][]types.ValidatorIndex,
) bool {
	for _, report := range workReports {
		hash := report.PackageSpec.Hash
		assignedValidators := GetAssignedValidators(report, assignmentMap)
		filteredJudgments := FilterJudgments(allJudgments, hash)

		if !IsWorkReportAudited(report, filteredJudgments, assignedValidators) {
			return false // One report not audited: block incomplete
		}
	}
	return true
}

func PublishAuditReport(audit types.AuditReport) error {
	// TODO: Implement the logic to publish the audit report
	// This could involve sending the report to a network, saving it to a database, etc.
	return nil
}

func SingleNodeAuditingAndPublish(validatorIndex int) error {
	// Get reports to audit
	Q := GetQ()

	// Collect audit assignment for local validator
	assignmentMap := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	positiveJudgers := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)

	// a₀: Initial deterministic assignment
	a0 := ComputeA0ForValidator(Q, validatorIndex)
	for _, item := range a0 {
		hash := item.Report.PackageSpec.Hash
		assignmentMap[hash] = append(assignmentMap[hash], types.ValidatorIndex(validatorIndex))
	}

	// Compute tranche index
	tranche := GetTranchIndex()

	// aₙ: Compute stochastic audit assignments (based on no-show)
	aN := ComputeAnForValidator(
		tranche,
		Q,
		assignmentMap,
		positiveJudgers,
		hash.Blake2bHash,
		validatorIndex,
	)

	// Update positiveJudgers map
	for _, a := range aN {
		if a.AuditResult {
			hash := a.Report.PackageSpec.Hash
			if _, ok := positiveJudgers[hash]; !ok {
				positiveJudgers[hash] = make(map[types.ValidatorIndex]bool)
			}
			positiveJudgers[hash][a.ValidatorID] = true
		}
	}

	// Sign judgements
	signed := BuildJudgements(tranche, aN, hash.Blake2bHash, validatorIndex)

	// Publish to audit pool
	for _, audit := range signed {
		err := PublishAuditReport(audit)
		if err != nil {
			return fmt.Errorf("failed to publish audit report for core %d: %w", audit.CoreID, err)
		}
	}

	return nil
}
