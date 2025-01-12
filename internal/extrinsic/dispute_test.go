package extrinsic

import (
	"bytes"
	"fmt"
	"testing"

	store "github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

var Kappa = types.ValidatorsData{
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		Ed25519:      types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
		Ed25519:      types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
}

var Lambda = types.ValidatorsData{
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0xaa2b95f7572875b0d0f186552ae745ba8222fc0b5bd456554bfe51c68938f8bc")),
		Ed25519:      types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0xf16e5352840afb47e206b5c89f560f2611835855cf2e6ebad1acc9520a72591d")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x837ce344bc9defceb0d7de7e9e9925096768b7adb4dad932e532eb6551e0ea02")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x5e465beb01dbafe160ce8216047f2155dd0569f058afd52dcea601025a8d161d")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x48e5fcdce10e0b64ec4eebd0d9211c7bac2f27ce54bca6f7776ff6fee86ab3e3")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x5c7f34a4bd4f2d04076a8c6f9060a0c8d2c6bdd082ceb3eda7df381cb260faff")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x3d5e5a51aab2b048f8686ecd79712a80e3265a114cc73f14bdb2a59233fb66d0")),
		Ed25519:      types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
	{
		Bandersnatch: types.BandersnatchPublic(HexToBytes("0x7f6190116d118d643a98878e294ccf62b509e214299931aad8ff9764181a4e33")),
		Ed25519:      types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
		Bls:          types.BlsPublic(HexToBytes("0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
		Metadata:     types.ValidatorMetadata(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
	},
}

func TestDisputeWorkFlow_progress_invalidates_avail_assignments_1(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0x3c6680931983df80bbd1cb77a0db0303e61550ee0a868d8edd69302d8f45c29f")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      true,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x4d3e576eaa0d449465115bedc8640c452813bd74385003e212b0ff679b9fbf4951ff1a624fcdf4539926cc00f719078d5828f7aecee8918a62c720ead07c240f")),
					},
					{
						Vote:      true,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0xc316345dbf3e4de359e4ddbb94ae8460230e9df90be6f57d92bc177f74da74337fb5461d51389ade2a3cfb55d8980b5e7ffb3063f5d6997b4b115dc9a369f301")),
					},
					{
						Vote:      true,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x8ef8805d2d8ad355c1805c6cbc378bfd549ed2e84ebcda971cc783a967f76a247d6ca435c1188c704f6045bdd080885a9611c3fe812a910c61a4f6016134d508")),
					},
					{
						Vote:      true,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0xc354b7fca4817783c68c2ca47c45028a1fac29cf6ed7b6df53951bf29b85871d3d1ab00163f085198efbbeb27328ab1af2c1b50ebddfc39475e6a0678e093a0d")),
					},
					{
						Vote:      true,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0xf36e93889f6df8103404f019c0b408a5079c85cb06a680c4fd70acc087cae0d563a5d4e5cdd7ff2fb1260b885b23ee1fbfc58c80be0301a95f9f271964eeba02")),
					},
				},
			},
			{
				Target: types.OpaqueHash(HexToBytes("0x94a9424dec0e513afb0a9187c0456b7760f02ee5969130fde29c8683d62f74fb")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      false,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0xccb6a65353f79cc9d65ba9e8b14cd19f24104cd5194c9d2cc81798d982dd6f50777e368364a9df062bb70c0bc9f66ce1391e2bb98cd76f891167cb2f90be770e")),
					},
					{
						Vote:      false,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0x3b487b6e435b45f11d01fa9bd2d3a34863b2ec4d2c2b63f99cebd6262270e642cce938a7f52e6b5eabd3eb80ca4f4c4b88dc6f147de42e7f7abf8bd0ce0ed00f")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x3070c1929f539dd8251542cc4b9ef79511684614d520410ca080c75be23dbd3d9f8308b68e78879354de377e66c413dcd170d742ed234a794478d6c50dc84801")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x9d8e8608c4a581ae559f2ed2265347183aac86888988a4f233e78b753b50181e153fc718354c0d964a3b0399886c18d0343ed7cfbda722aa51b3c18e3d1b4a09")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0x73c4c5f659227fd0b63e0929c764490726e1346bb6540c6cf9ead7bc74473c41829f92b43f78e0516603787ed80766e98e4daeb51187723b2f3a40f2c06b3209")),
					},
				},
			},
		},
		Culprits: []types.Culprit{
			{
				Target:    types.WorkReportHash(HexToBytes("0x94a9424dec0e513afb0a9187c0456b7760f02ee5969130fde29c8683d62f74fb")),
				Key:       types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
				Signature: types.Ed25519Signature(HexToBytes("0xa4b97289d848afdbf0424322a66231e0fb49357fe6e1a0db547cd6392fa12b00f037eb564615421c3744e054b8b23fc97d5a17e114d799d23501ef52a1e6f90e")),
			},
			{
				Target:    types.WorkReportHash(HexToBytes("0x94a9424dec0e513afb0a9187c0456b7760f02ee5969130fde29c8683d62f74fb")),
				Key:       types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
				Signature: types.Ed25519Signature(HexToBytes("0x67b501fc9b014f01075330c624317027f0d3b9986c8420c4c755344b1fa971bb054e56fab0fb486b6610a13a2f954adaef06f7582d28772c53cebe2e2882d104")),
			},
		},
		Faults: []types.Fault{
			{
				Target:    types.WorkReportHash(HexToBytes("0x3c6680931983df80bbd1cb77a0db0303e61550ee0a868d8edd69302d8f45c29f")),
				Vote:      false,
				Key:       types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
				Signature: types.Ed25519Signature(HexToBytes("0x11ae0a70ed9403160acb08701cde224162fe6f7957dd8ebd71652a58007dfc398459ea94f3ab98935bbe21588ed76ba2cfd65993a0e1775d45aab773147cca03")),
			},
		},
	}

	expectedPsi := types.DisputesRecords{
		Good: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x3c6680931983df80bbd1cb77a0db0303e61550ee0a868d8edd69302d8f45c29f")),
		},
		Bad: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x94a9424dec0e513afb0a9187c0456b7760f02ee5969130fde29c8683d62f74fb")),
		},
		Wonky: []types.WorkReportHash{},
		Offenders: []types.Ed25519Public{
			types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
		},
	}

	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)
	store.GetInstance().GetPriorStates().SetRho(
		types.AvailabilityAssignments{
			{
				Report: types.WorkReport{
					PackageSpec: types.WorkPackageSpec{
						Hash:         types.WorkPackageHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
						Length:       types.U32(0),
						ErasureRoot:  types.ErasureRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						ExportsRoot:  types.ExportsRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						ExportsCount: types.U16(0),
					},
					Context: types.RefineContext{
						Anchor:           types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						StateRoot:        types.StateRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						BeefyRoot:        types.BeefyRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						LookupAnchor:     types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						LookupAnchorSlot: types.TimeSlot(0),
						Prerequisites:    []types.OpaqueHash{},
					},
					CoreIndex:         types.CoreIndex(0),
					AuthorizerHash:    types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					AuthOutput:        types.ByteSequence(HexToBytes("0x030201")),
					SegmentRootLookup: types.SegmentRootLookup{},
					Results: []types.WorkResult{
						{
							ServiceId:     types.ServiceId(0),
							CodeHash:      types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
							PayloadHash:   types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
							AccumulateGas: types.Gas(42),
							Result: types.WorkExecResult{
								types.WorkExecResultType("ok"): HexToBytes("0x010203"),
							},
						},
					},
				},
				Timeout: types.TimeSlot(42),
			},
			{
				Report: types.WorkReport{
					PackageSpec: types.WorkPackageSpec{
						Hash:         types.WorkPackageHash(HexToBytes("0xe12c22d4f162d9a012c9319233da5d3e923cc5e1029b8f90e47249c9ab256b35")),
						Length:       0,
						ErasureRoot:  types.ErasureRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						ExportsRoot:  types.ExportsRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						ExportsCount: types.U16(0),
					},
					Context: types.RefineContext{
						Anchor:           types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						StateRoot:        types.StateRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						BeefyRoot:        types.BeefyRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						LookupAnchor:     types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						LookupAnchorSlot: types.TimeSlot(0),
						Prerequisites:    []types.OpaqueHash{},
					},
					CoreIndex:         types.CoreIndex(0),
					AuthorizerHash:    types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					AuthOutput:        types.ByteSequence(HexToBytes("0x030201")),
					SegmentRootLookup: types.SegmentRootLookup{},
					Results: []types.WorkResult{
						{
							ServiceId:     types.ServiceId(1),
							CodeHash:      types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
							PayloadHash:   types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
							AccumulateGas: types.Gas(42),
							Result: types.WorkExecResult{
								types.WorkExecResultType("ok"): HexToBytes("0x010203"),
							},
						},
					},
				},
				Timeout: types.TimeSlot(42),
			},
		},
	)

	expectedRho := types.AvailabilityAssignments{
		nil,
		{
			Report: types.WorkReport{
				PackageSpec: types.WorkPackageSpec{
					Hash:         types.WorkPackageHash(HexToBytes("0xe12c22d4f162d9a012c9319233da5d3e923cc5e1029b8f90e47249c9ab256b35")),
					Length:       0,
					ErasureRoot:  types.ErasureRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					ExportsRoot:  types.ExportsRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					ExportsCount: types.U16(0),
				},
				Context: types.RefineContext{
					Anchor:           types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					StateRoot:        types.StateRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					BeefyRoot:        types.BeefyRoot(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					LookupAnchor:     types.HeaderHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
					LookupAnchorSlot: types.TimeSlot(0),
					Prerequisites:    []types.OpaqueHash{},
				},
				CoreIndex:         types.CoreIndex(0),
				AuthorizerHash:    types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
				AuthOutput:        types.ByteSequence(HexToBytes("0x030201")),
				SegmentRootLookup: types.SegmentRootLookup{},
				Results: []types.WorkResult{
					{
						ServiceId:     types.ServiceId(1),
						CodeHash:      types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						PayloadHash:   types.OpaqueHash(HexToBytes("0x0000000000000000000000000000000000000000000000000000000000000000")),
						AccumulateGas: types.Gas(42),
						Result: types.WorkExecResult{
							types.WorkExecResultType("ok"): HexToBytes("0x010203"),
						},
					},
				},
			},
			Timeout: types.TimeSlot(42),
		},
	}
	expectedOutput := []types.Ed25519Public{
		types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
		types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
	}

	// initialize the store
	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{})

	output, err := Disputes(disputeExtrinsic)

	if err != nil {
		t.Errorf("Disputes failed: %v", err)
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi
	intermediateRho := store.GetInstance().GetIntermediateStates().GetRhoDagger()
	if intermediateRho[0] != expectedRho[0] {
		t.Errorf("expected rho[0] to be nil, got: %v", intermediateRho[0])
	}
	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}
	for i, offender := range output {
		if !bytes.Equal(offender[:], expectedOutput[i][:]) {
			t.Errorf("offenders mark does not match")
		}
	}
}

