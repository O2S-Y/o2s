// Package ui provides the shared visual language for O2S CLI:
// colour palettes, named themes, and reusable Lipgloss styles
// every command and TUI screen pulls from.
package ui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// Palette is a named colour set. We keep it intentionally small so
// every theme stays cohesive across all commands.
type Palette struct {
	Name      string
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Accent    lipgloss.AdaptiveColor
	Success   lipgloss.AdaptiveColor
	Warning   lipgloss.AdaptiveColor
	Danger    lipgloss.AdaptiveColor
	Muted     lipgloss.AdaptiveColor
	Text      lipgloss.AdaptiveColor
	Bg        lipgloss.AdaptiveColor
}

var palettes = map[string]Palette{
	"default": {
		Name:      "default",
		Primary:   lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"},
		Secondary: lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#38BDF8"},
		Accent:    lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"},
		Success:   lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"},
		Warning:   lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"},
		Danger:    lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"},
		Muted:     lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"},
		Text:      lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"},
		Bg:        lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0B0F19"},
	},
	"neon": {
		Name:      "neon",
		Primary:   lipgloss.AdaptiveColor{Light: "#FF006E", Dark: "#FF4FA3"},
		Secondary: lipgloss.AdaptiveColor{Light: "#3A86FF", Dark: "#5FA8FF"},
		Accent:    lipgloss.AdaptiveColor{Light: "#FFBE0B", Dark: "#FFD23F"},
		Success:   lipgloss.AdaptiveColor{Light: "#06FFA5", Dark: "#39FFB0"},
		Warning:   lipgloss.AdaptiveColor{Light: "#FB5607", Dark: "#FF7A3D"},
		Danger:    lipgloss.AdaptiveColor{Light: "#FF006E", Dark: "#FF3D7F"},
		Muted:     lipgloss.AdaptiveColor{Light: "#7B7B92", Dark: "#9B9BB1"},
		Text:      lipgloss.AdaptiveColor{Light: "#0F0F23", Dark: "#F0F0FF"},
		Bg:        lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0F0F23"},
	},
	"mono": {
		Name:      "mono",
		Primary:   lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"},
		Secondary: lipgloss.AdaptiveColor{Light: "#374151", Dark: "#D1D5DB"},
		Accent:    lipgloss.AdaptiveColor{Light: "#4B5563", Dark: "#9CA3AF"},
		Success:   lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"},
		Warning:   lipgloss.AdaptiveColor{Light: "#374151", Dark: "#E5E7EB"},
		Danger:    lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"},
		Muted:     lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"},
		Text:      lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"},
		Bg:        lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"},
	},
	"dracula": {
		Name:      "dracula",
		Primary:   lipgloss.AdaptiveColor{Light: "#BD93F9", Dark: "#BD93F9"},
		Secondary: lipgloss.AdaptiveColor{Light: "#8BE9FD", Dark: "#8BE9FD"},
		Accent:    lipgloss.AdaptiveColor{Light: "#FFB86C", Dark: "#FFB86C"},
		Success:   lipgloss.AdaptiveColor{Light: "#50FA7B", Dark: "#50FA7B"},
		Warning:   lipgloss.AdaptiveColor{Light: "#F1FA8C", Dark: "#F1FA8C"},
		Danger:    lipgloss.AdaptiveColor{Light: "#FF5555", Dark: "#FF5555"},
		Muted:     lipgloss.AdaptiveColor{Light: "#6272A4", Dark: "#6272A4"},
		Text:      lipgloss.AdaptiveColor{Light: "#F8F8F2", Dark: "#F8F8F2"},
		Bg:        lipgloss.AdaptiveColor{Light: "#282A36", Dark: "#282A36"},
	},
}

// Theme exposes the resolved palette + ready-to-use styles.
type Theme struct {
	P Palette

	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Heading     lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Danger      lipgloss.Style
	Info        lipgloss.Style
	Key         lipgloss.Style
	Value       lipgloss.Style
	Badge       lipgloss.Style
	Pill        lipgloss.Style
	Panel       lipgloss.Style
	PanelTitle  lipgloss.Style
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableAlt    lipgloss.Style
	Selected    lipgloss.Style
}

func newTheme(p Palette) *Theme {
	t := &Theme{P: p}
	t.Title = lipgloss.NewStyle().Foreground(p.Primary).Bold(true)
	t.Subtitle = lipgloss.NewStyle().Foreground(p.Secondary).Italic(true)
	t.Heading = lipgloss.NewStyle().Foreground(p.Primary).Bold(true).Underline(true)
	t.Body = lipgloss.NewStyle().Foreground(p.Text)
	t.Muted = lipgloss.NewStyle().Foreground(p.Muted)
	t.Success = lipgloss.NewStyle().Foreground(p.Success).Bold(true)
	t.Warning = lipgloss.NewStyle().Foreground(p.Warning).Bold(true)
	t.Danger = lipgloss.NewStyle().Foreground(p.Danger).Bold(true)
	t.Info = lipgloss.NewStyle().Foreground(p.Secondary)
	t.Key = lipgloss.NewStyle().Foreground(p.Accent).Bold(true)
	t.Value = lipgloss.NewStyle().Foreground(p.Text)
	t.Badge = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(p.Primary).
		Padding(0, 1).
		Bold(true)
	t.Pill = lipgloss.NewStyle().
		Foreground(p.Primary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Primary).
		Padding(0, 1)
	t.Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Primary).
		Padding(1, 2)
	t.PanelTitle = lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		Padding(0, 1)
	t.TableHeader = lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(p.Muted).
		Padding(0, 1)
	t.TableRow = lipgloss.NewStyle().Foreground(p.Text).Padding(0, 1)
	t.TableAlt = lipgloss.NewStyle().Foreground(p.Text).Padding(0, 1).Faint(true)
	t.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(p.Primary).
		Bold(true)
	return t
}

// Defaults the active theme & TTY detection.
var (
	active     = newTheme(palettes["default"])
	stdoutTTY  = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	noColorEnv = os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("O2S_NO_COLOR"), "true")
)

// SetTheme switches the active palette by name. Unknown names fall back to default.
func SetTheme(name string) {
	if p, ok := palettes[strings.ToLower(name)]; ok {
		active = newTheme(p)
		return
	}
	active = newTheme(palettes["default"])
}

// SetNoColor disables colour output globally (useful for pipes / CI).
// We toggle NO_COLOR which both lipgloss and termenv respect, instead of
// reaching into a specific lipgloss API surface that has shifted across
// versions. Setting this before any styled render is what matters.
func SetNoColor(v bool) {
	if v {
		_ = os.Setenv("NO_COLOR", "1")
	}
}

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool { return stdoutTTY && !noColorEnv }

// T returns the active theme.
func T() *Theme { return active }

// Themes returns the list of available theme names (sorted-ish for help text).
func Themes() []string {
	return []string{"default", "neon", "mono", "dracula"}
}
