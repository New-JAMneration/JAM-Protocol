package merklization

import "github.com/New-JAMneration/JAM-Protocol/internal/types"

// D.1. Serialization. The serialization of state primarily involves placing all the various components of σ into a single
// mapping from 31-octet sequence state-keys to octet sequences of indefinite length. The state-key is constructed from a
// hash component and a chapter component, equivalent to either the index of a state component or, in the case of the
// inner dictionaries of δ, a service index.

// (D.1) type 1
func C(stateIndex types.U8) types.StateKey {
	stateWrapper := StateWrapper{StateIndex: stateIndex}
	stateKey := stateWrapper.StateKeyConstruct()
	return stateKey
}

// (D.1) type 2
func DecodeServiceIdFromType2(stateKey types.StateKey) (types.ServiceId, error) {
	// Decode the service Id from the state key
	// service id = [k1, k3, k5, k7]
	encodedServiceId := types.ByteSequence{stateKey[1], stateKey[3], stateKey[5], stateKey[7]}
	decoder := types.NewDecoder()
	var serviceId types.ServiceId
	err := decoder.Decode(encodedServiceId, &serviceId)
	if err != nil {
		return types.ServiceId(0), err
	}

	return serviceId, nil
}

// (D.1) type 3
func decodeServiceIdFromType3(stateKey types.StateKey) (types.ServiceId, error) {
	// Decode the service Id from the state key
	// service id = [k0, k2, k4, k6]
	encodedServiceId := types.ByteSequence{stateKey[0], stateKey[2], stateKey[4], stateKey[6]}
	decoder := types.NewDecoder()
	var serviceId types.ServiceId
	err := decoder.Decode(encodedServiceId, &serviceId)
	if err != nil {
		return types.ServiceId(0), err
	}

	return serviceId, nil
}