func TestDisputeWorkFlow_progress_with_bad_signatures_1(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      false,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x647c04630e911a432f99e6c1108bcf4c06496754033b77c5eb8271a5d06a85e1884db0fb977e232e416643ccfaf4f334e99f3b8d9cdfc65a8e4ecbc9db284005")),
					},
					{
						Vote:      false,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0xb95288deca20fcb649a3515ece2f1d147f8c0f3acef20967cd05a7b770c96e2e0928a056af1aa233c45b0a154e31dae842a50ff48f249cc364af20f282db950a")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x2f8476e2c06dec1fd24130363f922f5419d91cd250d0d9448e8db471e08377a8ee927c8ca9c45ff47796cf0dfb35e4aff4fee96cae8dbe40fbd48d1cd59c7f0c")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0xdf20eab8438b43d774ac84d4225607cdd4159ee495991c89c9dc302e7d826b53e23b1f266a2dcf915dee277a0bfa0b93c957504503213c3f57a4c08b5ce07f0b")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0xdf20eab8438b43d774ac84d4225607cdd4159ee495991c89c9dc302e7d826b53e23b1f266a2dcf915dee277a0bfa0b93c957504503213c3f57a4c08b5ce07f0b")),
					},
				},
			},
		},
		Culprits: []types.Culprit{
			{
				Target:    types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Key:       types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
				Signature: types.Ed25519Signature(HexToBytes("0xc935d19a67d96edd5bff539e11a0340153acb119d4bb83e7f60aba87e6fec4acb236d19a97476924d311390accede172a045978dcb0a65adb528ff0e7f7cf609")),
			},
			{
				Target:    types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Key:       types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
				Signature: types.Ed25519Signature(HexToBytes("0x7ec23531ebc5aa88883a8b8f1f7b1f05ea18fa12f735de02a167309fe41bb0a87f5ff05d7c00c15fbc19f181badd53e0ec25d9a62e742d239c2058f4964b040f")),
			},
		},
		Faults: []types.Fault{},
	}

	expectedPsi := types.DisputesRecords{
		Good:      []types.WorkReportHash{},
		Bad:       []types.WorkReportHash{},
		Wonky:     []types.WorkReportHash{},
		Offenders: []types.Ed25519Public{},
	}

	// initialize the store
	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)

	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{})

	_, err := Disputes(disputeExtrinsic)

	if err == nil {
		t.Errorf("expected an error but got nil")
	} else {
		expectedError := "bad_signature"
		if err.Error() != expectedError {
			t.Errorf("expected error: %v, got: %v", expectedError, err)
		}
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi
	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}
}

