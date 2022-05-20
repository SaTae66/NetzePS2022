package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"math"
	"net"
	"os"
	"path"
	"satae66.dev/netzeps2022/network"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type Settings struct {
	networkTimeout time.Duration // timeout as time.Duration after which the connection is closed and the transmission is aborted
}

type Receiver struct {
	settings Settings
	outPath  string // path of directory in which to store transmissions

	keepRunning bool

	conn          *net.UDPConn
	transmissions map[uint8]*network.TransmissionIN
}

func NewReceiver(networkTimeout int, outPath string, addr *net.UDPAddr) (*Receiver, error) {
	if networkTimeout < 1 {
		return nil, errors.New("timeout must be at least 1 second")
	}
	if addr == nil {
		return nil, errors.New("addr must not be nil")
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Receiver{
		settings: Settings{
			networkTimeout: time.Duration(networkTimeout) * time.Second,
		},
		outPath:       outPath,
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
		r.closeIdleConnections()
		msg, addr, err := r.nextUDPMessage()
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// timeout happened
			continue
		} else if err != nil {
			status <- err
			continue
		}

		err = r.handlePacket(msg, addr)
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
	rawBytes := make([]byte, math.MaxUint16-8)

	// make timeout to be able to react to a Stop() call and not block until next UDPPacket
	err := r.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return nil, nil, err
	}

	n, _, _, addr, err := r.conn.ReadMsgUDP(rawBytes, nil)
	if err != nil {
		return nil, nil, err
	}

	return bytes.NewReader(rawBytes[:n]), addr, nil
}

func (r *Receiver) handlePacket(udpMessage *bytes.Reader, addr *net.UDPAddr) (err error) {
	header, err := packets.ParseHeader(udpMessage)
	if err != nil {
		return err
	}

	transmission := r.transmissions[header.StreamUID]
	if transmission == nil {
		if header.PacketType != packets.Info {
			return nil //ignore unexpected packets (out of order or timed out connections)
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
	//TODO: move this to TransmissionIN?
	t.StartTime = time.Now()
	err := r.initFileIO(path.Join(r.outPath, path.Clean(p.Filename)), t)
	if err != nil {
		return err
	}

	t.TotalSize = p.Filesize
	t.SeqNr++
	return nil
}

func (r *Receiver) handleData(p packets.DataPacket, t *network.TransmissionIN) error {
	//TODO: move this to TransmissionIN?
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
	//TODO: move this to TransmissionIN?
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
	//TODO: move this to TransmissionIN?
	header.PacketType = packets.Ack
	_, _, err := r.conn.WriteMsgUDP(header.ToBytes(), nil, addr)
	if err != nil {
		return err
	}

	return nil
}

func (r *Receiver) closeIdleConnections() {
	for i := 0; i < 256; i++ {
		uid := uint8(i)
		curTransmission := r.transmissions[uid]
		if curTransmission == nil {
			continue
		}
		if time.Now().After(curTransmission.LastUpdated.Add(r.settings.networkTimeout)) {
			_, _ = fmt.Fprintf(errorLog, "transmission %d timed out\n", i)
			r.closeTransmission(uid)
		}
	}
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

	t.File = bufio.NewWriterSize(file, math.MaxUint16-8)
	return nil
}
