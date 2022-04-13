package packets

import (
	"errors"
)

type DataPacket struct {
	header Header

	data []byte
}

func NewDataPacket(header Header, data []byte) DataPacket {
	return DataPacket{
		header: header,

		data: data,
	}
}

func (p DataPacket) Size() int {
	return p.header.Size() + len(p.data)
}

func (p DataPacket) ToBytes() []byte {
	raw := p.header.ToBytes()

	raw = append(raw, p.data...)

	return raw
}

func (p DataPacket) GetType() PacketType {
	return Data
}

func ParseDataPacket(data []byte) (DataPacket, error) {
	if len(data) < (DataPacket{}.Size()) {
		return DataPacket{}, errors.New("not enough data")
	}

	header, err := ParseHeader(data)
	if err != nil {
		return DataPacket{}, err
	}
	data = data[Header{}.Size():]

	return DataPacket{
		header: header,
		data:   data,
	}, nil
}