func TestDisputeWorkFlow_progress_with_bad_signatures_2(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      false,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x647c04630e911a432f99e6c1108bcf4c06496754033b77c5eb8271a5d06a85e1884db0fb977e232e416643ccfaf4f334e99f3b8d9cdfc65a8e4ecbc9db284005")),
					},
					{
						Vote:      false,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0xb95288deca20fcb649a3515ece2f1d147f8c0f3acef20967cd05a7b770c96e2e0928a056af1aa233c45b0a154e31dae842a50ff48f249cc364af20f282db950a")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x2f8476e2c06dec1fd24130363f922f5419d91cd250d0d9448e8db471e08377a8ee927c8ca9c45ff47796cf0dfb35e4aff4fee96cae8dbe40fbd48d1cd59c7f0c")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x8aa6778092622ad32db2aa10d060a2ad9f7b819af62ff5ada7f88616f81a8ffd0889c64cd857785c1d1a28e39fe2164b71555ecad03ad286f2952144a862c205")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0xdf20eab8438b43d774ac84d4225607cdd4159ee495991c89c9dc302e7d826b53e23b1f266a2dcf915dee277a0bfa0b93c957504503213c3f57a4c08b5ce07f0b")),
					},
				},
			},
		},
		Culprits: []types.Culprit{
			{
				Target:    types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Key:       types.Ed25519Public(HexToBytes("0xb3e0e096b02e2ec98a3441410aeddd78c95e27a0da6f411a09c631c0f2bea6e9")),
				Signature: types.Ed25519Signature(HexToBytes("0xc935d19a67d96edd5bff539e11a0340153acb119d4bb83e7f60aba87e6fec4acb236d19a97476924d311390accede172a045978dcb0a65adb528ff0e7f7cf609")),
			},
			{
				Target:    types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
				Key:       types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
				Signature: types.Ed25519Signature(HexToBytes("0x7ec23531ebc5aa88883a8b8f1f7b1f05ea18fa12f735de02a167309fe41bb0a87f5ff05d7c00c15fbc19f181badd53e0ec25d9a62e742d239c2058f4964b040f")),
			},
		},
		Faults: []types.Fault{},
	}

	expectedPsi := types.DisputesRecords{
		Good: []types.WorkReportHash{},
		Bad: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
		},
		Wonky: []types.WorkReportHash{},
		Offenders: []types.Ed25519Public{
			types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
		},
	}

	// initialize the store
	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)

	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{
		types.WorkReportHash(HexToBytes("0x0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")),
	})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{
		types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
	})

	_, err := Disputes(disputeExtrinsic)

	if err == nil {
		t.Errorf("expected an error but got nil")
	} else {
		expectedError := "bad_signature"
		if err.Error() != expectedError {
			t.Errorf("expected error: %v, got: %v", expectedError, err)
		}
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi
	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}
}

