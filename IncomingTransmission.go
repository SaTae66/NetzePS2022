package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/twmb/murmur3"
	"math/rand"
	"os"
	"path"
	"satae66.dev/netzeps2022/network/packets"
)

type IncomingTransmission struct {
	curSeqNr uint32
	uid      uint8

	filesize uint64
	file     *bufio.Writer

	hash murmur3.Hash128

	receiver     *Receiver
	packetBuffer []packets.Packet
}

func (t *IncomingTransmission) handleInfo(p packets.InfoPacket) error {
	if t.curSeqNr != 0 {
		return errors.New("did not expect info packet")
	}
	t.filesize = p.Filesize
	filename := p.Filename

	randomize := false
	if filename == "" {
		filename = "transmission_"
		randomize = true
	} else {
		_, err := os.Open(path.Join(t.receiver.outpath, filename))
		if err == nil {
			filename += "_"
			randomize = true
		}
	}

	if randomize {
		TRIES := 100
		for i := 0; i < TRIES-1; i++ {
			newFilename := filename + fmt.Sprint(rand.Intn(1000000))
			_, err := os.Open(path.Join(t.receiver.outpath, filename))
			if errors.Is(err, os.ErrNotExist) {
				filename = newFilename
				break
			}
		}
		newFilename := filename + fmt.Sprint(rand.Intn(1000000))
		_, err := os.Open(path.Join(t.receiver.outpath, filename))
		if errors.Is(err, os.ErrNotExist) {
			filename = newFilename
		} else {
			return errors.New("no suitable filename found")
		}
	}

	file, err := os.Create(path.Join(t.receiver.outpath))
	if err != nil {
		return err
	}

	t.file = bufio.NewWriter(file)

	return nil
}

func (t *IncomingTransmission) handleData(h packets.Header, p packets.DataPacket) error {
	//TODO handle seqNr overflow
	packetOffset := h.SequenceNr - t.curSeqNr
	if packetOffset > 10 {
		return errors.New("packets too much out of order")
	}
	if packetOffset != 0 {
		t.receiver.packetBuffer[packetOffset] = p
		return nil
	}
	_, err := t.hash.Write(p.Data)
	if err != nil {
		return err
	}

	_, err = t.file.Write(p.Data)
	if err != nil {
		return err
	}

	//TODO buffer more data before flush
	err = t.file.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (t *IncomingTransmission) handleFinalize(p packets.FinalizePacket) error {
	return nil
}
