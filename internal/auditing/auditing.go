package auditing

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/header"
	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
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
// CollectAuditReportCandidates constructs the audit report candidates Q (formula 17.1 ~ 17.2).
func CollectAuditReportCandidates() []*types.WorkReport {
	cs := blockchain.GetInstance()

	// ρ(rho): Current assignment map (per core)
	rho := cs.GetPriorStates().GetRho()

	// W: Available work reports
	W := cs.GetIntermediateStates().GetAvailableWorkReports()

	// Create a set of available work package hashes
	available := make(map[types.WorkPackageHash]bool)
	for _, report := range W {
		available[report.PackageSpec.Hash] = true
	}

	Q := make([]*types.WorkReport, types.CoresCount)
	for index, assignment := range rho {
		// check if core has an assignment
		if assignment != nil {
			report := assignment.Report
			// Only keep it if the assigned report is still available in W (ρ[c]w ∈ W)
			if available[report.PackageSpec.Hash] {
				Q[index] = &report
			}
			// else: Q[index] stays nil (∅ otherwise c <− NC)
		}
	}
	return Q
}

// GenerateValidatorAuditSeed computes the initial audit seed s0 for a validator, following Formula (17.3)-(17.4).
// Returns the VRF output (s0) as BandersnatchVrfSignature.
func GenerateValidatorAuditSeed(validatorIndex types.ValidatorIndex) (types.BandersnatchVrfSignature, error) {
	cs := blockchain.GetInstance()
	priorStates := cs.GetPriorStates()

	entropyHash, err := ComputeAuthorEntropyVrfOutput() // Y(Hᵥ): VRF output of block author's entropy
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to get Y(Hᵥ): %w", err)
	}

	// Construct context XU ⌢ Y(Hᵥ)
	context := append(types.ByteSequence(types.JamAudit[:]), entropyHash...)

	// Sign the context with validator's key and extract VRF output
	validatorKey := priorStates.GetKappa()[validatorIndex].Bandersnatch
	validatorVRF, err := safrole.CreateVRFHandler(validatorKey)
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to create validator VRF handler: %w", err)
	}

	// Sign the context (empty message)
	vrfSignature, err := validatorVRF.IETFSign(context, []byte(""))
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to sign audit context: %w", err)
	}

	// Derive VRF output from signature
	s0, err := validatorVRF.VRFIetfOutput(vrfSignature[:])
	if err != nil {
		return types.BandersnatchVrfSignature{}, fmt.Errorf("failed to compute VRF output: %w", err)
	}

	return types.BandersnatchVrfSignature(s0), nil
}

// ComputeA0ForValidator generates the initial audit assignment a0 for a given validator,
// based on formulas (17.3)~(17.7):
//
//	(17.3) s0 = VRF⟨XU ⌢ Y(Hᵥ)⟩ using validator key
//	(17.7) r = 𝒴(s0)    → VRF output over s₀
//	(17.6) p = Shuffle([0..CoresCount), r)
//	(17.5) a0 = top 10 of (c, Q[c]) where Q[c] ≠ ∅
//
// Returns a list of AuditReports
func ComputeInitialAuditAssignment(Q []*types.WorkReport, validatorIndex types.ValidatorIndex) ([]types.AuditReport, error) {
	cs := blockchain.GetInstance()

	// Get initial audit seed s0 (17.3)
	s0, err := GenerateValidatorAuditSeed(validatorIndex)
	if err != nil {
		return nil, fmt.Errorf("ComputeA0ForValidator: failed to get s0: %w", err)
	}

	// Compute r = 𝒴(s0) — derive audit random seed (17.7)
	validatorKey := cs.GetPriorStates().GetKappa()[validatorIndex].Bandersnatch
	handler, err := safrole.CreateVRFHandler(validatorKey)
	if err != nil {
		return nil, fmt.Errorf("ComputeA0ForValidator: failed to create VRF handler for validator: %w", err)
	}

	vrfOutput, err := handler.VRFIetfOutput(s0[:])
	if err != nil {
		return nil, fmt.Errorf("ComputeA0ForValidator: failed to get VRF output from s₀: %w", err)
	}

	// Generate core shuffle p = F([0..N], r) (17.6)
	coreIndices := make([]types.U32, types.CoresCount)
	for i := range coreIndices {
		coreIndices[i] = types.U32(i)
	}
	shuffled := shuffle.Shuffle(coreIndices, types.OpaqueHash(vrfOutput))

	// Step 4: Select top 10 assigned reports (17.5)
	var a0 []types.AuditReport
	for _, coreIdx := range shuffled {
		report := Q[coreIdx]
		if report != nil {
			a0 = append(a0, types.AuditReport{
				CoreID:      types.CoreIndex(coreIdx),
				Report:      *report,
				AuditResult: false,
			})
			if len(a0) == 10 {
				break
			}
		}
	}

	return a0, nil
}

