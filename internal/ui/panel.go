package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Panel renders content inside a rounded brand-coloured border with a title.
func Panel(title, content string) string {
	t := T()
	if !IsTTY() {
		return fmt.Sprintf("[ %s ]\n%s", title, content)
	}
	titleBar := t.PanelTitle.Render(title)
	body := lipgloss.NewStyle().Padding(0, 1).Render(content)
	return t.Panel.Render(titleBar + "\n" + body)
}

// KV renders an aligned key/value list (used by `sys info`, dashboards, etc).
func KV(pairs [][2]string) string {
	t := T()
	maxK := 0
	for _, p := range pairs {
		if w := lipgloss.Width(p[0]); w > maxK {
			maxK = w
		}
	}
	var b strings.Builder
	for _, p := range pairs {
		key := t.Key.Render(p[0]) + strings.Repeat(" ", maxK-lipgloss.Width(p[0])+2)
		b.WriteString(key)
		b.WriteString(t.Value.Render(p[1]))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Status produces a coloured "ok", "warn" or "err" badge.
func Status(kind, label string) string {
	t := T()
	switch kind {
	case "ok", "success":
		return t.Success.Render("● ") + label
	case "warn", "warning":
		return t.Warning.Render("● ") + label
	case "err", "error", "danger":
		return t.Danger.Render("● ") + label
	default:
		return t.Muted.Render("● ") + label
	}
}

// Columns lays out two strings side-by-side, each in its own panel-shaped block.
func Columns(left, right string, leftWidth int) string {
	if !IsTTY() {
		return left + "\n\n" + right
	}
	l := lipgloss.NewStyle().Width(leftWidth).Render(left)
	r := lipgloss.NewStyle().Render(right)
	return lipgloss.JoinHorizontal(lipgloss.Top, l, "  ", r)
}
