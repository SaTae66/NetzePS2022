package packets

import (
	"encoding/binary"
	"errors"
)

type InfoPacket struct {
	header Header

	filesize uint64
	filename string
}

func NewInfoPacket(header Header, filesize uint64, filename string) InfoPacket {
	return InfoPacket{
		filesize: filesize,
		filename: filename,
		header:   header,
	}
}

func (p InfoPacket) Size() int {
	return p.header.Size() + 8 + len(p.filename)
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
	if len(data) < (InfoPacket{}.Size()) {
		return InfoPacket{}, errors.New("not enough data")
	}

	header, err := ParseHeader(data)
	if err != nil {
		return InfoPacket{}, err
	}
	data = data[Header{}.Size():]

	return InfoPacket{
		header:   header,
		filesize: binary.LittleEndian.Uint64(data[:8]),
		filename: string(data[8:]),
	}, nil
}
