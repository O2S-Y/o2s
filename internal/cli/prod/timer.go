package prod

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func timerCmd() *cobra.Command {
	var label string
	c := &cobra.Command{
		Use:   "timer <duration>",
		Short: "Countdown timer (e.g. `o2s prod timer 25m --label pomodoro`)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dur, err := time.ParseDuration(args[0])
			if err != nil {
				return fmt.Errorf("bad duration %q: %w", args[0], err)
			}
			if !ui.IsTTY() {
				ui.Println(fmt.Sprintf("waiting %s...", dur))
				time.Sleep(dur)
				ui.Success("done")
				return nil
			}
			model := newTimerModel(dur, label)
			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
	c.Flags().StringVarP(&label, "label", "l", "timer", "label shown above the countdown")
	return c
}

type timerTick time.Time

type timerModel struct {
	end     time.Time
	dur     time.Duration
	label   string
	width   int
	height  int
	done    bool
	startAt time.Time
}

func newTimerModel(dur time.Duration, label string) timerModel {
	now := time.Now()
	return timerModel{end: now.Add(dur), dur: dur, label: label, startAt: now}
}

func (m timerModel) Init() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg { return timerTick(t) })
}

func (m timerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if k := msg.String(); k == "q" || k == "ctrl+c" || k == "esc" {
			return m, tea.Quit
		}
	case timerTick:
		if time.Now().After(m.end) {
			m.done = true
			return m, tea.Quit
		}
		return m, tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg { return timerTick(t) })
	}
	return m, nil
}

func (m timerModel) View() string {
	t := ui.T()
	remaining := time.Until(m.end)
	if remaining < 0 {
		remaining = 0
	}
	pct := 1 - remaining.Seconds()/m.dur.Seconds()
	bar := progressBar(pct, 40)
	timeStr := lipgloss.NewStyle().Foreground(t.P.Primary).Bold(true).Render(formatHMS(remaining))
	body := strings.Join([]string{
		t.Subtitle.Render(m.label),
		"",
		timeStr,
		"",
		bar,
		"",
		t.Muted.Render("press q to stop"),
	}, "\n")
	if m.done {
		body = strings.Join([]string{
			t.Success.Render("✔ time's up — " + m.label),
			"",
			t.Muted.Render("(window will close in a moment)"),
		}, "\n")
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		t.Panel.Render(body))
}

func progressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(float64(width) * pct)
	t := ui.T()
	left := lipgloss.NewStyle().Foreground(t.P.Primary).Render(strings.Repeat("█", filled))
	right := lipgloss.NewStyle().Foreground(t.P.Muted).Render(strings.Repeat("░", width-filled))
	return left + right
}

func formatHMS(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}
