package cli

import (
	"fmt"
	"strings"
)

const (
	horizontal = "-"
	vertical   = "|"
)

func printHeader() {
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

func printHeading() {
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
