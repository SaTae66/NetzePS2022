package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// InfoPacketSize represents the minimum payload size of a InfoPacket
const InfoPacketSize = 8

type InfoPacket struct {
	Header

	Filesize uint64
	Filename string
}

func NewInfoPacket(filesize uint64, filename string) InfoPacket {
	return InfoPacket{
		Filesize: filesize,
		Filename: filename,
	}
}

func ParseInfoPacket(r *bytes.Reader) (InfoPacket, error) {
	if r.Len() < InfoPacketSize {
		return InfoPacket{}, errors.New("not enough data")
	}
	buf := make([]byte, r.Len())
	_, err := r.Read(buf)
	if err != nil {
		return InfoPacket{}, err
	}

	return InfoPacket{
		Filesize: binary.LittleEndian.Uint64(buf[:8]),
		Filename: string(buf[8:]),
	}, nil
}

func (p InfoPacket) ToBytes() []byte {
	raw := make([]byte, 8)
	binary.LittleEndian.PutUint64(raw[:], p.Filesize)
	return append(raw, []byte(p.Filename)...)
}

func (p InfoPacket) Type() PacketType {
	return Info
}
