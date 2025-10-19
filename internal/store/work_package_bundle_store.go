package store

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// WorkPackageBundleStore manages the storage and retrieval of work package bundles.
type WorkPackageBundleStore struct {
	client *RedisClient
}

// NewWorkPackageBundleStore creates a new work package bundle store.
func NewWorkPackageBundleStore(client *RedisClient) *WorkPackageBundleStore {
	return &WorkPackageBundleStore{client: client}
}

// Save stores a work package bundle with its erasure root as the key.
func (s *WorkPackageBundleStore) Save(bundle *types.WorkPackageBundle) error {
	erasureRoot, err := s.computeErasureRoot(bundle)
	if err != nil {
		return fmt.Errorf("failed to compute erasure root: %w", err)
	}

	encoder := types.NewEncoder()
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return fmt.Errorf("failed to encode bundle: %w", err)
	}

	key := "wp_bundle:" + hex.EncodeToString(erasureRoot[:])
	ttl := types.SegmentErasureTTL
	return s.client.PutWithTTL(key, bundleBytes, ttl)
}

// Get retrieves a work package bundle by its erasure root.
func (s *WorkPackageBundleStore) Get(erasureRoot []byte) (*types.WorkPackageBundle, error) {
	key := "wp_bundle:" + hex.EncodeToString(erasureRoot)
	bundleBytes, err := s.client.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle from store: %w", err)
	}
	if bundleBytes == nil {
		return nil, fmt.Errorf("work package bundle not found for erasure root: %x", erasureRoot)
	}

	var bundle types.WorkPackageBundle
	decoder := types.NewDecoder()
	err = decoder.Decode(bundleBytes, &bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bundle: %w", err)
	}

	return &bundle, nil
}

// computeErasureRoot computes the erasure root for a work package bundle.
func (s *WorkPackageBundleStore) computeErasureRoot(bundle *types.WorkPackageBundle) (types.OpaqueHash, error) {
	encoder := types.NewEncoder()
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return types.OpaqueHash{}, fmt.Errorf("failed to encode bundle: %w", err)
	}

	var exportsData []types.ExportSegment
	for _, row := range bundle.ImportSegments {
		exportsData = append(exportsData, row...)
	}

	erasureRoot, err := s.computeErasureRootFromData(bundleBytes, exportsData)
	if err != nil {
		return types.OpaqueHash{}, fmt.Errorf("failed to compute erasure root from data: %w", err)
	}

	return erasureRoot, nil
}

// computeErasureRootFromData computes the erasure root from bundle bytes and export segments.
func (s *WorkPackageBundleStore) computeErasureRootFromData(bundleBytes []byte, exportsData []types.ExportSegment) (types.OpaqueHash, error) {
	// TODO: import the work_package package and use its functions directly
	// For now, we'll create a mock erasure root
	mockErasureRoot := hash.Blake2bHash(types.ByteSequence(bundleBytes))
	return mockErasureRoot, nil
}

// SaveWithMetadata stores a work package bundle with additional metadata.
func (s *WorkPackageBundleStore) SaveWithMetadata(bundle *types.WorkPackageBundle, metadata map[string]interface{}) error {
	erasureRoot, err := s.computeErasureRoot(bundle)
	if err != nil {
		return fmt.Errorf("failed to compute erasure root: %w", err)
	}

	encoder := types.NewEncoder()
	bundleBytes, err := encoder.Encode(bundle)
	if err != nil {
		return fmt.Errorf("failed to encode bundle: %w", err)
	}

	combined := struct {
		Bundle   []byte                 `json:"bundle"`
		Metadata map[string]interface{} `json:"metadata"`
		Created  time.Time              `json:"created"`
	}{
		Bundle:   bundleBytes,
		Metadata: metadata,
		Created:  time.Now(),
	}

	combinedBytes, err := json.Marshal(combined)
	if err != nil {
		return fmt.Errorf("failed to marshal combined data: %w", err)
	}

	key := "wp_bundle:" + hex.EncodeToString(erasureRoot[:])
	ttl := types.SegmentErasureTTL
	return s.client.PutWithTTL(key, combinedBytes, ttl)
}

// GetWithMetadata retrieves a work package bundle with its metadata.
func (s *WorkPackageBundleStore) GetWithMetadata(erasureRoot []byte) (*types.WorkPackageBundle, map[string]interface{}, error) {
	key := "wp_bundle:" + hex.EncodeToString(erasureRoot)
	combinedBytes, err := s.client.Get(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get bundle from store: %w", err)
	}
	if combinedBytes == nil {
		return nil, nil, fmt.Errorf("work package bundle not found for erasure root: %x", erasureRoot)
	}

	// Try to decode as combined structure first
	var combined struct {
		Bundle   []byte                 `json:"bundle"`
		Metadata map[string]interface{} `json:"metadata"`
		Created  time.Time              `json:"created"`
	}

	if err := json.Unmarshal(combinedBytes, &combined); err == nil {
		var bundle types.WorkPackageBundle
		decoder := types.NewDecoder()
		err = decoder.Decode(combined.Bundle, &bundle)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode bundle: %w", err)
		}
		return &bundle, combined.Metadata, nil
	}

	// Fallback: try to decode as just bundle bytes
	var bundle types.WorkPackageBundle
	decoder := types.NewDecoder()
	err = decoder.Decode(combinedBytes, &bundle)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode bundle: %w", err)
	}

	return &bundle, nil, nil
}
