package fuzz

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type MessageType uint8

const (
	MessageType_PeerInfo     MessageType = 0
	MessageType_ImportBlock  MessageType = 1
	MessageType_SetState     MessageType = 2
	MessageType_GetState     MessageType = 3
	MessageType_State        MessageType = 4
	MessageType_StateRoot    MessageType = 5
	MessageType_ErrorMessage MessageType = 6
)

const (
	// Size constants for serialization
	sizeOfUint8  = 1 // uint8 occupies 1 byte
	sizeOfUint32 = 4 // uint32 occupies 4 bytes
)

type (
	Version struct {
		Major uint8
		Minor uint8
		Patch uint8
	}

	Features uint32

	PeerInfo struct {
		FuzzVersion  uint8
		FuzzFeatures Features
		AppVersion   Version
		JamVersion   Version
		AppName      string
	}

	ImportBlock types.Block

	ErrorMessage struct {
		Error string
	}

	SetState struct {
		Header types.Header
		State  types.StateKeyVals
	}

	GetState types.HeaderHash

	StateRoot types.StateRoot

	State types.StateKeyVals

	Message struct {
		Type MessageType

		PeerInfo    *PeerInfo
		ImportBlock *ImportBlock
		SetState    *SetState
		GetState    *GetState
		StateRoot   *StateRoot
		State       *State

		// For ImportBlock
		Error *ErrorMessage
	}
)

func (v *Version) FromString(s string) error {
	strParts := strings.Split(s, ".")
	if len(strParts) != 3 {
		return ErrInvalidVersion
	}

	var parts [3]uint8
	for i, strPart := range strParts {
		part, err := strconv.ParseUint(strPart, 10, 8)
		if err != nil {
			return err
		}

		parts[i] = uint8(part)
	}

	v.Major = parts[0]
	v.Minor = parts[1]
	v.Patch = parts[2]

	return nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) AppendBinary(data []byte) ([]byte, error) {
	return append(data, v.Major, v.Minor, v.Patch), nil
}

func (v *Version) ReadFrom(reader io.Reader) (int64, error) {
	buffer := make([]byte, 3)

	n, err := io.ReadFull(reader, buffer)
	if err != nil {
		return int64(n), err
	}

	v.Major = buffer[0]
	v.Minor = buffer[1]
	v.Patch = buffer[2]

	return int64(n), nil
}

func (m *ErrorMessage) MarshalBinary() ([]byte, error) {
	var buffer []byte
	buffer = append(buffer, uint8(len(m.Error)))
	buffer = append(buffer, []byte(m.Error)...)
	return buffer, nil
}

func (m *ErrorMessage) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	l, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	errorBuffer := make([]byte, uint8(l))
	_, err = io.ReadFull(buffer, errorBuffer)
	if err != nil {
		return err
	}

	m.Error = string(errorBuffer)

	return nil
}

func (m *Features) MarshalBinary() ([]byte, error) {
	return marshalUint32(uint32(*m)), nil
}

func (m *Features) UnmarshalBinary(data []byte) error {
	*m = Features(unmarshalUint32(data))
	return nil
}

func (m *PeerInfo) FromConfig() error {
	fmt.Println(config.Config)
	return m.FromValues(
		config.Config.Info.Name,
		config.Config.Info.AppVersion,
		config.Config.Info.JamVersion,
		config.Config.Info.FuzzVersion,
		config.Config.Info.FuzzFeatures,
	)
}

func (m *PeerInfo) FromValues(name, strAppVersion, strJamVersion string, fuzzVersion uint8, fuzzFeatures uint32) error {
	var appVersion, jamVersion Version

	log.Println("here 1")
	log.Println(strAppVersion)

	err := appVersion.FromString(strAppVersion)
	if err != nil {
		return err
	}

	log.Println("here 2")

	err = jamVersion.FromString(strJamVersion)
	if err != nil {
		return err
	}

	log.Println("here 3")

	m.FuzzVersion = fuzzVersion
	m.FuzzFeatures = Features(fuzzFeatures)
	m.JamVersion = jamVersion
	m.AppVersion = appVersion
	if name != "" {
		m.AppName = name
	} else {
		m.AppName = "JAM-Protocol"
	}
	return nil
}

func (m *PeerInfo) MarshalBinary() ([]byte, error) {
	var buffer []byte

	// Append size of fuzz version
	buffer = append(buffer, byte(sizeOfUint8))
	buffer = append(buffer, marshalUint8(m.FuzzVersion)...)

	// Append size of fuzz features
	buffer = append(buffer, byte(sizeOfUint32))
	features, err := m.FuzzFeatures.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buffer = append(buffer, features...)

	// Append size of jam version
	buffer, err = m.JamVersion.AppendBinary(buffer)
	if err != nil {
		return nil, err
	}

	// Append size of app version
	buffer, err = m.AppVersion.AppendBinary(buffer)
	if err != nil {
		return nil, err
	}
	// Append size of app name
	buffer = append(buffer, uint8(len(m.AppName)))
	buffer = append(buffer, []byte(m.AppName)...)

	return buffer, nil
}

