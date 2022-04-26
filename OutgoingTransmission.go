package main

import (
	"errors"
	"github.com/twmb/murmur3"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type OutgoingTransmission struct {
	seqNr uint32
	uid   uint8

	hash murmur3.Hash128

	transmitter *Transmitter
}

func (t *OutgoingTransmission) sendPacket(header packets.Header, packet packets.Packet) error {
	deadline := time.Now().Add(time.Duration(t.transmitter.timeout) * time.Second)
	err := t.transmitter.conn.SetWriteDeadline(deadline)
	if err != nil {
		return err
	}

	rawData := append(header.ToBytes(), packet.ToBytes()...)
	if len(rawData) > t.transmitter.maxPacketSize {
		return errors.New("packet size exceeding limit")
	}

	_, _, err = t.transmitter.conn.WriteMsgUDP(rawData, nil, nil)
	if err == nil {
		t.seqNr++
	}
	return err
}

func (t *OutgoingTransmission) sendInfo(filesize uint64, filename string) error {
	h := packets.NewHeader(t.seqNr, t.uid, packets.Info)
	p := packets.NewInfoPacket(filesize, filename)

	return t.sendPacket(h, p)
}

func (t *OutgoingTransmission) sendData(data []byte) error {
	h := packets.NewHeader(t.seqNr, t.uid, packets.Data)
	p := packets.NewDataPacket(data)
	err := t.sendPacket(h, p)
	if err != nil {
		return err
	}

	_, err = t.hash.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (t *OutgoingTransmission) sendFinalize(checksum [16]byte) error {
	h := packets.NewHeader(t.seqNr, t.uid, packets.Finalize)
	p := packets.NewFinalizePacket(checksum)

	return t.sendPacket(h, p)
}
