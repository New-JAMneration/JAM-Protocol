package stf

import (
	"fmt"
	"os"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/accumulation"
	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/recent_history"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/timing"
)

// Set STF_TIMING=1 to enable verbose per-block timing output
var verboseTiming = os.Getenv("STF_TIMING") == "1"

// ResetCustomTimings resets the global timing collector
func ResetCustomTimings() {
	timing.ResetGlobal()
}

// STFTiming holds timing information for each STF step
type STFTiming struct {
	ValidateNonVRFHeader time.Duration
	UpdateBetaH          time.Duration
	UpdateDisputes       time.Duration
	UpdateSafrole        time.Duration
	ValidateHeaderVrf    time.Duration
	ValidateExtrinsic    time.Duration
	UpdateAssurances     time.Duration
	UpdateReports        time.Duration
	UpdateAccumulate     time.Duration
	UpdateHistory        time.Duration
	UpdatePreimages      time.Duration
	UpdateAuthorizations time.Duration
	UpdateStatistics     time.Duration
	Total                time.Duration
}

// String returns a formatted string of all timings
func (t STFTiming) String() string {
	return fmt.Sprintf(`STF Timing Breakdown:
  ValidateNonVRFHeader: %12v
  UpdateBetaH:          %12v
  UpdateDisputes:       %12v
  UpdateSafrole:        %12v
  ValidateHeaderVrf:    %12v
  ValidateExtrinsic:    %12v
  UpdateAssurances:     %12v
  UpdateReports:        %12v
  UpdateAccumulate:     %12v
  UpdateHistory:        %12v
  UpdatePreimages:      %12v
  UpdateAuthorizations: %12v
  UpdateStatistics:     %12v
  ─────────────────────────────────
  Total:                %12v`,
		t.ValidateNonVRFHeader,
		t.UpdateBetaH,
		t.UpdateDisputes,
		t.UpdateSafrole,
		t.ValidateHeaderVrf,
		t.ValidateExtrinsic,
		t.UpdateAssurances,
		t.UpdateReports,
		t.UpdateAccumulate,
		t.UpdateHistory,
		t.UpdatePreimages,
		t.UpdateAuthorizations,
		t.UpdateStatistics,
		t.Total,
	)
}

// RunSTFWithTiming executes the STF and returns detailed timing information.
// Use this for performance analysis and optimization review.
func RunSTFWithTiming() (bool, error, STFTiming) {
	var timing STFTiming
	totalStart := time.Now()

	var (
		err              error
		cs               = blockchain.GetInstance()
		priorState       = cs.GetPriorStates().GetState()
		header           = cs.GetLatestBlock().Header
		extrinsic        = cs.GetLatestBlock().Extrinsic
		unmatchedKeyVals = cs.GetPriorStateUnmatchedKeyVals()
	)

	// Update timeslot
	cs.GetPosteriorStates().SetTau(header.Slot)

	// Validate Non-VRF Header
	start := time.Now()
	if header.Parent != (types.HeaderHash{}) {
		err = ValidateNonVRFHeader(header, &priorState, extrinsic)
		if err != nil {
			timing.ValidateNonVRFHeader = time.Since(start)
			timing.Total = time.Since(totalStart)
			return true, err, timing
		}
	}
	timing.ValidateNonVRFHeader = time.Since(start)

	// Update BetaH
	start = time.Now()
	recent_history.STFBetaH2BetaHDagger()
	timing.UpdateBetaH = time.Since(start)

	// Update Disputes
	start = time.Now()
	err = UpdateDisputes()
	timing.UpdateDisputes = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Safrole
	start = time.Now()
	err = UpdateSafrole()
	timing.UpdateSafrole = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	postState := cs.GetPosteriorStates().GetState()

	// Validate Header VRF
	start = time.Now()
	err = ValidateHeaderVrf(header, &priorState, &postState)
	timing.ValidateHeaderVrf = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Validate Extrinsic
	start = time.Now()
	err = ValidateExtrinsic(extrinsic, &priorState, unmatchedKeyVals)
	timing.ValidateExtrinsic = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Assurances
	start = time.Now()
	err = UpdateAssurances()
	timing.UpdateAssurances = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Reports
	start = time.Now()
	err = UpdateReports()
	timing.UpdateReports = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Accumulate
	start = time.Now()
	err = UpdateAccumlate()
	timing.UpdateAccumulate = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update History
	start = time.Now()
	err = recent_history.STFBetaHDagger2BetaHPrime()
	timing.UpdateHistory = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Preimages
	start = time.Now()
	err = accumulation.ProcessPreimageExtrinsics()
	timing.UpdatePreimages = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Authorizations
	start = time.Now()
	err = UpdateAuthorizations()
	timing.UpdateAuthorizations = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	// Update Statistics
	start = time.Now()
	err = UpdateStatistics()
	timing.UpdateStatistics = time.Since(start)
	if err != nil {
		timing.Total = time.Since(totalStart)
		return true, err, timing
	}

	timing.Total = time.Since(totalStart)
	return false, nil, timing
}

// TimingSummary holds aggregated timing statistics
type TimingSummary struct {
	BlockCount int
	Total      STFTiming
	Average    STFTiming
}

