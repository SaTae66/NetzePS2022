package packets

type PacketType byte

const (
	Info     PacketType = 0x00
	Data                = 0x01
	Error               = 0xFD
	Ack                 = 0xFE
	Finalize            = 0xFF
)

type Packet interface {
	ToBytes() []byte
	Type() PacketType
}
