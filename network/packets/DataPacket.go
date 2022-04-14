package packets

import (
	"bytes"
	"errors"
	"satae66.dev/netzeps2022/util"
)

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
	return util.ReadToStruct[DataPacket](r)
}

func (p DataPacket) ToBytes() []byte {
	var raw []byte
	raw = append(raw, p.Data...)
	return raw
}

func (p DataPacket) GetType() PacketType {
	return Data
}
