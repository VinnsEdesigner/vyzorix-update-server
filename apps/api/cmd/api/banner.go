// Package main provides the Vyzorix API server entry point.
package main

import (
	"fmt"
	"os"
)

// ANSI color codes for terminal output.
const (
	cyan    = "\033[36m"
	magenta = "\033[35m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	green   = "\033[32m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	reset   = "\033[0m"
)

const bannerWidth = 61

// Version is the build stamp, overridden via -ldflags "-X main.Version".
var Version = "dev"

// asciiBanner is the VYZORIX ASCII art banner.
var asciiBanner = []string{
	"+" + repeatChar("-", bannerWidth) + "+",
	"|" + padCenter(" _   _           _        ____") + "|",
	"|" + padCenter("|_| |_|   ___   | |__    |  _|  ___  ___") + "|",
	"|" + padCenter("| | | |  / _ \\  | '_ \\  | |_  / _ \\/ __|") + "|",
	"|" + padCenter("| |_| | | (_) | | |_) | |  _|  __/\\__ \\") + "|",
	"|" + padCenter("|___|_|  \\___/  |_.__/   |_|   \\___||___/") + "|",
	"|" + padCenter("") + "|",
	"|" + padCenter("GOLANG SERVER v"+Version) + "|",
	"+" + repeatChar("-", bannerWidth) + "+",
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}

func padCenter(s string) string {
	width := bannerWidth
	if len(s) >= width {
		return s
	}
	padding := width - len(s)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return repeatChar(" ", leftPad) + s + repeatChar(" ", rightPad)
}

// PrintBanner displays the VYZORIX ASCII art banner.
func PrintBanner(mode string) {
	for _, line := range asciiBanner {
		fmt.Println(magenta + bold + line + reset)
	}

	modeColor := yellow
	if mode == "production" {
		modeColor = red
	}
	fmt.Printf("  %sMode:%s %s%s[%s]%s\n", dim, reset, modeColor, bold, mode, reset)
	fmt.Printf("  %s%s\n", dim, repeatChar("=", 59))
}

// PrintSection prints a section header.
func PrintSection(label string) {
	fmt.Printf("\n%s[%s]%s\n", cyan, label, reset)
}

// PrintStatus prints a status line with color coding.
func PrintStatus(label, value string) {
	fmt.Printf("  %s%s %s: %s%s%s\n", green, reset, label, green, value, reset)
}

// PrintWarning prints a warning message.
func PrintWarning(label, value string) {
	fmt.Printf("  %s%s %s: %s%s%s\n", yellow, reset, label, yellow, value, reset)
}

func getEnv() string {
	if os.Getenv("NODE_ENV") == "production" || os.Getenv("GIN_MODE") == "release" {
		return "production"
	}
	return "development"
}
