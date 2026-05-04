package sys

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func topCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "top",
		Short: "Live process viewer (Bubble Tea)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !ui.IsTTY() {
				return fmt.Errorf("`o2s sys top` requires an interactive terminal")
			}
			p := tea.NewProgram(initialTopModel(), tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}
}

type procRow struct {
	PID  int32
	Name string
	CPU  float64
	Mem  float64
	RSS  uint64
}

type tickMsg time.Time

type topModel struct {
	rows     []procRow
	cpuPct   float64
	memUsed  uint64
	memTotal uint64
	memPct   float64
	width    int
	height   int
	limit    int
	sortBy   string // "cpu" | "mem" | "name"
	last     time.Time
}

func initialTopModel() topModel {
	return topModel{limit: 20, sortBy: "cpu", last: time.Now()}
}

func (m topModel) Init() tea.Cmd {
	return tea.Batch(refreshCmd(), tickEvery(time.Second))
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func refreshCmd() tea.Cmd {
	return func() tea.Msg { return collectTop() }
}

type topSnapshot struct {
	procs    []procRow
	cpuPct   float64
	memUsed  uint64
	memTotal uint64
	memPct   float64
}

func collectTop() topSnapshot {
	snap := topSnapshot{}
	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		snap.cpuPct = pcts[0]
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		snap.memUsed = vm.Used
		snap.memTotal = vm.Total
		snap.memPct = vm.UsedPercent
	}
	procs, err := process.Processes()
	if err == nil {
		rows := make([]procRow, 0, len(procs))
		for _, p := range procs {
			name, _ := p.Name()
			cpu, _ := p.CPUPercent()
			memPct, _ := p.MemoryPercent()
			memInfo, _ := p.MemoryInfo()
			var rss uint64
			if memInfo != nil {
				rss = memInfo.RSS
			}
			rows = append(rows, procRow{
				PID: p.Pid, Name: name, CPU: cpu, Mem: float64(memPct), RSS: rss,
			})
		}
		snap.procs = rows
	}
	return snap
}

func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "c":
			m.sortBy = "cpu"
		case "m":
			m.sortBy = "mem"
		case "n":
			m.sortBy = "name"
		case "r":
			return m, refreshCmd()
		}
	case tickMsg:
		m.last = time.Time(msg)
		return m, tea.Batch(refreshCmd(), tickEvery(time.Second*2))
	case topSnapshot:
		m.rows = msg.procs
		m.cpuPct = msg.cpuPct
		m.memUsed = msg.memUsed
		m.memTotal = msg.memTotal
		m.memPct = msg.memPct
		m.sortRows()
	}
	return m, nil
}

func (m *topModel) sortRows() {
	switch m.sortBy {
	case "cpu":
		sort.Slice(m.rows, func(i, j int) bool { return m.rows[i].CPU > m.rows[j].CPU })
	case "mem":
		sort.Slice(m.rows, func(i, j int) bool { return m.rows[i].Mem > m.rows[j].Mem })
	case "name":
		sort.Slice(m.rows, func(i, j int) bool { return strings.ToLower(m.rows[i].Name) < strings.ToLower(m.rows[j].Name) })
	}
}

func (m topModel) View() string {
	t := ui.T()
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		t.Title.Render("o2s top"),
		"  ",
		t.Muted.Render(fmt.Sprintf("(%d processes — sort by %s — c/m/n/r/q)", len(m.rows), m.sortBy)),
	)
	stats := ui.KV([][2]string{
		{"cpu", fmt.Sprintf("%.1f%%", m.cpuPct)},
		{"mem", fmt.Sprintf("%s / %s (%.1f%%)", humanize.IBytes(m.memUsed), humanize.IBytes(m.memTotal), m.memPct)},
		{"updated", m.last.Format("15:04:05")},
	})

	limit := m.limit
	if m.height > 12 {
		limit = m.height - 12
	}
	if limit > len(m.rows) {
		limit = len(m.rows)
	}
	tbl := ui.Table{
		Headers: []string{"PID", "NAME", "CPU%", "MEM%", "RSS"},
		Aligns:  []string{"right", "left", "right", "right", "right"},
		Rows:    make([][]string, 0, limit),
	}
	for i := 0; i < limit; i++ {
		r := m.rows[i]
		tbl.Rows = append(tbl.Rows, []string{
			fmt.Sprintf("%d", r.PID),
			truncate(r.Name, 24),
			fmt.Sprintf("%.1f", r.CPU),
			fmt.Sprintf("%.1f", r.Mem),
			humanize.IBytes(r.RSS),
		})
	}
	footer := t.Muted.Render("press q to quit")
	return strings.Join([]string{header, "", stats, "", tbl.Render(), "", footer}, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
