package packets

import (
	"encoding/binary"
	"errors"
)

type Header struct {
	sequenceNr uint32
	streamUID  uint8
	packetType uint8
}

func NewHeader(sequenceNr uint32, streamUID uint8, packetType uint8) Header {
	return Header{
		sequenceNr: sequenceNr,
		streamUID:  streamUID,
		packetType: packetType,
	}
}

func (h Header) Size() int {
	return 4 + 1 + 1
}

func (h Header) ToBytes() []byte {
	raw := make([]byte, 6)

	binary.LittleEndian.PutUint32(raw[:4], h.sequenceNr)
	raw[4] = h.streamUID
	raw[5] = h.packetType

	return raw
}

func ParseHeader(data []byte) (Header, error) {
	if len(data) < (Header{}.Size()) {
		return Header{}, errors.New("not enough data")
	}

	return Header{
		sequenceNr: binary.LittleEndian.Uint32(data[:4]),
		streamUID:  data[4],
		packetType: data[5],
	}, nil
}
