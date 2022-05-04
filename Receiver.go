package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"net"
	"os"
	"path"
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
	transmissions map[uint8]*TransmissionIN
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
		transmissions: make(map[uint8]*TransmissionIN),
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

func (r *Receiver) getTransmission(uid uint8) *TransmissionIN {
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
		outPath:     r.settings.outPath,
		timeout:     nil,
		bufferLimit: r.settings.bufferLimit,
		buffer:      make(map[uint32]*packets.DataPacket),
	}
	r.transmissions[uid] = &newTransmission
	return &newTransmission
}

func (r *Receiver) removeTransmission(uid uint8) {
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
	transmission := r.getTransmission(header.StreamUID)

	defer func() {
		if err != nil {
			r.removeTransmission(transmission.uid)
		}
		transmission.seqNr++
	}()

	switch header.PacketType {
	case packets.Info:
		if transmission.initialised || header.SequenceNr != 0 || transmission.seqNr != 0 {
			return fmt.Errorf("packet with header %v malformed", header)
		}
		infoPacket, err := packets.ParseInfoPacket(udpMessage)
		if err != nil {
			return err
		}
		err = r.handleInfo(infoPacket, transmission)
		if err != nil {
			return err
		}
		break
	case packets.Data:
		dataPacket, err := packets.ParseDataPacket(udpMessage)
		if err != nil {
			return err
		}
		if header.SequenceNr != transmission.seqNr {
			if len(transmission.buffer) >= transmission.bufferLimit {
				return errors.New("packet buffer full")
			}
			dataPacket.SetHeader(header)
			transmission.buffer[header.SequenceNr] = &dataPacket
			break
		}
		err = r.handleData(dataPacket, transmission)
		if err != nil {
			return err
		}
		break
	case packets.Finalize:
		finalizePacket, err := packets.ParseFinalizePacket(udpMessage)
		if err != nil {
			return err
		}
		if header.SequenceNr != transmission.seqNr {
			finalizePacket.SetHeader(header)
			transmission.finalize = &finalizePacket
			break
		}
		err = r.handleFinalize(finalizePacket, transmission)
		if err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("packet with header %v malformed; expected data or finalize packet", header)
	}

	return nil
}

func (r *Receiver) handleInfo(p packets.InfoPacket, t *TransmissionIN) error {
	fmt.Printf("started transmission: %d\n", time.Now().UnixMilli())

	t.startTime = time.Now()
	t.timeout = time.After(r.settings.networkTimeout * time.Second)

	err := r.initFileIO(path.Join(t.outPath, p.Filename), t)
	if err != nil {
		return err
	}

	t.totalSize = p.Filesize
	t.initialised = true
	return nil
}

func (r *Receiver) handleData(p packets.DataPacket, t *TransmissionIN) error {
	_, err := t.fileIO.Write(p.Data)
	if err != nil {
		return err
	}

	err = t.fileIO.Flush()
	if err != nil {
		return err
	}

	_, err = t.hash.Write(p.Data)
	if err != nil {
		return err
	}

	t.transmittedSize += uint64(len(p.Data))
	return nil
}

func (r *Receiver) handleFinalize(p packets.FinalizePacket, t *TransmissionIN) error {
	actualHash := make([]byte, 0)
	actualHash = t.hash.Sum(actualHash)

	expectedHash := p.Checksum[:]

	diff := bytes.Compare(actualHash, expectedHash)
	if diff != 0 {
		return fmt.Errorf("integrity check failed; expected:<%x> actual:<%x>", expectedHash, actualHash)
	}

	_ = t.fileIO.Flush()
	r.removeTransmission(t.uid)
	fmt.Printf("finished transmission: %d\n", time.Now().UnixMilli())
	return nil
}

func (r *Receiver) handleBuffer(t *TransmissionIN) error {
	for p, exists := t.buffer[t.seqNr]; exists; {
		err := r.handlePacket(p.Header, bytes.NewReader(p.Data))
		if err != nil {
			return err
		}
		t.buffer[t.seqNr] = nil
		t.seqNr++
	}

	if t.finalize != nil && t.finalize.Header.SequenceNr == t.seqNr {
		err := r.handleFinalize(*t.finalize, t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Receiver) initFileIO(filePath string, t *TransmissionIN) error {
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

	t.fileIO = bufio.ReadWriter{Writer: bufio.NewWriter(file)}
	return nil
}
