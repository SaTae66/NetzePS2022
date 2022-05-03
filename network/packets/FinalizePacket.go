package packets

import (
	"bytes"
	"errors"
	"fmt"
)

// FinalizePacketSize represents the payload size of a FinalizePacket
const FinalizePacketSize = 16

type FinalizePacket struct {
	Header

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

	checksum := [16]byte{}
	n, err := r.Read(checksum[:])
	if err != nil {
		return FinalizePacket{}, err
	}
	if n != 16 {
		return FinalizePacket{}, fmt.Errorf("expected 16 bytes checksum; got %d bytes", n)
	}

	return FinalizePacket{
		Checksum: checksum,
	}, nil
}

func (p FinalizePacket) ToBytes() []byte {
	var raw []byte
	raw = append(raw, p.Checksum[:]...)
	return raw
}

func (p FinalizePacket) Type() PacketType {
	return Finalize
}
