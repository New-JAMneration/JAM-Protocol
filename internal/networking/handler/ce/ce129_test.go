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

	numBoundaryNodes := binary.LittleEndian.Uint32(boundaryBlob[:4])
	t.Logf("Number of boundary nodes: %d", numBoundaryNodes)

	offset := 4
	for i := uint32(0); i < numBoundaryNodes; i++ {
		if offset+4 > len(boundaryBlob) {
			t.Fatalf("Response truncated in boundary nodes section")
		}
		nodeLen := binary.LittleEndian.Uint32(boundaryBlob[offset : offset+4])
		offset += 4 + int(nodeLen)
	}

	numValues := binary.LittleEndian.Uint32(keyValuesBlob[:4])
	t.Logf("Number of state values returned: %d", numValues)

	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}

	// Parse and verify each state value (inside keyValuesBlob)
	offset = 4
	for i := uint32(0); i < numValues; i++ {
		if offset+4 > len(keyValuesBlob) {
			t.Fatalf("Response truncated in state values section")
		}
		valueLen := binary.LittleEndian.Uint32(keyValuesBlob[offset : offset+4])
		offset += 4

		if offset+int(valueLen) > len(keyValuesBlob) {
			t.Fatalf("Response truncated in state value data")
		}
		valueData := keyValuesBlob[offset : offset+int(valueLen)]
		t.Logf("State value %d: length=%d data=%x", i, valueLen, valueData)
		offset += int(valueLen)
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

	numBoundaryNodes := binary.LittleEndian.Uint32(boundaryBlob[:4])
	t.Logf("Number of boundary nodes: %d", numBoundaryNodes)

	offset := 4
	for i := uint32(0); i < numBoundaryNodes; i++ {
		if offset+4 > len(boundaryBlob) {
			t.Fatalf("Response truncated in boundary nodes section at node %d", i)
		}
		nodeLen := binary.LittleEndian.Uint32(boundaryBlob[offset : offset+4])
		t.Logf("Boundary node %d length: %d", i, nodeLen)
		offset += 4
		if offset+int(nodeLen) > len(boundaryBlob) {
			t.Fatalf("Response truncated in boundary node data at node %d", i)
		}
		nodeData := boundaryBlob[offset : offset+int(nodeLen)]
		t.Logf("Boundary node %d data: %x", i, nodeData)
		offset += int(nodeLen)
	}

	if len(keyValuesBlob) < 4 {
		t.Fatalf("Key/values blob too short at offset %d", 4+boundaryLen+4)
	}
	numValues := binary.LittleEndian.Uint32(keyValuesBlob[:4])

	if numValues == 0 {
		t.Errorf("Expected some state values, got 0")
	}

	offset = 4 // within keyValuesBlob
	var parsedValues []string
	for i := uint32(0); i < numValues; i++ {
		if offset+4 > len(keyValuesBlob) {
			t.Fatalf("Response truncated in state values section at value %d offset %d", i, offset)
		}
		valueLen := binary.LittleEndian.Uint32(keyValuesBlob[offset : offset+4])
		t.Logf("State value %d length: %d (at offset %d)", i, valueLen, offset)
		offset += 4

		if offset+int(valueLen) > len(keyValuesBlob) {
			t.Fatalf("Response truncated in state value data at value %d offset %d", i, offset)
		}
		valueData := keyValuesBlob[offset : offset+int(valueLen)]
		t.Logf("State value %d data: %x", i, valueData)
		parsedValues = append(parsedValues, fmt.Sprintf("%x", valueData))
		offset += int(valueLen)
	}

	t.Logf("Successfully received and parsed %d state values", numValues)
	if len(parsedValues) > 0 {
		t.Logf("Parsed values: %v", parsedValues)
	}
}
