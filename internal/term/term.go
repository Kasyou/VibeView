// Package term provides simple ANSI terminal colors.
package term

import "os"

var useColor = isTerminal()

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[36m"
	Pink   = "\033[35m"
)

func color(c string, s string) string {
	if !useColor {
		return s
	}
	return c + s + Reset
}

func BoldText(s string) string  { return color(Bold, s) }
func DimText(s string) string   { return color(Dim, s) }
func RedText(s string) string   { return color(Red, s) }
func GreenText(s string) string { return color(Green, s) }
func CyanText(s string) string  { return color(Blue, s) }
func PinkText(s string) string  { return color(Pink, s) }
