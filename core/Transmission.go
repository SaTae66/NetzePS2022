package core

import (
	"bufio"
	"github.com/twmb/murmur3"
	"net"
	"time"
)

type Transmission struct {
	SeqNr     uint32           // sequence-number of the next transmitted packet
	NetworkIO net.UDPConn      // udp connection of the transmission
	FileIO    bufio.ReadWriter // buffered filesystem ReadWriter of the transmission

	TransmittedSize uint64 // size of the already transmitted data
	TotalSize       uint64 // total size of the file that is to be transmitted

	Uid       uint8           // unique id of the transmission
	StartTime time.Time       // point of time at which the first data-packet was transmitted
	Hash      murmur3.Hash128 // Hash-object used to calculate the Hash of the file
}
