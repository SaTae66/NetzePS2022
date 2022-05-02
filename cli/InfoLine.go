package cli

import (
	"fmt"
	"strings"
	"time"
)

type InfoLine struct {
	id       uint8  // uid of the stream
	progress int    // progress of the stream in percent
	speed    string // speed of the stream in bytes/second
	eta      string // duration after which the steam is expected to be finished
}

func NewInfoLine(id uint8, progress int, speed uint32, eta time.Duration) *InfoLine {
	if id < 0 {
		return nil
	}

	return &InfoLine{
		id:       id,
		progress: parseProgress(progress),
		speed:    parseSpeed(speed),
		eta:      parseEta(eta),
	}
}

func parseProgress(progress int) int {
	if progress < 0 || progress > 100 {
		return 0
	}
	return progress
}

func parseSpeed(speed uint32) string {
	var realSpeed int
	var unit string

	if speed >= 1000000000 { // gB
		realSpeed = int(speed) / 1000000000
		unit = "gB"
	} else if speed >= 1000000 { // mB
		realSpeed = int(speed) / 1000000
		unit = "mB"
	} else if speed >= 1000 { // kB
		realSpeed = int(speed) / 1000
		unit = "kB"
	} else { // B
		realSpeed = int(speed) / 1
		unit = " B"
	}
	return fmt.Sprintf("%3d%s/s", realSpeed, unit)
}

func parseEta(eta time.Duration) string {
	return fmt.Sprintf("%s", eta)
}

func (l *InfoLine) UpdateValues(progress int, speed uint32, eta time.Duration) {
	l.progress = parseProgress(progress)
	l.speed = parseSpeed(speed)
	l.eta = parseEta(eta)
}

func (l *InfoLine) print() {
	line := strings.Builder{}

	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %03d ", l.id))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %3d%%  [%-10s] ", l.progress, strings.Repeat("#", l.progress/10)))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", l.speed))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %11s ", l.eta))
	line.WriteString(vertical)

	fmt.Printf("%s\n", line.String())
}
