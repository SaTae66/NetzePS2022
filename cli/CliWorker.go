package cli

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"math"
	"os"
	"satae66.dev/netzeps2022/core"
	"time"
)

type UIDrawer struct {
	run bool

	sleepPeriod int // time between ui refreshes in milliseconds

	anchor *map[uint8]*core.TransmissionIN // anchor back to the transmission map

	currentInfoLines map[uint8]*InfoLine // list of current info lines
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
		sleepPeriod:      1000 / refreshPerSecond,
		anchor:           anchor,
		currentInfoLines: make(map[uint8]*InfoLine),
	}, nil
}

func (w *UIDrawer) Start(inputHandler chan string) {
	w.run = true
	// switch stdin into 'raw' mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	printHeader()
	printHeading()
	printFooter()

	clearHeaderAfterLastTransmissionGone := false
	clipboard := ""
	for w.run {
		// TODO: only update fields that need update

		// clear line
		fmt.Printf("\r\033[K")

		netTransmissionShrinking := 0
		curAnchorLen := len(*w.anchor)
		for i := 0; i < 256; i++ {
			index := uint8(i)
			curTransmission := (*w.anchor)[index]
			if curTransmission == nil {
				if w.currentInfoLines[index] != nil {
					w.currentInfoLines[index] = nil
					netTransmissionShrinking++ // #transmissions shrinks by 1
				}
				continue
			}
			clearHeaderAfterLastTransmissionGone = true
			uid := curTransmission.Uid
			progress := calcProgress(curTransmission.TransmittedSize, curTransmission.TotalSize)
			speed := calcSpeed(curTransmission.TransmittedSize, int(math.Floor(time.Since(curTransmission.StartTime).Seconds())))
			eta := calcEta(curTransmission.TransmittedSize, curTransmission.TotalSize, speed)

			x := w.currentInfoLines[index]
			if x != nil {
				x.UpdateValues(progress, speed, eta)
				x.print()
				fmt.Printf("\033[1B") // move cursor down 1 line
			} else {
				x = NewInfoLine(uid, progress, speed, eta)
				w.currentInfoLines[index] = x
				netTransmissionShrinking-- // #transmissions grows by 1
				x.print()
				printFooter()
			}
		}

		fmt.Printf("\r\033[K") // clear line
		fmt.Printf(">%s", clipboard)

		// handle command input
		timeToUpdateCLI := time.After(time.Duration(w.sleepPeriod) * time.Millisecond)
		gotInput := make(chan bool, 1)
		select {
		case <-timeToUpdateCLI:
		case <-gotInput:
		}
		go func(fin chan bool) {
			b := make([]byte, 1)
			_, err := os.Stdin.Read(b)
			if err != nil {
				fmt.Printf("CliWorker down %v\n", err)
				return
			}

			// windows \r\n
			in := string(b)
			if in == "\r" {
				in = "\n"
				fmt.Printf("\033[1D") // remove \n because input is expected to be \r\n (windows newline)
			}

			if in == "\n" {
				inputHandler <- clipboard
				clipboard = ""
				fin <- true
			} else if in == "\x08" {
				clipboard = clipboard[0:int(math.Max(0, float64(len(clipboard)-1)))]
				if len(clipboard) > 0 {
					fmt.Printf("\033[1D \u001B[1D")
				}
			} else {
				clipboard += in
			}
		}(gotInput)

		fmt.Printf("\r\033[K") // clear line

		//move up to the first InfoLine
		if n := curAnchorLen * 2; n != 0 {
			fmt.Printf("\033[%dA", n)
		} else {
			if clearHeaderAfterLastTransmissionGone {
				fmt.Printf("\033[1B")
				fmt.Printf("\r\033[K") // clear line
				fmt.Printf("\033[1A")
				clearHeaderAfterLastTransmissionGone = false
			}
		}
	}
}

func (w *UIDrawer) Stop() {
	w.run = false
}

func calcProgress(totalSent uint64, totalSize uint64) int {
	return int(math.Ceil(float64(totalSent) / float64(totalSize) * 100))
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
