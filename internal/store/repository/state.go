package repository

import (
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

func (repo *Repository) SaveStateRootByHeaderHash(w database.Writer, headerHash types.HeaderHash, stateRoot types.StateRoot) error {

	fmt.Println("HERE SAVE STATE ROOT")
	fmt.Println(string(stateRootKey(headerHash)))

	return w.Put(stateRootKey(headerHash), stateRoot[:])
}

func (repo *Repository) GetStateRootByHeaderHash(r database.Reader, headerHash types.HeaderHash) (types.StateRoot, error) {
	fmt.Println("HERE GET STATE ROOT")
	fmt.Println(string(stateRootKey(headerHash)))

	data, found, err := r.Get(stateRootKey(headerHash))
	if err != nil {
		return types.StateRoot{}, err
	}
	if !found {
		return types.StateRoot{}, fmt.Errorf("state root not found for header hash %x", headerHash)
	}

	var stateRoot types.StateRoot
	copy(stateRoot[:], data)
	return stateRoot, nil
}

func (repo *Repository) SaveStateData(w database.Writer, stateRoot types.StateRoot, stateKeyVals types.StateKeyVals) error {
	data, err := repo.encoder.Encode(&stateKeyVals)
	if err != nil {
		return fmt.Errorf("failed to encode state data: %w", err)
	}
	return w.Put(stateDataKey(stateRoot), data)
}

func (repo *Repository) GetStateData(r database.Reader, stateRoot types.StateRoot) (types.StateKeyVals, error) {
	data, found, err := r.Get(stateDataKey(stateRoot))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("state data not found for state root %x", stateRoot)
	}

	var stateKeyVals types.StateKeyVals
	err = repo.decoder.Decode(data, &stateKeyVals)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state data: %w", err)
	}

	return stateKeyVals, nil
}

func (repo *Repository) GetStateDataByHeaderHash(r database.Reader, headerHash types.HeaderHash) (types.StateKeyVals, error) {
	stateRoot, err := repo.GetStateRootByHeaderHash(r, headerHash)
	if err != nil {
		return nil, err
	}

	state, err := repo.GetStateData(r, stateRoot)
	if err != nil {
		return nil, err
	}

	return state, nil
}
