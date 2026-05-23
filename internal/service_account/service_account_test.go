package service_account

import (
	"reflect"
	"testing"

	types "github.com/New-JAMneration/JAM-Protocol/internal/types"
	utils "github.com/New-JAMneration/JAM-Protocol/internal/utilities"
	hash "github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

// seedGlobalKV mirrors the legacy StorageDict / LookupDict map entries into
// globalKV via the Insert* methods so the resulting ServiceAccount behaves
// like one populated incrementally through the post-globalKV API. Test
// fixtures that still construct accounts with struct literals call this
// helper once after construction.
func seedGlobalKV(t *testing.T, account *types.ServiceAccount, serviceID types.ServiceID) {
	t.Helper()
	if err := account.MigrateLegacyMapsToGlobalKV(serviceID); err != nil {
		t.Fatalf("MigrateLegacyMapsToGlobalKV: %v", err)
	}
}

func TestFetchCodeByHash(t *testing.T) {
	// set up test data
	var (
		mockMetadata = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
		mockCode     = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
	)

	// encode metaCode
	encoder := types.NewEncoder()
	encodedMetaCode, err := encoder.Encode(&types.MetaCode{
		Metadata: mockMetadata,
		Code:     mockCode,
	})
	if err != nil {
		t.Errorf("Error encoding MetaCode: %v", err)
	}

	mockCodeHash := hash.Blake2bHash(encodedMetaCode)

	// create ServiceAccount
	mockAccount := types.ServiceAccount{
		PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
			mockCodeHash: encodedMetaCode,
		},
	}

	// FetchCodeByHash reads PreimageLookup directly (no globalKV access on
	// the happy path), so seedGlobalKV isn't strictly required here. We
	// still call it for parity with the other tests / to exercise
	// ValidatePreimageLookupDict's fallback path correctly.
	if mockAccount.LookupDict == nil {
		mockAccount.LookupDict = make(types.LookupMetaMapEntry)
	}
	mockAccount.LookupDict[types.LookupMetaMapkey{
		Hash:   mockCodeHash,
		Length: types.U32(len(encodedMetaCode)),
	}] = types.TimeSlotSet{}
	seedGlobalKV(t, &mockAccount, types.ServiceID(0))

	metadata, code, err := FetchCodeByHash(types.ServiceID(0), mockAccount, mockCodeHash)
	if err != nil {
		t.Fatalf("FetchCodeByHash failed: %v", err)
	}

	// check if code is equal to mockCode
	if code == nil {
		t.Errorf("FetchCodeByHash failed: non exist code for codeHash %v", mockCodeHash)
	} else if !reflect.DeepEqual(code, mockCode) {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockCode, code)
	}
	if metadata == nil {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockMetadata, metadata)
	} else if !reflect.DeepEqual(metadata, mockMetadata) {
		t.Errorf("FetchCodeByHash failed: expected %v, got %v", mockMetadata, metadata)
	}
}

func TestValidatePreimageLookupDict(t *testing.T) {
	// ∀a ∈ A, (h ↦ p) ∈ a_p ⇒ h = H(p) ∧ (h, |p|) ∈ K(a_l)

	// set up test data
	var (
		// mockCodeHash = hash(mockCode) -> preimage of mockCodeHash = mockCode
		// mockCode     = types.ByteSequence("0x92cdf578c47085a5992256f0dcf97d0b19f1")
		mockCode     = types.ByteSequence("0x5b5477bef56d05dd59b758c2c4672c88aa8a71a2949f3921f37a25a9a167aeba")
		mockCodeHash = hash.Blake2bHash(utils.ByteSequenceWrapper{Value: mockCode}.Serialize())
		// mockCodeHash_bs  = types.ByteSequence(mockCodeHash[:])
		// mockCodeHash_str = hex.EncodeToString(mockCodeHash_bs)
		preimage = mockCode

		// create ServiceAccount
		mockAccount = types.ServiceAccount{
			// h = H(p)
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: preimage,
			},
			// (h, |p|) ∈ K(a_l)
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(preimage))}: {},
			},
		}
	)

	// Mirror the legacy LookupDict / PreimageLookup contents into globalKV so
	// ValidatePreimageLookupDict's a_l[h, |p|] membership check works
	// post-refactor.
	seedGlobalKV(t, &mockAccount, types.ServiceID(0))

	err := ValidatePreimageLookupDict(types.ServiceID(0), mockAccount)
	if err != nil {
		t.Errorf("ValidateAccount failed: %v", err)
	}
}

