package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// asciiLogo is the bold "O2S" wordmark used as the brand banner.
// Generated with figlet "ANSI Shadow" font.
const asciiLogo = ` ██████╗ ██████╗ ███████╗
██╔═══██╗╚════██╗██╔════╝
██║   ██║ █████╔╝███████╗
██║   ██║██╔═══╝ ╚════██║
╚██████╔╝███████╗███████║
 ╚═════╝ ╚══════╝╚══════╝`

// Banner returns the multi-coloured brand banner.
// The ASCII art is rendered with a horizontal gradient between
// the active palette's primary and secondary colours.
func Banner() string {
	t := T()
	if !IsTTY() {
		return strings.TrimRight(asciiLogo, "\n")
	}

	lines := strings.Split(strings.Trim(asciiLogo, "\n"), "\n")
	out := make([]string, 0, len(lines)+2)
	for i, ln := range lines {
		// Alternate between primary and secondary so each line has a slightly
		// different vibe — close enough to a gradient at zero CPU cost.
		col := t.P.Primary
		if i%2 == 1 {
			col = t.P.Secondary
		}
		out = append(out, lipgloss.NewStyle().Foreground(col).Bold(true).Render(ln))
	}
	tagline := t.Muted.Render("a beautiful, powerful CLI for devs, sysadmins & humans")
	return strings.Join(out, "\n") + "\n" + tagline
}

// Headline renders a short title with the brand colour and an underline rule.
func Headline(text string) string {
	t := T()
	if !IsTTY() {
		return text
	}
	rule := strings.Repeat("─", lipgloss.Width(text)+2)
	return t.Title.Render(" "+text+" ") + "\n" + t.Muted.Render(rule)
}
