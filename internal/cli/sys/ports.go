package sys

import (
	"fmt"
	"sort"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

type portRow struct {
	Proto   string `json:"proto"`
	Address string `json:"address"`
	Port    uint32 `json:"port"`
	State   string `json:"state"`
	PID     int32  `json:"pid"`
	Process string `json:"process"`
}

func portsCmd() *cobra.Command {
	var listening bool
	c := &cobra.Command{
		Use:   "ports",
		Short: "List listening sockets and the processes that own them",
		RunE: func(cmd *cobra.Command, _ []string) error {
			conns, err := net.Connections("inet")
			if err != nil {
				return fmt.Errorf("read connections: %w", err)
			}
			rows := make([]portRow, 0, len(conns))
			for _, c := range conns {
				if listening && c.Status != "LISTEN" {
					continue
				}
				name := "-"
				if c.Pid > 0 {
					if p, err := process.NewProcess(c.Pid); err == nil {
						if n, err := p.Name(); err == nil {
							name = n
						}
					}
				}
				proto := "tcp"
				if c.Type == 2 {
					proto = "udp"
				}
				rows = append(rows, portRow{
					Proto: proto, Address: c.Laddr.IP, Port: c.Laddr.Port,
					State: c.Status, PID: c.Pid, Process: name,
				})
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].Port < rows[j].Port })

			tbl := ui.Table{
				Headers: []string{"PROTO", "ADDR", "PORT", "STATE", "PID", "PROCESS"},
				Aligns:  []string{"left", "left", "right", "left", "right", "left"},
			}
			for _, r := range rows {
				tbl.Rows = append(tbl.Rows, []string{
					r.Proto,
					r.Address,
					fmt.Sprintf("%d", r.Port),
					r.State,
					fmt.Sprintf("%d", r.PID),
					r.Process,
				})
			}
			pretty := ui.Headline("Open Ports") + "\n" + tbl.Render()
			return ui.Render(rows, pretty)
		},
	}
	c.Flags().BoolVarP(&listening, "listening", "l", false, "show only listening sockets")
	return c
}
