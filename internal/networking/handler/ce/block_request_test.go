package ce

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// --- fakeBlockchain is a fake implementation of blockchain.Blockchain.
type fakeBlockchain struct {
	blocks            map[types.HeaderHash]types.Block
	blockNumberToHash map[uint32][]types.HeaderHash
	hashToblockNumber map[types.HeaderHash]uint32
	genesis           types.HeaderHash
	currentHead       types.HeaderHash
}

func (f *fakeBlockchain) GetBlockNumber(hash types.HeaderHash) (uint32, error) {
	blocknumber, ok := f.hashToblockNumber[hash]
	if !ok {
		return 0, fmt.Errorf("block not found: %v", hex.EncodeToString(hash[:]))
	}
	return blocknumber, nil
}

func (f *fakeBlockchain) GetBlockHashByNumber(number uint32) (res []types.HeaderHash, err error) {
	res, ok := f.blockNumberToHash[number]
	if !ok {
		return nil, errors.New("block not found")
	}
	return res, nil
}

func (f *fakeBlockchain) GetBlock(hash types.HeaderHash) (types.Block, error) {
	blk, ok := f.blocks[hash]
	if !ok {
		return types.Block{}, fmt.Errorf("block not found: %v", hex.EncodeToString(hash[:]))
	}
	return blk, nil
}

func (f *fakeBlockchain) GenesisBlockHash() types.HeaderHash {
	return f.genesis
}

func (f *fakeBlockchain) GetCurrentHead() (types.Block, error) {
	return f.blocks[f.currentHead], nil
}

func (f *fakeBlockchain) SetCurrentHead(hash types.HeaderHash) {
	f.currentHead = hash
}

// SetupFakeBlockchain creates a simple chain:
// genesis (slot 0, parent = genesis) → block1 (slot 1, parent = genesis)
// → block2 (slot 2, parent = block1) → block3 (slot 3, parent = block2)
func SetupFakeBlockchain() *fakeBlockchain {
	var genesisHash, block1Hash, block2Hash, block3Hash types.HeaderHash
	genesisHash[0] = 0
	block1Hash[0] = 1
	block2Hash[0] = 2
	block3Hash[0] = 3

	genesisBlock := types.Block{
		Header: types.Header{
			Slot:   0,
			Parent: types.HeaderHash{},
		},
		Extrinsic: types.Extrinsic{},
	}
	// Block1: parent = genesis.
	block1 := types.Block{
		Header: types.Header{
			Slot:   1,
			Parent: genesisHash,
		},
		Extrinsic: types.Extrinsic{},
	}
	// Block2: parent = block1.
	block2 := types.Block{
		Header: types.Header{
			Slot:   2,
			Parent: block1Hash,
		},
		Extrinsic: types.Extrinsic{},
	}
	// Block3: parent = block2.
	block3 := types.Block{
		Header: types.Header{
			Slot:   3,
			Parent: block2Hash,
		},
		Extrinsic: types.Extrinsic{},
	}
	fb := &fakeBlockchain{
		blocks:            make(map[types.HeaderHash]types.Block),
		blockNumberToHash: make(map[uint32][]types.HeaderHash),
		hashToblockNumber: make(map[types.HeaderHash]uint32),
		genesis:           genesisHash,
	}

	fb.blocks[genesisHash] = genesisBlock
	fb.blocks[block1Hash] = block1
	fb.blocks[block2Hash] = block2
	fb.blocks[block3Hash] = block3

	// blockNumberToHash
	fb.blockNumberToHash[0] = []types.HeaderHash{genesisHash}
	fb.blockNumberToHash[1] = []types.HeaderHash{block1Hash}
	fb.blockNumberToHash[2] = []types.HeaderHash{block2Hash}
	fb.blockNumberToHash[3] = []types.HeaderHash{block3Hash}

	// hashToblockNumber
	fb.hashToblockNumber[genesisHash] = 0
	fb.hashToblockNumber[block1Hash] = 1
	fb.hashToblockNumber[block2Hash] = 2
	fb.hashToblockNumber[block3Hash] = 3

	fb.SetCurrentHead(block3Hash)

	return fb
}

// Helper function to frame a payload with a 4-byte little-endian length prefix.
func framePayload(payload []byte) []byte {
	buf := new(bytes.Buffer)
	lenPrefix := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenPrefix, uint32(len(payload)))
	buf.Write(lenPrefix)
	buf.Write(payload)
	return buf.Bytes()
}

