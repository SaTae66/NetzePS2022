package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"satae66.dev/netzeps2022/util"
)

const InfoPacketSize = 2

type InfoPacket struct {
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
	return util.ReadToStruct[InfoPacket](r)
}

func (p InfoPacket) ToBytes() []byte {
	raw := make([]byte, 8)
	binary.LittleEndian.PutUint64(raw[:], p.Filesize)
	return append(raw, []byte(p.Filename)...)
}

func (p InfoPacket) GetType() PacketType {
	return Info
}