// (17.8) let n = (T − P ⋅ Ht) / A
// GetTranchIndex computes tranche index from wall-clock time and block slot
func GetTranchIndex() types.U64 {
	T := types.U64(header.GetCurrentTimeInSecond())                                 // T current time (seconds)
	Ht := types.U64(blockchain.GetInstance().GetProcessingBlockPointer().GetSlot()) // Ht slot number from block header
	P := types.U64(types.SlotPeriod)                                                // P: seconds per slot
	A := types.U64(types.TranchePeriod)                                             // A: seconds per tranche
	n := (T - P*Ht) / A                                                             // n = (T - P ⋅ Ht) / A
	return n
}

// BuildAnnouncement generates the announcement signature S
// over the validator's audit assignment aₙ at tranche index n,
// following formula:
// S ≡ Eκ[v]e ⟨XI + n ⌢ xn ⌢ H(H)⟩
// This Getfunction will change in future. Just remind
func BuildAnnouncement(
	n types.U8, // tranche index
	an []types.AuditReport, // an: assignment at tranche n
	hashFunc func(types.ByteSequence) types.OpaqueHash, // H(w): hash function
	validatorIndex types.ValidatorIndex,
	validatorPrivKey ed25519.PrivateKey, // κ[v]ᵉ: Ed25519 private key
) (types.Ed25519Signature, error) {
	// (17.10) Compute xn = concat of E([E2(c) ⌢ H(w)] for all (c, w) ∈ an)
	var xnPayload types.ByteSequence
	for _, pair := range an {
		coreID := utilities.SerializeFixedLength(types.U64(pair.CoreID), 2) // E2(c)
		hashW := hashFunc(utilities.WorkReportSerialization(pair.Report))   // H(w)
		xnPayload = append(xnPayload, coreID...)
		xnPayload = append(xnPayload, hashW[:]...)
	}
	xn := utilities.SerializeByteSequence(xnPayload)

	// (17.11) XI = $jam_announce
	XI := types.ByteSequence(types.JamAnnounce[:])

	// Get H(H): hash of the intermediate header
	header := blockchain.GetInstance().GetProcessingBlockPointer().GetHeader()
	serializedHeader, err := utilities.HeaderSerialization(header)
	if err != nil {
		return types.Ed25519Signature{}, err
	}
	headerHash := hashFunc(serializedHeader)

	// (17.9) context = ⟨XI ⌢ n ⌢ xn ⌢ H(H)⟩
	context := XI
	context = append(context, []byte{uint8(n)}...) // ⌢ n
	context = append(context, xn...)               // ⌢ xn
	context = append(context, headerHash[:]...)    // ⌢ H(H)

	// Sign context with validator Ed25519 private key: S = Sign(context)
	signature := ed25519.Sign(validatorPrivKey, context)
	return types.Ed25519Signature(signature), nil
}

// (17.12) GetAssignedValidators returns the set Aₙ(w) of validators assigned to work-report w.
func GetAssignedValidators(
	w types.WorkReport,
	An types.AssignmentMap, // An: assignment map
) []types.ValidatorIndex {
	if assigned, ok := An[w.PackageSpec.Hash]; ok {
		return assigned
	}
	return []types.ValidatorIndex{} // ∅ if not found
}

