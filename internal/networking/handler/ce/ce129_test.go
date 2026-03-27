package ce

import (
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

// TestHandleStateRequest tests the HandleStateRequest function directly
func TestHandleStateRequest(t *testing.T) {
	fakeBC := SetupFakeBlockchain()
	genesisHash := fakeBC.GenesisBlockHash()

	state, err := fakeBC.GetStateAt(genesisHash)
	if err != nil {
		t.Logf("GetStateAt error: %v", err)
	} else {
		t.Logf("State values at genesis: %d", len(state))
		for _, val := range state {
			t.Logf("State key-value: key=%x value=%s", val.Key, val.Value)
		}
	}

	mockStream := newMockStream([]byte{})

	req := CE129Payload{
		HeaderHash: fakeBC.GenesisBlockHash(),
		KeyStart:   types.StateKey{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		KeyEnd:     types.StateKey{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		MaxSize:    1000,
	}

	handleErr := HandleStateRequest(fakeBC, req, &quic.Stream{Stream: mockStream})
	if handleErr != nil {
		t.Fatalf("HandleStateRequest failed: %v", handleErr)
	}

	response := mockStream.w.Bytes()
	// Response is two messages: [4][boundaryBlob][4][keyValuesBlob]
	if len(response) < 8 {
		t.Fatalf("Response too short")
	}
	boundaryLen := binary.LittleEndian.Uint32(response[:4])
	if 4+int(boundaryLen)+4 > len(response) {
		t.Fatalf("Response truncated after boundary message")
	}
	boundaryBlob := response[4 : 4+boundaryLen]
	keyValuesLen := binary.LittleEndian.Uint32(response[4+boundaryLen : 4+boundaryLen+4])
	if 4+int(boundaryLen)+4+int(keyValuesLen) > len(response) {
		t.Fatalf("Response truncated in key/values message")
	}
	keyValuesBlob := response[4+boundaryLen+4 : 4+boundaryLen+4+keyValuesLen]

	// Parse boundaryBlob: whole [BoundaryNode] sequence (no count/length prefix)
	decoder := types.NewDecoder()
	offset := 0
	var numBoundaryNodes int
	for offset < len(boundaryBlob) {
		var node types.BoundaryNode
		n, err := decoder.DecodeWithConsumed(boundaryBlob[offset:], &node)
		if err != nil {
			t.Fatalf("Failed to decode boundary node at offset %d: %v", offset, err)
		}
		offset += n
		numBoundaryNodes++
		t.Logf("Boundary node %d: Key=%x Hash=%x", numBoundaryNodes-1, node.Key, node.Hash)
	}
	t.Logf("Number of boundary nodes: %d", numBoundaryNodes)

	// Parse keyValuesBlob: whole [Key++Value] sequence (no count/length prefix)
	var stateValues types.StateKeyVals
	offset = 0
	for offset < len(keyValuesBlob) {
		var kv types.StateKeyVal
		n, err := decoder.DecodeWithConsumed(keyValuesBlob[offset:], &kv)
		if err != nil {
			t.Fatalf("Failed to decode state value at offset %d: %v", offset, err)
		}
		offset += n
		stateValues = append(stateValues, kv)
		t.Logf("State value %d: key=%x value=%s", len(stateValues)-1, kv.Key, kv.Value)
	}

	if len(stateValues) == 0 {
		t.Errorf("Expected some state values, got 0")
	}
}

// TestHandleStateRequestStream tests the stream-based handler

// TestCE129RequestWithPeer tests with peer-based setup
func TestCE129RequestWithPeer(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	fakeBC := SetupFakeBlockchain()

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

	conn, err := clientPeer.Connect(serverNetAddr, quic.Validator)
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

	// Send protocol ID then one length-prefixed message.
	reqPayload := make([]byte, 0, 1+4+len(encodedReq))
	reqPayload = append(reqPayload, 129)
	reqPayload = append(reqPayload, framePayload(encodedReq)...)

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
	t.Logf("Raw response bytes: %x", respData)

	// Response is two length-prefixed messages: [4][boundaryBlob][4][keyValuesBlob]
	if len(respData) < 8 {
		t.Fatalf("Response too short, got %d bytes", len(respData))
	}
	boundaryLen := binary.LittleEndian.Uint32(respData[:4])
	if 4+int(boundaryLen)+4 > len(respData) {
		t.Fatalf("Response truncated after boundary message")
	}
	boundaryBlob := respData[4 : 4+boundaryLen]
	keyValuesLen := binary.LittleEndian.Uint32(respData[4+boundaryLen : 4+boundaryLen+4])
	if 4+int(boundaryLen)+4+int(keyValuesLen) > len(respData) {
		t.Fatalf("Response truncated in key/values message")
	}
	keyValuesBlob := respData[4+boundaryLen+4 : 4+boundaryLen+4+keyValuesLen]

	// Parse boundaryBlob: whole [BoundaryNode] sequence (no count/length prefix)
	decoder := types.NewDecoder()
	offset := 0
	for offset < len(boundaryBlob) {
		var node types.BoundaryNode
		n, err := decoder.DecodeWithConsumed(boundaryBlob[offset:], &node)
		if err != nil {
			t.Fatalf("Failed to decode boundary node at offset %d: %v", offset, err)
		}
		offset += n
		t.Logf("Boundary node decoded: Key=%x Hash=%x", node.Key, node.Hash)
	}

	// Parse keyValuesBlob: whole [Key++Value] sequence (no count/length prefix)
	var stateValues types.StateKeyVals
	offset = 0
	for offset < len(keyValuesBlob) {
		var kv types.StateKeyVal
		n, err := decoder.DecodeWithConsumed(keyValuesBlob[offset:], &kv)
		if err != nil {
			t.Fatalf("Failed to decode state value at offset %d: %v", offset, err)
		}
		offset += n
		stateValues = append(stateValues, kv)
	}

	if len(stateValues) == 0 {
		t.Errorf("Expected some state values, got 0")
	}

	t.Logf("Successfully received and parsed %d state values", len(stateValues))
	if len(stateValues) > 0 {
		var parsedValues []string
		for _, kv := range stateValues {
			parsedValues = append(parsedValues, fmt.Sprintf("%x", kv.Value))
		}
		t.Logf("Parsed values: %v", parsedValues)
	}
}
