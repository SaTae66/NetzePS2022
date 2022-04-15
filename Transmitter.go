package main

import (
	"encoding/binary"
	"errors"
	"github.com/twmb/murmur3"
	"net"
	"os"
	"satae66.dev/netzeps2022/network/packets"
)

type Transmitter struct {
	maxPacketSize int
	transmissions map[uint8]bool

	conn *net.UDPConn
}

func NewTransmitter(maxPacketSize int) (Transmitter, error) {
	if maxPacketSize < packets.HeaderSize+1 {
		return Transmitter{}, errors.New("maxPacketSize must be at least HeaderSize+1")
	}

	return Transmitter{
		maxPacketSize: maxPacketSize,
		transmissions: make(map[uint8]bool, 0),
	}, nil
}

type Transmission struct {
	seqNr uint32
	uid   uint8

	hash murmur3.Hash128

	transmitter *Transmitter
}

func (t *Transmitter) newTransmission() (*Transmission, error) {
	var uid int
	var inUse bool
	for uid = 0; uid < 256; uid++ {
		inUse = t.transmissions[uint8(uid)]
		if !inUse {
			break
		}
	}
	if inUse {
		return nil, errors.New("no new transmissions available")
	}

	newTransmission := Transmission{
		seqNr:       0,
		uid:         uint8(uid),
		hash:        murmur3.New128(),
		transmitter: t,
	}
	t.transmissions[uint8(uid)] = true
	return &newTransmission, nil
}

func (t *Transmitter) endTransmission(uid uint8) {
	t.transmissions[uid] = false
}

func (t *Transmitter) SendFileTo(file *os.File, addr *net.UDPAddr) error {
	fInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fInfo.IsDir() {
		return errors.New("expected file not directory")
	}

	transmission, err := t.newTransmission()
	defer t.endTransmission(transmission.uid)
	if err != nil {
		return err
	}

	t.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	err = transmission.sendInfo(uint64(fInfo.Size()), fInfo.Name())
	if err != nil {
		return err
	}

	err = transmission.sendData([]byte("hello"))
	if err != nil {
		return err
	}

	checksum := make([]byte, 16)
	x1, x2 := transmission.hash.Sum128()
	binary.LittleEndian.PutUint64(checksum[:8], x1)
	binary.LittleEndian.PutUint64(checksum[8:], x2)
	err = transmission.sendFinalize(*(*[16]byte)(checksum))
	if err != nil {
		return err
	}

	return nil
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
