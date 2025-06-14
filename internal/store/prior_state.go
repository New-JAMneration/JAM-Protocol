package store

import (
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// PriorStates represents a thread-safe global state container
type PriorStates struct {
	mu    sync.RWMutex
	state *types.State
}

// NewPriorStates creates a new instance of PriorStates
func NewPriorStates() *PriorStates {
	return &PriorStates{
		state: &types.State{
			Theta:      make([]types.ReadyQueueItem, types.EpochLength),
			Xi:         make(types.AccumulatedQueue, types.EpochLength),
			LastAccOut: make(types.AccumulatedServiceOutput),
		},
	}
}

// GetState returns the current state
func (s *PriorStates) GetState() types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.state
}

// GenerateGenesisState generates a genesis state
func (s *PriorStates) GenerateGenesisState(state types.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = &state
}

// SetAlpha sets the alpha value
func (s *PriorStates) SetAlpha(alpha types.AuthPools) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Alpha = alpha
}

// GetAlpha returns the alpha value
func (s *PriorStates) GetAlpha() types.AuthPools {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Alpha
}

// SetBeta sets the beta value
func (s *PriorStates) SetBeta(beta types.Beta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta = beta
}

func (s *PriorStates) SetBetaH(betaH types.BlocksHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta.History = betaH
}

func (s *PriorStates) SetBetaB(betaB types.Mmr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Beta.BeefyBelt = betaB
}

// GetBeta returns the beta value
func (s *PriorStates) GetBeta() types.Beta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Beta
}

// SetGamma sets the gamma value
func (s *PriorStates) SetGamma(gamma types.Gamma) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma = gamma
}

// GetGamma returns the gamma value
func (s *PriorStates) GetGamma() types.Gamma {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma
}

// SetGammaK sets the gammaK value
func (s *PriorStates) SetGammaK(gammaK types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaK = gammaK
}

// GetGammaK returns the gammaK value
func (s *PriorStates) GetGammaK() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaK
}

// SetGammaZ sets the gammaZ value
func (s *PriorStates) SetGammaZ(gammaZ types.BandersnatchRingCommitment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaZ = gammaZ
}

// GetGammaZ returns the gammaZ value
func (s *PriorStates) GetGammaZ() types.BandersnatchRingCommitment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaZ
}

// SetGammaS sets the gammaS value
func (s *PriorStates) SetGammaS(gammaS types.TicketsOrKeys) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaS = gammaS
}

// GetGammaS returns the gammaS value
func (s *PriorStates) GetGammaS() types.TicketsOrKeys {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaS
}

// SetGammaA sets the gammaA value
func (s *PriorStates) SetGammaA(gammaA types.TicketsAccumulator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Gamma.GammaA = gammaA
}

// GetGammaA returns the gammaA value
func (s *PriorStates) GetGammaA() types.TicketsAccumulator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Gamma.GammaA
}

// SetDelta sets the delta value
func (s *PriorStates) SetDelta(delta types.ServiceAccountState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Delta = delta
}

// GetDelta returns the delta value
func (s *PriorStates) GetDelta() types.ServiceAccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Delta
}

// SetEta sets the eta value
func (s *PriorStates) SetEta(eta types.EntropyBuffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Eta = eta
}

// GetEta returns the eta value
func (s *PriorStates) GetEta() types.EntropyBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Eta
}

// SetIota sets the iota value
func (s *PriorStates) SetIota(variota types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Iota = variota
}

// GetIota returns the iota value
func (s *PriorStates) GetIota() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Iota
}

// SetKappa sets the kappa value
func (s *PriorStates) SetKappa(kappa types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Kappa = kappa
}

// GetKappa returns the kappa value
func (s *PriorStates) GetKappa() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Kappa
}

// SetLambda sets the lambda value
func (s *PriorStates) SetLambda(lambda types.ValidatorsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Lambda = lambda
}

// GetLambda returns the lambda value
func (s *PriorStates) GetLambda() types.ValidatorsData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Lambda
}

// SetRho sets the rho value
func (s *PriorStates) SetRho(rho types.AvailabilityAssignments) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rho = rho
}

// GetRho returns the rho value
func (s *PriorStates) GetRho() types.AvailabilityAssignments {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Rho
}

// SetTau sets the tau value
func (s *PriorStates) SetTau(tau types.TimeSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Tau = tau
}

// GetTau returns the tau value
func (s *PriorStates) GetTau() types.TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Tau
}

// SetVarphi sets the varphi value
func (s *PriorStates) SetVarphi(varphi types.AuthQueues) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Varphi = varphi
}

// GetVarphi returns the varphi value
func (s *PriorStates) GetVarphi() types.AuthQueues {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Varphi
}

// SetChi sets the chi value
func (s *PriorStates) SetChi(chi types.Privileges) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi = chi
}

// GetChi returns the chi value
func (s *PriorStates) GetChi() types.Privileges {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi
}

