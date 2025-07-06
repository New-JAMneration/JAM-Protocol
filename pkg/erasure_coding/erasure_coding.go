package erasurecoding

/*
#cgo CFLAGS: -I${SRCDIR}/reed-solomon-ffi
#cgo LDFLAGS: -L${SRCDIR}/reed-solomon-ffi/target/release -lreed_solomon_ffi
#include "reedsolomon.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func EncodeDataShards(data []byte, dataShard, parityShard int) ([][]byte, error) {
	// Call FFI for flattened shards
	flat, err := EncodeData(data, dataShard, parityShard)
	if err != nil {
		return nil, err
	}

	numShards := dataShard + parityShard
	shardSize := len(flat) / numShards
	if len(flat)%numShards != 0 {
		return nil, fmt.Errorf("unexpected output size %d is not divisible by %d shards", len(flat), numShards)
	}

	// Slice to shards [][]byte
	shards := make([][]byte, numShards)
	for i := 0; i < numShards; i++ {
		start := i * shardSize
		end := start + shardSize
		shardCopy := make([]byte, shardSize)
		copy(shardCopy, flat[start:end])
		shards[i] = shardCopy
	}
	return shards, nil
}

func EncodeData(data []byte, dataShards, parityShards int) ([]byte, error) {
	var outPtr *C.uchar
	var outLen C.size_t

	if len(data) == 0 {
		return nil, errors.New("input data is empty")
	}

	ret := C.rs_encode(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		C.size_t(dataShards),
		C.size_t(parityShards),
		(**C.uchar)(unsafe.Pointer(&outPtr)),
		(*C.size_t)(unsafe.Pointer(&outLen)),
	)

	if ret != 0 {
		return nil, errors.New("erasure_encode failed with code " + string(ret))
	}

	// Copy the Rust-allocated memory into Go slice
	result := C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen))

	// Free the Rust-allocated memory
	C.free(unsafe.Pointer(outPtr))

	return result, nil
}

type Shard struct {
	Index int
	Data  [2]byte
}

func DecodeShards(flatten []byte, indices []int, dataShards, parityShards, shardSize int) ([]byte, error) {
	if len(flatten) == 0 || len(indices) == 0 {
		return nil, errors.New("no shards provided")
	}
	if len(flatten)%shardSize != 0 {
		return nil, fmt.Errorf("flatten data length %d not divisible by shardSize %d", len(flatten), shardSize)
	}
	shardCount := len(indices)

	// Allocate indices C array
	cIdxSize := C.size_t(shardCount) * C.size_t(unsafe.Sizeof(C.size_t(0)))
	cIdxPtr := C.malloc(cIdxSize)
	if cIdxPtr == nil {
		return nil, errors.New("malloc indices failed")
	}
	defer C.free(cIdxPtr)
	cIdxSlice := (*[1 << 30]C.size_t)(cIdxPtr)[:shardCount:shardCount]
	for i, idx := range indices {
		cIdxSlice[i] = C.size_t(idx)
	}

	var outPtr *C.uint8_t
	var outLen C.size_t

	ret := C.rs_decode(
		(*C.uchar)(unsafe.Pointer(&flatten[0])),
		(*C.size_t)(cIdxPtr),
		C.size_t(shardCount),
		C.size_t(shardSize),
		C.size_t(dataShards),
		C.size_t(parityShards),
		&outPtr,
		&outLen,
	)
	if ret != 0 {
		return nil, fmt.Errorf("erasure_decode failed with code %d", ret)
	}
	defer C.free(unsafe.Pointer(outPtr))

	out := C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen))
	return out, nil
}
