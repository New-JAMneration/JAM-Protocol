package ce

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// fakeBlockchainWithState extends fakeBlockchain to support state queries
type fakeBlockchainWithState struct {
	*fakeBlockchain
	stateData map[types.HeaderHash]types.StateKeyVals
}

func (f *fakeBlockchainWithState) GetStateAt(hash types.HeaderHash) (types.StateKeyVals, error) {
	state, ok := f.stateData[hash]
	if !ok {
		return nil, fmt.Errorf("state not found for hash: %x", hash)
	}
	return state, nil
}

func (f *fakeBlockchainWithState) GetStateRange(hash types.HeaderHash, startKey, endKey types.StateKey, maxSize uint32) (types.StateKeyVals, error) {
	state, err := f.GetStateAt(hash)
	if err != nil {
		return nil, err
	}

	var result types.StateKeyVals
	totalSize := uint32(0)

	for _, stateVal := range state {
		// Check if the key is in the requested range
		if bytes.Compare(stateVal.Key[:], startKey[:]) >= 0 && bytes.Compare(stateVal.Key[:], endKey[:]) <= 0 {
			// Estimate size (key + value length + value)
			estimatedSize := uint32(len(stateVal.Key)) + 4 + uint32(len(stateVal.Value))
			if totalSize+estimatedSize > maxSize {
				break
			}
			result = append(result, stateVal)
			totalSize += estimatedSize
		}
	}

	return result, nil
}

// SetupFakeBlockchainWithState creates a blockchain with mock state data
func SetupFakeBlockchainWithState() *fakeBlockchainWithState {
	fb := SetupFakeBlockchain()

	// Create some mock state data
	stateData := make(map[types.HeaderHash]types.StateKeyVals)

	// Add state data for each block
	for hash, block := range fb.blocks {
		var stateVals types.StateKeyVals

		// Create some mock state values
		for i := 0; i < 3; i++ {
			key := types.StateKey{}
			key[0] = byte(i + 1)
			key[1] = byte(block.Header.Slot)

			value := types.ByteSequence(fmt.Sprintf("value_%d_%d", block.Header.Slot, i))

			stateVals = append(stateVals, types.StateKeyVal{
				Key:   key,
				Value: value,
			})
		}

		stateData[hash] = stateVals
	}

	return &fakeBlockchainWithState{
		fakeBlockchain: fb,
		stateData:      stateData,
	}
}

// TestHandleStateRequest tests the HandleStateRequest function directly
func TestHandleStateRequest(t *testing.T) {
	fakeBC := SetupFakeBlockchainWithState()

	// Create a mock stream
	mockStream := newMockStream([]byte{})

	// Create a test request
	req := CE129Payload{
		HeaderHash: fakeBC.GenesisBlockHash(),
		KeyStart:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		KeyEnd:     types.StateKey{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		MaxSize:    1000,
	}

	// Call the handler
	err := HandleStateRequest(fakeBC, req, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleStateRequest failed: %v", err)
	}

	// Read the response
	response := mockStream.w.Bytes()

	// Parse the response
	if len(response) < 4 {
		t.Fatalf("Response too short")
	}

	numValues := binary.LittleEndian.Uint32(response[:4])
	t.Logf("Number of state values returned: %d", numValues)

	// Verify we got some state values
	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}
}

// TestHandleStateRequestStream tests the stream-based handler
func TestHandleStateRequestStream(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	fakeBC := SetupFakeBlockchainWithState()

	// Create a mock stream with request data
	req := CE129Payload{
		HeaderHash: fakeBC.GenesisBlockHash(),
		KeyStart:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		KeyEnd:     types.StateKey{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		MaxSize:    1000,
	}

	// Encode the request
	reqPayload := make([]byte, 98)
	copy(reqPayload[:32], req.HeaderHash[:])
	copy(reqPayload[32:63], req.KeyStart[:])
	copy(reqPayload[63:94], req.KeyEnd[:])
	binary.LittleEndian.PutUint32(reqPayload[94:98], req.MaxSize)

	mockStream := newMockStream(reqPayload)

	// Call the stream handler
	err := HandleStateRequestStream(fakeBC, &quic.Stream{Stream: mockStream})
	if err != nil {
		t.Fatalf("HandleStateRequestStream failed: %v", err)
	}

	// Read the response
	response := mockStream.w.Bytes()

	// Parse the response
	if len(response) < 4 {
		t.Fatalf("Response too short")
	}

	numValues := binary.LittleEndian.Uint32(response[:4])
	t.Logf("Number of state values returned: %d", numValues)

	// Verify we got some state values
	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}
}

