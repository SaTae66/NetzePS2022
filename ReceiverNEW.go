package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"net"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type ReceiverSettings struct {
	maxPacketSize  int           // maximum size of a packet of this transmission that is allowed
	networkTimeout time.Duration // timeout as time.Duration after which the connection is closed and the transmission is aborted
	bufferLimit    int           // maximum size of the packet buffer in packets
}

type ReceiverNEW struct {
	settings ReceiverSettings

	keepRunning bool

	conn          *net.UDPConn
	transmissions map[uint8]*TransmissionIN
}

func NewReceiverNEW(maxPacketSize int, networkTimeout int, bufferLimit int, addr *net.UDPAddr) (*ReceiverNEW, error) {
	if maxPacketSize < packets.HeaderSize+1 {
		return nil, errors.New("maxPacketSize must be at least the size of the header +1")
	}
	if networkTimeout < 1 {
		return nil, errors.New("timeout must be at least 1 second")
	}
	if bufferLimit < 1 {
		return nil, errors.New("bufferLimit must be at least 1 packet")
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &ReceiverNEW{
		settings: ReceiverSettings{
			maxPacketSize:  maxPacketSize,
			networkTimeout: time.Duration(networkTimeout) * time.Second,
			bufferLimit:    bufferLimit,
		},
		conn:          conn,
		transmissions: make(map[uint8]*TransmissionIN),
	}, nil
}

func (r *ReceiverNEW) Start() error {
	r.keepRunning = true

	for r.keepRunning {
		nextPacket, err := r.nextPacket()
		if err != nil {
			return err
		}

		header, err := packets.ParseHeader(nextPacket)
		if err != nil {
			return err
		}

		transmission := r.getTransmission(header.StreamUID)
		err = transmission.HandlePacket(header, nextPacket)
		if err != nil {
			if err == TransmissionSuccessful {
				// TODO: handle successful transmission
			} else if err == TransmissionFailed {
				// TODO: handle failed transmission
			} else {
				// TODO: handle error
			}
			fmt.Printf("%v\n", err)
			r.removeTransmission(header.StreamUID, nil)
		}
	}

	return nil
}

func (r *ReceiverNEW) Stop() {
	r.keepRunning = false
}

func (r *ReceiverNEW) removeTransmission(uid uint8, v *TransmissionIN) {
	r.transmissions[uid] = v
}

func (r *ReceiverNEW) getTransmission(uid uint8) *TransmissionIN {
	incomingTransmission := r.transmissions[uid]
	if incomingTransmission != nil {
		return incomingTransmission
	}

	newTransmission := TransmissionIN{
		Transmission: Transmission{
			seqNr:           0,
			networkIO:       net.UDPConn{},
			fileIO:          bufio.ReadWriter{},
			transmittedSize: 0,
			totalSize:       0,
			uid:             uid,
			startTime:       time.Time{},
			hash:            murmur3.New128(),
		},
		bufferLimit: r.settings.bufferLimit,
		buffer:      make(map[uint32]*packets.DataPacketAndHeader),
	}
	r.transmissions[uid] = &newTransmission
	return &newTransmission
}

func (r *ReceiverNEW) nextPacket() (*bytes.Reader, error) {
	err := r.conn.SetReadDeadline(time.Now().Add(r.settings.networkTimeout))
	if err != nil {
		return nil, err
	}

	rawBytes := make([]byte, r.settings.maxPacketSize)
	n, _, _, _, err := r.conn.ReadMsgUDP(rawBytes, nil)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rawBytes[:n]), nil
}