// (17.13) ∀(c, w) ∈ a0 ∶ v ∈ q0(w)
func UpdateAssignmentMap(
	A0 []types.AuditReport,
	An types.AssignmentMap, // An: assignment map
) types.AssignmentMap {
	for _, audit := range A0 {
		hash := audit.Report.PackageSpec.Hash
		if _, ok := An[hash]; !ok {
			An[hash] = []types.ValidatorIndex{}
		}
		An[hash] = append(An[hash], audit.ValidatorID)
	}
	return An
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

// (17.14) Y(Hᵥ) ∈ F[] κ[v]b ⟨XU ⌢ H(Hv)⟩
// GetYHv computes the VRF output Y(Hᵥ) using the block author's key and the block header's entropy source.
func ComputeAuthorEntropyVrfOutput() ([]byte, error) {
	// Compute Y(Hᵥ) — entropy hashed by block author's key
	cs := blockchain.GetInstance()
	priorStates := cs.GetPriorStates()
	header := cs.GetProcessingBlockPointer().GetHeader()
	authorKey := priorStates.GetKappa()[header.AuthorIndex].Bandersnatch
	authorVRF, err := safrole.CreateVRFHandler(authorKey)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create author VRF handler: %w", err)
	}

	entropyHash, err := authorVRF.VRFIetfOutput(header.EntropySource[:])
	if err != nil {
		return []byte{}, fmt.Errorf("failed to compute Y(Hᵥ): %w", err)
	}
	return entropyHash, nil
}

