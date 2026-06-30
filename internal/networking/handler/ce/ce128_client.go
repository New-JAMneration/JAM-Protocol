package ce

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// RequestBlocks opens a CE 128 stream on conn and returns the requested blocks.
func RequestBlocks(ctx context.Context, conn *quic.Connection, req CE128Payload) ([]types.Block, error) {
	if conn == nil {
		return nil, fmt.Errorf("nil connection")
	}

	qstream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("open CE 128 stream: %w", err)
	}
	stream := &quic.Stream{Stream: qstream}

	if err := stream.WriteStreamKind(byte(BlockRequest)); err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("write stream kind: %w", err)
	}

	handler := NewDefaultCERequestHandler()
	payload, err := handler.Encode(BlockRequest, &req)
	if err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("encode CE 128 request: %w", err)
	}
	if err := stream.WriteMessage(payload); err != nil {
		_ = stream.Close()
		return nil, fmt.Errorf("write CE 128 request: %w", err)
	}

	decoder := types.NewDecoder()
	blocks := make([]types.Block, 0, req.MaxBlocks)
	for {
		data, err := stream.ReadMessage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return blocks, fmt.Errorf("read CE 128 response: %w", err)
		}
		var block types.Block
		if err := decoder.Decode(data, &block); err != nil {
			return blocks, fmt.Errorf("decode block: %w", err)
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
