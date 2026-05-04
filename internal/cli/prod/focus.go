package prod

import (
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/store"
	"github.com/O2S-Y/o2s/internal/ui"
)

// focusCmd is a Pomodoro-style full-screen focus session bound to a chosen todo.
func focusCmd() *cobra.Command {
	var (
		dur     time.Duration
		breakD  time.Duration
		rounds  int
	)
	c := &cobra.Command{
		Use:   "focus",
		Short: "Pomodoro focus session bound to a todo",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := store.Open()
			if err != nil {
				return err
			}
			defer s.Close()
			todos := loadAllTodos(s)
			active := todos[:0]
			for _, t := range todos {
				if !t.Done {
					active = append(active, t)
				}
			}
			sort.Slice(active, func(i, j int) bool { return active[i].CreatedAt.After(active[j].CreatedAt) })

			titles := []string{"(no todo — just focus)"}
			for _, t := range active {
				titles = append(titles, t.Title)
			}
			picked, err := ui.Select("Pick what you'll focus on", titles)
			if err != nil {
				return err
			}
			if !ui.IsTTY() {
				ui.Println("focus mode requires a TTY")
				return nil
			}

			m := focusModel{
				task:    picked,
				focus:   dur,
				brk:     breakD,
				rounds:  rounds,
				phase:   "focus",
				end:     time.Now().Add(dur),
				current: 1,
			}
			p := tea.NewProgram(m, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
	c.Flags().DurationVarP(&dur, "duration", "d", 25*time.Minute, "focus block length")
	c.Flags().DurationVarP(&breakD, "break", "b", 5*time.Minute, "break length")
	c.Flags().IntVarP(&rounds, "rounds", "r", 4, "number of focus rounds")
	return c
}

type focusTick time.Time

type focusModel struct {
	task    string
	focus   time.Duration
	brk     time.Duration
	rounds  int
	current int
	phase   string // "focus" | "break" | "done"
	end     time.Time
	width   int
	height  int
}

func (m focusModel) Init() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return focusTick(t) })
}

func (m focusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ", "s":
			if m.phase == "focus" {
				m.phase = "break"
				m.end = time.Now().Add(m.brk)
			} else if m.phase == "break" {
				m.phase = "focus"
				m.end = time.Now().Add(m.focus)
			}
		}
	case focusTick:
		if time.Now().After(m.end) {
			switch m.phase {
			case "focus":
				if m.current >= m.rounds {
					m.phase = "done"
				} else {
					m.phase = "break"
					m.end = time.Now().Add(m.brk)
				}
			case "break":
				m.current++
				m.phase = "focus"
				m.end = time.Now().Add(m.focus)
			}
		}
		if m.phase == "done" {
			return m, tea.Quit
		}
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return focusTick(t) })
	}
	return m, nil
}

func (m focusModel) View() string {
	t := ui.T()
	remaining := time.Until(m.end)
	if remaining < 0 {
		remaining = 0
	}
	dur := m.focus
	if m.phase == "break" {
		dur = m.brk
	}
	pct := 1 - remaining.Seconds()/dur.Seconds()
	header := t.Title.Render("FOCUS")
	if m.phase == "break" {
		header = t.Info.Render("BREAK")
	}
	if m.phase == "done" {
		header = t.Success.Render("✔ session complete")
	}

	body := strings.Join([]string{
		header,
		"",
		t.Subtitle.Render("Task: ") + t.Body.Render(m.task),
		t.Muted.Render(roundLine(m.current, m.rounds)),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(t.P.Primary).Render(formatHMS(remaining)),
		"",
		progressBar(pct, 50),
		"",
		t.Muted.Render("space/s = skip phase   q = quit"),
	}, "\n")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		t.Panel.Render(body))
}

func roundLine(cur, total int) string {
	out := strings.Builder{}
	for i := 1; i <= total; i++ {
		if i <= cur {
			out.WriteString("●")
		} else {
			out.WriteString("○")
		}
		if i < total {
			out.WriteString(" ")
		}
	}
	return out.String()
}
