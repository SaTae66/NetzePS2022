package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"math/rand"
	"net"
	"os"
	"path"
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
		seqNr:    0,
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
		if transmission.seqNr != 0 {
			return errors.New("did not expect info packet")
		}
		transmission.filesize = p.Filesize
		filename := p.Filename

		if filename == "" {
			goto RandomName
		}
		_, err = os.Open(path.Join(r.outpath, filename))
		if err == nil {
			goto FoundName
		}

	RandomName:
		TRIES := 100
		for i := 0; i < TRIES; i++ {
			newFilename := filename + fmt.Sprint(rand.Int())
			_, err := os.Open(path.Join(r.outpath, filename))
			if errors.Is(err, os.ErrNotExist) {
				filename = newFilename
				goto FoundName
			}
		}
		return errors.New("no suitable filename found")

	FoundName:
		file, err := os.Create(path.Join(r.outpath))
		if err != nil {
			return err
		}

		transmission.file = bufio.NewWriter(file)
		break
	case packets.Data:
		p, err := packets.ParseDataPacket(reader)
		if err != nil {
			return err
		}
		if h.SequenceNr < transmission.seqNr {
			//TODO ignore
		}
		packetOffset := h.SequenceNr - transmission.seqNr
		if packetOffset > 10 {
			return errors.New("packets too much out of order")
		}
		if packetOffset != 0 {
			r.packetBuffer[packetOffset] = p
			break
		}
		_, err = transmission.hash.Write(p.Data)
		if err != nil {
			return err
		}

		_, err = transmission.file.Write(p.Data)
		if err != nil {
			return err
		}
		err = transmission.file.Flush()
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

	transmission.seqNr++
	return nil
}

func (r *Receiver) receivePacket() (*bytes.Reader, error) {
	buf := make([]byte, r.maxPacketSize)
	n, _, _, _, err := r.conn.ReadMsgUDP(buf, nil)

	return bytes.NewReader(buf[:n]), err
}
