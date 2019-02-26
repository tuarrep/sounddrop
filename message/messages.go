package message

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
)

// Receiver interface of a internal message receiver
type Receiver interface {
	GetChan() chan proto.Message
}

// ServiceNumber service identifier on mesh network. Insure protocol compatibility of all devices on mesh
const ServiceNumber uint32 = 0xECC377BC

// Messages opCodes
const (
	AnnounceMessage         = 0x00
	DeviceAllowedMessage    = 0x10
	DeviceDisallowedMessage = 0x11
	StreamDataMessage       = 0x20
	PeerOnlineMessage       = 0xF0
	PeerOfflineMessage      = 0xF1
	WriteRequestMessage     = 0xF2
)

// FromBuffer get message instance from raw bytes buffer
func FromBuffer(buffer []byte) (proto.Message, error) {
	if len(buffer) < 1 {
		return nil, nil
	}

	var message proto.Message

	opCode := buffer[0]
	switch opCode {
	case AnnounceMessage:
		message = &Announce{}
	case DeviceAllowedMessage:
		message = &DeviceAllowed{}
	case DeviceDisallowedMessage:
		message = &DeviceDisallowed{}
	case StreamDataMessage:
		message = &StreamData{}
	default:
		return nil, fmt.Errorf("invalid OP code %d", opCode)
	}

	err := proto.Unmarshal(buffer[1:], message)
	return message, err
}

// ToBuffer get bytes buffer from message instance
func ToBuffer(message proto.Message) ([]byte, error) {
	opcode, err := FindOpCode(message)
	if err != nil {
		return nil, err
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	return append([]byte{opcode}, data...), nil
}

// FindOpCode message opCode from message type
func FindOpCode(message proto.Message) (byte, error) {
	var opcode byte

	switch message.(type) {
	case *Announce:
		opcode = AnnounceMessage
	case *DeviceAllowed:
		opcode = DeviceAllowedMessage
	case *DeviceDisallowed:
		opcode = DeviceDisallowedMessage
	case *StreamData:
		opcode = StreamDataMessage
	case *PeerOnline:
		opcode = PeerOnlineMessage
	case *PeerOffline:
		opcode = PeerOfflineMessage
	case *WriteRequest:
		opcode = WriteRequestMessage
	default:
		return 0x00, fmt.Errorf("invalid message type %s", reflect.TypeOf(message).String())
	}

	return opcode, nil
}
