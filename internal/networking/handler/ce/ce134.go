package ce

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/store/keystore"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/work_package"
)

type SegmentRootMapping struct {
	WorkPackageHash types.WorkPackageHash
	SegmentRoot     types.OpaqueHash
}

type CE134WorkPackageShare struct {
	CoreIndex           types.CoreIndex
	SegmentRootMappings []SegmentRootMapping
	Bundle              []byte
}

// Helper to decode segment-root mappings from the stream
func readSegmentRootMappings(r io.Reader) ([]SegmentRootMapping, error) {
	var mappings []SegmentRootMapping
	var countBuf [1]byte
	if _, err := io.ReadFull(r, countBuf[:]); err != nil {
		return nil, err
	}
	count := int(countBuf[0])
	for i := 0; i < count; i++ {
		var wpHash types.WorkPackageHash
		var segRoot types.OpaqueHash
		if _, err := io.ReadFull(r, wpHash[:]); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(r, segRoot[:]); err != nil {
			return nil, err
		}
		mappings = append(mappings, SegmentRootMapping{wpHash, segRoot})
	}
	return mappings, nil
}

// Handler for CE 134: Guarantor <-> Guarantor work-package sharing
// Accepts any io.ReadWriteCloser for testability, and allows injection of a PVMExecutor and store maps for unit tests
func HandleWorkPackageShare(
	blockchain blockchain.Blockchain,
	stream io.ReadWriteCloser,
	keypair keystore.KeyPair,
	pvmExecutor work_package.PVMExecutor,
	erasureMap *store.SegmentErasureMap,
	segmentRootLookup *store.HashSegmentMap,
) error {
	// 1. Read core index (2 bytes, little endian)
	coreIndexBuf := make([]byte, 2)
	if _, err := io.ReadFull(stream, coreIndexBuf); err != nil {
		return err
	}
	coreIndex := types.CoreIndex(binary.LittleEndian.Uint16(coreIndexBuf))

	// 2. Read segment-root mappings
	_, err := readSegmentRootMappings(stream)
	if err != nil {
		return fmt.Errorf("failed to read segment-root mappings: %w", err)
	}
	// (Note: mappings are not used in this minimal handler; add logic as needed)

	// 3. Read work-package bundle (rest of stream until FIN)
	bundle := make([]byte, 65536)
	n, err := stream.Read(bundle)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read bundle: %w", err)
	}
	bundle = bundle[:n]

	controller := work_package.NewSharedController(bundle, erasureMap, segmentRootLookup, coreIndex)
	if pvmExecutor != nil {
		controller.PVM = pvmExecutor
	}
	workReport, err := controller.Process()
	if err != nil {
		return fmt.Errorf("work-package verification failed: %w", err)
	}

	workReportHash := workReport.PackageSpec.Hash

	sig, err := keypair.Sign(workReportHash[:])
	if err != nil {
		return fmt.Errorf("failed to sign work-report hash: %w", err)
	}

	resp := append(workReportHash[:], sig...)
	if _, err := stream.Write(resp); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}
	return stream.Close()
}
