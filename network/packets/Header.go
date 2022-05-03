package packets

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const HeaderSize = 4 + 1 + 1

type Header struct {
	SequenceNr uint32
	StreamUID  uint8
	PacketType PacketType
}

func NewHeader(sequenceNr uint32, streamUID uint8, packetType PacketType) Header {
	return Header{
		SequenceNr: sequenceNr,
		StreamUID:  streamUID,
		PacketType: packetType,
	}
}

func ParseHeader(r *bytes.Reader) (Header, error) {
	data := make([]byte, HeaderSize)
	n, err := r.Read(data)
	if err != nil {
		return Header{}, err
	}

	if n < HeaderSize {
		return Header{}, errors.New("not enough data")
	}

	return Header{
		SequenceNr: binary.LittleEndian.Uint32(data[:4]),
		StreamUID:  data[4],
		PacketType: PacketType(data[5]),
	}, nil
}

func (h *Header) SetHeader(data Header) {
	h.StreamUID = data.StreamUID
	h.SequenceNr = data.SequenceNr
	h.PacketType = data.PacketType
}

func (h *Header) ToBytes() []byte {
	raw := make([]byte, HeaderSize)

	binary.LittleEndian.PutUint32(raw[:4], h.SequenceNr)
	raw[4] = h.StreamUID
	raw[5] = uint8(h.PacketType)

	return raw
}
