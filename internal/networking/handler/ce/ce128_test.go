package ce

import (
	"bytes"
	"context"
	"crypto/ed25519"
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
	"net"
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

func (f *fakeBlockchain) GetStateAt(hash types.HeaderHash) (types.StateKeyVals, error) {
	// For testing, return some fake state data
	// This would normally query the actual state at the given block hash
	return types.StateKeyVals{
		{
			Key:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Value: types.ByteSequence{1, 2, 3, 4, 5},
		},
		{
			Key:   types.StateKey{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Value: types.ByteSequence{6, 7, 8, 9, 10},
		},
	}, nil
}

func (f *fakeBlockchain) GetStateRange(hash types.HeaderHash, startKey types.StateKey, endKey types.StateKey, maxSize uint32) (types.StateKeyVals, error) {
	// For testing, return some fake state data in the specified range
	// This would normally query the actual state range at the given block hash
	stateVals := types.StateKeyVals{
		{
			Key:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Value: types.ByteSequence{1, 2, 3, 4, 5},
		},
		{
			Key:   types.StateKey{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Value: types.ByteSequence{6, 7, 8, 9, 10},
		},
		{
			Key:   types.StateKey{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Value: types.ByteSequence{11, 12, 13, 14, 15},
		},
	}

	// Filter by key range if needed (simplified for testing)
	if startKey[0] > 0 {
		// Filter out keys less than startKey
		filtered := types.StateKeyVals{}
		for _, val := range stateVals {
			if val.Key[0] >= startKey[0] {
				filtered = append(filtered, val)
			}
		}
		stateVals = filtered
	}

	if endKey[0] > 0 {
		// Filter out keys greater than endKey
		filtered := types.StateKeyVals{}
		for _, val := range stateVals {
			if val.Key[0] <= endKey[0] {
				filtered = append(filtered, val)
			}
		}
		stateVals = filtered
	}

	// Limit by maxSize
	if uint32(len(stateVals)) > maxSize {
		stateVals = stateVals[:maxSize]
	}

	return stateVals, nil
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

	clientTLS, err := quic.NewTLSConfig(false, false)
	clientTLS.InsecureSkipVerify = true
	if err != nil {
		t.Fatalf("Client TLS config error: %v", err)
	}

	listener, err := quic.NewListener("localhost:0", false, quic.NewTLSConfig, nil)
	if err != nil {
		t.Fatalf("Listener error: %v", err)
	}
	defer listener.Close()

	addr := listener.ListenAddress()
	t.Logf("Server listening on %s", addr)

	fakeBC := SetupFakeBlockchain()

	go func() {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			t.Errorf("Server accept error: %v", err)
			return
		}
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			t.Errorf("Stream accept error: %v", err)
			return
		}
		wrappedStream := &quic.Stream{Stream: stream}
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

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		t.Fatalf("Client open stream error: %v", err)
	}

	// CE 128
	// Create a block request payload:
	// 32 bytes header hash (using genesis), 1 byte direction (0 for ascending), 4 bytes max blocks.
	var req CE128Payload
	req.HeaderHash = fakeBC.GenesisBlockHash()
	req.Direction = 0 // ascending exclusive
	req.MaxBlocks = 3

	reqPayload := make([]byte, 32+1+4)
	copy(reqPayload[:32], req.HeaderHash[:])
	reqPayload[32] = req.Direction
	binary.LittleEndian.PutUint32(reqPayload[33:37], req.MaxBlocks)

	_, err = stream.Write(reqPayload)
	if err != nil {
		t.Fatalf("Client write error: %v", err)
	}
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
		if err := decoder.Decode(blkData, &block); err != nil {
			t.Fatalf("failed to decode block: %v", err)
		}
		respBlocks = append(respBlocks, block)
	}

	// For an ascending exclusive request starting from genesis,
	// we expect three blocks: block1, block2, block3.
	if len(respBlocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(respBlocks))
	}

	// Check chain validity:
	// block1: parent should be genesis
	if respBlocks[0].Header.Parent != fakeBC.genesis {
		t.Errorf("expected block1 parent to be genesis, got %v", respBlocks[0].Header.Parent)
	}

	// block2: parent should be block1
	var block1Hash [32]byte
	block1Hash[0] = 1
	if respBlocks[1].Header.Parent != block1Hash {
		t.Errorf("expected block2 parent to be block1, got %v", respBlocks[1].Header.Parent)
	}

	// block3: parent should be block2
	var block2Hash [32]byte
	block2Hash[0] = 2
	if respBlocks[2].Header.Parent != block2Hash {
		t.Errorf("expected block3 parent to be block2, got %v", respBlocks[2].Header.Parent)
	}
}

