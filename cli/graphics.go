package cli

import (
	"fmt"
	"strings"
)

const (
	horizontal = "-"
	vertical   = "|"
)

type InfoLine struct {
	Id       int
	Progress int
	Speed    string
	Eta      string
}

func NewInfoLine(id int, progress int, speed string, eta string) *InfoLine {
	if id < 0 || id > 999 {
		return nil
	}
	if progress < 0 || progress > 100 {
		return nil
	}

	return &InfoLine{
		Id:       id,
		Progress: progress,
		Speed:    speed,
		Eta:      eta,
	}
}

func (l *InfoLine) print() {
	line := strings.Builder{}

	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %03d ", l.Id))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %3d%%  [%-10s] ", l.Progress, strings.Repeat("#", l.Progress/10)))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", l.Speed))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", l.Eta))
	line.WriteString(vertical)

	fmt.Printf("%s\n", line.String())
}

func Draw(lines []*InfoLine) {
	printHeader()
	for _, line := range lines {
		line.print()
		printFooter()
	}
}

func printHeader() {
	printDefaultLine()
	printInfoLine()
	printDefaultLine()
}

func printFooter() {
	printDefaultLine()
}

func printDefaultLine() {
	line := strings.Builder{}

	line.WriteString(vertical)
	line.WriteString(strings.Repeat(horizontal, 5))
	line.WriteString(vertical)
	line.WriteString(strings.Repeat(horizontal, 20))
	line.WriteString(vertical)
	line.WriteString(strings.Repeat(horizontal, 9))
	line.WriteString(vertical)
	line.WriteString(strings.Repeat(horizontal, 13))
	line.WriteString(vertical)

	fmt.Printf("%s\n", line.String())
}

func printInfoLine() {
	line := strings.Builder{}

	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", "UID"))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf("      %s      ", "PROGRESS"))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf("  %s  ", "SPEED"))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf("     %s     ", "EST"))
	line.WriteString(vertical)

	fmt.Printf("%s\n", line.String())
}
