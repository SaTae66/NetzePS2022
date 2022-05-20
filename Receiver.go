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
	"satae66.dev/netzeps2022/network"
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
	transmissions map[uint8]*network.TransmissionIN
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
	if addr == nil {
		return nil, errors.New("addr must not be nil")
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
		transmissions: make(map[uint8]*network.TransmissionIN),
	}, nil
}

func (r *Receiver) Start(status chan error) {
	r.keepRunning = true

	go r.run(status)
}

func (r *Receiver) run(status chan error) {
	for r.keepRunning {
		nextPacket, addr, err := r.nextUDPMessage()
		if err != nil {
			status <- err
			continue
		}

		header, err := packets.ParseHeader(nextPacket)
		if err != nil {
			status <- err
			continue
		}

		err = r.handlePacket(header, nextPacket, addr)
		if err != nil {
			status <- err
			continue
		}
	}
}

func (r *Receiver) Stop() {
	r.keepRunning = false
}

func (r *Receiver) openNewTransmission(uid uint8) *network.TransmissionIN {
	newTransmission := network.TransmissionIN{
		Transmission: network.Transmission{
			Uid:  uid,
			Hash: murmur3.New128(),
		},
	}
	r.transmissions[uid] = &newTransmission
	return &newTransmission
}

func (r *Receiver) closeTransmission(uid uint8) {
	delete(r.transmissions, uid)
}

func (r *Receiver) nextUDPMessage() (*bytes.Reader, *net.UDPAddr, error) {
	rawBytes := make([]byte, r.settings.maxPacketSize)
	n, _, _, addr, err := r.conn.ReadMsgUDP(rawBytes, nil)
	if err != nil {
		return nil, nil, err
	}

	return bytes.NewReader(rawBytes[:n]), addr, nil
}

func (r *Receiver) handlePacket(header packets.Header, udpMessage *bytes.Reader, addr *net.UDPAddr) (err error) {
	transmission := r.transmissions[header.StreamUID]
	if transmission == nil {
		if header.PacketType != packets.Info {
			return nil //ignore unexpected packets (out of order or timed out connections
		}
		transmission = r.openNewTransmission(header.StreamUID)
	}

	defer func() {
		transmission.LastUpdated = time.Now()
		if err != nil {
			//TODO: send error packet
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
		dataPacket, err := packets.ParseDataPacket(udpMessage)
		if err != nil {
			return err
		}
		dataPacket.SetHeader(header)
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
		finalizePacket.SetHeader(header)
		err = r.handleFinalize(finalizePacket, transmission)
		if err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("malformed packet with header %v", header)
	}

	err = r.sendAck(header, addr)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) handleInfo(p packets.InfoPacket, t *network.TransmissionIN) error {
	t.StartTime = time.Now()
	err := r.initFileIO(path.Join(r.settings.outPath, path.Clean(p.Filename)), t)
	if err != nil {
		return err
	}

	t.TotalSize = p.Filesize
	t.SeqNr++
	return nil
}

func (r *Receiver) handleData(p packets.DataPacket, t *network.TransmissionIN) error {
	_, err := t.File.Write(p.Data)
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

func (r *Receiver) handleFinalize(p packets.FinalizePacket, t *network.TransmissionIN) error {
	_ = t.File.Flush()

	actualHash := make([]byte, 0)
	actualHash = t.Hash.Sum(actualHash)

	expectedHash := p.Checksum[:]

	diff := bytes.Compare(actualHash, expectedHash)
	if diff != 0 {
		return fmt.Errorf("integrity check failed; expected:<%x> actual:<%x>", expectedHash, actualHash)
	}

	r.closeTransmission(t.Uid)

	// PRINTING
	_, _ = fmt.Fprintf(measureLog, "%d\n", time.Since(t.StartTime).Milliseconds())
	_ = measureLog.Flush()
	return nil
}

func (r *Receiver) sendAck(header packets.Header, addr *net.UDPAddr) error {
	header.PacketType = packets.Ack
	_, _, err := r.conn.WriteMsgUDP(header.ToBytes(), nil, addr)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) initFileIO(filePath string, t *network.TransmissionIN) error {
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

	t.File = bufio.NewWriterSize(file, r.settings.maxPacketSize)
	return nil
}
