package packets

type PacketType byte

const (
	Info     PacketType = 0x00
	Data                = 0x01
	Finalize            = 0xFF
)

type DataPacketAndHeader struct {
	Header Header
	Packet DataPacket
}

type FinalizePacketAndHeader struct {
	Header Header
	Packet FinalizePacket
}

type Packet interface {
	ToBytes() []byte
	Type() PacketType
}
