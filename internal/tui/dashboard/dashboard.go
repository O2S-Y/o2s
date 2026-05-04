// Package dashboard renders the home screen the user sees when running `o2s`
// with no arguments. It's an at-a-glance Bubble Tea view that mixes
// system stats, recent todos and the brand banner.
package dashboard

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/O2S-Y/o2s/internal/config"
	"github.com/O2S-Y/o2s/internal/store"
	"github.com/O2S-Y/o2s/internal/ui"
)

// Run starts the interactive dashboard.
// Falls back to a printed banner + summary when stdout is not a TTY.
func Run(cfg config.Config) error {
	if !ui.IsTTY() {
		return printStatic(cfg)
	}
	m := newModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func printStatic(cfg config.Config) error {
	greeting := "hello"
	if cfg.Name != "" {
		greeting = "hello, " + cfg.Name
	}
	ui.Println(ui.Banner())
	ui.Println("")
	ui.Println(ui.Headline(greeting))
	ui.Println("")
	ui.Println(ui.KV([][2]string{
		{"theme", cfg.Theme},
		{"runtime", runtime.Version() + " on " + runtime.GOOS + "/" + runtime.GOARCH},
		{"hint", "run `o2s --help` to see what's available"},
	}))
	return nil
}

type tickMsg time.Time

type model struct {
	cfg     config.Config
	width   int
	height  int
	cpuPct  float64
	memPct  float64
	memUsed uint64
	memTot  uint64
	uptime  string
	todos   []todoItem
	clock   time.Time
}

type todoItem struct {
	Title    string
	Done     bool
	Priority string
}

func newModel(cfg config.Config) model {
	return model{cfg: cfg, clock: time.Now()}
}

func (m model) Init() tea.Cmd { return tea.Batch(refreshCmd(), tickEvery(time.Second)) }

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type snapshot struct {
	cpuPct  float64
	memPct  float64
	memUsed uint64
	memTot  uint64
	uptime  string
	todos   []todoItem
}

func refreshCmd() tea.Cmd {
	return func() tea.Msg {
		s := snapshot{}
		if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
			s.cpuPct = pcts[0]
		}
		if vm, err := mem.VirtualMemory(); err == nil {
			s.memPct = vm.UsedPercent
			s.memUsed = vm.Used
			s.memTot = vm.Total
		}
		if hi, err := host.Info(); err == nil {
			s.uptime = humanize.RelTime(time.Now().Add(-time.Duration(hi.Uptime)*time.Second), time.Now(), "ago", "")
		}
		st, err := store.Open()
		if err == nil {
			defer st.Close()
			_ = st.Iter(store.BucketTodos, func(_ string, raw []byte) error {
				var t struct {
					Title     string    `json:"title"`
					Done      bool      `json:"done"`
					Priority  string    `json:"priority"`
					CreatedAt time.Time `json:"created_at"`
				}
				if err := jsonUnmarshal(raw, &t); err == nil {
					s.todos = append(s.todos, todoItem{Title: t.Title, Done: t.Done, Priority: t.Priority})
				}
				return nil
			})
			sort.Slice(s.todos, func(i, j int) bool {
				if s.todos[i].Done != s.todos[j].Done {
					return !s.todos[i].Done
				}
				return false
			})
		}
		return s
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, refreshCmd()
		}
	case tickMsg:
		m.clock = time.Time(msg)
		return m, tea.Batch(refreshCmd(), tickEvery(2*time.Second))
	case snapshot:
		m.cpuPct = msg.cpuPct
		m.memPct = msg.memPct
		m.memUsed = msg.memUsed
		m.memTot = msg.memTot
		m.uptime = msg.uptime
		m.todos = msg.todos
	}
	return m, nil
}

func (m model) View() string {
	t := ui.T()
	greet := "Welcome back"
	if m.cfg.Name != "" {
		greet = "Welcome back, " + m.cfg.Name
	}

	left := strings.Join([]string{
		ui.Banner(),
		"",
		t.Subtitle.Render(greet),
		t.Muted.Render(m.clock.Format("Mon 2 Jan · 15:04:05")),
		"",
		t.Heading.Render("System"),
		ui.KV([][2]string{
			{"cpu", fmt.Sprintf("%.0f%%", m.cpuPct)},
			{"mem", fmt.Sprintf("%s / %s (%.0f%%)", humanize.IBytes(m.memUsed), humanize.IBytes(m.memTot), m.memPct)},
			{"uptime", m.uptime},
			{"theme", m.cfg.Theme},
		}),
	}, "\n")

	right := strings.Join([]string{
		t.Heading.Render("Todos"),
		renderTodos(m.todos),
		"",
		t.Heading.Render("Try"),
		t.Muted.Render("• o2s sys top"),
		t.Muted.Render("• o2s files tree"),
		t.Muted.Render("• o2s prod focus"),
		t.Muted.Render("• o2s dev init"),
	}, "\n")

	leftBox := t.Panel.Render(left)
	rightBox := t.Panel.Render(right)
	cols := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox)
	footer := t.Muted.Render("r refresh   q quit")
	return cols + "\n" + footer
}

func renderTodos(items []todoItem) string {
	t := ui.T()
	if len(items) == 0 {
		return t.Muted.Render("nothing yet — add with `o2s prod todo add`")
	}
	max := 6
	if len(items) < max {
		max = len(items)
	}
	out := make([]string, 0, max)
	for i := 0; i < max; i++ {
		mark := t.Warning.Render("○")
		if items[i].Done {
			mark = t.Success.Render("●")
		}
		out = append(out, mark+" "+items[i].Title)
	}
	if len(items) > max {
		out = append(out, t.Muted.Render(fmt.Sprintf("…and %d more", len(items)-max)))
	}
	return strings.Join(out, "\n")
}
