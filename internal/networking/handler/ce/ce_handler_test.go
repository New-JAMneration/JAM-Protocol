package ce

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"io"
	"net"
	"os"
	"testing"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func TestCE128RequestWithPeer(t *testing.T) {
	os.Setenv("USE_MINI_REDIS", "true")
	defer store.CloseMiniRedis()

	fakeBC := SetupFakeBlockchain()

	ceRequestHandler := NewDefaultCERequestHandler()

	upHandler := quic.NewDefaultUPHandler()

	ceHandler := quic.NewDefaultCEHandler(fakeBC)

	ceHandler.RegisterCEHandler(128, func(bc blockchain.Blockchain, stream *quic.Stream) error {
		return HandleBlockRequestStream(bc, stream)
	})

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

	blockReq := &BlockRequestMessage{
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
