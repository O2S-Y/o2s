package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Table renders a simple aligned table without external dependencies.
// It zebra-stripes rows and styles the header with the brand palette.
type Table struct {
	Headers []string
	Rows    [][]string
	// Aligns is one of "left" | "right" per column. Empty means left.
	Aligns []string
}

func (t Table) Render() string {
	if len(t.Rows) == 0 && len(t.Headers) == 0 {
		return ""
	}
	cols := len(t.Headers)
	for _, r := range t.Rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	widths := make([]int, cols)
	for i, h := range t.Headers {
		if w := lipgloss.Width(h); w > widths[i] {
			widths[i] = w
		}
	}
	for _, r := range t.Rows {
		for i, c := range r {
			if w := lipgloss.Width(c); w > widths[i] {
				widths[i] = w
			}
		}
	}

	pad := func(s string, w int, i int) string {
		align := "left"
		if i < len(t.Aligns) {
			align = t.Aligns[i]
		}
		gap := w - lipgloss.Width(s)
		if gap <= 0 {
			return s
		}
		if align == "right" {
			return strings.Repeat(" ", gap) + s
		}
		return s + strings.Repeat(" ", gap)
	}

	th := T()
	var b strings.Builder
	if len(t.Headers) > 0 {
		parts := make([]string, cols)
		for i := 0; i < cols; i++ {
			h := ""
			if i < len(t.Headers) {
				h = t.Headers[i]
			}
			parts[i] = th.TableHeader.Render(pad(h, widths[i], i))
		}
		b.WriteString(strings.Join(parts, "  "))
		b.WriteString("\n")
	}
	for ri, r := range t.Rows {
		parts := make([]string, cols)
		style := th.TableRow
		if ri%2 == 1 {
			style = th.TableAlt
		}
		for i := 0; i < cols; i++ {
			c := ""
			if i < len(r) {
				c = r[i]
			}
			parts[i] = style.Render(pad(c, widths[i], i))
		}
		b.WriteString(strings.Join(parts, "  "))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
