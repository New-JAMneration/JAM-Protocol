package quic

import (
	"io"
	"time"

	"github.com/quic-go/quic-go"
)

// Stream encapsulates a quic.Stream to provide a higher-level abstraction
// for reading from and writing to QUIC streams.
type Stream struct {
	Stream quic.Stream
}

// Read reads data from the underlying QUIC stream into the provided buffer.
func (s *Stream) Read(p []byte) (int, error) {
	return s.Stream.Read(p)
}

// Write writes data from the provided buffer to the underlying QUIC stream.
func (s *Stream) Write(p []byte) (int, error) {
	return s.Stream.Write(p)
}

// Close closes the underlying QUIC stream.
func (s *Stream) Close() error {
	return s.Stream.Close()
}

// SetReadDeadline sets the deadline for future Read calls.
func (s *Stream) SetReadDeadline(t time.Time) error {
	return s.Stream.SetReadDeadline(t)
}

// Copy is a helper function that copies data from the src stream to the dest stream
// until an EOF is encountered.
func Copy(dest quic.Stream, src quic.Stream) error {
	_, err := io.Copy(dest, src)
	return err
}
