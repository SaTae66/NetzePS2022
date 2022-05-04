package main

import (
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"io"
	"net"
	"os"
	"satae66.dev/netzeps2022/network/packets"
	"time"
)

type Transmitter struct {
	conn *net.UDPConn

	maxPacketSize int
	timeout       int

	transmissions map[uint8]bool
}

func NewTransmitter(maxPacketSize int, timeout int) (Transmitter, error) {
	if maxPacketSize < packets.HeaderSize+1 {
		return Transmitter{}, errors.New("maxPacketSize must be at least HeaderSize+1")
	}

	return Transmitter{
		maxPacketSize: maxPacketSize,
		timeout:       timeout,

		transmissions: make(map[uint8]bool, 0),
	}, nil
}

func (t *Transmitter) newTransmission() (*OutgoingTransmission, error) {
	var uid int
	var inUse bool
	for uid = 0; uid < 256; uid++ {
		inUse = t.transmissions[uint8(uid)]
		if !inUse {
			break
		}
	}
	if inUse {
		return nil, errors.New("no new transmissions available")
	}

	newTransmission := OutgoingTransmission{
		seqNr:       0,
		uid:         uint8(uid),
		hash:        murmur3.New128(),
		transmitter: t,
	}
	t.transmissions[uint8(uid)] = true
	return &newTransmission, nil
}

func (t *Transmitter) endTransmission(uid uint8) {
	t.transmissions[uid] = false
}

func (t *Transmitter) SendFileTo(file *os.File, addr *net.UDPAddr) error {
	fInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fInfo.IsDir() {
		return errors.New("expected file not directory")
	}

	transmission, err := t.newTransmission()
	defer t.endTransmission(transmission.uid)
	if err != nil {
		return err
	}

	transmission.totalSize = uint64(fInfo.Size())

	t.conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	err = transmission.sendInfo(uint64(fInfo.Size()), fInfo.Name())
	if err != nil {
		return err
	}

	fmt.Printf("started sending transmission(%d): %d\n", transmission.uid, time.Now().UnixMilli())

	fin := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-fin:
				return
			default:
				fmt.Printf("\r%f%s\r", float64(transmission.bytesSent)/float64(fInfo.Size())*100, "%")
				time.Sleep(1 * time.Second)
			}
		}
	}()
	defer func() { fin <- true }()

	buf := make([]byte, t.maxPacketSize-packets.HeaderSize)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		err = transmission.sendData(buf[:n])
		if err != nil {
			return err
		}

		transmission.bytesSent += uint64(n)

		if transmission.bytesSent == transmission.totalSize {
			break
		}

		//time.Sleep(10 * time.Microsecond)
	}

	checksum := transmission.hash.Sum(nil)
	err = transmission.sendFinalize(*(*[16]byte)(checksum))
	if err != nil {
		return err
	}

	fmt.Printf("finished sending transmission(%d): %d\n", transmission.uid, time.Now().UnixMilli())

	return nil
}
