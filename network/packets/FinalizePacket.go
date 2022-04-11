package packets

import "satae66.dev/netzeps2022/network"

type FinalizePacket struct {
	checksum [16]byte

	header network.Header
}

func NewFinalizePacket(header network.Header, checksum [16]byte) FinalizePacket {
	return FinalizePacket{
		checksum: checksum,
		header:   header,
	}
}

func (p FinalizePacket) ToBytes() []byte {
	raw := p.header.ToBytes()

	raw = append(raw, p.checksum[:]...)

	return raw
}

func (p FinalizePacket) GetType() PacketType {
	return Finalize
}
