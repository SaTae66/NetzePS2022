package packets

import (
	"encoding/binary"
	"errors"
	"satae66.dev/netzeps2022/network"
)

type InfoPacket struct {
	filesize uint64
	filename string

	header network.Header
}

func NewInfoPacket(header network.Header, filesize uint64, filename string) InfoPacket {
	return InfoPacket{
		filesize: filesize,
		filename: filename,
		header:   header,
	}
}

func (p InfoPacket) ToBytes() []byte {
	raw := p.header.ToBytes()

	size := [8]byte{}
	binary.LittleEndian.PutUint64(size[:], p.filesize)
	raw = append(raw, size[:]...)

	raw = append(raw, []byte(p.filename)...)

	return raw
}

func (p InfoPacket) GetType() PacketType {
	return Info
}

func ParseInfoPacket(data []byte) (InfoPacket, error) {
	if len(data) < 8 {
		return InfoPacket{}, errors.New("to few data")
	}

	return InfoPacket{
		filesize: binary.LittleEndian.Uint64(data[:8]),
		filename: string(data[8:]),
	}, nil
}