func TestDisputeWorkFlow_ProgressWithCulprit1(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      false,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x826a4bbe7ee3400ffe0f64bdd87ae65aa50d98f48ad6a60da927636cd430ae5d3914d3bc6b87c47c94a9cc5bef84bf30be5534e5c649fc2cd4434918a37a2301")),
					},
					{
						Vote:      false,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0x726e970fff9e9a05a891fe46ec3371b099c7a637fccc0314bdf42f254f868baee58cf902cbd6eda9871f8ac7687aa2a381eaa70e4b0a9a4b7640ecac9b88300e")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x960493c94e625e296d48756050b0f92217ecc4059f369597defccae78d92b0e6628faf6a216f44f95bfc6b2f5f5a192c8c163ceb8b147de21373a658c9f34706")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x2e40c056ccbe227ae2e10c1f07800b7981fba2d9ca80eaec293251b9287e2ed55dcc177cadb1a51929fa0caffe9fb1833aa7566dd4be09ac28eb215428a4a509")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0xeaa2e3d9e334b116fd122d4d79e87955ad1994e005548a6451255f840fab5e19899e1efdff6868500b57beb32449340cbb53073c2d51fc5bce4f42915b1bd60a")),
					},
				},
			},
		},
		Culprits: []types.Culprit{},
		Faults:   []types.Fault{},
	}

	expectedPsi := types.DisputesRecords{
		Good:      []types.WorkReportHash{},
		Bad:       []types.WorkReportHash{},
		Wonky:     []types.WorkReportHash{},
		Offenders: []types.Ed25519Public{},
	}

	// initialize the store
	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)

	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{})

	_, err := Disputes(disputeExtrinsic)

	if err == nil {
		t.Errorf("expected an error but got nil")
	} else {
		expectedError := "not_enough_culprits"
		if err.Error() != expectedError {
			t.Errorf("expected error: %v, got: %v", expectedError, err)
		}
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi
	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}
}

