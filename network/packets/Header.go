package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"satae66.dev/netzeps2022/util"
)

const HeaderSize = 4 + 1 + 1

type Header struct {
	SequenceNr uint32
	StreamUID  uint8
	PacketType uint8
}

func NewHeader(sequenceNr uint32, streamUID uint8, packetType uint8) Header {
	return Header{
		SequenceNr: sequenceNr,
		StreamUID:  streamUID,
		PacketType: packetType,
	}
}

func ParseHeader(r *bytes.Reader) (Header, error) {
	if r.Len() < HeaderSize {
		return Header{}, errors.New("not enough data")
	}
	return util.ReadToStruct[Header](r)
}

func (h Header) ToBytes() []byte {
	raw := make([]byte, HeaderSize)

	binary.LittleEndian.PutUint32(raw[:4], h.SequenceNr)
	raw[4] = h.StreamUID
	raw[5] = h.PacketType

	return raw
}
