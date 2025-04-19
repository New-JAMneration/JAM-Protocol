package jamtests

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// ANSI color codes
var (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

var debugMode = false

// var debugMode = true

func cLog(color string, string string) {
	if debugMode {
		fmt.Printf("%s%s%s\n", color, string, Reset)
	}
}

type PreimageTestCase struct {
	Input     PreimageInput  `json:"input"`
	PreState  PreimageState  `json:"pre_state"`
	Output    PreimageOutput `json:"output"`
	PostState PreimageState  `json:"post_state"`
}

type PreimageInput struct {
	Preimages types.PreimagesExtrinsic `json:"preimages"`
	Slot      types.TimeSlot           `json:"slot"`
}

type PreimageOutputData struct {
	Null *struct{}
}

type PreimageOutput struct {
	Ok  interface{}        `json:"ok,omitempty"`
	Err *PreimageErrorCode `json:"err,omitempty"`
}

type PreimagesMapEntry struct {
	Hash types.OpaqueHash   `json:"hash"`
	Blob types.ByteSequence `json:"blob"`
}

type LookupMetaMapkey struct {
	Hash   types.OpaqueHash `json:"hash"`
	Length types.U32        `json:"length"`
}

type LookupMetaMapEntry struct {
	Key LookupMetaMapkey `json:"key"`
	Val []types.TimeSlot `json:"value"`
}

type Account struct {
	Preimages  []PreimagesMapEntry  `json:"preimages"`
	LookupMeta []LookupMetaMapEntry `json:"lookup_meta"`
}

type AccountsMapEntry struct {
	Id   types.ServiceId `json:"id"`
	Data Account         `json:"data"`
}

type PreimageState struct {
	Accounts []AccountsMapEntry `json:"accounts"`
}

type PreimageErrorCode types.ErrorCode

func (p *PreimageErrorCode) Error() string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *p)
}

const (
	PreimageUnneeded         PreimageErrorCode = iota // 0
	PreimagesNotSortedUnique PreimageErrorCode = 1
)

var preimageErrorMap = map[string]PreimageErrorCode{
	"preimage_unneeded":           PreimageUnneeded,
	"preimages_not_sorted_unique": PreimagesNotSortedUnique,
}

func (e *PreimageErrorCode) UnmarshalJSON(data []byte) error {
	var str string

	if err := json.Unmarshal(data, &str); err == nil {
		if val, ok := preimageErrorMap[str]; ok {
			*e = val
			return nil
		}
		return errors.New("invalid error code name: " + str)
	}
	return errors.New("invalid error code format, expected string")
}

func (p *PreimagesMapEntry) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash string `json:"hash"`
		Blob string `json:"blob"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}

	p.Hash = types.OpaqueHash(hashBytes)

	blobBytes, err := hex.DecodeString(temp.Blob[2:])
	if err != nil {
		return err
	}
	p.Blob = types.ByteSequence(blobBytes)

	return nil
}

func (l *LookupMetaMapkey) UnmarshalJSON(data []byte) error {
	var temp struct {
		Hash   string    `json:"hash"`
		Length types.U32 `json:"length"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	hashBytes, err := hex.DecodeString(temp.Hash[2:])
	if err != nil {
		return err
	}

	l.Hash = types.OpaqueHash(hashBytes)

	l.Length = temp.Length

	return nil
}