// CalculateTimingSummary calculates timing statistics from multiple blocks
func CalculateTimingSummary(timings []STFTiming) TimingSummary {
	if len(timings) == 0 {
		return TimingSummary{}
	}

	var total STFTiming
	for _, t := range timings {
		total.ValidateNonVRFHeader += t.ValidateNonVRFHeader
		total.UpdateBetaH += t.UpdateBetaH
		total.UpdateDisputes += t.UpdateDisputes
		total.UpdateSafrole += t.UpdateSafrole
		total.ValidateHeaderVrf += t.ValidateHeaderVrf
		total.ValidateExtrinsic += t.ValidateExtrinsic
		total.UpdateAssurances += t.UpdateAssurances
		total.UpdateReports += t.UpdateReports
		total.UpdateAccumulate += t.UpdateAccumulate
		total.UpdateHistory += t.UpdateHistory
		total.UpdatePreimages += t.UpdatePreimages
		total.UpdateAuthorizations += t.UpdateAuthorizations
		total.UpdateStatistics += t.UpdateStatistics
		total.Total += t.Total
	}

	n := time.Duration(len(timings))
	avg := STFTiming{
		ValidateNonVRFHeader: total.ValidateNonVRFHeader / n,
		UpdateBetaH:          total.UpdateBetaH / n,
		UpdateDisputes:       total.UpdateDisputes / n,
		UpdateSafrole:        total.UpdateSafrole / n,
		ValidateHeaderVrf:    total.ValidateHeaderVrf / n,
		ValidateExtrinsic:    total.ValidateExtrinsic / n,
		UpdateAssurances:     total.UpdateAssurances / n,
		UpdateReports:        total.UpdateReports / n,
		UpdateAccumulate:     total.UpdateAccumulate / n,
		UpdateHistory:        total.UpdateHistory / n,
		UpdatePreimages:      total.UpdatePreimages / n,
		UpdateAuthorizations: total.UpdateAuthorizations / n,
		UpdateStatistics:     total.UpdateStatistics / n,
		Total:                total.Total / n,
	}

	return TimingSummary{
		BlockCount: len(timings),
		Total:      total,
		Average:    avg,
	}
}

// FormatSummaryTable returns a formatted table string for the timing summary
func (s TimingSummary) FormatSummaryTable() string {
	if s.BlockCount == 0 {
		return "No timing data collected"
	}

	// Calculate percentage of total for each step
	total := float64(s.Average.Total)
	pct := func(d time.Duration) float64 {
		if total == 0 {
			return 0
		}
		return float64(d) / total * 100
	}

	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════════════╗
║                    STF TIMING SUMMARY (%d blocks)                    
╠══════════════════════════════════════════════════════════════════════╣
║  Step                      │    Average    │   Total     │   %%     ║
╠══════════════════════════════════════════════════════════════════════╣
║  ValidateNonVRFHeader      │ %12v │ %11v │ %5.1f%%  ║
║  UpdateBetaH               │ %12v │ %11v │ %5.1f%%  ║
║  UpdateDisputes            │ %12v │ %11v │ %5.1f%%  ║
║  UpdateSafrole             │ %12v │ %11v │ %5.1f%%  ║
║  ValidateHeaderVrf         │ %12v │ %11v │ %5.1f%%  ║
║  ValidateExtrinsic         │ %12v │ %11v │ %5.1f%%  ║
║  UpdateAssurances          │ %12v │ %11v │ %5.1f%%  ║
║  UpdateReports             │ %12v │ %11v │ %5.1f%%  ║
║  UpdateAccumulate          │ %12v │ %11v │ %5.1f%%  ║
║  UpdateHistory             │ %12v │ %11v │ %5.1f%%  ║
║  UpdatePreimages           │ %12v │ %11v │ %5.1f%%  ║
║  UpdateAuthorizations      │ %12v │ %11v │ %5.1f%%  ║
║  UpdateStatistics          │ %12v │ %11v │ %5.1f%%  ║
╠══════════════════════════════════════════════════════════════════════╣
║  TOTAL                     │ %12v │ %11v │ 100.0%%  ║
╚══════════════════════════════════════════════════════════════════════╝`,
		s.BlockCount,
		s.Average.ValidateNonVRFHeader, s.Total.ValidateNonVRFHeader, pct(s.Average.ValidateNonVRFHeader),
		s.Average.UpdateBetaH, s.Total.UpdateBetaH, pct(s.Average.UpdateBetaH),
		s.Average.UpdateDisputes, s.Total.UpdateDisputes, pct(s.Average.UpdateDisputes),
		s.Average.UpdateSafrole, s.Total.UpdateSafrole, pct(s.Average.UpdateSafrole),
		s.Average.ValidateHeaderVrf, s.Total.ValidateHeaderVrf, pct(s.Average.ValidateHeaderVrf),
		s.Average.ValidateExtrinsic, s.Total.ValidateExtrinsic, pct(s.Average.ValidateExtrinsic),
		s.Average.UpdateAssurances, s.Total.UpdateAssurances, pct(s.Average.UpdateAssurances),
		s.Average.UpdateReports, s.Total.UpdateReports, pct(s.Average.UpdateReports),
		s.Average.UpdateAccumulate, s.Total.UpdateAccumulate, pct(s.Average.UpdateAccumulate),
		s.Average.UpdateHistory, s.Total.UpdateHistory, pct(s.Average.UpdateHistory),
		s.Average.UpdatePreimages, s.Total.UpdatePreimages, pct(s.Average.UpdatePreimages),
		s.Average.UpdateAuthorizations, s.Total.UpdateAuthorizations, pct(s.Average.UpdateAuthorizations),
		s.Average.UpdateStatistics, s.Total.UpdateStatistics, pct(s.Average.UpdateStatistics),
		s.Average.Total, s.Total.Total,
	)
}

// PrintTimingSummary prints a summary of timing results for multiple blocks
func PrintTimingSummary(timings []STFTiming) {
	summary := CalculateTimingSummary(timings)
	fmt.Println(summary.FormatSummaryTable())

	// Print custom timings from the timing package
	timing.PrintGlobalSummary()
}

// LogTimingSummary returns the summary string for use with t.Log in tests
func LogTimingSummary(timings []STFTiming) string {
	summary := CalculateTimingSummary(timings)
	return summary.FormatSummaryTable()
}