// generateTLSConfig returns a real TLS configuration with a self-signed certificate.
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate RSA key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-proto"},
	}
}

// TestRealQuicStreamBlockRequest uses a real QUIC connection to test the block request handler.
func TestRealQuicStreamBlockRequest(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true") // Set environment variable to enable test mode
	defer store.CloseMiniRedis()

	// Setup TLS configurations.
	clientTLS, err := quic.NewTLSConfig(false, false)
	clientTLS.InsecureSkipVerify = true
	if err != nil {
		t.Fatalf("Client TLS config error: %v", err)
	}

	// Listen on an ephemeral port.
	listener, err := quic.NewListener("localhost:0", false, quic.NewTLSConfig, nil)
	if err != nil {
		t.Fatalf("Listener error: %v", err)
	}
	defer listener.Close()

	// Get listener address.
	addr := listener.ListenAddress()
	t.Logf("Server listening on %s", addr)

	// Setup our fake blockchain.
	fakeBC := SetupFakeBlockchain()

	// Start server goroutine.
	go func() {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			t.Errorf("Server accept error: %v", err)
			return
		}
		// Accept a stream.
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			t.Errorf("Stream accept error: %v", err)
			return
		}
		// Wrap the quic-go stream into our custom quic.Stream type.
		wrappedStream := &quic.Stream{Stream: stream}
		// Process the block request.
		err = HandleBlockRequestStream(fakeBC, wrappedStream)
		if err != nil {
			t.Errorf("HandleBlockRequestStream error: %v", err)
		}
	}()

	// Client: dial the server.
	conn, err := quic.Dial(context.Background(), addr, clientTLS, nil, quic.Validator)
	if err != nil {
		t.Fatalf("Client dial error: %v", err)
	}
	defer conn.Close()

	// Open a stream.
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		t.Fatalf("Client open stream error: %v", err)
	}

	// CE 128
	// Create a block request payload:
	// 32 bytes header hash (using genesis), 1 byte direction (0 for ascending), 4 bytes max blocks.
	var req CE128Payload
	req.HeaderHash = fakeBC.GenesisBlockHash() // starting from genesis
	req.Direction = 0                          // ascending exclusive
	req.MaxBlocks = 3

	reqPayload := make([]byte, 32+1+4)
	copy(reqPayload[:32], req.HeaderHash[:])
	reqPayload[32] = req.Direction
	binary.LittleEndian.PutUint32(reqPayload[33:37], req.MaxBlocks)

	// Write the request payload.
	_, err = stream.Write(framePayload(reqPayload))
	if err != nil {
		t.Fatalf("Client write error: %v", err)
	}
	// Signal that no more data will be sent.
	err = stream.Close()
	if err != nil {
		t.Fatalf("CloseWrite error: %v", err)
	}

	// Now read the response from the stream.
	respData, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("Client read error: %v", err)
	}

	// Decode the response: each block is framed with a 4-byte little-endian length prefix.
	r := bytes.NewReader(respData)
	var respBlocks []types.Block
	decoder := types.NewDecoder()
	for {
		var size uint32
		if err := binary.Read(r, binary.LittleEndian, &size); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("failed to read length prefix: %v", err)
		}
		blkData := make([]byte, size)
		if _, err := io.ReadFull(r, blkData); err != nil {
			t.Fatalf("failed to read block data: %v", err)
		}
		block := types.Block{}
		decoder.Decode(blkData, &block)
		respBlocks = append(respBlocks, block)
	}

	// For an ascending exclusive request starting from genesis,
	// we expect three blocks: block1, block2, block3.
	if len(respBlocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(respBlocks))
	}

	// Check chain validity:
	var block1Hash [32]byte
	block1Hash[0] = 1
	if respBlocks[1].Header.Parent != fakeBC.genesis {
		t.Errorf("expected block1 parent to be genesis, got %v", respBlocks[1].Header.Parent)
	}

	var block2Hash [32]byte
	block2Hash[0] = 2
	log.Printf("block2: %v", respBlocks[2])
	if respBlocks[2].Header.Parent != block1Hash {
		t.Errorf("expected block2 parent to be block1, got %v", respBlocks[2].Header.Parent)
	}
}
