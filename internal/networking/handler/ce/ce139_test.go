package ce

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

func TestHandleSegmentShardRequest(t *testing.T) {
	// Erasure root (32 bytes) + shard index (4 bytes) + segment indices length (2 bytes) + segment indices + FIN
	erasureRoot := make([]byte, 32)
	for i := range erasureRoot {
		erasureRoot[i] = byte(i)
	}

	shardIndex := uint32(0)
	segmentIndices := []uint16{0, 1}
	segmentIndicesLen := uint16(len(segmentIndices))

	var buf bytes.Buffer

	buf.Write(erasureRoot)

	shardIndexBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shardIndexBytes, shardIndex)
	buf.Write(shardIndexBytes)

	segmentIndicesLenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(segmentIndicesLenBytes, segmentIndicesLen)
	buf.Write(segmentIndicesLenBytes)

	for _, index := range segmentIndices {
		indexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(indexBytes, index)
		buf.Write(indexBytes)
	}

	buf.Write([]byte("FIN"))

	mockStream := newMockStream(buf.Bytes())

	err := HandleSegmentShardRequest_Assurer(nil, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleSegmentShardRequest failed: %v", err)
	}

	response := mockStream.w.Bytes()

	// Check that we got a response with actual segment shards
	// The response should contain the concatenated segment shards (2 segments * 32 bytes = 64 bytes minimum)
	if len(response) < 64 {
		t.Errorf("Expected response to contain segment shards, got length %d", len(response))
	}

	if !bytes.HasSuffix(response, []byte("FIN")) {
		t.Error("Response does not end with FIN")
	}
}

func TestLookupWorkPackageBundleWithRedis(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer os.Unsetenv("USE_MINI_REDIS")

	store.ResetInstance()

	erasureRoot := make([]byte, 32)
	for i := range erasureRoot {
		erasureRoot[i] = byte(i)
	}

	bundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		t.Fatalf("lookupWorkPackageBundle failed: %v", err)
	}
	if bundle == nil {
		t.Fatal("Expected bundle to be returned, got nil")
	}

	storeInstance := store.GetInstance()
	if storeInstance == nil {
		t.Fatal("Store instance should be initialized")
	}

	workPackageBundleStore := storeInstance.GetWorkPackageBundleStore()
	if workPackageBundleStore != nil {
		t.Fatal("Expected work package bundle store to be nil initially")
	}

	bundle, err = lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		t.Fatalf("lookupWorkPackageBundle failed: %v", err)
	}
	if bundle == nil {
		t.Fatal("Expected bundle to be returned, got nil")
	}

	if bundle.Package.Authorization == nil {
		t.Error("Expected bundle to have authorization data")
	}
	if len(bundle.Package.Items) == 0 {
		t.Error("Expected bundle to have work items")
	}

	redisBackend, err := store.GetRedisBackend()
	if err != nil {
		t.Fatalf("Failed to get Redis backend: %v", err)
	}

	workPackageBundleStore = store.NewWorkPackageBundleStore(redisBackend.GetClient())
	storeInstance.SetWorkPackageBundleStore(workPackageBundleStore)

	testBundle := CreateTestWorkPackageBundle()

	err = workPackageBundleStore.Save(testBundle)
	if err != nil {
		t.Fatalf("Failed to save test bundle: %v", err)
	}

	retrievedBundle, err := lookupWorkPackageBundle(erasureRoot)
	if err != nil {
		t.Fatalf("lookupWorkPackageBundle failed after storing bundle: %v", err)
	}
	if retrievedBundle == nil {
		t.Fatal("Expected retrieved bundle to be returned, got nil")
	}

	if len(retrievedBundle.Package.Items) != len(testBundle.Package.Items) {
		t.Errorf("Expected %d work items, got %d", len(testBundle.Package.Items), len(retrievedBundle.Package.Items))
	}

	store.CloseMiniRedis()
}
