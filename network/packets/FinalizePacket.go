package packets

import (
	"bytes"
	"errors"
	"satae66.dev/netzeps2022/util"
)

const FinalizePacketSize = 16

type FinalizePacket struct {
	Checksum [16]byte
}

func NewFinalizePacket(checksum [16]byte) FinalizePacket {
	return FinalizePacket{
		Checksum: checksum,
	}
}

func ParseFinalizePacket(r *bytes.Reader) (FinalizePacket, error) {
	if r.Len() < FinalizePacketSize {
		return FinalizePacket{}, errors.New("not enough data")
	}
	return util.ReadToStruct[FinalizePacket](r)
}

func (p FinalizePacket) ToBytes() []byte {
	var raw []byte
	raw = append(raw, p.Checksum[:]...)
	return raw
}

func (p FinalizePacket) Type() PacketType {
	return Finalize
}