// SetManagerServiceIndex sets the managerServiceIndex value
func (s *PriorStates) SetManagerServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Bless = serviceId
}

// GetManagerServiceIndex returns the managerServiceIndex value
func (s *PriorStates) GetManagerServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Bless
}

// SetAlterPhiServiceIndex sets the alterPhiServiceIndex value
func (s *PriorStates) SetAlterPhiServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Assign = serviceId
}

// GetAlterPhiServiceIndex returns the alterPhiServiceIndex value
func (s *PriorStates) GetAlterPhiServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Assign
}

// SetAlterIotaServiceIndex sets the alterIotaServiceIndex value
func (s *PriorStates) SetAlterIotaServiceIndex(serviceId types.ServiceId) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.Designate = serviceId
}

// GetAlterIotaServiceIndex returns the alterIotaServiceIndex value
func (s *PriorStates) GetAlterIotaServiceIndex() types.ServiceId {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.Designate
}

// SetAutoAccumulateGasLimits sets the autoAccumulateGasLimits value
func (s *PriorStates) SetAutoAccumulateGasLimits(autoAccumulateGasLimits types.AlwaysAccumulateMap) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Chi.AlwaysAccum = autoAccumulateGasLimits
}

// GetAutoAccumulateGasLimits returns the autoAccumulateGasLimits value
func (s *PriorStates) GetAutoAccumulateGasLimits() types.AlwaysAccumulateMap {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Chi.AlwaysAccum
}

// SetPsi sets the psi value
func (s *PriorStates) SetPsi(psi types.DisputesRecords) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi = psi
}

// GetPsi returns the psi value
func (s *PriorStates) GetPsi() types.DisputesRecords {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi
}

// AddPsiOffenders adds a new offender
func (s *PriorStates) AddPsiOffenders(offender types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = append(s.state.Psi.Offenders, offender)
}

// SetPsiG sets the psiG value
func (s *PriorStates) SetPsiG(psiG []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Good = psiG
}

// GetPsiG returns the psiG value
func (s *PriorStates) GetPsiG() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Good
}

// SetPsiB sets the psiB value
func (s *PriorStates) SetPsiB(psiB []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Bad = psiB
}

// GetPsiB returns the psiB value
func (s *PriorStates) GetPsiB() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Bad
}

// SetPsiW sets the psiW value
func (s *PriorStates) SetPsiW(psiW []types.WorkReportHash) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Wonky = psiW
}

// GetPsiW returns the psiW value
func (s *PriorStates) GetPsiW() []types.WorkReportHash {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Wonky
}

// SetPsiO sets the psiO value
func (s *PriorStates) SetPsiO(psiO []types.Ed25519Public) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Psi.Offenders = psiO
}

// GetPsiO returns the psiO value
func (s *PriorStates) GetPsiO() []types.Ed25519Public {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Psi.Offenders
}

// SetPi sets the pi value
func (s *PriorStates) SetPi(pi types.Statistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi = pi
}

// GetPi returns the pi value
func (s *PriorStates) GetPi() types.Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi
}

// SetPiCurrent sets the pi.Current value
func (s *PriorStates) SetPiCurrent(current types.ValidatorsStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.ValsCurr = current
}

// GetPiCurrent returns the pi.Current value
func (s *PriorStates) GetPiCurrent() types.ValidatorsStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.ValsCurr
}

// SetPiLast sets the pi Last.value
func (s *PriorStates) SetPiLast(last types.ValidatorsStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.ValsLast = last
}

// GetPiLast returns the pi.Last value
func (s *PriorStates) GetPiLast() types.ValidatorsStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.ValsLast
}

// Set cores statisitcs
func (s *PriorStates) SetCoresStatistics(coresStatistics types.CoresStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.Cores = coresStatistics
}

// Get cores statisitcs
func (s *PriorStates) GetCoresStatistics() types.CoresStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.Cores
}

// Set services statistics
func (s *PriorStates) SetServicesStatistics(servicesStatistics types.ServicesStatistics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Pi.Services = servicesStatistics
}

// Get services statistics
func (s *PriorStates) GetServicesStatistics() types.ServicesStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Pi.Services
}

// SetTheta sets the theta value
func (s *PriorStates) SetTheta(theta types.ReadyQueue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Theta = theta
}

// GetTheta returns the theta value
func (s *PriorStates) GetTheta() types.ReadyQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Theta
}

// SetXi sets the xi value
func (s *PriorStates) SetXi(xi types.AccumulatedQueue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Xi = xi
}

// GetXi returns the xi value
func (s *PriorStates) GetXi() types.AccumulatedQueue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Xi
}

func (s *PriorStates) SetLastAccOut(c types.AccumulatedServiceOutput) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.LastAccOut = c
}

func (s *PriorStates) GetLastAccOut() types.AccumulatedServiceOutput {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.LastAccOut
}