func TestHistoricalLookupFunction(t *testing.T) {
	t.Run("basic success case", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata  = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode      = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			mockTimestamp = types.TimeSlot(42)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {mockTimestamp},
			},
		}

		// test HistoricalLookup
		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, mockTimestamp, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("HistoricalLookup failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}
	})

	t.Run("code hash not exist case", func(t *testing.T) {
		mockCodeHash := types.OpaqueHash{}
		mockTimestamp := types.TimeSlot(42)
		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{},
			LookupDict:     map[types.LookupMetaMapkey]types.TimeSlotSet{},
		}

		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, mockTimestamp, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil for non-existent hash, got metadata=%v, code=%v", metadata, code)
		}
	})

	t.Run("timestamp invalid case - empty timeslot set", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata  = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode      = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			mockTimestamp = types.TimeSlot(42)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {}, // empty timeslot set
			},
		}

		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, mockTimestamp, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil for invalid timestamp (empty set), got metadata=%v, code=%v", metadata, code)
		}
	})

	t.Run("timestamp invalid case - length 1 timeslot set", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode     = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			lowerTime    = types.TimeSlot(40)
			upperTime    = types.TimeSlot(60)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		// case: timestamp less than set element (should return nil)
		mockAccount1 := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {upperTime}, // timestamp is higher than requested
			},
		}

		seedGlobalKV(t, &mockAccount1, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount1, lowerTime, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp < only time in set, got metadata=%v, code=%v", metadata, code)
		}

		// case: timestamp greater than or equal to set element (should succeed)
		mockAccount2 := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {lowerTime}, // 時間與請求相同
			},
		}

		seedGlobalKV(t, &mockAccount2, types.ServiceID(0))
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount2, lowerTime, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}
	})

	t.Run("timestamp invalid case - length 2 timeslot set", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode     = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			lowerTime    = types.TimeSlot(40)
			upperTime    = types.TimeSlot(60)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		// case: timestamp less than set element (should return nil)
		mockAccount1 := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {lowerTime, upperTime},
			},
		}

		seedGlobalKV(t, &mockAccount1, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount1, lowerTime-1, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp < lower bound, got metadata=%v, code=%v", metadata, code)
		}

		// case: timestamp equals lower bound (should succeed)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount1, lowerTime, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}

		// case: timestamp within range (should succeed)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount1, lowerTime+1, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}

		// case: timestamp equals upper bound (should return nil, because upper bound is open interval)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount1, upperTime, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp = upper bound, got metadata=%v, code=%v", metadata, code)
		}

		// case: timestamp greater than upper bound (should return nil)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount1, upperTime+1, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp > upper bound, got metadata=%v, code=%v", metadata, code)
		}
	})

	t.Run("timestamp invalid case - length 3 timeslot set", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode     = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			lowerTime    = types.TimeSlot(40)
			upperTime    = types.TimeSlot(60)
			thirdTime    = types.TimeSlot(80)
			middleTime1  = types.TimeSlot(50)
			middleTime2  = types.TimeSlot(70)
			highTime     = types.TimeSlot(90)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		// create ServiceAccount, time set is [lowerTime, upperTime, thirdTime]
		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {lowerTime, upperTime, thirdTime},
			},
		}

		// case 1: timestamp less than lower bound (should return nil)
		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, lowerTime-1, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp < lower bound, got metadata=%v, code=%v", metadata, code)
		}

		// case 2: timestamp within first range (should succeed)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount, middleTime1, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}

		// case 3: timestamp between ranges (should return nil)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount, middleTime2, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when timestamp between ranges, got metadata=%v, code=%v", metadata, code)
		}

		// case 4: timestamp equals third time slot (should succeed)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount, thirdTime, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}

		// case 5: timestamp greater than third time slot (should succeed)
		bytes = HistoricalLookup(types.ServiceID(0), mockAccount, highTime, mockCodeHash)
		metadata, code, err = DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if !reflect.DeepEqual(code, mockCode) {
			t.Errorf("HistoricalLookup failed: expected code %v, got %v", mockCode, code)
		}
		if !reflect.DeepEqual(metadata, mockMetadata) {
			t.Errorf("HistoricalLookup failed: expected metadata %v, got %v", mockMetadata, metadata)
		}
	})

	t.Run("timestamp invalid case - length greater than 3 timeslot set", func(t *testing.T) {
		// set up test data
		var (
			mockMetadata  = types.ByteSequence([]byte{0x01, 0x02, 0x03, 0x04, 0x05})
			mockCode      = types.ByteSequence([]byte{0x06, 0x07, 0x08, 0x09, 0x0A})
			mockTimestamp = types.TimeSlot(42)
		)

		// encode metaCode
		testMetaCode := types.MetaCode{
			Metadata: mockMetadata,
			Code:     mockCode,
		}
		encoder := types.NewEncoder()
		encodedMetaCode, err := encoder.Encode(&testMetaCode)
		if err != nil {
			t.Fatalf("Error encoding MetaCode: %v", err)
		}
		mockCodeHash := hash.Blake2bHash(encodedMetaCode)

		// create ServiceAccount, time slot set length is 4
		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: encodedMetaCode,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(encodedMetaCode))}: {10, 20, 30, 40},
			},
		}

		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, mockTimestamp, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err != nil {
			t.Fatalf("DecodeMetaCode failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil for time sets with length > 3, got metadata=%v, code=%v", metadata, code)
		}
	})

	t.Run("decode failure case", func(t *testing.T) {
		// set up invalid encoded data
		mockCodeHash := types.OpaqueHash{}
		mockTimestamp := types.TimeSlot(42)
		invalidEncodedData := types.ByteSequence([]byte{0x01}) // invalid encoded data

		mockAccount := types.ServiceAccount{
			PreimageLookup: map[types.OpaqueHash]types.ByteSequence{
				mockCodeHash: invalidEncodedData,
			},
			LookupDict: map[types.LookupMetaMapkey]types.TimeSlotSet{
				{Hash: mockCodeHash, Length: types.U32(len(invalidEncodedData))}: {mockTimestamp},
			},
		}

		// when decode fails, function should return nil
		seedGlobalKV(t, &mockAccount, types.ServiceID(0))
		bytes := HistoricalLookup(types.ServiceID(0), mockAccount, mockTimestamp, mockCodeHash)
		metadata, code, err := DecodeMetaCode(bytes)
		if err == nil {
			t.Fatalf("DecodeMetaCode should failed: %v", err)
		}
		if metadata != nil || code != nil {
			t.Errorf("HistoricalLookup should return nil when decoding fails, got metadata=%v, code=%v", metadata, code)
		}
	})
}