// TestRealQuicStreamStateRequest tests with real QUIC connection
func TestRealQuicStreamStateRequest(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	// Setup TLS configurations
	clientTLS, err := quic.NewTLSConfig(false, false)
	clientTLS.InsecureSkipVerify = true
	if err != nil {
		t.Fatalf("Client TLS config error: %v", err)
	}

	// Listen on an ephemeral port
	listener, err := quic.NewListener("localhost:0", false, quic.NewTLSConfig, nil)
	if err != nil {
		t.Fatalf("Listener error: %v", err)
	}
	defer listener.Close()

	addr := listener.ListenAddress()
	t.Logf("Server listening on %s", addr)

	// Setup our fake blockchain with state
	fakeBC := SetupFakeBlockchainWithState()

	// Start server goroutine
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
		err = HandleStateRequestStream(fakeBC, wrappedStream)
		if err != nil {
			t.Errorf("HandleStateRequestStream error: %v", err)
		}
	}()

	// Client: dial the server
	conn, err := quic.Dial(context.Background(), addr, clientTLS, nil, quic.Validator)
	if err != nil {
		t.Fatalf("Client dial error: %v", err)
	}
	defer conn.Close()

	// Open a stream
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		t.Fatalf("Client open stream error: %v", err)
	}

	// Create a state request payload
	var req CE129Payload
	req.HeaderHash = fakeBC.GenesisBlockHash()
	copy(req.KeyStart[:], []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	copy(req.KeyEnd[:], []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	req.MaxSize = 1000

	// Encode the request payload
	reqPayload := make([]byte, 98)
	copy(reqPayload[:32], req.HeaderHash[:])
	copy(reqPayload[32:63], req.KeyStart[:])
	copy(reqPayload[63:94], req.KeyEnd[:])
	binary.LittleEndian.PutUint32(reqPayload[94:98], req.MaxSize)

	// Write the request payload
	_, err = stream.Write(reqPayload)
	if err != nil {
		t.Fatalf("Client write error: %v", err)
	}

	err = stream.Close()
	if err != nil {
		t.Fatalf("CloseWrite error: %v", err)
	}

	// Read the response
	respData, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("Client read error: %v", err)
	}

	// Parse the response
	if len(respData) < 4 {
		t.Fatalf("Response too short")
	}

	numValues := binary.LittleEndian.Uint32(respData[:4])
	t.Logf("Number of state values returned: %d", numValues)

	// Verify we got some state values
	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}

	// Parse individual state values
	offset := 4
	for i := uint32(0); i < numValues; i++ {
		if offset+4 > len(respData) {
			t.Fatalf("Response truncated")
		}

		valueLength := binary.LittleEndian.Uint32(respData[offset : offset+4])
		offset += 4

		if offset+int(valueLength) > len(respData) {
			t.Fatalf("Value data truncated")
		}

		valueData := respData[offset : offset+int(valueLength)]
		t.Logf("State value %d: length=%d, data=%x", i, valueLength, valueData)
		offset += int(valueLength)
	}
}

// TestCE129RequestWithPeer tests with peer-based setup
func TestCE129RequestWithPeer(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	fakeBC := SetupFakeBlockchainWithState()

	ceRequestHandler := NewDefaultCERequestHandler()
	upHandler := quic.NewDefaultUPHandler()
	ceHandler := quic.NewDefaultCEHandler(fakeBC)

	// Register CE129 handler
	ceHandler.RegisterCEHandler(129, HandleStateRequestStream)

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

	// Create state request
	stateReq := &CE129Payload{
		HeaderHash: fakeBC.GenesisBlockHash(),
		KeyStart:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		KeyEnd:     types.StateKey{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		MaxSize:    1000,
	}

	encodedReq, err := ceRequestHandler.Encode(StateRequest, stateReq)
	if err != nil {
		t.Fatalf("Failed to encode state request: %v", err)
	}

	reqPayload := make([]byte, 1+len(encodedReq))
	reqPayload[0] = 129 // CE129 protocol ID
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

	// Parse the response
	if len(respData) < 4 {
		t.Fatalf("Response too short")
	}

	numValues := binary.LittleEndian.Uint32(respData[:4])
	t.Logf("Number of state values returned: %d", numValues)

	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}

	t.Logf("Successfully received %d state values", numValues)
}