func TestDisputeWorkFlow_ProgressWithVerdicts4(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      true,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x0b1e29dbda5e3bba5dde21c81a8178b115ebf0cf5920fe1a38e897ecadd91718e34bf01c9fc7fdd0df31d83020231b6e8338c8dc204b618cbde16a03cb269d05")),
					},
					{
						Vote:      true,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0x0d44746706e09ff6b6f2929e736c2f868a4d17939af6d37ca7d3c7f6d4914bd095a6fd4ff48c320b673e2de92bfdb5ed9f5c0c40749816ab4171a2272386fc05")),
					},
					{
						Vote:      true,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x0d5d39f2239b775b22aff53b74a0d708a9b9363ed5017170f0abebc8ffd97fc1cc3cf597c578b555ad5abab26e09ecda727c2909feae99587c6354b86e4cc50c")),
					},
					{
						Vote:      true,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x701d277fa78993b343a5d4367f1c2a2fb7ddb77f0246bf9028196feccbb7c0f2bd994966b3e9b1e51ff5dd63d8aa5e2331432b9cca4a125552c4700d51814a04")),
					},
					{
						Vote:      true,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0x08d96d2e49546931dc3de989a69aa0ae3547d67a038bdaa84f7e549da8318d48aab72b4b30ecc0c588696305fce3e2c4657f409463f6a05c52bf641f2684460f")),
					},
				},
			},
			{
				Target: types.OpaqueHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      false,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0xd76bba06ffb8042bedce3f598e22423660e64f2108566cbd548f6d2c42b1a39607a214bddfa7ccccf83fe993728a58393c64283b8a9ab8f3dff49cbc3cc2350e")),
					},
					{
						Vote:      false,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0x77edbe63b2cfab4bda9227bc9fcc8ac4aa8157616c3d8dff9f90fe88cc998fef871a57bbc43eaa1bdee241a1f903ffb42e39a4207c0752d9352f7d98835eda0a")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x1843d18350a8ddee1502bc47cbd1dd30a3354f24bf7e095ad848e8f0744afc4b04a224b5b2143297d571309799c3c0a17f1b7d7782aaeb8f4991cf5dd749310b")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x561bab9479abe38ed5d9609e92145fa689995ef2b71e94577a60eee6177663e8fd1f5bacd1f1afdadce1ea48598ad10a0893e733c34ab6f4aa821b0fdbdf0201")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0xb579159c1ab983583ed8d95bf8632ac7d3be51bdff3d5221258105b801782a5146e08247c269c7bcec10bec76c7d648704e7e6bf3ace77951e828f23894b500c")),
					},
				},
			},
		},
		Culprits: []types.Culprit{
			{
				Target:    types.WorkReportHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
				Key:       types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
				Signature: types.Ed25519Signature(HexToBytes("0xa6a135b2f36906be1c00cd0e48425a38cbde296a5ff73d6de6d3b0e4c26f1761adbf563961da0d3611c24ee8f5c5781647f327513912cb58f1de4bc72b5e6f01")),
			},
			{
				Target:    types.WorkReportHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
				Key:       types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
				Signature: types.Ed25519Signature(HexToBytes("0x940439909168820e32e5788b293786e1e02e7377e32260a96997cb991638c8c88980d0a7a6f26c7fb9bb81282129fdaa09c87932db02cbfd9955dc1940b90a03")),
			},
		},
		Faults: []types.Fault{
			{
				Target:    types.WorkReportHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
				Vote:      false,
				Key:       types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
				Signature: types.Ed25519Signature(HexToBytes("0x826a4bbe7ee3400ffe0f64bdd87ae65aa50d98f48ad6a60da927636cd430ae5d3914d3bc6b87c47c94a9cc5bef84bf30be5534e5c649fc2cd4434918a37a2301")),
			},
		},
	}

	expectedPsi := types.DisputesRecords{
		Good: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
		},
		Bad: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
		},
		Wonky: []types.WorkReportHash{},
		Offenders: []types.Ed25519Public{
			types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
		},
	}
	expectedOutput := []types.Ed25519Public{
		types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
		types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
	}

	// initialize the store
	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)

	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{})

	output, err := Disputes(disputeExtrinsic)

	if err != nil {
		t.Errorf("Disputes failed: %v", err)
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi

	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}

	for i, offender := range output {
		if !bytes.Equal(offender[:], expectedOutput[i][:]) {
			t.Errorf("offenders mark does not match")
		}
	}
}

