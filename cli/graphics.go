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
	Id       string
	Progress string
	Speed    string
	Eta      string
}

func NewInfoLine(id string, progress string, speed string, eta string) InfoLine {

	return InfoLine{
		Id:       id,
		Progress: progress,
		Speed:    speed,
		Eta:      eta,
	}
}

func (l *InfoLine) print() {
	line := strings.Builder{}

	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", l.Id))
	line.WriteString(vertical)
	line.WriteString(fmt.Sprintf(" %s ", l.Progress))
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
