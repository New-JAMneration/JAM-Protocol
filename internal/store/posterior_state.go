package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// PosteriorStates represents a thread-safe global state container
type PosteriorStates struct {
	mu    sync.RWMutex
	state *types.State
}

// NewPosteriorStates returns a new instance of PosteriorStates
func NewPosteriorStates() *PosteriorStates {
	return &PosteriorStates{
		state: &types.State{
			Theta: make([]types.ReadyQueueItem, types.EpochLength),
			Xi:    make(types.AccumulatedQueue, types.EpochLength),
		},
	}
}

// GetState returns the current state
func (s *PosteriorStates) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

// GenerateGenesisState generates a genesis state
func (s *PosteriorStates) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

// SetAlpha sets the alpha value
func (s *PosteriorStates) SetAlpha(alpha types.AuthPools) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Alpha = alpha
}

// GetAlpha returns the alpha value
func (s *PosteriorStates) GetAlpha() types.AuthPools {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Alpha
}

func (s *PosteriorStates) SetBeta(beta types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}

// GetBeta returns the beta value
func (s *PosteriorStates) GetBeta() types.BlocksHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Beta
}

// SetGamma sets the gamma value
func (s *PosteriorStates) SetGamma(gamma types.Gamma) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma = gamma
}

// GetGamma returns the gamma value
func (s *PosteriorStates) GetGamma() types.Gamma {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma
}

// SetGammaK sets the gammaK value
func (s *PosteriorStates) SetGammaK(gammaK types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaK = gammaK
}

// GetGammaK returns the gammaK value
func (s *PosteriorStates) GetGammaK() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaK
}

// SetGammaZ sets the gammaZ value
func (s *PosteriorStates) SetGammaZ(gammaZ types.BandersnatchRingCommitment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaZ = gammaZ
}

// GetGammaZ returns the gammaZ value
func (s *PosteriorStates) GetGammaZ() types.BandersnatchRingCommitment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaZ
}

// SetGammaS sets the gammaS value
func (s *PosteriorStates) SetGammaS(gammaS types.TicketsOrKeys) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaS = gammaS
}

// GetGammaS returns the gammaS value
func (s *PosteriorStates) GetGammaS() types.TicketsOrKeys {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaS
}

// SetGammaA sets the gammaA value
func (s *PosteriorStates) SetGammaA(gammaA types.TicketsAccumulator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaA = gammaA
}

// GetGammaA returns the gammaA value
func (s *PosteriorStates) GetGammaA() types.TicketsAccumulator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaA
}

// SetDelta sets the delta value
func (s *PosteriorStates) SetDelta(delta types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Delta = delta
}

// GetDelta returns the delta value
func (s *PosteriorStates) GetDelta() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Delta
}

// SetEta sets the eta value
func (s *PosteriorStates) SetEta(eta types.EntropyBuffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Eta = eta
}

// GetEta returns the eta value
func (s *PosteriorStates) GetEta() types.EntropyBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Eta
}

// SetEta0 sets the eta0 value
func (s *PosteriorStates) SetEta0(entropy types.Entropy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Eta[0] = entropy
}

// SetIota sets the iota value
func (s *PosteriorStates) SetIota(variota types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Iota = variota
}

// GetIota returns the iota value
func (s *PosteriorStates) GetIota() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Iota
}

// SetKappa sets the kappa value
func (s *PosteriorStates) SetKappa(kappa types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa
}

// GetKappa returns the kappa value
func (s *PosteriorStates) GetKappa() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Kappa
}

// SetLambda sets the lambda value
func (s *PosteriorStates) SetLambda(lambda types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

// GetLambda returns the lambda value
func (s *PosteriorStates) GetLambda() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Lambda
}

// SetRho sets the rho value
func (s *PosteriorStates) SetRho(rho types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rho = rho
}

// GetRho returns the rho value
func (s *PosteriorStates) GetRho() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Rho
}

// SetTau sets the tau value
func (s *PosteriorStates) SetTau(tau types.TimeSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Tau = tau
}

// GetTau returns the tau value
func (s *PosteriorStates) GetTau() types.TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Tau
}

// SetVarphi sets the varphi value
func (s *PosteriorStates) SetVarphi(varphi types.AuthQueues) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Varphi = varphi
}

