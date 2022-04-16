package main

import (
	"bytes"
	"errors"
	"github.com/twmb/murmur3"
	"satae66.dev/netzeps2022/network/packets"
)

type Transmission struct {
	seqNr uint32
	uid   uint8

	hash murmur3.Hash128

	transmitter *Transmitter
	receiver    *Receiver
}

func (t *Transmission) sendPacket(header packets.Header, packet packets.Packet) error {
	rawData := append(header.ToBytes(), packet.ToBytes()...)
	if len(rawData) > t.transmitter.maxPacketSize {
		return errors.New("packet size exceeding limit")
	}

	_, _, err := t.transmitter.conn.WriteMsgUDP(rawData, nil, nil)
	if err == nil {
		t.seqNr++
	}
	return err
}

func (t *Transmission) sendInfo(filesize uint64, filename string) error {
	h := packets.NewHeader(t.seqNr, t.uid, packets.Info)
	p := packets.NewInfoPacket(filesize, filename)

	return t.sendPacket(h, p)
}

func (t *Transmission) sendData(data []byte) error {
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

func (t *Transmission) sendFinalize(checksum [16]byte) error {
	h := packets.NewHeader(t.seqNr, t.uid, packets.Finalize)
	p := packets.NewFinalizePacket(checksum)

	return t.sendPacket(h, p)
}

func (t *Transmission) receivePacket() (packets.Header, packets.Packet, error) {
	buf := make([]byte, t.receiver.maxPacketSize)
	n, _, _, _, err := t.receiver.conn.ReadMsgUDP(buf, nil)

	r := bytes.NewReader(buf[:n])
	h, err := packets.ParseHeader(r)
	if err != nil {
		return packets.Header{}, nil, err
	}

	var p packets.Packet
	switch h.PacketType {
	case packets.Info:
		p, err = packets.ParseInfoPacket(r)
		break
	case packets.Data:
		p, err = packets.ParseDataPacket(r)
		break
	case packets.Finalize:
		p, err = packets.ParseFinalizePacket(r)
		break
	}

	return h, p, err
}