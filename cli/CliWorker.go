package cli

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"satae66.dev/netzeps2022/core"
	"strings"
	"time"
)

type UIDrawer struct {
	run bool

	sleepPeriod int // time between ui refreshes in milliseconds

	anchor *map[uint8]*core.TransmissionIN // anchor back to the transmission map
}

func NewCliWorker(refreshPerSecond int, anchor *map[uint8]*core.TransmissionIN) (*UIDrawer, error) {
	if refreshPerSecond < 1 {
		return nil, errors.New("ui must be refreshed at least once per second")
	}
	if refreshPerSecond > 100 {
		return nil, errors.New("ui must NOT be refreshed more than 100 times per second")
	}
	if anchor == nil {
		return nil, errors.New("anchor must NOT be nil")
	}

	return &UIDrawer{
		sleepPeriod: 1000 / refreshPerSecond,
		anchor:      anchor,
	}, nil
}

func (w *UIDrawer) Start(commandLine chan string) {
	w.run = true
	reader := bufio.NewReader(os.Stdin)
	cmdInput := make(chan string, 1)

	for w.run {
		// TODO: update ui
		printHeader()
		printHeading()
		printFooter()
		amount := 3
		for i := uint8(0); i < 255; i++ {
			curTransmission := (*w.anchor)[i]
			if curTransmission == nil {
				continue
			}
			//calc stuff
			uid := curTransmission.Uid
			progress := calcProgress(curTransmission.TransmittedSize, curTransmission.TotalSize)
			speed := calcSpeed(curTransmission.TransmittedSize, int(math.Floor(time.Since(curTransmission.StartTime).Seconds())))
			eta := calcEta(curTransmission.TransmittedSize, curTransmission.TotalSize, speed)

			printHeader()
			NewInfoLine(uid, progress, speed, eta).print()
			printFooter()
			amount += 3
		}

		alive := true
		line := ""
		go func() {
			fmt.Printf(">")
			readString, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("CliWorker down %v\n", err)
				alive = false
			}
			cmdInput <- readString
		}()
		for alive {
			select {
			case <-time.After(1 * time.Second):
				alive = false
			case line = <-cmdInput:
				alive = false
				commandLine <- line
				amount += strings.Count(line, "\n") + 1
				//fmt.Printf("\u001B[1A%s\r", strings.Repeat(" ", len(line)))
			}
		}
		// save input
		time.Sleep(time.Duration(w.sleepPeriod) * time.Millisecond)
		fmt.Printf("\r\033[%dA", amount)
	}
}

func (w *UIDrawer) Stop() {
	w.run = false
}

func calcProgress(totalSent uint64, totalSize uint64) int {
	return int(float64(totalSent) / float64(totalSize) * 100)
}

func calcSpeed(totalSent uint64, timeElapsed int) uint32 {
	return uint32(totalSent / uint64(math.Max(1, float64(timeElapsed))))
}

func calcEta(totalSent uint64, totalSize uint64, speed uint32) time.Duration {
	secLeft := (totalSize - totalSent) / uint64(math.Max(1, float64(speed)))
	eta, err := time.ParseDuration(fmt.Sprintf("%ds", secLeft))
	if err != nil {
		return time.Duration(0)
	}
	return eta
}
