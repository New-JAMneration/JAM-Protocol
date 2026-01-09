package local

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/keystore"
)

type LocalKeyStore struct {
	basePath string
	mu       sync.RWMutex
}

// keyData structure for JSON storage
type keyData struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// NewLocalKeyStore creates a new local file-based keystore at the specified path
func NewLocalKeyStore(basePath string) (*LocalKeyStore, error) {
	if err := os.MkdirAll(filepath.Join(basePath, "keystore"), 0700); err != nil {
		return nil, fmt.Errorf("failed to create keystore directory: %w", err)
	}
	return &LocalKeyStore{basePath: basePath}, nil
}

// getKeyStoreFilePath returns the path to the keys index file for a given key type
func (l *LocalKeyStore) getKeyStoreFilePath(t keystore.KeyType) string {
	return filepath.Join(l.basePath, "keystore", string(t), "keys.json")
}

// getKeyTypeDir returns the directory path for a key type
func (l *LocalKeyStore) getKeyTypeDir(t keystore.KeyType) string {
	return filepath.Join(l.basePath, "keystore", string(t))
}

// loadKeys loads the keys index from disk
func (l *LocalKeyStore) loadKeys(t keystore.KeyType) (map[string]keyData, error) {
	keysFile := l.getKeyStoreFilePath(t)
	data, err := os.ReadFile(keysFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]keyData), nil
		}
		return nil, fmt.Errorf("failed to read keys file: %w", err)
	}

	var keys map[string]keyData
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keys: %w", err)
	}
	return keys, nil
}

// saveKeys saves the keys to disk
func (l *LocalKeyStore) saveKeys(t keystore.KeyType, keys map[string]keyData) error {
	keyTypeDir := l.getKeyTypeDir(t)
	if err := os.MkdirAll(keyTypeDir, 0700); err != nil {
		return fmt.Errorf("failed to create key type directory: %w", err)
	}

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %w", err)
	}

	keysFile := l.getKeyStoreFilePath(t)
	if err := os.WriteFile(keysFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write keys file: %w", err)
	}
	return nil
}

// Generate generates a new keypair and stores it
func (l *LocalKeyStore) Generate(t keystore.KeyType) (keystore.KeyPair, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var kp keystore.KeyPair
	var err error

	switch t {
	case keystore.KeyTypeEd25519:
		kp, err = keystore.NewEd25519KeyPair()
	case keystore.KeyTypeBandersnatch:
		kp, err = keystore.NewBandersnatchKeyPair()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	keys, err := l.loadKeys(t)
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(kp.PublicKey())
	privHex := hex.EncodeToString(kp.PrivateKey())

	keys[pubHex] = keyData{
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}

	if err := l.saveKeys(t, keys); err != nil {
		return nil, err
	}

	return kp, nil
}

// Import imports a keypair from a seed and stores it
func (l *LocalKeyStore) Import(t keystore.KeyType, seed []byte) (keystore.KeyPair, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var kp keystore.KeyPair
	var err error

	switch t {
	case keystore.KeyTypeEd25519:
		kp, err = keystore.ImportEd25519KeyPair(seed)
	case keystore.KeyTypeBandersnatch:
		kp, err = keystore.ImportBandersnatchKeyPair(seed)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	keys, err := l.loadKeys(t)
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(kp.PublicKey())
	privHex := hex.EncodeToString(kp.PrivateKey())

	keys[pubHex] = keyData{
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}

	if err := l.saveKeys(t, keys); err != nil {
		return nil, err
	}

	return kp, nil
}

// Contains checks if a public key exists in the keystore
func (l *LocalKeyStore) Contains(t keystore.KeyType, publicKey []byte) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys, err := l.loadKeys(t)
	if err != nil {
		return false, err
	}

	pubHex := hex.EncodeToString(publicKey)
	_, exists := keys[pubHex]
	return exists, nil
}

// Get retrieves a keypair by public key
func (l *LocalKeyStore) Get(t keystore.KeyType, pubKey []byte) (keystore.KeyPair, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys, err := l.loadKeys(t)
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(pubKey)
	kd, exists := keys[pubHex]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	privBytes, err := hex.DecodeString(kd.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid stored key: %w", err)
	}

	switch t {
	case keystore.KeyTypeEd25519:
		return keystore.FromEd25519PrivateKey(privBytes)
	case keystore.KeyTypeBandersnatch:
		return keystore.ImportBandersnatchKeyPair(privBytes)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
}

// List returns all keypairs of a given type
func (l *LocalKeyStore) List(t keystore.KeyType) ([]keystore.KeyPair, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var err error
	keys, err := l.loadKeys(t)
	if err != nil {
		return nil, err
	}

	var kps []keystore.KeyPair
	var kp keystore.KeyPair

	for _, kd := range keys {
		privBytes, err := hex.DecodeString(kd.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key hex: %w", err)
		}

		switch t {
		case keystore.KeyTypeEd25519:
			kp, err = keystore.FromEd25519PrivateKey(privBytes)
		case keystore.KeyTypeBandersnatch:
			kp, err = keystore.ImportBandersnatchKeyPair(privBytes)
		default:
			return nil, fmt.Errorf("unsupported key type: %s", t)
		}
		if err != nil {
			return nil, err
		}
		kps = append(kps, kp)
	}
	return kps, nil
}

// Delete removes a keypair from the keystore
func (l *LocalKeyStore) Delete(t keystore.KeyType, pubKey []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	keys, err := l.loadKeys(t)
	if err != nil {
		return err
	}

	pubHex := hex.EncodeToString(pubKey)
	delete(keys, pubHex)

	if err := l.saveKeys(t, keys); err != nil {
		return fmt.Errorf("failed to save keys after deletion: %w", err)
	}
	return nil
}

// ImportValidatorKeysFromSeed imports both Ed25519 and Bandersnatch keys
// from a JIP-5 seed into the local keystore.
func (l *LocalKeyStore) ImportValidatorKeysFromSeed(seed []byte) error {
	// Derive the secret seeds directly
	ed25519SecretSeed, bandersnatchSecretSeed, _, _, err := keystore.DeriveValidatorKeys(seed)
	if err != nil {
		return fmt.Errorf("failed to derive validator keys: %w", err)
	}

	// Import Ed25519 key using the derived seed (32 bytes)
	if _, err := l.Import(keystore.KeyTypeEd25519, ed25519SecretSeed); err != nil {
		return fmt.Errorf("failed to import Ed25519 key: %w", err)
	}

	// Import Bandersnatch key using the derived secret seed (32 bytes)
	if _, err := l.Import(keystore.KeyTypeBandersnatch, bandersnatchSecretSeed); err != nil {
		return fmt.Errorf("failed to import Bandersnatch key: %w", err)
	}

	return nil
}
