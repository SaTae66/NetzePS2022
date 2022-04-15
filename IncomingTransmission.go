package main

import (
	"bufio"
	"github.com/twmb/murmur3"
	"satae66.dev/netzeps2022/network/packets"
)

type IncomingTransmission struct {
	seqNr uint32
	uid   uint8

	filesize uint64
	file     *bufio.Writer

	hash murmur3.Hash128

	receiver     *Receiver
	packetBuffer []packets.Packet
}

func (i *IncomingTransmission) getInfo() {

}
