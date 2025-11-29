package merklization

import (
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// C(1)
func decodeAlpha(encodedValue types.ByteSequence) (types.AuthPools, error) {
	output := types.AuthPools{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.AuthPools{}, err
	}

	return output, nil
}

// C(2)
func decodeVarphi(encodedValue types.ByteSequence) (types.AuthQueues, error) {
	output := types.AuthQueues{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.AuthQueues{}, err
	}

	return output, nil
}

// C(3)
func decodeBeta(encodedValue types.ByteSequence) (types.RecentBlocks, error) {
	output := types.RecentBlocks{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.RecentBlocks{}, err
	}

	return output, nil
}

// C(4)
func decodeGamma(encodedValue types.ByteSequence) (types.Gamma, error) {
	output := types.Gamma{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.Gamma{}, err
	}

	return output, nil
}

// C(5)
func decodePsi(encodedValue types.ByteSequence) (types.DisputesRecords, error) {
	output := types.DisputesRecords{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.DisputesRecords{}, err
	}

	return output, nil
}

// C(6)
func decodeEta(encodedValue types.ByteSequence) (types.EntropyBuffer, error) {
	output := types.EntropyBuffer{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.EntropyBuffer{}, err
	}

	return output, nil
}

// C(7)
func decodeIota(encodedValue types.ByteSequence) (types.ValidatorsData, error) {
	output := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.ValidatorsData{}, err
	}

	return output, nil
}

// C(8)
func decodeKappa(encodedValue types.ByteSequence) (types.ValidatorsData, error) {
	output := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.ValidatorsData{}, err
	}

	return output, nil
}

// C(9)
func decodeLambda(encodedValue types.ByteSequence) (types.ValidatorsData, error) {
	output := types.ValidatorsData{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.ValidatorsData{}, err
	}

	return output, nil
}

// C(10)
func decodeRho(encodedValue types.ByteSequence) (types.AvailabilityAssignments, error) {
	output := types.AvailabilityAssignments{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.AvailabilityAssignments{}, err
	}

	return output, nil
}

// C(11)
func decodeTau(encodedValue types.ByteSequence) (types.TimeSlot, error) {
	output := types.TimeSlot(0)
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.TimeSlot(0), err
	}

	return output, nil
}

// C(12)
func decodeChi(encodedValue types.ByteSequence) (types.Privileges, error) {
	output := types.Privileges{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.Privileges{}, err
	}

	return output, nil
}

// C(13)
func decodePi(encodedValue types.ByteSequence) (types.Statistics, error) {
	output := types.Statistics{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.Statistics{}, err
	}

	return output, nil
}

// C(14)
func decodeTheta(encodedValue types.ByteSequence) (types.ReadyQueue, error) {
	output := types.ReadyQueue{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.ReadyQueue{}, err
	}

	return output, nil
}

// C(15)
func decodeXi(encodedValue types.ByteSequence) (types.AccumulatedQueue, error) {
	output := types.AccumulatedQueue{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.AccumulatedQueue{}, err
	}

	return output, nil
}

// C(16)
// theta LastAccOut
func decodeThetaAccOut(encodedValue types.ByteSequence) (types.LastAccOut, error) {
	output := types.LastAccOut{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.LastAccOut{}, err
	}

	return output, nil
}

// C(255, s) -> ac ⌢ E8(ab, ag, am, ao) ⌢ E4(ai)
func DecodeServiceInfo(encodedValue types.ByteSequence) (types.ServiceInfo, error) {
	output := types.ServiceInfo{}
	decoder := types.NewDecoder()
	err := decoder.Decode(encodedValue, &output)
	if err != nil {
		return types.ServiceInfo{}, err
	}

	return output, nil
}
