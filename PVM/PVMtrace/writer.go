package PVMtrace

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

// streamWriter wraps a gzip-compressed file writer with buffering and SHA-256 sidecar.
type streamWriter struct {
	path     string
	file     *os.File
	buf      *bufio.Writer
	gz       *gzip.Writer
	hash     *sha256Writer
	written  int64
	closed   bool
}

// sha256Writer is an io.Writer that tees all writes through a SHA-256 hasher.
type sha256Writer struct {
	w    io.Writer
	hash [32]byte
	h    interface{ Sum([]byte) []byte; Write([]byte) (int, error); Reset() }
}

func newSha256Writer(w io.Writer) *sha256Writer {
	h := sha256.New()
	return &sha256Writer{w: w, h: h}
}

func (s *sha256Writer) Write(p []byte) (int, error) {
	s.h.Write(p)
	return s.w.Write(p)
}

func (s *sha256Writer) Sum() [32]byte {
	var out [32]byte
	copy(out[:], s.h.Sum(nil))
	return out
}

func newStreamWriter(dir, name string, gzipLevel, bufSizeMB int) (*streamWriter, error) {
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	bufSize := bufSizeMB * 1024 * 1024
	if bufSize <= 0 {
		bufSize = 4 * 1024 * 1024
	}
	buf := bufio.NewWriterSize(f, bufSize)
	hashW := newSha256Writer(buf)
	gz, err := gzip.NewWriterLevel(hashW, gzipLevel)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("gzip writer %s: %w", name, err)
	}
	return &streamWriter{
		path: path,
		file: f,
		buf:  buf,
		gz:   gz,
		hash: hashW,
	}, nil
}

func (sw *streamWriter) Write(p []byte) (int, error) {
	n, err := sw.gz.Write(p)
	atomic.AddInt64(&sw.written, int64(n))
	return n, err
}

func (sw *streamWriter) Close() error {
	if sw.closed {
		return nil
	}
	sw.closed = true
	if err := sw.gz.Close(); err != nil {
		sw.file.Close()
		return err
	}
	if err := sw.buf.Flush(); err != nil {
		sw.file.Close()
		return err
	}
	if err := sw.file.Close(); err != nil {
		return err
	}
	sum := sw.hash.Sum()
	shaPath := sw.path + ".sha256"
	return os.WriteFile(shaPath, []byte(hex.EncodeToString(sum[:])+"\n"), 0644)
}
