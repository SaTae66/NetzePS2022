package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

var TransmissionSuccessful = errors.New("transmission finished with success")
var TransmissionFailed = errors.New("transmission finished with error")

type TransmissionIN struct {
	Transmission
	initialised bool
	outPath     string

	bufferLimit int
	buffer      map[uint32]*packets.DataPacketAndHeader
	finalize    *packets.FinalizePacketAndHeader
}

func (t *TransmissionIN) HandlePacket(header packets.Header, nextPacket *bytes.Reader) error {
	switch header.PacketType {
	case packets.Info:
		if t.initialised || header.SequenceNr != 0 || t.seqNr != 0 {
			return fmt.Errorf("packet with header %v malformed; expected info packet", header)
		}
		infoPacket, err := packets.ParseInfoPacket(nextPacket)
		if err != nil {
			return err
		}
		err = t.handleInfo(infoPacket)
		if err != nil {
			return err
		}
		t.seqNr++
		break
	case packets.Data:
		p, err := packets.ParseDataPacket(nextPacket)
		if err != nil {
			return err
		}
		if header.SequenceNr != t.seqNr {
			if len(t.buffer) >= t.bufferLimit {
				return errors.New("packet buffer full")
			}
			t.buffer[header.SequenceNr] = &packets.DataPacketAndHeader{
				Header: header,
				Packet: p,
			}
			break
		}
		err = t.handleData(p)
		if err != nil {
			return err
		}
		t.seqNr++
		break
	case packets.Finalize:
		p, err := packets.ParseFinalizePacket(nextPacket)
		if err != nil {
			return err
		}
		if header.SequenceNr != t.seqNr {
			t.finalize = &packets.FinalizePacketAndHeader{
				Header: header,
				Packet: p,
			}
			break
		}
		err = t.handleFinalize(p)
		if err != nil {
			return err
		}
		t.seqNr++
		break
	default:
		return fmt.Errorf("packet with header %v malformed; expected data or finalize packet", header)
	}

	return t.handleBuffer()
}

func (t *TransmissionIN) handleInfo(packet packets.InfoPacket) error {
	t.startTime = time.Now()

	err := t.initFileIO(path.Join(t.outPath, packet.Filename))
	if err != nil {
		return err
	}

	t.totalSize = packet.Filesize
	t.initialised = true
	return nil
}

func (t *TransmissionIN) handleData(packet packets.DataPacket) error {
	_, err := t.fileIO.Write(packet.Data)
	if err != nil {
		return err
	}

	err = t.fileIO.Flush()
	if err != nil {
		return err
	}

	_, err = t.hash.Write(packet.Data)
	if err != nil {
		return err
	}

	t.transmittedSize += uint64(len(packet.Data))
	return nil
}

func (t *TransmissionIN) handleFinalize(packet packets.FinalizePacket) error {
	hash := make([]byte, 0)
	hash = t.hash.Sum(hash)

	expectedHash := packet.Checksum[:]

	diff := bytes.Compare(hash, expectedHash)
	if diff != 0 {
		return TransmissionFailed
	}

	return TransmissionSuccessful
}

// util

func (t *TransmissionIN) handleBuffer() error {

	for p, exists := t.buffer[t.seqNr]; exists; {
		err := t.HandlePacket(p.Header, bytes.NewReader(p.Packet.Data))
		if err != nil {
			return err
		}
		t.buffer[t.seqNr] = nil
		t.seqNr++
	}

	if t.finalize != nil && t.finalize.Header.SequenceNr == t.seqNr {
		err := t.handleFinalize(t.finalize.Packet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TransmissionIN) initFileIO(filePath string) error {
	_, err := os.Open(filePath)
	if os.IsExist(err) {
		return errors.New("file already exists at specified path")
	}
	if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	t.fileIO = bufio.ReadWriter{Reader: bufio.NewReader(file)}
	return nil
}
