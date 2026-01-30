package store

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func (repo *Repository) GetHashSegmentMap(r database.Reader) (map[types.OpaqueHash]types.OpaqueHash, error) {
	result := make(map[types.OpaqueHash]types.OpaqueHash)

	data, found, err := r.Get(hashSegmentMapKey())
	if err != nil {
		return result, err
	}
	if !found || data == nil {
		return result, nil
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hash segment map: %w", err)
	}

	for k, v := range raw {
		parts := strings.SplitN(k, "_", 2)
		if len(parts) != 2 {
			continue
		}
		var wpHash, segmentRoot types.OpaqueHash
		wpBytes, err := hex.DecodeString(parts[1])
		if err != nil {
			continue
		}
		rootBytes, err := hex.DecodeString(v)
		if err != nil {
			continue
		}
		copy(wpHash[:], wpBytes)
		copy(segmentRoot[:], rootBytes)
		result[wpHash] = segmentRoot
	}
	return result, nil
}

func (repo *Repository) SetHashSegmentMap(w database.Writer, hashSegmentMap map[string]string) error {
	encoded, err := json.Marshal(hashSegmentMap)
	if err != nil {
		return fmt.Errorf("failed to marshal hash segment map: %w", err)
	}
	return w.Put(hashSegmentMapKey(), encoded)
}

func (repo *Repository) SetHashSegmentMapWithLimit(r database.Reader, w database.Writer, wpHash, segmentRoot types.OpaqueHash) (map[types.OpaqueHash]types.OpaqueHash, error) {
	existingBytes, found, err := r.Get(hashSegmentMapKey())
	dict := make(map[string]string)

	if err == nil && found && existingBytes != nil {
		if err := json.Unmarshal(existingBytes, &dict); err != nil {
			return nil, fmt.Errorf("failed to unmarshal existing hash segment map: %w", err)
		}
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	dict[timestamp+"_"+hex.EncodeToString(wpHash[:])] = hex.EncodeToString(segmentRoot[:])

	if len(dict) > 8 {
		keys := make([]string, 0, len(dict))
		for k := range dict {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		delete(dict, keys[0])
	}

	encoded, err := json.Marshal(dict)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hash segment map: %w", err)
	}
	if err := w.Put(hashSegmentMapKey(), encoded); err != nil {
		return nil, fmt.Errorf("failed to put hash segment map: %w", err)
	}

	final := make(map[types.OpaqueHash]types.OpaqueHash)
	for k, v := range dict {
		parts := strings.SplitN(k, "_", 2)
		if len(parts) != 2 {
			continue
		}
		var wph, sr types.OpaqueHash
		wpBytes, err := hex.DecodeString(parts[1])
		if err != nil {
			continue
		}
		segBytes, err := hex.DecodeString(v)
		if err != nil {
			continue
		}
		copy(wph[:], wpBytes)
		copy(sr[:], segBytes)
		final[wph] = sr
	}
	return final, nil
}
