package network

import (
	"bufio"
	"time"
)

type TransmissionIN struct {
	Transmission

	File *bufio.Writer

	LastUpdated time.Time
}
