package packets

import "satae66.dev/netzeps2022/network"

type DataPacket struct {
	data []byte

	header network.Header
}

func NewDataPacket(header network.Header, data []byte) DataPacket {
	return DataPacket{
		data:   data,
		header: header,
	}
}

func (p DataPacket) ToBytes() []byte {
	raw := p.header.ToBytes()

	raw = append(raw, p.data...)

	return raw
}

func (p DataPacket) GetType() PacketType {
	return Data
}
