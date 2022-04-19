package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"net"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type Receiver struct {
	conn *net.UDPConn

	maxPacketSize int
	timeout       int

	outpath string

	transmissions map[uint8]*IncomingTransmission
}

func NewReceiver(maxPacketSize int, filedir string, timeout int, addr *net.UDPAddr) (Receiver, error) {
	if maxPacketSize < packets.HeaderSize+1 {
		return Receiver{}, errors.New("maxPacketSize must be at least HeaderSize+1")
	}
	if timeout < 1 {
		return Receiver{}, errors.New("timeout must be at least 1 second")
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return Receiver{}, err
	}

	return Receiver{
		conn: conn,

		maxPacketSize: maxPacketSize,

		outpath: filedir,
		timeout: timeout,

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

func (r *Receiver) TransferFile() (err error) {
	defer func(err error) {
		if err == nil {
			return
		}
		for _, t := range r.transmissions {
			fmt.Printf("transmission dump: %v", t.packetBuffer)
			r.endTransmission(t.uid)
		}
	}(err)

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
		if transmission.filesize != 0 {
			// ignore info packets if transmission already started by info packet
			break
		}
		p, err := packets.ParseInfoPacket(reader)
		if err != nil {
			return err
		}
		err = transmission.handleInfo(p)
		if err != nil {
			return err
		}
		transmission.curSeqNr++
		break
	case packets.Data:
		p, err := packets.ParseDataPacket(reader)
		if err != nil {
			return err
		}
		if h.SequenceNr != transmission.curSeqNr {
			if len(transmission.packetBuffer) >= 10 {
				return errors.New("packetBuffer overflow")
			}
			transmission.packetBuffer[h.SequenceNr] = &p
			break
		}
		err = transmission.handleData(p)
		if err != nil {
			return err
		}
		transmission.curSeqNr++
		break
	case packets.Finalize:
		p, err := packets.ParseFinalizePacket(reader)
		if err != nil {
			return err
		}
		if h.SequenceNr != transmission.curSeqNr {
			if len(transmission.finalizeBuffer) >= 1 {
				return errors.New("finalizeBuffer overflow")
			}
			transmission.finalizeBuffer[h.SequenceNr] = &p
			break
		}
		err = transmission.handleFinalize(p)
		if err != nil {
			return err
		}
		transmission.curSeqNr++
		return nil
	}

	// handle packet cache
	for true {
		p := transmission.packetBuffer[transmission.curSeqNr]
		if p == nil {
			break
		}
		err = transmission.handleData(*p)
		if err != nil {
			return err
		}
		transmission.packetBuffer[transmission.curSeqNr] = nil
		transmission.curSeqNr++
	}

	f := transmission.finalizeBuffer[transmission.curSeqNr]
	// what if curSeqNr = SeqNr of Packet before Overflow and Finalize is due to in 1 SeqNr cycle
	if f != nil {
		return transmission.handleFinalize(*f)
	}

	return r.TransferFile()
}

func (r *Receiver) receivePacket() (*bytes.Reader, error) {
	buf := make([]byte, r.maxPacketSize)
	deadline := time.Now().Add(time.Duration(r.timeout) * time.Second)
	for time.Now().Before(deadline) {
		fmt.Printf("Listening...\n")

		n, _, _, _, err := r.conn.ReadMsgUDP(buf, nil)
		return bytes.NewReader(buf[:n]), err
	}
	return nil, errors.New("connection timeout")
}
