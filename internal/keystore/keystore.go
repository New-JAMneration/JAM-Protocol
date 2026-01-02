package keystore

type KeyType string

const (
	KeyTypeEd25519      KeyType = "ed25519"
	KeyTypeBLS          KeyType = "bls"
	KeyTypeBandersnatch KeyType = "bandersnatch"
)

type KeyPair interface {
	Type() KeyType
	PublicKey() []byte
	PrivateKey() []byte
	Sign(message []byte) ([]byte, error)
	Verify(message []byte, signature []byte) bool
}

type KeyStore interface {
	Generate(t KeyType) (KeyPair, error)
	Import(t KeyType, seed []byte) (KeyPair, error)
	Contains(t KeyType, publicKey []byte) (bool, error)
	Get(t KeyType, publicKey []byte) (KeyPair, error)
	List(t KeyType) ([]KeyPair, error)
	Delete(t KeyType, publicKey []byte) error
}