// UnmarshalJSON PreimageInput
func (i *PreimageInput) UnmarshalJSON(data []byte) error {
	var temp struct {
		Preimages types.PreimagesExtrinsic `json:"preimages"`
		Slot      types.TimeSlot           `json:"slot"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Preimages) != 0 {
		i.Preimages = temp.Preimages
	}

	i.Slot = temp.Slot

	return nil
}

// UnmarshalJSON LookupMetaMapEntry
func (l *LookupMetaMapEntry) UnmarshalJSON(data []byte) error {
	var temp struct {
		Key LookupMetaMapkey `json:"key"`
		Val []types.TimeSlot `json:"value"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	l.Key = temp.Key
	if len(temp.Val) != 0 {
		l.Val = temp.Val
	}

	return nil
}

func (a *Account) UnmarshalJSON(data []byte) error {
	var temp struct {
		Preimages  []PreimagesMapEntry  `json:"preimages"`
		LookupMeta []LookupMetaMapEntry `json:"lookup_meta"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if len(temp.Preimages) != 0 {
		a.Preimages = temp.Preimages
	}

	if len(temp.LookupMeta) != 0 {
		a.LookupMeta = temp.LookupMeta
	}

	return nil
}

// PreimageInput
func (i *PreimageInput) Decode(d *types.Decoder) error {
	var err error

	if err = i.Preimages.Decode(d); err != nil {
		return err
	}

	if err = i.Slot.Decode(d); err != nil {
		return err
	}

	return nil
}

// PreimageOutput
func (o *PreimageOutput) Decode(d *types.Decoder) error {
	cLog(Yellow, "Decoding PreimageOutput")
	var err error

	okOrErr, err := d.ReadPointerFlag()
	if err != nil {
		return err
	}

	isOk := okOrErr == 0
	if isOk {
		cLog(Yellow, "PreimageOutput is ok")

		return nil
	} else {
		cLog(Yellow, "PreimageOutput is err")
		cLog(Yellow, "Decoding PreimageErrorCode")

		// Read a byte as error code
		errByte, err := d.ReadErrorByte()
		if err != nil {
			return err
		}

		o.Err = (*PreimageErrorCode)(&errByte)

		cLog(Yellow, fmt.Sprintf("PreimageErrorCode: %v", *o.Err))
	}

	return nil
}

// LookupMetaMapkey
func (l *LookupMetaMapkey) Decode(d *types.Decoder) error {
	cLog(White, "Decoding LookupMetaMapkey")

	var err error

	if err = l.Hash.Decode(d); err != nil {
		return err
	}

	if err = l.Length.Decode(d); err != nil {
		return err
	}

	return nil
}

// LookupMetaMapEntry
func (l *LookupMetaMapEntry) Decode(d *types.Decoder) error {
	cLog(White, "Decoding LookupMetaMapEntry")

	var err error

	if err = l.Key.Decode(d); err != nil {
		return err
	}

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length != 0 {
		// make a slice of length length
		l.Val = make([]types.TimeSlot, length)
		for i := uint64(0); i < length; i++ {
			cLog(White, fmt.Sprintf("Decoding TimeSlot %d", i))

			if err = l.Val[i].Decode(d); err != nil {
				return err
			}
		}
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Decode(d *types.Decoder) error {
	cLog(Cyan, "Decoding PreimagesMapEntry")

	var err error

	if err = p.Hash.Decode(d); err != nil {
		return err
	}

	if err = p.Blob.Decode(d); err != nil {
		return err
	}

	return nil
}

// Account
func (a *Account) Decode(d *types.Decoder) error {
	cLog(Gray, "Decoding Account")

	var err error

	preimagesLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if preimagesLength != 0 {
		// make a slice of length length
		a.Preimages = make([]PreimagesMapEntry, preimagesLength)
		for i := uint64(0); i < preimagesLength; i++ {
			cLog(Gray, fmt.Sprintf("Decoding Preimage %d", i))

			if err = a.Preimages[i].Decode(d); err != nil {
				return err
			}
		}
	}

	lookupMetaLength, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if lookupMetaLength != 0 {
		// make a slice of length length
		a.LookupMeta = make([]LookupMetaMapEntry, lookupMetaLength)
		for i := uint64(0); i < lookupMetaLength; i++ {
			cLog(Gray, fmt.Sprintf("Decoding LookupMeta %d", i))

			if err = a.LookupMeta[i].Decode(d); err != nil {
				return err
			}
		}
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Decode(d *types.Decoder) error {
	cLog(Magenta, "Decoding AccountsMapEntry")

	var err error

	if err = a.Id.Decode(d); err != nil {
		return err
	}

	if err = a.Data.Decode(d); err != nil {
		return err
	}

	return nil
}

// PreimageState
func (p *PreimageState) Decode(d *types.Decoder) error {
	cLog(Blue, "Decoding PreimageState")

	var err error

	length, err := d.DecodeLength()
	if err != nil {
		return err
	}

	if length == 0 {
		return nil
	}

	// make a slice of length length
	p.Accounts = make([]AccountsMapEntry, length)
	for i := uint64(0); i < length; i++ {
		cLog(Blue, fmt.Sprintf("Decoding Account %d", i))

		if err = p.Accounts[i].Decode(d); err != nil {
			return err
		}
	}

	return nil
}

// PreiamgeTestCase
func (t *PreimageTestCase) Decode(d *types.Decoder) error {
	var err error

	if err = t.Input.Decode(d); err != nil {
		return err
	}

	if err = t.PreState.Decode(d); err != nil {
		return err
	}

	if err = t.Output.Decode(d); err != nil {
		return err
	}

	if err = t.PostState.Decode(d); err != nil {
		return err
	}

	return nil
}

// Encode
type Encodable interface {
	Encode(e *types.Encoder) error
}

// PreimageInput
func (i *PreimageInput) Encode(e *types.Encoder) error {
	var err error

	if err = i.Preimages.Encode(e); err != nil {
		return err
	}

	if err = i.Slot.Encode(e); err != nil {
		return err
	}

	return nil
}

// PreimageOutput
func (o *PreimageOutput) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding PreimageOutput")
	var err error

	if o.Err != nil {
		cLog(Cyan, "PreimageOutput is Err")
		if err = e.WriteByte(1); err != nil {
			return err
		}

		if err = e.WriteByte(byte(*o.Err)); err != nil {
			return err
		}

		return nil
	}

	if o.Ok == nil {
		cLog(Cyan, "PreimageOutput is Ok")
		if err = e.WriteByte(0); err != nil {
			return err
		}

		// Ok value is empty

		return nil
	}

	return nil
}

// PreimagesMapEntry
func (p *PreimagesMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding PreimagesMapEntry")
	var err error

	if err = p.Hash.Encode(e); err != nil {
		return err
	}

	if err = p.Blob.Encode(e); err != nil {
		return err
	}

	return nil
}

// LookupMetaMapkey
func (l *LookupMetaMapkey) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding LookupMetaMapkey")
	var err error

	if err = l.Hash.Encode(e); err != nil {
		return err
	}

	if err = l.Length.Encode(e); err != nil {
		return err
	}

	return nil
}

// LookupMetaMapEntry
func (l *LookupMetaMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding LookupMetaMapEntry")
	var err error

	if err = l.Key.Encode(e); err != nil {
		return err
	}

	if err = e.EncodeLength(uint64(len(l.Val))); err != nil {
		return err
	}

	for _, timeSlot := range l.Val {
		if err = timeSlot.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// Account
func (a *Account) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding Account")
	var err error

	if err = e.EncodeLength(uint64(len(a.Preimages))); err != nil {
		return err
	}

	for _, preimage := range a.Preimages {
		if err = preimage.Encode(e); err != nil {
			return err
		}
	}

	if err = e.EncodeLength(uint64(len(a.LookupMeta))); err != nil {
		return err
	}

	for _, lookupMeta := range a.LookupMeta {
		if err = lookupMeta.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// AccountsMapEntry
func (a *AccountsMapEntry) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding AccountsMapEntry")
	var err error

	if err = a.Id.Encode(e); err != nil {
		return err
	}

	if err = a.Data.Encode(e); err != nil {
		return err
	}

	return nil
}

// PreimageState
func (p *PreimageState) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding PreimageState")
	var err error

	if err = e.EncodeLength(uint64(len(p.Accounts))); err != nil {
		return err
	}

	for _, account := range p.Accounts {
		if err = account.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

// PreimageTestCase
func (t *PreimageTestCase) Encode(e *types.Encoder) error {
	cLog(Cyan, "Encoding PreimageTestCase")
	var err error

	if err = t.Input.Encode(e); err != nil {
		return err
	}

	if err = t.PreState.Encode(e); err != nil {
		return err
	}

	if err = t.Output.Encode(e); err != nil {
		return err
	}

	if err = t.PostState.Encode(e); err != nil {
		return err
	}

	return nil
}

// TODO: Implement Dump method
func (p *PreimageTestCase) Dump() error {
	return nil
}

func (p *PreimageTestCase) GetPostState() interface{} {
	return p.PostState
}

func (p *PreimageTestCase) GetOutput() interface{} {
	return p.Output
}

func (p *PreimageTestCase) ExpectError() error {
	if p.Output.Err == nil {
		return nil
	}
	return p.Output.Err
}

func (p *PreimageTestCase) Validate() error {
	return nil
}
