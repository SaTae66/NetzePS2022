package packets

import "bytes"

// ErrorPacketSize represents the minimum payload size of a ErrorPacket
const ErrorPacketSize = 1

type ErrorPacket struct {
	Header

	reason string
}

func NewErrorPacket(reason string) ErrorPacket {
	return ErrorPacket{reason: reason}
}

func ParseErrorPacket(r *bytes.Reader) (ErrorPacket, error) {
	buf := make([]byte, r.Len())
	_, err := r.Read(buf)
	if err != nil {
		return ErrorPacket{}, err
	}

	return ErrorPacket{reason: string(buf)}, nil
}

func (p ErrorPacket) ToBytes() []byte {
	return []byte(p.reason)
}

func (p ErrorPacket) Type() PacketType {
	return Error
}
