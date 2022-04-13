package packets

import (
	"errors"
)

type FinalizePacket struct {
	header Header

	checksum [16]byte
}

func NewFinalizePacket(header Header, checksum [16]byte) FinalizePacket {
	return FinalizePacket{
		header: header,

		checksum: checksum,
	}
}

func (p FinalizePacket) Size() int {
	return p.header.Size() + 16
}

func (p FinalizePacket) ToBytes() []byte {
	raw := p.header.ToBytes()

	raw = append(raw, p.checksum[:]...)

	return raw
}

func (p FinalizePacket) GetType() PacketType {
	return Finalize
}

func ParseFinalizePacket(data []byte) (FinalizePacket, error) {
	if len(data) < (FinalizePacket{}.Size()) {
		return FinalizePacket{}, errors.New("not enough data")
	}

	header, err := ParseHeader(data)
	if err != nil {
		return FinalizePacket{}, err
	}
	data = data[Header{}.Size():]

	return FinalizePacket{
		header:   header,
		checksum: *(*[16]byte)(data),
	}, nil
}
