package core

import (
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type TransmissionIN struct {
	Transmission
	IsInitialised bool
	OutPath       string

	LastUpdated time.Time

	BufferLimit int
	Buffer      map[uint32]*packets.DataPacket
	Finalize    *packets.FinalizePacket
}
