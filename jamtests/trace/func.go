package jamtests

import (
	"fmt"
	"log"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/internal/store"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/merklization"
)

func (s *TraceTestCase) Dump() error {
	// Add block, state
	st := store.GetInstance()
	st.AddBlock(s.Block)

	return nil
}

func (s *TraceTestCase) GetPostState() interface{} {
	return s.PostState
}

func (s *TraceTestCase) GetOutput() interface{} {
	return nil
}

func (s *TraceTestCase) ExpectError() error {
	return nil
}

func (s *TraceTestCase) Validate() error {
	stateRoot := merklization.MerklizationState(store.GetInstance().GetPosteriorStates().GetState())

	if stateRoot != s.PostState.StateRoot {
		diff, err := s.CmpKeyVal(stateRoot)
		if err != nil {
			return fmt.Errorf("compare key-val error: %v", err)
		}

		return diff
	}

	return nil
}

func (s *TraceTestCase) CmpKeyVal(stateRoot types.StateRoot) (statDiff error, err error) {
	keyVals, err := merklization.StateEncoder(store.GetInstance().GetPosteriorStates().GetState())
	if err != nil {
		return nil, fmt.Errorf("state encode keyVals failed")
	}

	keyValDiffs, err := merklization.GetStateKeyValsDiff(s.PostState.KeyVals, keyVals)
	if err != nil {
		return nil, fmt.Errorf("get state keyValsDiff failed")
	}

	// decoder := types.NewDecoder()
	var serviceList []types.U32
	var errorStateSlice []string
	for _, keyVal := range keyValDiffs {
		// C(1)-C(16)
		if state, keyExists := keyValMap[keyVal.Key]; keyExists {
			errorStateSlice = append(errorStateSlice, state)
			log.Println("state: ", keyValMap[keyVal.Key])
			// log.Println("actual key-val-diff : ", keyVal.ActualValue)
			// log.Println("expect key-val-diff : ", keyVal.ExpectedValue)
			if keyValMap[keyVal.Key] == "pi" {
				expectedStruct, err := merklization.SingleKeyValToState(keyVal.Key, keyVal.ExpectedValue)
				if err != nil {
					log.Println("failed to parse single key-val to state")
					return nil, err
				}
				fmt.Printf("statistics:%+v\n", store.GetInstance().GetPosteriorStates().GetPi())
				fmt.Println("---------------------------------")
				fmt.Printf("execptedst:%+v\n", expectedStruct)
			}
			continue
		}
		// C(255)
		if keyVal.Key[0] == byte(255) {
			serviceId, err := merklization.DecodeServiceIdFromType2(keyVal.Key)
			if err != nil {
				return nil, err
			}
			log.Println("serviceID: ", serviceId)
			log.Println("actual key-val-diff : ", keyVal.ActualValue)
			log.Println("expect key-val-diff : ", keyVal.ExpectedValue)
			serviceList = append(serviceList, types.U32(serviceId))
			continue
		}
		// a_s,  a_p,  a_l

	}
	/*
		for _, v := range keyValDiffs {
			newDecoder := types.NewDecoder()
			state := types.State{}
			keyType := keyValMap[v.Key]
			stateType, err := getStateField(state, keyType)
			if err != nil {
				return nil, err
			}

			expectedValue := reflect.New(reflect.TypeOf(stateType)).Interface()
			err = newDecoder.Decode(v.ExpectedValue, expectedValue)
			if err != nil {
				return nil, fmt.Errorf("decode expectedValue failed: %v", err)
			}

			actualValue := reflect.New(reflect.TypeOf(stateType)).Interface()
			err = newDecoder.Decode(v.ActualValue, actualValue)
			if err != nil {
				return nil, fmt.Errorf("decode actualValue failed: %v", err)
			}
		}
	*/
	var diff string

	if len(serviceList) > 0 || len(errorStateSlice) > 0 {
		var deltaDiffStr string
		if len(serviceList) > 0 {
			deltaDiffStr = fmt.Sprintf("delta:%v", serviceList)
		}
		errorStateSlice = append(errorStateSlice, deltaDiffStr)
		diff = strings.Join(errorStateSlice, ", ")
	} else {
		diff = "check account-storage, account-lookupDict, account-primageLookupDict"
	}

	return fmt.Errorf("%s", diff), nil
}

var keyValMap = map[types.StateKey]string{
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "alpha",
	{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "delta",
	{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "beta",
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "gamma",
	{5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "psi",
	{6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "eta",
	{7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "iota",
	{8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "kappa",
	{9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}:  "lambda",
	{10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "rho",
	{11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "tau",
	{12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "Chi",
	{13, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "pi",
	{14, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "theta",
	{15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "xi",
	{16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}: "vartheta",
}
