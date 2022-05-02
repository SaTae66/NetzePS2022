package main

import (
	"bufio"
	"github.com/twmb/murmur3"
	"net"
	"time"
)

type Transmission struct {
	seqNr     uint32           // sequence-number of the next transmitted packet
	networkIO net.UDPConn      // udp connection of the transmission
	fileIO    bufio.ReadWriter // buffered filesystem ReadWriter of the transmission

	transmittedSize uint64 // size of the already transmitted data
	totalSize       uint64 // total size of the file that is to be transmitted

	uid       uint8           // unique id of the transmission
	startTime time.Time       // point of time at which the first data-packet was transmitted
	hash      murmur3.Hash128 // hash-object used to calculate the hash of the file
}