func TestDisputeWorkFlow_ProgressWithVerdicts6(t *testing.T) {
	disputeExtrinsic := types.DisputesExtrinsic{
		Verdicts: []types.Verdict{
			{
				Target: types.OpaqueHash(HexToBytes("0xe12c22d4f162d9a012c9319233da5d3e923cc5e1029b8f90e47249c9ab256b35")),
				Age:    0,
				Votes: []types.Judgement{
					{
						Vote:      true,
						Index:     0,
						Signature: types.Ed25519Signature(HexToBytes("0x98b07efd76fec53dbfb3ac4cb521c06945a852bbb12c74158cb660f31dcf09cbc297815f0baf7a3c982f3a2bd42242bc1f60230f374106627c66b0bfae0c3400")),
					},
					{
						Vote:      true,
						Index:     1,
						Signature: types.Ed25519Signature(HexToBytes("0xce37e99cce4a9dd4621970a54efb78e69694eea0f9a9ed24dde41b481b0a5984d8705f19beb18396eda280af75e43f45726290ed59dac27f53ba5514c70f7b07")),
					},
					{
						Vote:      false,
						Index:     2,
						Signature: types.Ed25519Signature(HexToBytes("0x9ae03b97c94551c2416494df1e2de19d8f16e029cd93e50c115ed15b51c8a2a23b6e397380597feb0ea5b5f4835e4f73f2e0276aa3b505b484eba6146655300e")),
					},
					{
						Vote:      false,
						Index:     3,
						Signature: types.Ed25519Signature(HexToBytes("0x46ae920062dd4722b67c535d9d30abc493547e553044d9f2ce709910c6eaae15a21cf0da0d0c1d650022c0302800bccd2bd01afabbfebbe5a07ff79878e0020e")),
					},
					{
						Vote:      false,
						Index:     4,
						Signature: types.Ed25519Signature(HexToBytes("0x5b5d5ba6b7e93827c04d1a27ff9542dfac63c80c5377446e0bb3afb919103383a418ac70b7049ff94f10e1589a1da9018a9a983f466aa1d8143f1eb5e11ee30d")),
					},
				},
			},
		},
		Culprits: []types.Culprit{},
		Faults:   []types.Fault{},
	}

	expectedPsi := types.DisputesRecords{
		Good: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
		},
		Bad: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
		},
		Wonky: []types.WorkReportHash{
			types.WorkReportHash(HexToBytes("0xe12c22d4f162d9a012c9319233da5d3e923cc5e1029b8f90e47249c9ab256b35")),
		},
		Offenders: []types.Ed25519Public{
			types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
			types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
			types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
		},
	}

	expectedOutput := []types.Ed25519Public{}

	// initialize the store
	store.GetInstance().GetPriorStates().SetKappa(Kappa)
	store.GetInstance().GetPriorStates().SetLambda(Lambda)

	store.GetInstance().GetPosteriorStates().SetPsiG([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiB([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPosteriorStates().SetPsiO([]types.Ed25519Public{})

	store.GetInstance().GetPriorStates().SetPsiG([]types.WorkReportHash{
		types.WorkReportHash(HexToBytes("0x11da6d1f761ddf9bdb4c9d6e5303ebd41f61858d0a5647a1a7bfe089bf921be9")),
	})
	store.GetInstance().GetPriorStates().SetPsiB([]types.WorkReportHash{
		types.WorkReportHash(HexToBytes("0x7b0aa1735e5ba58d3236316c671fe4f00ed366ee72417c9ed02a53a8019e85b8")),
	})
	store.GetInstance().GetPriorStates().SetPsiW([]types.WorkReportHash{})
	store.GetInstance().GetPriorStates().SetPsiO([]types.Ed25519Public{
		types.Ed25519Public(HexToBytes("0x22351e22105a19aabb42589162ad7f1ea0df1c25cebf0e4a9fcd261301274862")),
		types.Ed25519Public(HexToBytes("0x3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")),
		types.Ed25519Public(HexToBytes("0xe68e0cf7f26c59f963b5846202d2327cc8bc0c4eff8cb9abd4012f9a71decf00")),
	})

	output, err := Disputes(disputeExtrinsic)

	if err != nil {
		t.Errorf("Disputes failed: %v", err)
	}

	posteriorPsi := store.GetInstance().GetPosteriorState().Psi

	if err := compareDisputesRecords(posteriorPsi, expectedPsi); err != nil {
		t.Errorf("posteriorPsi does not match expectedPsi: %v", err)
	}

	for i, offender := range output {
		if !bytes.Equal(offender[:], expectedOutput[i][:]) {
			t.Errorf("offenders mark does not match")
		}
	}
}

func compareDisputesRecords(a, b types.DisputesRecords) error {
	if len(a.Good) != len(b.Good) {
		return fmt.Errorf("length mismatch in Good: %d != %d", len(a.Good), len(b.Good))
	}
	if len(a.Bad) != len(b.Bad) {
		return fmt.Errorf("length mismatch in Bad: %d != %d", len(a.Bad), len(b.Bad))
	}
	if len(a.Wonky) != len(b.Wonky) {
		return fmt.Errorf("length mismatch in Wonky: %d != %d", len(a.Wonky), len(b.Wonky))
	}

	for i := range a.Good {
		if a.Good[i] != b.Good[i] {
			return fmt.Errorf("mismatch in Good at index %d: %x != %x", i, a.Good[i], b.Good[i])
		}
	}

	for i := range a.Bad {
		if a.Bad[i] != b.Bad[i] {
			return fmt.Errorf("mismatch in Bad at index %d: %x != %x", i, a.Bad[i], b.Bad[i])
		}
	}

	for i := range a.Wonky {
		if a.Wonky[i] != b.Wonky[i] {
			return fmt.Errorf("mismatch in Wonky at index %d: %x != %x", i, a.Wonky[i], b.Wonky[i])
		}
	}

	if !arraysContainSameElements(a.Offenders, b.Offenders) {
		return fmt.Errorf("arrays do not contain the same elements")
	}
	return nil
}

func arraysContainSameElements(a, b []types.Ed25519Public) bool {
	if len(a) != len(b) {
		return false
	}

	counts := make(map[types.Ed25519Public]int)
	for _, item := range a {
		counts[item]++
	}
	for _, item := range b {
		if counts[item] == 0 {
			return false
		}
		counts[item]--
	}
	return true
}
