package fuzz

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/New-JAMneration/JAM-Protocol/config"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

type MessageType uint8

const (
	MessageType_PeerInfo    MessageType = 0
	MessageType_ImportBlock MessageType = 1
	MessageType_SetState    MessageType = 2
	MessageType_GetState    MessageType = 3
	MessageType_State       MessageType = 4
	MessageType_StateRoot   MessageType = 5
)

type (
	Version struct {
		Major uint8
		Minor uint8
		Patch uint8
	}

	PeerInfo struct {
		Name       string
		AppVersion Version
		JamVersion Version
	}

	ImportBlock types.Block

	SetState struct {
		Header types.Header
		State  types.StateKeyVals
	}

	GetState types.HeaderHash

	StateRoot types.StateRoot

	State types.StateKeyVals

	Message struct {
		Type    MessageType
		payload any
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

func (m *PeerInfo) FromConfig() error {
	return m.FromValues(
		config.Config.Info.Name,
		config.Config.Info.AppVersion,
		config.Config.Info.JamVersion,
	)
}

func (m *PeerInfo) FromValues(name, strAppVersion, strJamVersion string) error {
	var appVersion, jamVersion Version

	err := appVersion.FromString(strAppVersion)
	if err != nil {
		return err
	}

	err = jamVersion.FromString(strJamVersion)
	if err != nil {
		return err
	}

	m.Name = name
	m.AppVersion = appVersion
	m.JamVersion = jamVersion

	return nil
}

func (m *PeerInfo) MarshalBinary() ([]byte, error) {
	var buffer []byte

	buffer = append(buffer, uint8(len(m.Name)))
	buffer = append(buffer, []byte(m.Name)...)

	buffer, err := m.AppVersion.AppendBinary(buffer)
	if err != nil {
		return nil, err
	}

	buffer, err = m.JamVersion.AppendBinary(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func (m *PeerInfo) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	l, err := buffer.ReadByte()
	if err != nil {
		return err
	}

	nameBuffer := make([]byte, uint8(l))
	_, err = io.ReadFull(buffer, nameBuffer)
	if err != nil {
		return err
	}

	var appVersion, jamVersion Version

	_, err = appVersion.ReadFrom(buffer)
	if err != nil {
		return err
	}

	_, err = jamVersion.ReadFrom(buffer)
	if err != nil {
		return err
	}

	m.Name = string(nameBuffer)
	m.AppVersion = appVersion
	m.JamVersion = jamVersion

	return nil
}

func (m *ImportBlock) MarshalBinary() ([]byte, error) {
	return nil, ErrNotImpl
}

func (m *ImportBlock) UnmarshalBinary(data []byte) error {
	return ErrNotImpl
}

func (m *SetState) MarshalBinary() ([]byte, error) {
	return nil, ErrNotImpl
}

func (m *SetState) UnmarshalBinary(data []byte) error {
	return ErrNotImpl
}

func (m *GetState) MarshalBinary() ([]byte, error) {
	return nil, ErrNotImpl
}

func (m *GetState) UnmarshalBinary(data []byte) error {
	return ErrNotImpl
}

func (m *State) MarshalBinary() ([]byte, error) {
	return nil, ErrNotImpl
}

func (m *State) UnmarshalBinary(data []byte) error {
	return ErrNotImpl
}

func (m *StateRoot) MarshalBinary() ([]byte, error) {
	return nil, ErrNotImpl
}

func (m *StateRoot) UnmarshalBinary(data []byte) error {
	return ErrNotImpl
}

// returns ErrInvalidPayloadType if the payload type doesn't match the expected type based on message type
// returns ErrInvalidMessageType if the message type is not one of the supported values
func (m *Message) Validate() error {
	switch m.Type {
	case MessageType_PeerInfo:
		return checkPayloadType[PeerInfo](m.payload)
	case MessageType_ImportBlock:
		return checkPayloadType[ImportBlock](m.payload)
	case MessageType_SetState:
		return checkPayloadType[SetState](m.payload)
	case MessageType_GetState:
		return checkPayloadType[GetState](m.payload)
	case MessageType_State:
		return checkPayloadType[State](m.payload)
	case MessageType_StateRoot:
		return checkPayloadType[StateRoot](m.payload)
	default:
		return ErrInvalidMessageType
	}
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
	totalBytesRead += 4 // note that bytes read is discarded in binary.Read implementation so we never have that information

	err = binary.Read(reader, binary.LittleEndian, &m.Type)
	if err != nil {
		return totalBytesRead, err
	}
	totalBytesRead += 1 // note that bytes read is discarded in binary.Read implementation so we never have that information

	switch m.Type {
	case MessageType_PeerInfo:
		m.payload = new(PeerInfo)
	case MessageType_ImportBlock:
		m.payload = new(ImportBlock)
	case MessageType_SetState:
		m.payload = new(SetState)
	case MessageType_GetState:
		m.payload = new(GetState)
	case MessageType_State:
		m.payload = new(State)
	case MessageType_StateRoot:
		m.payload = new(StateRoot)
	default:
		return totalBytesRead, ErrInvalidMessageType
	}

	payload := make([]byte, encodedMessageLength-1)
	bytesRead, err := io.ReadFull(reader, payload)
	totalBytesRead += int64(bytesRead)
	if err != nil {
		return totalBytesRead, err
	}

	unmarshaler, valid := m.payload.(encoding.BinaryUnmarshaler)
	if !valid {
		return totalBytesRead, ErrInvalidPayloadType
	}

	err = unmarshaler.UnmarshalBinary(payload)

	return totalBytesRead, err
}

func (m *Message) MarshalBinary() ([]byte, error) {
	marshaler, valid := m.payload.(encoding.BinaryMarshaler)
	if !valid {
		return nil, ErrInvalidPayloadType
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

func (m *Message) PeerInfo() (*PeerInfo, error) {
	return castMessage[PeerInfo](m, MessageType_PeerInfo)
}

func (m *Message) ImportBlock() (*ImportBlock, error) {
	return castMessage[ImportBlock](m, MessageType_ImportBlock)
}

func (m *Message) GetState() (*GetState, error) {
	return castMessage[GetState](m, MessageType_GetState)
}

func (m *Message) SetState() (*SetState, error) {
	return castMessage[SetState](m, MessageType_SetState)
}

func (m *Message) State() (*State, error) {
	return castMessage[State](m, MessageType_State)
}

func (m *Message) StateRoot() (*StateRoot, error) {
	return castMessage[StateRoot](m, MessageType_StateRoot)
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return val
}

func castMessage[T any](m *Message, expectedMessageType MessageType) (*T, error) {
	if m.Type != expectedMessageType {
		return nil, ErrInvalidMessageType
	}

	payload, valid := m.payload.(*T)
	if !valid {
		return nil, ErrInvalidPayloadType
	}

	return payload, nil
}

func checkPayloadType[T any](payload any) error {
	_, valid := payload.(*T)
	if !valid {
		return ErrInvalidPayloadType
	}

	return nil
}
