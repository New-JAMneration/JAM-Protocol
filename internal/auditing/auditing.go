package auditing

import (
	"crypto/ed25519"

	"github.com/New-JAMneration/JAM-Protocol/internal/safrole"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/shuffle"
)

// (17.1) Q ∈ ⟦W?⟧C(17.1)
// (17.2) Q ≡ ρ[c]w if ρ[c]w ∈ W
//
//	∅ otherwise 		c <− NC
//
// Construct Q: A list of length NC, one report per core if assigned
func GetQ() []*types.WorkReport {
	rho := store.GetInstance().GetPosteriorStates().GetState().Rho

	// Initialize Q with length = NC, default nil (∅)
	Q := make([]*types.WorkReport, len(rho))

	for index, assignment := range rho {
		// If availability assignment exists for this core, set Q[c] = ρ[c]ʷ
		if assignment != nil {
			Q[index] = &assignment.Report
		}
		// else: leave as nil (∅)
	}
	return Q
}

// (17.3) s0 ∈ F[]  ⟨XU ⌢ Y(Hv)⟩
//
//	κ[v]b
//
// (17.4) U = $jam_audit
// Generate s₀ from jam_audit and block author entropy using this validator's key
func GetS0(validatorIndex int) types.BandersnatchVrfSignature {
	s := store.GetInstance()
	posteriorState := s.GetPosteriorStates()
	header := s.GetIntermediateHeader()

	// Get block author's entropy hash Y(Hᵥ)
	authorKey := posteriorState.GetKappa()[header.AuthorIndex].Bandersnatch
	authorHandler, _ := safrole.CreateVRFHandler(authorKey)
	yHv, _ := authorHandler.VRFIetfOutput(header.EntropySource[:])

	// Combine context: Xᵁ = $jam_audit ⌢ Y(Hᵥ)
	context := append(types.ByteSequence(types.JamAudit[:]), yHv...)

	// s₀ = VRF⟨context⟩ using own key
	validatorKey := posteriorState.GetKappa()[validatorIndex].Bandersnatch
	validatorHandler, _ := safrole.CreateVRFHandler(validatorKey)
	s0, _ := validatorHandler.VRFIetfOutput(context)

	return types.BandersnatchVrfSignature(s0)
}

// (17.5) a0 = {(c, w) | (c, w) ∈ p⋅⋅⋅+10, w ≠ ∅}
// (17.6) where p = F([(c, Qc) | c c <− NC], r)
// (17.7) and r = Y(s0)
//
//	Generate initial audit assignment set a0 for validator v
func ComputeA0ForValidator(Q []*types.WorkReport, validatorIndex int) (output []types.AuditReport) { // TODO: get s0 from local data?
	// (17.3) get s0
	s := store.GetInstance()
	s0 := GetS0(validatorIndex)

	// (17.7) compute r = 𝒴(s₀)
	validatorKey := s.GetPosteriorStates().GetKappa()[validatorIndex].Bandersnatch
	handler, _ := safrole.CreateVRFHandler(validatorKey)
	r, _ := handler.VRFIetfOutput(s0[:]) // (17.7)

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
func GetTranchIndex(T types.U32) types.U32 { // TODO, how to get T
	H_t := store.GetInstance().GetIntermediateHeader().Slot
	n := (T - types.U32(types.SlotPeriod)*types.U32(H_t)) / types.TranchePeriod
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
	ed25519Pub := store.GetInstance().GetPosteriorStates().GetKappa()[validatorIndex].Ed25519
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

// (17.18) n = {Sκ[v]e (Xe(w) ⌢ H(w)) S (c, w) ∈ an}
func BuildJudgements(
	tranche types.U32,
	reports []types.AuditReport, // (c, w) ∈ aₙ
	hashFunc func(types.ByteSequence) types.OpaqueHash,
	validator_index int,
) []types.Ed25519Signature {
	var judgments []types.Ed25519Signature

	for _, item := range reports {
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
		judgments = append(judgments, types.Ed25519Signature(signature))
	}

	return judgments
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
	n types.U32, /*tranche*/
	Q []*types.WorkReport, // Q
	priorAssignment map[types.WorkPackageHash][]types.ValidatorIndex, // A_n-1
	positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool, // J⊤
	hashFunc func(types.ByteSequence) types.OpaqueHash, // H(w)
	validator_index int,
) []types.AuditReport {
	var result []types.AuditReport

	for _, Report := range Q {
		if Report == nil {
			continue
		}
		report := *Report
		// (17.16) mₙ = |Aₙ₋₁(w) ∖ J⊤(w)|
		m_n := 0 // no-show assigned validators
		wHash := report.PackageSpec.Hash
		assigned := priorAssignment[wHash]
		judged := positiveJudgers[wHash]

		for _, v := range assigned {
			if !judged[v] {
				m_n++
			}
		}
		if m_n == 0 {
			continue // report no longer needs audit
		}
		// Construct VRF input seed
		var context types.ByteSequence
		context = append(context, types.JamAudit...)
		// Follow formula 6.22's implementation to get Y(H_v)
		s := store.GetInstance()

		posterior_state := s.GetPosteriorStates()
		header := s.GetIntermediateHeader()

		public_key := posterior_state.GetKappa()[header.AuthorIndex].Bandersnatch
		entropy_source := header.EntropySource
		handler, _ := safrole.CreateVRFHandler(public_key)
		// Y(Hv)
		vrfOutput, _ := handler.VRFIetfOutput(entropy_source[:])
		context = append(context, vrfOutput...)
		//  H(w)
		H_w := hashFunc(utilities.WorkReportSerialization(report))
		context = append(context, H_w[:]...)
		// n
		context = append(context, utilities.SerializeFixedLength(n, 4)...)

		validator_public_key := posterior_state.GetKappa()[validator_index].Bandersnatch
		handler2, _ := safrole.CreateVRFHandler(validator_public_key)

		sn_w, _ := handler2.VRFIetfOutput(context[:])

		vrfOutput3, _ := handler.VRFIetfOutput(sn_w[:])
		// threshold: 𝒴(sₙ(w))₀ < mₙ
		if int(vrfOutput3[0]) < m_n {
			result = append(result, types.AuditReport{
				CoreID:      report.CoreIndex,
				Report:      report,
				ValidatorID: types.ValidatorIndex(validator_index),
				AuditResult: false,
			})
		}
	}
	return result
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
	if len(positives) > (2*types.ValidatorsCount)/3 {
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