func (m *PeerInfo) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)

	// the first byte is the size of the fuzz version
	fuzzVersionSize, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	fuzzVersionBuffer := make([]byte, fuzzVersionSize)
	_, err = io.ReadFull(buffer, fuzzVersionBuffer)
	if err != nil {
		return err
	}
	fuzzVersion := unmarshalUint8(fuzzVersionBuffer)

	// fuzzfeature, 4 bytes
	fuzzFeaturesSize, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	fuzzFeaturesBuffer := make([]byte, fuzzFeaturesSize)
	_, err = io.ReadFull(buffer, fuzzFeaturesBuffer)
	if err != nil {
		return err
	}
	fuzzFeatures := unmarshalUint32(fuzzFeaturesBuffer)

	var appVersion, jamVersion Version

	_, err = appVersion.ReadFrom(buffer)
	if err != nil {
		return err
	}

	_, err = jamVersion.ReadFrom(buffer)
	if err != nil {
		return err
	}

	l, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	nameBuffer := make([]byte, uint8(l))
	_, err = io.ReadFull(buffer, nameBuffer)
	if err != nil {
		return err
	}

	m.AppName = string(nameBuffer)
	m.AppVersion = appVersion
	m.JamVersion = jamVersion
	m.FuzzVersion = fuzzVersion
	m.FuzzFeatures = Features(fuzzFeatures)

	return nil
}

func (m *ImportBlock) MarshalBinary() ([]byte, error) {
	encoder := types.NewEncoder()
	return encoder.Encode((*types.Block)(m))
}

func (m *ImportBlock) UnmarshalBinary(data []byte) error {
	decoder := types.NewDecoder()
	return decoder.Decode(data, (*types.Block)(m))
}

func (m *SetState) Encode(e *types.Encoder) error {
	if err := m.Header.Encode(e); err != nil {
		return err
	}

	if err := m.State.Encode(e); err != nil {
		return err
	}

	return nil
}

func (m *SetState) Decode(d *types.Decoder) error {
	if err := m.Header.Decode(d); err != nil {
		return err
	}

	if err := m.State.Decode(d); err != nil {
		return err
	}

	return nil
}

func (m *SetState) MarshalBinary() ([]byte, error) {
	encoder := types.NewEncoder()
	return encoder.Encode(m)
}

func (m *SetState) UnmarshalBinary(data []byte) error {
	decoder := types.NewDecoder()
	return decoder.Decode(data, m)
}

func (m *GetState) MarshalBinary() ([]byte, error) {
	encoder := types.NewEncoder()
	return encoder.Encode((*types.HeaderHash)(m))
}

func (m *GetState) UnmarshalBinary(data []byte) error {
	decoder := types.NewDecoder()
	return decoder.Decode(data, (*types.HeaderHash)(m))
}

func (m *State) MarshalBinary() ([]byte, error) {
	encoder := types.NewEncoder()
	return encoder.Encode((*types.StateKeyVals)(m))
}

func (m *State) UnmarshalBinary(data []byte) error {
	decoder := types.NewDecoder()
	return decoder.Decode(data, (*types.StateKeyVals)(m))
}

func (m *StateRoot) MarshalBinary() ([]byte, error) {
	encoder := types.NewEncoder()
	return encoder.Encode((*types.StateRoot)(m))
}

func (m *StateRoot) UnmarshalBinary(data []byte) error {
	decoder := types.NewDecoder()
	return decoder.Decode(data, (*types.StateRoot)(m))
}

func (m *Message) ReadFrom(reader io.Reader) (int64, error) {
	var encodedMessageLength uint32

	totalBytesRead := int64(0)

	// message := encodedMessageLength (4 bytes) | encodedMessage (encodedMessageLength bytes)
	// encodedMessage := message type (1 byte) | payload (encodedMessageLength - 1 bytes)
	err := binary.Read(reader, binary.LittleEndian, &encodedMessageLength)
	if err != nil {
		return totalBytesRead, err
	}
	totalBytesRead += 4

	err = binary.Read(reader, binary.LittleEndian, &m.Type)
	if err != nil {
		return totalBytesRead, err
	}
	totalBytesRead += 1

	payload := make([]byte, encodedMessageLength-1)
	bytesRead, err := io.ReadFull(reader, payload)
	totalBytesRead += int64(bytesRead)
	if err != nil {
		return totalBytesRead, err
	}

	var unmarshaler encoding.BinaryUnmarshaler

	switch m.Type {
	case MessageType_PeerInfo:
		m.PeerInfo = new(PeerInfo)
		unmarshaler = m.PeerInfo
	case MessageType_ImportBlock:
		m.ImportBlock = new(ImportBlock)
		unmarshaler = m.ImportBlock
	case MessageType_SetState:
		m.SetState = new(SetState)
		unmarshaler = m.SetState
	case MessageType_GetState:
		m.GetState = new(GetState)
		unmarshaler = m.GetState
	case MessageType_StateRoot:
		m.StateRoot = new(StateRoot)
		unmarshaler = m.StateRoot
	case MessageType_ErrorMessage:
		m.Error = new(ErrorMessage)
		unmarshaler = m.Error
	case MessageType_State:
		m.State = new(State)
		unmarshaler = m.State
	default:
		return totalBytesRead, ErrInvalidMessageType
	}

	err = unmarshaler.UnmarshalBinary(payload)

	return totalBytesRead, err
}

func (m *Message) MarshalBinary() ([]byte, error) {
	var marshaler encoding.BinaryMarshaler

	switch m.Type {
	case MessageType_PeerInfo:
		marshaler = m.PeerInfo
	case MessageType_ImportBlock:
		marshaler = m.ImportBlock
	case MessageType_SetState:
		marshaler = m.SetState
	case MessageType_GetState:
		marshaler = m.GetState
	case MessageType_StateRoot:
		marshaler = m.StateRoot
	case MessageType_State:
		marshaler = m.State
	case MessageType_ErrorMessage:
		marshaler = m.Error
	default:
		return nil, ErrInvalidMessageType
	}

	payload, err := marshaler.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encodedMessageLength := uint32(len(payload) + 1)

	buffer := make([]byte, 0, encodedMessageLength+4)

	buffer, err = binary.Append(buffer, binary.LittleEndian, encodedMessageLength)
	if err != nil {
		return nil, err
	}

	buffer = append(buffer, byte(m.Type))
	buffer = append(buffer, payload...)

	return buffer, nil
}
