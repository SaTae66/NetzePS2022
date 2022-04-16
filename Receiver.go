package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/twmb/murmur3"
	"net"
	"satae66.dev/netzeps2022/network/packets"
)

type Receiver struct {
	conn *net.UDPConn

	maxPacketSize int

	outpath string

	packetBuffer  [10]packets.Packet
	transmissions map[uint8]*IncomingTransmission
}

func NewReceiver(maxPacketSize int, addr *net.UDPAddr) (Receiver, error) {
	if maxPacketSize < packets.HeaderSize+1 {
		return Receiver{}, errors.New("maxPacketSize must be at least HeaderSize+1")
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return Receiver{}, err
	}

	return Receiver{
		conn: conn,

		maxPacketSize: maxPacketSize,

		packetBuffer:  *new([10]packets.Packet),
		transmissions: make(map[uint8]*IncomingTransmission),
	}, nil
}

func (r *Receiver) getTransmission(uid uint8) *IncomingTransmission {
	incomingTransmission := r.transmissions[uid]
	if incomingTransmission != nil {
		return incomingTransmission
	}

	newTransmission := IncomingTransmission{
		curSeqNr: 0,
		hash:     murmur3.New128(),
		receiver: r,
	}
	r.transmissions[uid] = &newTransmission

	return &newTransmission
}

func (r *Receiver) endTransmission(uid uint8) {
	r.transmissions[uid] = nil
}

func (r *Receiver) ListenMessage() error {
	reader, err := r.receivePacket()
	if err != nil {
		return err
	}

	h, err := packets.ParseHeader(reader)
	if err != nil {
		return err
	}

	transmission := r.getTransmission(h.StreamUID)

	switch h.PacketType {
	case packets.Info:
		p, err := packets.ParseInfoPacket(reader)
		if err != nil {
			return err
		}
		err = transmission.handleInfo(p)
		if err != nil {
			return err
		}
		break
	case packets.Data:
		p, err := packets.ParseDataPacket(reader)
		if err != nil {
			return err
		}
		err = transmission.handleData(h, p)
		if err != nil {
			return err
		}
		break
	case packets.Finalize:
		p, err := packets.ParseFinalizePacket(reader)
		if err != nil {
			return err
		}

		hashBuf := make([]byte, 16)
		x1, x2 := transmission.hash.Sum128()
		binary.LittleEndian.PutUint64(hashBuf[:8], x1)
		binary.LittleEndian.PutUint64(hashBuf[8:], x2)
		checksum := *(*[16]byte)(hashBuf)

		if bytes.Compare(checksum[:], p.Checksum[:]) != 0 {
			return errors.New("something went wrong, file hashes not equal")
		}
		return nil
	}

	//TODO handle packet cache

	transmission.curSeqNr++
	return nil
}

func (r *Receiver) receivePacket() (*bytes.Reader, error) {
	buf := make([]byte, r.maxPacketSize)
	n, _, _, _, err := r.conn.ReadMsgUDP(buf, nil)

	return bytes.NewReader(buf[:n]), err
}
