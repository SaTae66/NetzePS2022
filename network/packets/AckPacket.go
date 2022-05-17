package packets

import "bytes"

// AckPacketSize represents the minimum payload size of a AckPacket
const AckPacketSize = 1

type AckPacket struct {
	Header
}

func NewAckPacket() AckPacket {
	return AckPacket{}
}

func ParseAckPacket(r *bytes.Reader) (AckPacket, error) {
	return AckPacket{}, nil
}

func (p AckPacket) ToBytes() []byte {
	return []byte{}
}

func (p AckPacket) Type() PacketType {
	return Ack
}
