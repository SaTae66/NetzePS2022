package cli

import (
	"errors"
	"fmt"
	"math"
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

func (w *UIDrawer) Start(inputHandler chan string) {
	w.run = true
	/*
		// switch stdin into 'raw' mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	*/

	fmt.Print(GetSeparatorLine())
	fmt.Print(GetHeadingLine())
	fmt.Print(GetSeparatorLine())

	lineCount := 0
	printBuffer := strings.Builder{}

	for w.run {
		printBuffer.Reset()
		printBuffer.WriteString(strings.Repeat("\r\033[1A\033[K", lineCount))
		lineCount = 0

		for i := 0; i < 256; i++ {
			curTransmission := (*w.anchor)[uint8(i)]
			if curTransmission == nil {
				continue
			}

			uid := curTransmission.Uid
			progress := calcProgress(curTransmission.TransmittedSize, curTransmission.TotalSize)
			speed := calcSpeed(curTransmission.TransmittedSize, int(math.Floor(time.Since(curTransmission.StartTime).Seconds())))
			eta := calcEta(curTransmission.TransmittedSize, curTransmission.TotalSize, speed)

			printBuffer.WriteString(GetInfoLine(uid, progress, speed, eta))
			printBuffer.WriteString(GetSeparatorLine())
			lineCount += 2
		}

		fmt.Print(printBuffer.String())

		time.Sleep(time.Duration(w.sleepPeriod) * time.Millisecond)
	}

	fmt.Print("\033[2J")
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