// GetVarphi returns the varphi value
func (s *PosteriorStates) GetVarphi() types.AuthQueues {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Varphi
}

// SetChi sets the chi value
func (s *PosteriorStates) SetChi(chi types.Privileges) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi = chi
}

// GetChi returns the chi value
func (s *PosteriorStates) GetChi() types.Privileges {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi
}

// SetManagerServiceIndex sets the managerServiceIndex value
func (s *PosteriorStates) SetManagerServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Bless = serviceId
}

// GetManagerServiceIndex returns the managerServiceIndex value
func (s *PosteriorStates) GetManagerServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Bless
}

// SetAlterPhiServiceIndex sets the alterPhiServiceIndex value
func (s *PosteriorStates) SetAlterPhiServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Assign = serviceId
}

// GetAlterPhiServiceIndex returns the alterPhiServiceIndex value
func (s *PosteriorStates) GetAlterPhiServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Assign
}

// SetAlterIotaServiceIndex sets the alterIotaServiceIndex value
func (s *PosteriorStates) SetAlterIotaServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Designate = serviceId
}

// GetAlterIotaServiceIndex returns the alterIotaServiceIndex value
func (s *PosteriorStates) GetAlterIotaServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Designate
}

// SetAutoAccumulateGasLimits sets the autoAccumulateGasLimits value
func (s *PosteriorStates) SetAutoAccumulateGasLimits(autoAccumulateGasLimits types.AlwaysAccumulateMap) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.AlwaysAccum = autoAccumulateGasLimits
}

// GetAutoAccumulateGasLimits returns the autoAccumulateGasLimits value
func (s *PosteriorStates) GetAutoAccumulateGasLimits() types.AlwaysAccumulateMap {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.AlwaysAccum
}

// SetPsi sets the psi value
func (s *PosteriorStates) SetPsi(psi types.DisputesRecords) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi = psi
}

// GetPsi returns the psi value
func (s *PosteriorStates) GetPsi() types.DisputesRecords {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi
}

// SetPsiG sets the psiG value
func (s *PosteriorStates) SetPsiG(psiG []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Good = psiG
}

// GetPsiG returns the psiG value
func (s *PosteriorStates) GetPsiG() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Good
}

// SetPsiB sets the psiB value
func (s *PosteriorStates) SetPsiB(psiB []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Bad = psiB
}

// GetPsiB returns the psiB value
func (s *PosteriorStates) GetPsiB() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Bad
}

// SetPsiW sets the psiW value
func (s *PosteriorStates) SetPsiW(psiW []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Wonky = psiW
}

// GetPsiW returns the psiW value
func (s *PosteriorStates) GetPsiW() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Wonky
}

// SetPsiO sets the psiO value
func (s *PosteriorStates) SetPsiO(psiO []types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = psiO
}

// GetPsiO returns the psiO value
func (s *PosteriorStates) GetPsiO() []types.Ed25519Public {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Offenders
}

// SetPi sets the pi value
func (s *PosteriorStates) SetPi(pi types.Statistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi = pi
}

// GetPi returns the pi value
func (s *PosteriorStates) GetPi() types.Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi
}

// SetPiCurrent sets the pi.Current value
func (s *PosteriorStates) SetPiCurrent(current types.ValidatorsStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.ValsCurr = current
}

// GetPiCurrent returns the pi.Current value
func (s *PosteriorStates) GetPiCurrent() types.ValidatorsStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.ValsCurr
}

// SetPiLast sets the pi Last.value
func (s *PosteriorStates) SetPiLast(last types.ValidatorsStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.ValsLast = last
}

// GetPiLast returns the pi.Last value
func (s *PosteriorStates) GetPiLast() types.ValidatorsStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.ValsLast
}

// SetTheta sets the theta value
func (s *PosteriorStates) SetTheta(theta types.ReadyQueue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Theta = theta
}

// GetTheta returns the theta value
func (s *PosteriorStates) GetTheta() types.ReadyQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Theta
}

// SetXi sets the xi value
func (s *PosteriorStates) SetXi(xi types.AccumulatedQueue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Xi = xi
}

// GetXi returns the xi value
func (s *PosteriorStates) GetXi() types.AccumulatedQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Xi
}
