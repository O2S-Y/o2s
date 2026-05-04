package sys

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

type infoSnapshot struct {
	Hostname     string  `json:"hostname"`
	OS           string  `json:"os"`
	Platform     string  `json:"platform"`
	KernelVer    string  `json:"kernel_version"`
	Uptime       string  `json:"uptime"`
	CPU          string  `json:"cpu"`
	Cores        int     `json:"cores"`
	MemoryUsed   string  `json:"memory_used"`
	MemoryTotal  string  `json:"memory_total"`
	MemoryPct    float64 `json:"memory_pct"`
	DiskUsed     string  `json:"disk_used"`
	DiskTotal    string  `json:"disk_total"`
	DiskPct      float64 `json:"disk_pct"`
	IPv4         string  `json:"ipv4"`
	GoVersion    string  `json:"go_version"`
	Architecture string  `json:"arch"`
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Pretty system summary (OS, CPU, RAM, disk, IP)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			snap, err := collectInfo()
			if err != nil {
				return err
			}
			pretty := ui.Headline("System Info") + "\n" + ui.KV([][2]string{
				{"hostname", snap.Hostname},
				{"os", snap.OS + " " + snap.KernelVer},
				{"platform", snap.Platform},
				{"uptime", snap.Uptime},
				{"cpu", fmt.Sprintf("%s (%d cores)", snap.CPU, snap.Cores)},
				{"memory", fmt.Sprintf("%s / %s (%.0f%%)", snap.MemoryUsed, snap.MemoryTotal, snap.MemoryPct)},
				{"disk", fmt.Sprintf("%s / %s (%.0f%%)", snap.DiskUsed, snap.DiskTotal, snap.DiskPct)},
				{"ipv4", snap.IPv4},
				{"go", snap.GoVersion},
				{"arch", snap.Architecture},
			})
			return ui.Render(snap, pretty)
		},
	}
}

func collectInfo() (infoSnapshot, error) {
	var s infoSnapshot
	hi, err := host.Info()
	if err == nil {
		s.Hostname = hi.Hostname
		s.OS = hi.OS
		s.Platform = hi.Platform + " " + hi.PlatformVersion
		s.KernelVer = hi.KernelVersion
		s.Uptime = humanize.RelTime(time.Now().Add(-time.Duration(hi.Uptime)*time.Second), time.Now(), "ago", "")
	}
	cpus, _ := cpu.Info()
	if len(cpus) > 0 {
		s.CPU = cpus[0].ModelName
	}
	s.Cores = runtime.NumCPU()

	if vm, err := mem.VirtualMemory(); err == nil {
		s.MemoryUsed = humanize.IBytes(vm.Used)
		s.MemoryTotal = humanize.IBytes(vm.Total)
		s.MemoryPct = vm.UsedPercent
	}
	root := "/"
	if runtime.GOOS == "windows" {
		root = "C:\\"
	}
	if d, err := disk.Usage(root); err == nil {
		s.DiskUsed = humanize.IBytes(d.Used)
		s.DiskTotal = humanize.IBytes(d.Total)
		s.DiskPct = d.UsedPercent
	}
	s.IPv4 = primaryIPv4()
	s.GoVersion = runtime.Version()
	s.Architecture = runtime.GOOS + "/" + runtime.GOARCH
	return s, nil
}

func primaryIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "-"
	}
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 || ifc.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if v4 := ipn.IP.To4(); v4 != nil {
					return v4.String()
				}
			}
		}
	}
	return "-"
}
