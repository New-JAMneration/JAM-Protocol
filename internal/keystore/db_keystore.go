package keystore

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Compile-time interface check
var _ KeyStore = (*DBKeyStore)(nil)

type DBKeyStore struct {
	storage Storage
}

func NewDBKeyStore(storage Storage) *DBKeyStore {
	return &DBKeyStore{storage: storage}
}

// Key storage format:
//   - keystore:{keyType}:{pubKeyHex} => {keyType}:{privateKeyHex}
//     (prefix with keyType to support different key formats)
//   - keystore:{keyType}:keys => set(publicKeyHex)
const keyPrefix = "keystore"

func (r *DBKeyStore) Generate(t KeyType) (KeyPair, error) {
	var kp KeyPair
	var err error

	switch t {
	case KeyTypeEd25519:
		kp, err = NewEd25519KeyPair()
	case KeyTypeBandersnatch:
		kp, err = NewBandersnatchKeyPair()
	// TODO: BLS
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	return r.storeKeyPair(t, kp)
}

func (r *DBKeyStore) Import(t KeyType, seed []byte) (KeyPair, error) {
	var kp KeyPair
	var err error

	switch t {
	case KeyTypeEd25519:
		kp, err = ImportEd25519KeyPair(seed)
	case KeyTypeBandersnatch:
		kp, err = ImportBandersnatchKeyPair(seed)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	// Check if key already exists
	exists, err := r.Contains(t, kp.PublicKey())
	if err != nil {
		return nil, fmt.Errorf("failed to check key existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: key already exists", ErrKeyExists)
	}

	return r.storeKeyPair(t, kp)
}

// storeKeyPair stores a key pair atomically using transaction
func (r *DBKeyStore) storeKeyPair(t KeyType, kp KeyPair) (KeyPair, error) {
	pubHex := hex.EncodeToString(kp.PublicKey())
	privHex := hex.EncodeToString(kp.PrivateKey())

	// Store with keyType prefix for proper deserialization
	// Format: {keyType}:{privateKeyHex}
	storedValue := fmt.Sprintf("%s:%s", t, privHex)

	dataKey := fmt.Sprintf("%s:%s:%s", keyPrefix, t, pubHex)
	setKey := fmt.Sprintf("%s:%s:keys", keyPrefix, t)

	// Use transaction for atomicity
	tx, err := r.storage.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := tx.Put(dataKey, []byte(storedValue)); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	if err := tx.SetAdd(setKey, []byte(pubHex)); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to add to key set: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return kp, nil
}

func (r *DBKeyStore) Contains(t KeyType, pubKey []byte) (bool, error) {
	publicKeyHex := hex.EncodeToString(pubKey)
	setKey := fmt.Sprintf("%s:%s:keys", keyPrefix, t)

	return r.storage.SetIsMember(setKey, []byte(publicKeyHex))
}

func (r *DBKeyStore) Get(t KeyType, pubKey []byte) (KeyPair, error) {
	pubHex := hex.EncodeToString(pubKey)
	key := fmt.Sprintf("%s:%s:%s", keyPrefix, t, pubHex)

	encoded, err := r.storage.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Parse stored value: {keyType}:{privateKeyHex}
	storedValue := string(encoded)
	parts := strings.SplitN(storedValue, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid stored key format: %s", storedValue)
	}

	storedKeyType := KeyType(parts[0])
	privHex := parts[1]

	// Verify key type matches
	if storedKeyType != t {
		return nil, fmt.Errorf("key type mismatch: expected %s, got %s", t, storedKeyType)
	}

	privBytes, err := hex.DecodeString(privHex)
	if err != nil {
		return nil, fmt.Errorf("invalid stored key: %w", err)
	}

	switch t {
	case KeyTypeEd25519:
		return FromEd25519PrivateKey(privBytes)
	case KeyTypeBandersnatch:
		return ImportBandersnatchKeyPair(privBytes)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
}

func (r *DBKeyStore) List(t KeyType) ([]KeyPair, error) {
	setKey := fmt.Sprintf("%s:%s:keys", keyPrefix, t)
	pubKeys, err := r.storage.SetMembers(setKey)
	if err != nil {
		return nil, err
	}

	if len(pubKeys) == 0 {
		return []KeyPair{}, nil
	}

	// Build all data keys
	dataKeys := make([]string, len(pubKeys))
	for i, pubHexBytes := range pubKeys {
		pubHexStr := string(pubHexBytes)
		dataKeys[i] = fmt.Sprintf("%s:%s:%s", keyPrefix, t, pubHexStr)
	}

	// Get all keys in one batch operation
	keyData, err := r.storage.GetMultiple(dataKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	// Parse and reconstruct key pairs
	result := make([]KeyPair, 0, len(pubKeys))
	for _, pubHexBytes := range pubKeys {
		pubHexStr := string(pubHexBytes)
		dataKey := fmt.Sprintf("%s:%s:%s", keyPrefix, t, pubHexStr)

		encoded, exists := keyData[dataKey]
		if !exists {
			// Key might have been deleted, skip it
			continue
		}

		// Parse stored value
		storedValue := string(encoded)
		parts := strings.SplitN(storedValue, ":", 2)
		if len(parts) != 2 {
			continue // Skip invalid entries
		}

		storedKeyType := KeyType(parts[0])
		privHex := parts[1]

		if storedKeyType != t {
			continue // Skip mismatched types
		}

		privBytes, err := hex.DecodeString(privHex)
		if err != nil {
			continue // Skip invalid hex
		}

		var kp KeyPair
		switch t {
		case KeyTypeEd25519:
			kp, err = FromEd25519PrivateKey(privBytes)
		case KeyTypeBandersnatch:
			kp, err = ImportBandersnatchKeyPair(privBytes)
		default:
			continue
		}

		if err != nil {
			continue // Skip keys that can't be reconstructed
		}

		result = append(result, kp)
	}

	return result, nil
}

// Delete: atomic operation using transaction
func (r *DBKeyStore) Delete(t KeyType, pubKey []byte) error {
	pubHex := hex.EncodeToString(pubKey)
	dataKey := fmt.Sprintf("%s:%s:%s", keyPrefix, t, pubHex)
	setKey := fmt.Sprintf("%s:%s:keys", keyPrefix, t)

	// Use transaction for atomicity
	tx, err := r.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := tx.Delete(dataKey); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete data key: %w", err)
	}

	if err := tx.SetRemove(setKey, []byte(pubHex)); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove pubkey from set: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
