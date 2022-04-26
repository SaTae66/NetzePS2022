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

func (r *Receiver) ReceiverFile() (err error) {
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
			if transmission.finalizeBuffer.FinalizePacket != nil {
				return errors.New("received multiple finalize packets")
			}
			transmission.finalizeBuffer = struct {
				*packets.Header
				*packets.FinalizePacket
			}{Header: &h, FinalizePacket: &p}
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

	if transmission.finalizeBuffer.Header != nil && transmission.curSeqNr == transmission.finalizeBuffer.Header.SequenceNr {
		f := transmission.finalizeBuffer.FinalizePacket
		if f != nil {
			return transmission.handleFinalize(*f)
		}
	}

	return r.ReceiverFile()
}

func (r *Receiver) receivePacket() (*bytes.Reader, error) {
	deadline := time.Now().Add(time.Duration(r.timeout) * time.Second)
	err := r.conn.SetReadDeadline(deadline)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, r.maxPacketSize)
	n, _, _, _, err := r.conn.ReadMsgUDP(buf, nil)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf[:n]), nil
}
