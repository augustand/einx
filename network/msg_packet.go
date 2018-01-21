package network

import (
	//"github.com/Cyinx/einx/slog"
	"encoding/binary"
	"errors"
	"io"
)

type IReader interface {
	Read(b []byte) (n int, err error)
}

const (
	MSG_KEY_LENGTH    = 32
	MSG_TYPE_LENGTH   = 2
	MSG_HEADER_LENGTH = 4
	MSG_ID_LENGTH     = 2
)

// --------------------------------------------------------------------------------------------------------
// |                                header              | body                          |
// | type byte | body_length uint16 | packet_flag uint8 | msg_id uint16| msg_data []byte|
// --------------------------------------------------------------------------------------------------------
var msg_header_length int = MSG_HEADER_LENGTH

type PacketHeader struct {
	MsgType    byte
	BodyLength uint16
	PacketFlag uint8
}

func MsgHeaderLength() int {
	return msg_header_length
}

func ReadBinary(r io.Reader, data interface{}) error {
	return binary.Read(r, binary.LittleEndian, data)
}

func UnmarshalMsgBinary(packet *PacketHeader, b []byte) (uint16, interface{}, error) {
	var msg_id uint16 = 0
	if len(b) < MSG_TYPE_LENGTH {
		return 0, nil, errors.New("msg packet length error")
	}
	msg_id = uint16(b[0]) | (uint16(b[1]) << 8) //小端
	msg_body := b[MSG_TYPE_LENGTH:]
	var msg interface{}
	switch packet.MsgType {
	case 'P':
		msg = MsgProtoUnmarshal(msg_id, msg_body)
	case 'R':
		msg = nil
	default:
		break
	}
	return msg_id, msg, nil
}

func ReadMsgPacket(r io.Reader, msg_packet *PacketHeader, header_buffer []byte, b *[]byte) (uint16, interface{}, error) {
	if _, err := io.ReadFull(r, header_buffer); err != nil {
		return 0, nil, err
	}

	msg_packet.MsgType = header_buffer[0]
	msg_packet.BodyLength = uint16(header_buffer[1]) | (uint16(header_buffer[2]) << 8) //小端
	msg_packet.PacketFlag = header_buffer[3]

	if cap(*b) < int(msg_packet.BodyLength) {
		*b = make([]byte, msg_packet.BodyLength)
	} else {
		*b = (*b)[0:msg_packet.BodyLength]
	}

	if _, err := io.ReadFull(r, *b); err != nil {
		return 0, nil, err
	}
	return UnmarshalMsgBinary(msg_packet, *b)
}

func MarshalMsgBinary(msg_id ProtoTypeID, msg_buffer []byte, b *[]byte) bool {
	var msg_body_length int = len(msg_buffer) + MSG_ID_LENGTH
	var msg_length int = msg_body_length + MSG_HEADER_LENGTH

	if cap(*b) < msg_length {
		*b = make([]byte, msg_length)
	} else {
		*b = (*b)[:msg_length]
	}

	buffer := *b
	//packet header
	buffer[0] = 'P'
	buffer[1] = byte(msg_body_length & 0xFF)
	buffer[2] = byte(msg_body_length >> 8)
	buffer[3] = 0

	//msg wrapper

	buffer[4] = byte(msg_id & 0xFF)
	buffer[5] = byte(msg_id >> 8)

	copy(buffer[MSG_HEADER_LENGTH+MSG_ID_LENGTH:], msg_buffer)

	return true
}