package cli

import (
	"fmt"
	"strings"
	"time"
)

const infoLineFormat = "%s %03d %s %3d%%  [%-10s] %s %s %s %11s %s\n"

func GetInfoLine(id uint8, progress int, speed uint32, eta time.Duration) string {
	parsedProgress := parseProgress(progress)
	return fmt.Sprintf(infoLineFormat, vertical, id, vertical, parsedProgress, strings.Repeat("#", parsedProgress/10), vertical, parseSpeed(speed), vertical, parseEta(eta), vertical)
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
