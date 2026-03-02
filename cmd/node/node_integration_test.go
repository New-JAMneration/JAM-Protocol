package main

import (
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
	"github.com/stretchr/testify/require"
)

func TestNode_SetupJAMProtocol_SeedsGenesisFromChainSpec_ToRedis(t *testing.T) {
	chainPath := "./test_data/dev.chainspec.json"
	// chainPath := "./test_data/jamduna-spec.json"

	spec, err := blockchain.GetChainSpecFromJson(chainPath)
	require.NoError(t, err)

	pp, err := spec.ParseProtocolParameters()
	require.NoError(t, err)
	require.NoError(t, types.ApplyProtocolParameters(pp))

	hdrBytes, err := spec.GenesisHeaderBytes()
	require.NoError(t, err)
	hdr, err := blockchain.DecodeHeaderFromBin(hdrBytes)
	require.NoError(t, err)

	kvs, err := spec.GenesisStateKeyVals()
	require.NoError(t, err)

	wantInputKVs, wantRoot, err := blockchain.GetInstance().BuildStateRootInputKeyValsAndRoot(kvs)
	require.NoError(t, err)

	genesisHashBytes, err := hash.ComputeBlockHeaderHash(*hdr)
	require.NoError(t, err)
	wantGenesisHash := genesisHashBytes

	SetupJAMProtocol(chainPath)

	cs := blockchain.GetInstance()
	gotRoot, err := cs.GetStateRootByBlockHash(types.HeaderHash(wantGenesisHash))
	require.NoError(t, err)
	require.Equal(t, wantRoot, gotRoot)

	raw, err := cs.GetStateByBlockHash(types.HeaderHash(wantGenesisHash))
	require.NoError(t, err)
	require.NotEmpty(t, raw)

	decoded := raw

	require.Equal(t, len(wantInputKVs), len(decoded))
	for i := range decoded {
		require.Equal(t, wantInputKVs[i].Key, decoded[i].Key)
		require.Equal(t, wantInputKVs[i].Value, decoded[i].Value)
	}
}
