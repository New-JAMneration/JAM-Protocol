package stf

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

func ValidateBlock(block types.Block) error {
	header := block.Header
	extrinsics := block.Extrinsic

	if err := validateExtrinsicHash(header, extrinsics); err != nil {
		return err
	}

	return nil
}

func validateExtrinsicHash(header types.Header, extrinsics types.Extrinsic) error {
	calculatedExtrinsicHash, err := createExtrinsicHash(extrinsics)
	if err != nil {
		return err
	}

	if calculatedExtrinsicHash != header.ExtrinsicHash {
		error := fmt.Errorf("invalid extrinsic hash")
		return error
	}

	return nil
}

// CreateExtrinsicHash creates the extrinsic hash of the header.
// (5.4), (5.5), (5.6)
// H_x: extrinsic hash
func createExtrinsicHash(extrinsic types.Extrinsic) (types.OpaqueHash, error) {
	// Encode the extrinsic elements
	encodedTicketsExtrinsic, err := utilities.EncodeExtrinsicTickets(extrinsic.Tickets)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	encodedPreimagesExtrinsic, err := utilities.EncodeExtrinsicPreimages(extrinsic.Preimages)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	encodedGuaranteesExtrinsic, err := utilities.EncodeExtrinsicGuarantees(extrinsic.Guarantees)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	encodedAssurancesExtrinsic, err := utilities.EncodeExtrinsicAssurances(extrinsic.Assurances)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	encodedDisputesExtrinsic, err := utilities.EncodeExtrinsicDisputes(extrinsic.Disputes)
	if err != nil {
		return types.OpaqueHash{}, err
	}

	// Hash encoded elements
	encodedTicketsHash := hash.Blake2bHash(encodedTicketsExtrinsic)
	encodedPreimagesHash := hash.Blake2bHash(encodedPreimagesExtrinsic)
	encodedGuaranteesHash := hash.Blake2bHash(encodedGuaranteesExtrinsic)
	encodedAssurancesHash := hash.Blake2bHash(encodedAssurancesExtrinsic)
	encodedDisputesHash := hash.Blake2bHash(encodedDisputesExtrinsic)

	// Concatenate the encoded elements
	encodedHash := types.ByteSequence{}
	encodedHash = append(encodedHash, types.ByteSequence(encodedTicketsHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedPreimagesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedGuaranteesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedAssurancesHash[:])...)
	encodedHash = append(encodedHash, types.ByteSequence(encodedDisputesHash[:])...)

	// Hash the encoded elements
	extrinsicHash := hash.Blake2bHash(encodedHash)

	return extrinsicHash, nil
}
