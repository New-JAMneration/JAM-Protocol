package main

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/require"
)

func TestNode_SetupJAMProtocol_SeedsGenesisFromChainSpec_ToRedis(t *testing.T) {
	chainPath := "./test_data/dev.chainspec.json"
	// chainPath := "./test_data/jamduna-spec.json"

	spec, err := store.GetChainSpecFromJson(chainPath)
	require.NoError(t, err)

	pp, err := spec.ParseProtocolParameters()
	require.NoError(t, err)
	require.NoError(t, types.ApplyProtocolParameters(pp))

	hdrBytes, err := spec.GenesisHeaderBytes()
	require.NoError(t, err)
	hdr, err := store.DecodeHeaderFromBin(hdrBytes)
	require.NoError(t, err)

	kvs, err := spec.GenesisStateKeyVals()
	require.NoError(t, err)

	wantInputKVs, wantRoot, err := store.BuildStateRootInputKeyValsAndRoot(kvs)
	require.NoError(t, err)

	genesisHashBytes, err := hash.ComputeBlockHeaderHash(*hdr)
	require.NoError(t, err)
	wantGenesisHash := genesisHashBytes

	SetupJAMProtocol(chainPath)

	redisBackend, err := store.GetRedisBackend()
	require.NoError(t, err)

	ctx := context.Background()

	stateRootKey := "state_root:" + hex.EncodeToString(wantGenesisHash[:])
	gotRootBytes, err := redisBackend.DebugGetBytes(ctx, stateRootKey)
	require.NoError(t, err)
	require.Len(t, gotRootBytes, 32)

	var gotRoot types.StateRoot
	copy(gotRoot[:], gotRootBytes)
	require.Equal(t, wantRoot, gotRoot)

	stateDataKey := "state_data:" + hex.EncodeToString(wantRoot[:])
	raw, err := redisBackend.DebugGetBytes(ctx, stateDataKey)
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	var decoded types.StateKeyVals
	dec := types.NewDecoder()
	require.NoError(t, dec.Decode(raw, &decoded))

	require.Equal(t, len(wantInputKVs), len(decoded))
	for i := range decoded {
		require.Equal(t, wantInputKVs[i].Key, decoded[i].Key)
		require.Equal(t, wantInputKVs[i].Value, decoded[i].Value)
	}
}
