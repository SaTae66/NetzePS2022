package main

import (
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type TransmissionIN struct {
	Transmission
	initialised bool
	outPath     string

	timeout <-chan time.Time

	bufferLimit int
	buffer      map[uint32]*packets.DataPacket
	finalize    *packets.FinalizePacket
}
