package packets

type PacketType byte

const (
	Info     PacketType = 0x00
	Data                = 0x01
	Finalize            = 0xFF
)

type Packet interface {
	ToBytes() []byte
	GetType() PacketType
}