// (17.15) sn(w) ∈ F[] κ[v]b ⟨XU ⌢ Y(Hv) ⌢ H(w) n⟩
// (17.16) an ≡ { V/256F Y(sn(w))0 < mn | w ∈ Q, w ≠ ∅}
// where mn = SAn−1(w) ∖ J⊺(w)S
func ComputeAnForValidator(
	n types.U8,
	Q []*types.WorkReport,
	priorAssignments map[types.WorkPackageHash][]types.ValidatorIndex, // Aₙ₋₁(w)
	positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool, // J⊤(w)
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validator_index types.ValidatorIndex,
) ([]types.AuditReport, error) {
	var an []types.AuditReport

	priorStates := blockchain.GetInstance().GetPriorStates()

	// Y(Hᵥ): VRF output of block author's entropy
	Y_Hv, err := ComputeAuthorEntropyVrfOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get Y(Hᵥ): %w", err)
	}

	// validator handler
	validatorKey := priorStates.GetKappa()[validator_index].Bandersnatch
	vrfHandler, err := safrole.CreateVRFHandler(validatorKey)
	if err != nil {
		return nil, fmt.Errorf("CreateVRFHandler for validator: %w", err)
	}

	for _, wPtr := range Q {
		if wPtr == nil {
			continue
		}
		report := *wPtr
		reportHash := report.PackageSpec.Hash
		assignedValidators := priorAssignments[reportHash]
		positiveJudgedMap := positiveJudgers[reportHash]

		// mₙ = |Aₙ₋₁(w) ∖ J⊤(w)|
		noShowCount := 0
		for _, vid := range assignedValidators {
			if !positiveJudgedMap[vid] {
				noShowCount++
			}
		}
		if noShowCount == 0 {
			continue
		}

		// Build context ⟨XU ⌢ Y(Hv) ⌢ H(w) ⌢ n⟩
		conext := types.ByteSequence(types.JamAudit[:])                               // XU
		context := append(conext, Y_Hv...)                                            // Y(Hv)
		Hw := hashFunc(utilities.WorkReportSerialization(report))                     // H(w)
		context = append(context, Hw[:]...)                                           // H(w)
		context = append(context, utilities.SerializeFixedLength(types.U32(n), 4)...) // n

		// Compute sₙ(w)
		signature, err := vrfHandler.IETFSign(context, []byte(""))
		if err != nil {
			return nil, fmt.Errorf("signing sₙ(w) failed: %w", err)
		}
		sn_w, err := vrfHandler.VRFIetfOutput(signature[:])
		if err != nil {
			return nil, fmt.Errorf("VRF output Y(sₙ(w)) failed: %w", err)
		}

		//  V * Y(sn(w))0 / 256F < m
		// This report is assigned to the validator
		if int(sn_w[0])*types.ValidatorsCount/(256*types.BiasFactor) < noShowCount {
			an = append(an, types.AuditReport{
				CoreID:      report.CoreIndex,
				Report:      report,
				ValidatorID: types.ValidatorIndex(validator_index),
				AuditResult: false,
			})
		}
	}

	return an, nil
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
	tranche types.U8,
	auditReports []types.AuditReport, // (c, w) ∈ aₙ
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validator_index types.ValidatorIndex,
) []types.AuditReport {
	for index, audit := range auditReports {
		report := audit.Report
		// Xe
		var context types.ByteSequence
		if audit.AuditResult {
			context = []byte(types.JamValid)
		} else {
			context = []byte(types.JamInvalid)
		}

		// Hash the report content
		hashW := hashFunc(utilities.WorkReportSerialization(report)) // H(w)
		context = append(context, hashW[:]...)                       // Xe(w) ⌢ H(w)

		// Sign the message
		validator_key := blockchain.GetInstance().GetPriorStates().GetKappa()[validator_index].Ed25519
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

// NODE-TODO [CE145 send]: Send signed judgments to all validators.
// Ref: feat/jam-np-ce-handler — ce145.go HandleJudgmentAnnouncement_Auditor
func BroadcastAuditReport(audit []types.AuditReport) {
}

// NODE-TODO [CE144 send]: Broadcast audit announcement + VRF evidence to all validators.
// Ref: feat/jam-np-ce-handler — ce144.go HandleAuditAnnouncement_Send
func BroadcastAnnouncement(validatorIndex types.ValidatorIndex, tranche types.U8, assignment map[types.WorkPackageHash][]types.ValidatorIndex, signature types.Ed25519Signature) {
}

// Deprecated: UpdateAssignmentMapFromOtherNode — replaced by SyncAssignmentMapFromBus in audit_bus.go.
// Kept for backward compatibility; delegates to no-op if bus is nil.
func UpdateAssignmentMapFromOtherNode(assignmentMap map[types.WorkPackageHash][]types.ValidatorIndex) map[types.WorkPackageHash][]types.ValidatorIndex {
	return assignmentMap
}

// Deprecated: UpdatePositiveJudgersFromOtherNode — replaced by SyncPositiveJudgersFromBus in audit_bus.go.
func UpdatePositiveJudgersFromOtherNode(positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool) map[types.WorkPackageHash]map[types.ValidatorIndex]bool {
	return positiveJudgers
}

func UpdatePositiveJudgersFromAudit(audits []types.AuditReport, positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool) map[types.WorkPackageHash]map[types.ValidatorIndex]bool {
	for _, audit := range audits {
		if audit.AuditResult {
			h := audit.Report.PackageSpec.Hash
			if _, ok := positiveJudgers[h]; !ok {
				positiveJudgers[h] = make(map[types.ValidatorIndex]bool)
			}
			positiveJudgers[h][audit.ValidatorID] = true
		}
	}
	return positiveJudgers
}

// WaitNextTranche blocks until the deadline for the given tranche elapses, or
// ctx is cancelled (e.g. block fully audited, slot change).
//
// GP §17.7: tranche n occupies [slotStart + n*A, slotStart + (n+1)*A).
// This function waits until slotStart + (tranche+1)*A so the *next* tranche
// can begin with fresh CE144/CE145 data.
func WaitNextTranche(tranche types.U8, slotStartTime time.Time, ctx context.Context) error {
	deadline := slotStartTime.Add(time.Duration(tranche+1) * time.Duration(types.TranchePeriod) * time.Second)
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Deprecated: SyncPositiveJudgersFromOtherNodes — replaced by SyncPositiveJudgersFromBus.
func SyncPositiveJudgersFromOtherNodes(positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool, validatorIndex types.ValidatorIndex, tranche types.U8) map[types.WorkPackageHash]map[types.ValidatorIndex]bool {
	return positiveJudgers
}

// Deprecated: SyncAssignmentMapFromOtherNodes — replaced by SyncAssignmentMapFromBus.
func SyncAssignmentMapFromOtherNodes(
	assignmentMap map[types.WorkPackageHash][]types.ValidatorIndex,
	validatorIndex types.ValidatorIndex, tranche types.U8,
) map[types.WorkPackageHash][]types.ValidatorIndex {
	return assignmentMap
}

// GetJudgement is defined in judgement.go — implements GP §17.16–17.17.
// It fetches the original bundle, re-executes Ξ(p,c), and compares.

// SingleNodeAuditingAndPublish runs the full audit lifecycle for one block.
// It is designed to run as a goroutine — one per block. Pass a cancellable
// ctx so that it exits early when the block is superseded or fully audited
// from the outside.
//
// Flow:
//
//	tranche 0 (deterministic):
//	  ComputeInitialAuditAssignment → announce (CE144) → judge → broadcast (CE145) → sync → check
//	tranche n≥1 (stochastic, loop):
//	  wait 8s → sync CE144/CE145 → ComputeAnForValidator → judge → announce → broadcast → check
func SingleNodeAuditingAndPublish(
	validatorIndex types.ValidatorIndex,
	validatorPrivKey ed25519.PrivateKey,
	slotStartTime time.Time,
	bus *AuditMessageBus,
	ctx context.Context,
) error {
	// ── Step 1: Collect audit candidates Q (GP §17.1–17.2) ──
	Q := CollectAuditReportCandidates()
	var workReports []types.WorkReport
	for _, report := range Q {
		if report != nil {
			workReports = append(workReports, *report)
		}
	}
	if len(workReports) == 0 {
		return nil
	}

	assignmentMap := make(map[types.WorkPackageHash][]types.ValidatorIndex)
	positiveJudgers := make(map[types.WorkPackageHash]map[types.ValidatorIndex]bool)
	var allJudgments []types.AuditReport // accumulated across all tranches

	// ── Tranche 0: deterministic initial assignment (GP §17.3–17.7) ──
	a0, err := ComputeInitialAuditAssignment(Q, validatorIndex)
	if err != nil {
		return fmt.Errorf("failed to compute a₀: %w", err)
	}

	for _, audit := range a0 {
		h := audit.Report.PackageSpec.Hash
		assignmentMap[h] = append(assignmentMap[h], validatorIndex)
	}

	a0Ann, err := BuildAnnouncement(0, a0, hash.Blake2bHash, validatorIndex, validatorPrivKey)
	if err != nil {
		return err
	}
	BroadcastAnnouncement(validatorIndex, 0, assignmentMap, a0Ann)

	for i := range a0 {
		a0[i].AuditResult = GetJudgement(a0[i])
	}
	positiveJudgers = UpdatePositiveJudgersFromAudit(a0, positiveJudgers)

	signed := BuildJudgements(0, a0, hash.Blake2bHash, validatorIndex)
	allJudgments = append(allJudgments, signed...)
	BroadcastAuditReport(signed)

	// Drain any CE messages that arrived during tranche-0 execution.
	assignmentMap = SyncAssignmentMapFromBus(bus, assignmentMap)
	positiveJudgers = SyncPositiveJudgersFromBus(bus, positiveJudgers)

	if IsBlockAudited(workReports, allJudgments, assignmentMap) {
		return nil
	}

	// ── Tranche loop (n ≥ 1): wait → sync → compute → judge → broadcast → check ──
	for tranche := types.U8(1); ; tranche++ {
		// Wait until the current tranche period ends (8s boundary).
		// During this wait CE handlers keep pushing into the bus channels.
		if err := WaitNextTranche(tranche-1, slotStartTime, ctx); err != nil {
			return err
		}

		// Drain CE144/CE145 messages accumulated during the wait.
		assignmentMap = SyncAssignmentMapFromBus(bus, assignmentMap)
		positiveJudgers = SyncPositiveJudgersFromBus(bus, positiveJudgers)

		// Compute stochastic assignment based on latest no-show data.
		an, err := ComputeAnForValidator(tranche, Q, assignmentMap, positiveJudgers, hash.Blake2bHash, validatorIndex)
		if err != nil {
			return fmt.Errorf("failed to compute aₙ at tranche %d: %w", tranche, err)
		}

		for i := range an {
			an[i].AuditResult = GetJudgement(an[i])
		}

		anAnn, err := BuildAnnouncement(tranche, an, hash.Blake2bHash, validatorIndex, validatorPrivKey)
		if err != nil {
			return err
		}
		BroadcastAnnouncement(validatorIndex, tranche, assignmentMap, anAnn)

		positiveJudgers = UpdatePositiveJudgersFromAudit(an, positiveJudgers)
		signedAn := BuildJudgements(tranche, an, hash.Blake2bHash, validatorIndex)
		allJudgments = append(allJudgments, signedAn...)
		BroadcastAuditReport(signedAn)

		if IsBlockAudited(workReports, allJudgments, assignmentMap) {
			return nil
		}
	}
}