func TestCE128RequestWithPeer(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	fakeBC := SetupFakeBlockchain()

	ceRequestHandler := NewDefaultCERequestHandler()

	upHandler := quic.NewDefaultUPHandler()

	ceHandler := quic.NewDefaultCEHandler(fakeBC)

	ceHandler.RegisterCEHandler(128, HandleBlockRequestStream)

	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate public key: %v", err)
	}

	serverConfig := quic.PeerConfig{
		Role:          quic.Validator,
		Addr:          &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		GenesisHeader: fakeBC.GenesisBlockHash(),
		PublicKey:     publicKey,
		UPHandler:     upHandler,
		CEHandler:     ceHandler,
	}

	serverPeer, err := quic.NewPeer(serverConfig)
	if err != nil {
		t.Fatalf("Failed to create server peer: %v", err)
	}
	defer serverPeer.Listener.Close()

	serverPeer.SetTLSInsecureSkipVerify(true)

	serverAddr := serverPeer.Listener.ListenAddress()
	t.Logf("Server listening on %s", serverAddr)

	go func() {
		conn, err := serverPeer.Listener.Accept(context.Background())
		if err != nil {
			t.Errorf("Server accept error: %v", err)
			return
		}

		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			t.Errorf("Stream accept error: %v", err)
			return
		}

		wrappedStream := &quic.Stream{Stream: stream}

		err = serverPeer.CEHandler.HandleStream(wrappedStream)
		if err != nil {
			t.Errorf("CE handler error: %v", err)
		}
	}()

	clientPublicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate client public key: %v", err)
	}

	clientConfig := quic.PeerConfig{
		Role:          quic.Validator,
		Addr:          &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		GenesisHeader: fakeBC.GenesisBlockHash(),
		PublicKey:     clientPublicKey,
		UPHandler:     upHandler,
		CEHandler:     ceHandler,
	}

	clientPeer, err := quic.NewPeer(clientConfig)
	if err != nil {
		t.Fatalf("Failed to create client peer: %v", err)
	}
	defer clientPeer.Listener.Close()

	clientPeer.SetTLSInsecureSkipVerify(true)

	serverNetAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to resolve server address: %v", err)
	}

	conn, err := clientPeer.Connect(serverNetAddr, *serverPeer)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		t.Fatalf("Failed to open stream: %v", err)
	}

	blockReq := &CE128Payload{
		HeaderHash: fakeBC.GenesisBlockHash(),
		Direction:  0,
		MaxBlocks:  3,
	}

	encodedReq, err := ceRequestHandler.Encode(BlockRequest, blockReq)
	if err != nil {
		t.Fatalf("Failed to encode block request: %v", err)
	}

	reqPayload := make([]byte, 1+len(encodedReq))
	reqPayload[0] = 128 // CE128 protocol ID
	copy(reqPayload[1:], encodedReq)

	t.Logf("Encoded request payload: %x", reqPayload)

	_, err = stream.Write(reqPayload)
	if err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	err = stream.Close()
	if err != nil {
		t.Fatalf("Failed to close stream: %v", err)
	}

	respData, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	t.Logf("Received response data length: %d", len(respData))

	r := bytes.NewReader(respData)
	var respBlocks []types.Block
	decoder := types.NewDecoder()
	for {
		var size uint32
		if err := binary.Read(r, binary.LittleEndian, &size); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("Failed to read length prefix: %v", err)
		}

		blkData := make([]byte, size)
		if _, err := io.ReadFull(r, blkData); err != nil {
			t.Fatalf("Failed to read block data: %v", err)
		}

		block := types.Block{}
		err := decoder.Decode(blkData, &block)
		if err != nil {
			t.Fatalf("Failed to decode block: %v", err)
		}

		respBlocks = append(respBlocks, block)
	}

	// For an ascending exclusive request starting from genesis,
	// we expect three blocks: block1, block2, block3
	if len(respBlocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(respBlocks))
	}

	// Check chain validity
	var block1Hash [32]byte
	block1Hash[0] = 1
	if respBlocks[0].Header.Parent != fakeBC.genesis {
		t.Errorf("Expected block1 parent to be genesis, got %v", respBlocks[0].Header.Parent)
	}

	var block2Hash [32]byte
	block2Hash[0] = 2
	if respBlocks[1].Header.Parent != block1Hash {
		t.Errorf("Expected block2 parent to be block1, got %v", respBlocks[1].Header.Parent)
	}

	var block3Hash [32]byte
	block3Hash[0] = 3
	if respBlocks[2].Header.Parent != block2Hash {
		t.Errorf("Expected block3 parent to be block2, got %v", respBlocks[2].Header.Parent)
	}

	t.Logf("Successfully received %d blocks", len(respBlocks))
	for i, block := range respBlocks {
		t.Logf("Block %d: Slot=%d, Parent=%x", i, block.Header.Slot, block.Header.Parent[:])
	}
}
