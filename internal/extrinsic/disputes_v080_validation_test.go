package extrinsic

import (
	"crypto/ed25519"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestFaultControllerRequiresDecidedContradictoryVerdict(t *testing.T) {
	good := types.WorkReportHash{0x11}
	bad := types.WorkReportHash{0x22}
	wonky := types.WorkReportHash{0x33}
	unknown := types.WorkReportHash{0x44}

	tests := []struct {
		name    string
		target  types.WorkReportHash
		vote    bool
		wantErr bool
	}{
		{name: "good report with invalid vote", target: good, vote: false},
		{name: "good report with valid vote", target: good, vote: true, wantErr: true},
		{name: "bad report with valid vote", target: bad, vote: true},
		{name: "bad report with invalid vote", target: bad, vote: false, wantErr: true},
		{name: "wonky report", target: wonky, vote: true, wantErr: true},
		{name: "unknown report", target: unknown, vote: false, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blockchain.ResetInstance()
			t.Cleanup(blockchain.ResetInstance)
			posterior := blockchain.GetInstance().GetPosteriorStates()
			posterior.SetPsiG([]types.WorkReportHash{good})
			posterior.SetPsiB([]types.WorkReportHash{bad})
			posterior.SetPsiW([]types.WorkReportHash{wonky})

			controller := NewFaultController()
			controller.Faults = []types.Fault{{Target: tc.target, Vote: tc.vote}}
			err := controller.VerifyReportHashValidty()
			if tc.wantErr {
				if err == nil || err.Error() != "fault_verdict_wrong" {
					t.Fatalf("error = %v, want fault_verdict_wrong", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerdictVerifySignatureUsesActiveAndPreviousValidatorSets(t *testing.T) {
	currentPublic, currentPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate current key: %v", err)
	}
	previousPublic, previousPrivate, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate previous key: %v", err)
	}
	target := types.WorkReportHash{0xAA}

	tests := []struct {
		name       string
		age        types.U32
		privateKey ed25519.PrivateKey
		wantErr    bool
	}{
		{name: "current epoch uses active set", age: 2, privateKey: currentPrivate},
		{name: "previous epoch uses previous set", age: 1, privateKey: previousPrivate},
		{name: "previous epoch rejects active key", age: 1, privateKey: currentPrivate, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blockchain.ResetInstance()
			t.Cleanup(blockchain.ResetInstance)
			cs := blockchain.GetInstance()
			current := make(types.ValidatorsData, types.ValidatorsCount)
			previous := make(types.ValidatorsData, types.ValidatorsCount)
			copy(current[0].Ed25519[:], currentPublic)
			copy(previous[0].Ed25519[:], previousPublic)
			cs.GetPriorStates().SetKappa(current)
			cs.GetPriorStates().SetLambda(previous)
			cs.GetPriorStates().SetTau(types.TimeSlot(2 * types.EpochLength))

			message := append([]byte(types.JamValid), target[:]...)
			signature := ed25519.Sign(tc.privateKey, message)
			var encodedSignature types.Ed25519Signature
			copy(encodedSignature[:], signature)
			verdict := VerdictWrapper{Verdict: types.Verdict{
				Target: target,
				Age:    tc.age,
				Votes: []types.Judgement{{
					Vote:      true,
					Index:     0,
					Signature: encodedSignature,
				}},
			}}

			err := verdict.VerifySignature()
			if tc.wantErr {
				if err == nil || err.Error() != "bad_signature" {
					t.Fatalf("error = %v, want bad_signature", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerdictVerifySignatureRejectsPreviousEpochAtGenesis(t *testing.T) {
	blockchain.ResetInstance()
	t.Cleanup(blockchain.ResetInstance)
	cs := blockchain.GetInstance()
	cs.GetPriorStates().SetTau(0)

	verdict := VerdictWrapper{Verdict: types.Verdict{Age: ^types.U32(0)}}
	if err := verdict.VerifySignature(); err == nil || err.Error() != "bad_judgement_age" {
		t.Fatalf("error = %v, want bad_judgement_age", err)
	}
}

func TestDisputeControllerUpdatesOffenderState(t *testing.T) {
	blockchain.ResetInstance()
	t.Cleanup(blockchain.ResetInstance)
	cs := blockchain.GetInstance()
	existing := types.Ed25519Public{0x10}
	newFault := types.Ed25519Public{0x20}
	newCulprit := types.Ed25519Public{0x30}
	cs.GetPriorStates().SetPsi(types.DisputesRecords{Offenders: []types.Ed25519Public{existing}})

	controller := NewDisputeController(NewVerdictController(), NewFaultController(), NewCulpritController())
	controller.UpdatePsiO(
		[]types.Culprit{{Key: newCulprit}},
		[]types.Fault{{Key: newFault}},
	)

	got := cs.GetPosteriorStates().GetPsiO()
	want := []types.Ed25519Public{existing, newFault, newCulprit}
	if len(got) != len(want) {
		t.Fatalf("offender count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("offender[%d] = %x, want %x", i, got[i], want[i])
		}
	}
}
