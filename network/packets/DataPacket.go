package packets

import (
	"bytes"
	"errors"
)

// DataPacketSize represents the minimum payload size of a DataPacket
const DataPacketSize = 1

type DataPacket struct {
	Data []byte
}

func NewDataPacket(data []byte) DataPacket {
	return DataPacket{
		Data: data,
	}
}

func ParseDataPacket(r *bytes.Reader) (DataPacket, error) {
	if r.Len() < (DataPacketSize) {
		return DataPacket{}, errors.New("not enough data")
	}
	buf := make([]byte, r.Len())
	_, err := r.Read(buf)
	return DataPacket{Data: buf}, err
}

func (p DataPacket) ToBytes() []byte {
	var raw []byte
	raw = append(raw, p.Data...)
	return raw
}

func (p DataPacket) Type() PacketType {
	return Data
}
