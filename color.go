package main

import "fmt"

type Color int

const (
	Bold Color = 1

	Red   Color = 31
	Green Color = 32
	Cyan  Color = 36
)

func color(text string, color Color) string {
	return fmt.Sprintf("\033[%dm\033[%dm%s\033[0m", Bold, color, text)
}
