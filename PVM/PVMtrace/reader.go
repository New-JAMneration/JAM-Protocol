package PVMtrace

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TraceReader provides access to a trace directory's metadata and streams.
type TraceReader struct {
	Dir  string
	Info TraceInfo
}

// OpenTrace opens a trace directory and loads its metadata.
func OpenTrace(dir string) (*TraceReader, error) {
	infoPath := filepath.Join(dir, "meta", "info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("read info.json: %w", err)
	}
	var info TraceInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("parse info.json: %w", err)
	}
	return &TraceReader{Dir: dir, Info: info}, nil
}

// ProgramBin returns the raw program blob if available.
func (r *TraceReader) ProgramBin() ([]byte, error) {
	return os.ReadFile(filepath.Join(r.Dir, "meta", "program.bin"))
}

// HasStream checks whether a stream file exists.
func (r *TraceReader) HasStream(name string) bool {
	_, err := os.Stat(filepath.Join(r.Dir, name))
	return err == nil
}

// OpenStream opens a gzip record reader for the named stream.
func (r *TraceReader) OpenStream(name string, recordSize int) (*GzipRecordReader, error) {
	return NewGzipRecordReader(filepath.Join(r.Dir, name), recordSize)
}

// ReadHostCalls loads all host-call records. Tries host_calls.jsonl.gz first,
// then falls back to info.json embedded host_calls if the file doesn't exist.
func (r *TraceReader) ReadHostCalls() ([]HostCallRecord, error) {
	jsonlPath := filepath.Join(r.Dir, StreamHostCalls)
	if _, err := os.Stat(jsonlPath); err == nil {
		return readHostCallsJSONL(jsonlPath)
	}
	// No jsonl.gz file — not available
	return nil, nil
}

func readHostCallsJSONL(path string) ([]HostCallRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip open host_calls: %w", err)
	}
	defer gz.Close()

	var records []HostCallRecord
	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var rec HostCallRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return records, fmt.Errorf("parse host-call record: %w", err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return records, err
	}
	return records, nil
}
