package cli

import (
	"errors"
	"fmt"
	"golang.org/x/term"
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
	if refreshPerSecond > 1000 {
		return nil, errors.New("ui must NOT be refreshed more than 1000 times per second")
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
	// switch stdin into 'raw' mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	cmdInput := make(chan string, 1)
	line := ""

	for w.run {
		printHeader()
		printHeading()
		printFooter()
		// TODO: update ui
		amount := 3 // amount of lines printed

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

			//TODO: only update lines that need to be
		}

		waitForInput := true

		go func() {
			fmt.Printf(">%s", line)
			b := make([]byte, 1)
			for {
				_, err := os.Stdin.Read(b)
				if err != nil {
					fmt.Printf("CliWorker down %v\n", err)
					waitForInput = false
					break
				}

				if string(b) == "\x08" {
					line = line[0:int(math.Max(0, float64(len(line)-1)))]
				} else {
					newChar := string(b)
					if newChar == "\r" {
						newChar = "\n"
					}
					line += newChar
					if newChar == "\n" {
						break
					}
				}
			}
			cmdInput <- ""
		}()

		select {
		case <-time.After(time.Duration(w.sleepPeriod) * time.Millisecond):
			waitForInput = false
		case _ = <-cmdInput:
			waitForInput = false
			commandLine <- line
			amount += strings.Count(line, "\n")
			line = ""
		}

		//fmt.Printf("\033[2J\r")

		fmt.Printf("\r")
		fmt.Printf("\033[K")

		for i := 0; i < amount; i++ {
			fmt.Printf("\033[%dA", 1)
			fmt.Printf("\r")
			fmt.Printf("\033[K")
		}
		fmt.Printf("\r")
		if waitForInput {
			fmt.Printf("\r")
		}
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
