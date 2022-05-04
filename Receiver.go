package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"net"
	"os"
	"path"
	"satae66.dev/netzeps2022/core"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type ReceiverSettings struct {
	maxPacketSize  int           // maximum size of a packet of this transmission that is allowed
	networkTimeout time.Duration // timeout as time.Duration after which the connection is closed and the transmission is aborted
	bufferLimit    int           // maximum size of the packet buffer in packets
	outPath        string        // path of directory in which to store transmissions
}

type Receiver struct {
	settings ReceiverSettings

	keepRunning bool

	conn          *net.UDPConn
	transmissions map[uint8]*core.TransmissionIN
}

func NewReceiver(maxPacketSize int, networkTimeout int, bufferLimit int, outPath string, addr *net.UDPAddr) (*Receiver, error) {
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

	return &Receiver{
		settings: ReceiverSettings{
			maxPacketSize:  maxPacketSize,
			networkTimeout: time.Duration(networkTimeout) * time.Second,
			bufferLimit:    bufferLimit,
			outPath:        outPath,
		},
		conn:          conn,
		transmissions: make(map[uint8]*core.TransmissionIN),
	}, nil
}

func (r *Receiver) Start(status chan error) {
	r.keepRunning = true

	go func() {
		for r.keepRunning {
			nextPacket, err := r.nextUDPMessage()
			if err != nil {
				status <- err
				continue
			}

			header, err := packets.ParseHeader(nextPacket)
			if err != nil {
				status <- err
				continue
			}

			err = r.handlePacket(header, nextPacket)
			if err != nil {
				status <- err
				continue
			}
		}
	}()
}

func (r *Receiver) Stop() {
	r.keepRunning = false
}

func (r *Receiver) openNewTransmission(uid uint8) *core.TransmissionIN {
	newTransmission := core.TransmissionIN{
		Transmission: core.Transmission{
			SeqNr:           0,
			NetworkIO:       net.UDPConn{},
			FileIO:          bufio.ReadWriter{},
			TransmittedSize: 0,
			TotalSize:       0,
			Uid:             uid,
			StartTime:       time.Time{},
			Hash:            murmur3.New128(),
		},
		OutPath:     r.settings.outPath,
		Timeout:     nil,
		BufferLimit: r.settings.bufferLimit,
		Buffer:      make(map[uint32]*packets.DataPacket),
	}
	r.transmissions[uid] = &newTransmission
	return &newTransmission
}

func (r *Receiver) closeTransmission(uid uint8) {
	delete(r.transmissions, uid)
}

func (r *Receiver) nextUDPMessage() (*bytes.Reader, error) {
	rawBytes := make([]byte, r.settings.maxPacketSize)
	n, _, _, _, err := r.conn.ReadMsgUDP(rawBytes, nil)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(rawBytes[:n]), nil
}

func (r *Receiver) handlePacket(header packets.Header, udpMessage *bytes.Reader) (err error) {
	transmission := r.transmissions[header.StreamUID]
	if transmission == nil {
		if header.PacketType != packets.Info {
			return fmt.Errorf("unexpected packet with header %+v", header)
		}
		transmission = r.openNewTransmission(header.StreamUID)
	}

	defer func() {
		if err != nil {
			r.closeTransmission(transmission.Uid)
		}
	}()

	switch header.PacketType {
	case packets.Info:
		infoPacket, err := packets.ParseInfoPacket(udpMessage)
		if err != nil {
			return err
		}
		infoPacket.SetHeader(header)
		err = r.handleInfo(infoPacket, transmission)
		if err != nil {
			return err
		}
		break
	case packets.Data:
		if header.SequenceNr == binary.LittleEndian.Uint32([]byte{0xCE, 0x37, 0x00, 0x00}) {
			fmt.Print()
		}
		dataPacket, err := packets.ParseDataPacket(udpMessage)
		if err != nil {
			return err
		}
		dataPacket.SetHeader(header)
		err = r.handleData(dataPacket, transmission)
		if err != nil {
			return err
		}
		err = r.handleBuffer(transmission)
		if err != nil {
			return err
		}
		break
	case packets.Finalize:
		finalizePacket, err := packets.ParseFinalizePacket(udpMessage)
		if err != nil {
			return err
		}
		finalizePacket.SetHeader(header)
		err = r.handleFinalize(finalizePacket, transmission)
		if err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("malformed packet with header %v", header)
	}

	return nil
}

func (r *Receiver) handleInfo(p packets.InfoPacket, t *core.TransmissionIN) error {
	if t.IsInitialised || p.SequenceNr != 0 {
		return fmt.Errorf("unexpected packet with header %+v", p.Header)
	}
	// PRINTING
	fmt.Printf("started receiving transmission(%d): %d\n", t.Uid, time.Now().UnixMilli())

	t.StartTime = time.Now()
	t.Timeout = time.After(r.settings.networkTimeout * time.Second)

	err := r.initFileIO(path.Join(t.OutPath, p.Filename), t)
	if err != nil {
		return err
	}

	t.TotalSize = p.Filesize
	t.IsInitialised = true
	t.SeqNr++
	return nil
}

func (r *Receiver) handleData(p packets.DataPacket, t *core.TransmissionIN) error {
	if p.SequenceNr != t.SeqNr {
		if len(t.Buffer) >= t.BufferLimit {
			return errors.New("packet buffer full")
		}
		t.Buffer[p.SequenceNr] = &p
		return nil
	}

	_, err := t.FileIO.Write(p.Data)
	if err != nil {
		return err
	}

	err = t.FileIO.Flush()
	if err != nil {
		return err
	}

	_, err = t.Hash.Write(p.Data)
	if err != nil {
		return err
	}

	t.TransmittedSize += uint64(len(p.Data))
	t.SeqNr++
	return nil
}

func (r *Receiver) handleFinalize(p packets.FinalizePacket, t *core.TransmissionIN) error {
	if p.SequenceNr != t.SeqNr {
		t.Finalize = &p
		return nil
	}

	actualHash := make([]byte, 0)
	actualHash = t.Hash.Sum(actualHash)

	expectedHash := p.Checksum[:]

	diff := bytes.Compare(actualHash, expectedHash)
	if diff != 0 {
		return fmt.Errorf("integrity check failed; expected:<%x> actual:<%x>", expectedHash, actualHash)
	}

	_ = t.FileIO.Flush()
	r.closeTransmission(t.Uid)

	// PRINTING
	fmt.Printf("finished receiving transmission(%d): %d\n", t.Uid, time.Now().UnixMilli())
	return nil
}

func (r *Receiver) handleBuffer(t *core.TransmissionIN) error {
	for p, exists := t.Buffer[t.SeqNr]; exists; p, exists = t.Buffer[t.SeqNr] {
		delete(t.Buffer, t.SeqNr)
		err := r.handlePacket(p.Header, bytes.NewReader(p.Data))
		if err != nil {
			return err
		}
	}

	if t.Finalize != nil && t.Finalize.Header.SequenceNr == t.SeqNr {
		err := r.handleFinalize(*t.Finalize, t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Receiver) initFileIO(filePath string, t *core.TransmissionIN) error {
	_, err := os.Open(filePath)
	if os.IsExist(err) {
		return errors.New("file already exists at specified path")
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	t.FileIO = bufio.ReadWriter{Writer: bufio.NewWriter(file)}
	return nil
}
