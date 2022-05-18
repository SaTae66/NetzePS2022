package core

import (
	"fmt"
	"time"
)

type TransmissionCleaner struct {
	run bool

	sleepPeriod int // time between ui refreshes in milliseconds
	timeout     int // time after which Transmission is closed

	anchor *map[uint8]*TransmissionIN // anchor back to the transmission map
}

func NewTransmissionCleaner(refreshPerSecond int, timeout int, anchor *map[uint8]*TransmissionIN) (*TransmissionCleaner, error) {
	if anchor == nil {
		return nil, fmt.Errorf("")
	}

	return &TransmissionCleaner{
		sleepPeriod: 1000 / refreshPerSecond,
		timeout:     timeout,
		anchor:      anchor,
	}, nil
}

func (t *TransmissionCleaner) Start() {
	t.run = true

	for t.run {
		for i := uint8(0); i < 255; i++ {
			curTransmission := (*t.anchor)[i]
			if curTransmission == nil {
				continue
			}
			if time.Now().After(curTransmission.LastUpdated.Add(time.Duration(t.timeout) * time.Second)) {
				//fmt.Printf("transmission %d timed out\n", i)
				(*t.anchor)[i] = nil
			}
		}

		time.Sleep(time.Duration(t.sleepPeriod) * time.Millisecond)
	}
}

func (t *TransmissionCleaner) Stop() {
	t.run = false
}
