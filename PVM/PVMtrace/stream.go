package PVMtrace

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Stream record widths (bytes).
const (
	PCWidth      = 8  // u64
	OpcodeWidth  = 1  // u8
	GasWidth     = 8  // i64
	DstValWidth  = 8  // u64
	Src1ValWidth = 8  // u64
	Src2ValWidth = 8  // u64
	LoadWidth    = 12 // u32 addr + u64 val
	StoreWidth   = 12 // u32 addr + u64 val
)

// StreamName constants.
const (
	StreamPC      = "pc.gz"
	StreamOpcode  = "opcode.gz"
	StreamGas     = "gas.gz"
	StreamDstVal  = "dst_val.gz"
	StreamSrc1Val = "src1_val.gz"
	StreamSrc2Val = "src2_val.gz"
	StreamLoads   = "loads.gz"
	StreamStores  = "stores.gz"
	StreamHostCalls = "host_calls.jsonl.gz"
)

// AllStreamNames lists all per-instruction stream filenames.
var AllStreamNames = []string{
	StreamPC, StreamOpcode, StreamGas,
	StreamDstVal, StreamSrc1Val, StreamSrc2Val,
	StreamLoads, StreamStores,
}

// GzipRecordReader reads fixed-width records from a gzip stream.
type GzipRecordReader struct {
	file       *os.File
	gz         *gzip.Reader
	buf        *bufio.Reader
	recordSize int
}

func NewGzipRecordReader(path string, recordSize int) (*GzipRecordReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("gzip open %s: %w", path, err)
	}
	return &GzipRecordReader{
		file:       f,
		gz:         gz,
		buf:        bufio.NewReaderSize(gz, 256*1024),
		recordSize: recordSize,
	}, nil
}

// ReadRecord reads one record into dst. Returns io.EOF at end.
func (r *GzipRecordReader) ReadRecord(dst []byte) error {
	if len(dst) < r.recordSize {
		return fmt.Errorf("buffer too small: need %d, got %d", r.recordSize, len(dst))
	}
	_, err := io.ReadFull(r.buf, dst[:r.recordSize])
	return err
}

// ReadU64 reads one little-endian u64 record.
func (r *GzipRecordReader) ReadU64() (uint64, error) {
	var buf [8]byte
	if err := r.ReadRecord(buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

// ReadI64 reads one little-endian i64 record.
func (r *GzipRecordReader) ReadI64() (int64, error) {
	v, err := r.ReadU64()
	return int64(v), err
}

// ReadU8 reads one u8 record.
func (r *GzipRecordReader) ReadU8() (byte, error) {
	var buf [1]byte
	if err := r.ReadRecord(buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

// ReadMemRecord reads a 12-byte memory access record (u32 addr + u64 val).
func (r *GzipRecordReader) ReadMemRecord() (addr uint32, val uint64, err error) {
	var buf [12]byte
	if err = r.ReadRecord(buf[:]); err != nil {
		return
	}
	addr = binary.LittleEndian.Uint32(buf[0:4])
	val = binary.LittleEndian.Uint64(buf[4:12])
	return
}

func (r *GzipRecordReader) Close() error {
	r.gz.Close()
	return r.file.Close()
}

// CountRecords counts total records in a gzip stream without loading all into memory.
func CountRecords(path string, recordSize int) (int64, error) {
	r, err := NewGzipRecordReader(path, recordSize)
	if err != nil {
		return 0, err
	}
	defer r.Close()
	buf := make([]byte, recordSize)
	var count int64
	for {
		if err := r.ReadRecord(buf); err != nil {
			if err == io.EOF {
				return count, nil
			}
			return count, err
		}
		count++
	}
}
