package keystore

import (
	"encoding/hex"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
)

type RedisKeyStore struct {
	client *store.RedisClient
}

func NewRedisKeyStore(client *store.RedisClient) *RedisKeyStore {
	return &RedisKeyStore{client: client}
}

// Redis Key Design:
// - keystore:{keyType}:{pubKeyHex}     => string(privateKeyHex)
// - keystore:{keyType}:keys            => set(publicKeyHex)

func (r *RedisKeyStore) Generate(t KeyType) (KeyPair, error) {
	var kp KeyPair
	var err error

	switch t {
	case KeyTypeEd25519:
		kp, err = NewEd25519KeyPair()
	// TODO: BLS, Bandersnatch
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(kp.PublicKey())
	privHex := hex.EncodeToString(kp.PrivateKey())

	dataKey := fmt.Sprintf("keystore:%s:%s", t, pubHex)
	setKey := fmt.Sprintf("keystore:%s:keys", t)

	if err := r.client.Put(dataKey, []byte(privHex)); err != nil {
		return nil, err
	}
	if err := r.client.SAdd(setKey, []byte(pubHex)); err != nil {
		return nil, err
	}
	return kp, nil
}

func (r *RedisKeyStore) Import(t KeyType, seed []byte) (KeyPair, error) {
	var kp KeyPair
	var err error

	switch t {
	case KeyTypeEd25519:
		kp, err = ImportEd25519KeyPair(seed)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(kp.PublicKey())
	privHex := hex.EncodeToString(kp.PrivateKey())

	dataKey := fmt.Sprintf("keystore:%s:%s", t, pubHex)
	setKey := fmt.Sprintf("keystore:%s:keys", t)

	if err := r.client.Put(dataKey, []byte(privHex)); err != nil {
		return nil, err
	}
	if err := r.client.SAdd(setKey, []byte(pubHex)); err != nil {
		return nil, err
	}
	return kp, nil
}

func (r *RedisKeyStore) Contains(t KeyType, pubKey []byte) (bool, error) {
	publicKeyHex := hex.EncodeToString(pubKey)
	key := fmt.Sprintf("keystore:%s:keys", t)

	return r.client.SIsMember(key, []byte(publicKeyHex))
}

func (r *RedisKeyStore) Get(t KeyType, pubKey []byte) (KeyPair, error) {
	pubHex := hex.EncodeToString(pubKey)
	key := fmt.Sprintf("keystore:%s:%s", t, pubHex)
	encoded, err := r.client.Get(key)
	if err != nil {
		return nil, err
	}
	if encoded == nil {
		return nil, fmt.Errorf("key not found")
	}
	privBytes, err := hex.DecodeString(string(encoded))
	if err != nil {
		return nil, fmt.Errorf("invalid stored key: %w", err)
	}
	switch t {
	case KeyTypeEd25519:
		return FromEd25519PrivateKey(privBytes)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", t)
	}
}

func (r *RedisKeyStore) List(t KeyType) ([]KeyPair, error) {
	setKey := fmt.Sprintf("keystore:%s:keys", t)
	pubKeys, err := r.client.SMembers(setKey)
	if err != nil {
		return nil, err
	}

	var result []KeyPair
	for _, pubHexBytes := range pubKeys {
		pubHexStr := string(pubHexBytes)
		pubKey, err := hex.DecodeString(pubHexStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pub key hex %q: %w", pubHexStr, err)
		}

		kp, err := r.Get(t, pubKey)
		if err != nil {
			return nil, err
		}
		result = append(result, kp)
	}
	return result, nil
}

func (r *RedisKeyStore) Delete(t KeyType, pubKey []byte) error {
	pubHex := hex.EncodeToString(pubKey)
	dataKey := fmt.Sprintf("keystore:%s:%s", t, pubHex)
	setKey := fmt.Sprintf("keystore:%s:keys", t)

	if err := r.client.Delete(dataKey); err != nil {
		return fmt.Errorf("failed to delete data key: %w", err)
	}
	if err := r.client.SRem(setKey, []byte(pubHex)); err != nil {
		return fmt.Errorf("failed to remove pubkey from set: %w", err)
	}
	return nil
}