// TestServiceAccountFootprintFromCounters checks that ThresholdBalance reads
// a_i / a_o straight from the incremental counters (post-Step-3) and that
// the GP §9.8 formula at = max(0, BS + BI*ai + BL*ao - af) is honoured.
//
// The legacy GetServiceAccountDerivatives helper has been retired in Step 8a;
// this test replaces the older derivative coverage with a counter-based
// equivalent.
func TestServiceAccountFootprintFromCounters(t *testing.T) {
	var account types.ServiceAccount
	account.SetTotalNumberOfItems(2)
	account.SetTotalNumberOfOctets(81 + 9)

	if got, want := CalcKeys(account), types.U32(2); got != want {
		t.Fatalf("CalcKeys: got %d, want %d", got, want)
	}
	if got, want := CalcOctets(account), types.U64(90); got != want {
		t.Fatalf("CalcOctets: got %d, want %d", got, want)
	}

	at, err := account.ThresholdBalance()
	if err != nil {
		t.Fatalf("ThresholdBalance returned err: %v", err)
	}
	// a_f defaults to 0, so a_t = BS + BI*2 + BL*90.
	want := types.U64(types.BasicMinBalance) +
		types.U64(types.AdditionalMinBalancePerItem)*2 +
		types.U64(types.AdditionalMinBalancePerOctet)*90
	if at != want {
		t.Fatalf("ThresholdBalance: got %d, want %d", at, want)
	}

	// Now exercise the GP §9.8 max(0, ... - a_f) clause with a non-zero
	// gratis storage offset. Before the fix this branch incorrectly
	// returned `storage` instead of `storage - a_f`.
	t.Run("with non-zero deposit offset", func(t *testing.T) {
		account.ServiceInfo.DepositOffset = 50
		at, err := account.ThresholdBalance()
		if err != nil {
			t.Fatalf("ThresholdBalance returned err: %v", err)
		}
		if at != want-50 {
			t.Fatalf("ThresholdBalance with a_f=50: got %d, want %d", at, want-50)
		}
	})

	t.Run("deposit offset larger than storage clamps to zero", func(t *testing.T) {
		account.ServiceInfo.DepositOffset = types.U64(want) + 1000
		at, err := account.ThresholdBalance()
		if err != nil {
			t.Fatalf("ThresholdBalance returned err: %v", err)
		}
		if at != 0 {
			t.Fatalf("ThresholdBalance clamp: got %d, want 0", at)
		}
	})
}
