package sys

import (
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

// We use TCP "ping" rather than ICMP because raw sockets require admin
// privileges on Windows and most Linux installs. Connecting to port 80 (or a
// user-supplied port) gives a useful "is this host reachable + how fast?".
func pingCmd() *cobra.Command {
	var (
		count   int
		port    int
		timeout time.Duration
	)
	c := &cobra.Command{
		Use:   "ping <host>",
		Short: "TCP ping — measure reachability + latency without raw sockets",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]
			addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
			t := ui.T()
			ui.Println(ui.Headline("Pinging " + addr))
			results := make([]float64, 0, count)
			for i := 0; i < count; i++ {
				start := time.Now()
				conn, err := net.DialTimeout("tcp", addr, timeout)
				elapsed := time.Since(start)
				if err != nil {
					ui.Println(t.Danger.Render("✘ ") + fmt.Sprintf("seq=%d  error: %v", i+1, err))
					continue
				}
				_ = conn.Close()
				ms := float64(elapsed.Microseconds()) / 1000.0
				results = append(results, ms)
				ui.Println(t.Success.Render("✔ ") + fmt.Sprintf("seq=%d  time=%.2f ms", i+1, ms))
				time.Sleep(300 * time.Millisecond)
			}
			if len(results) == 0 {
				return fmt.Errorf("no successful pings")
			}
			min, max, avg := results[0], results[0], 0.0
			for _, v := range results {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
				avg += v
			}
			avg /= float64(len(results))
			summary := ui.KV([][2]string{
				{"sent", fmt.Sprintf("%d", count)},
				{"received", fmt.Sprintf("%d", len(results))},
				{"loss", fmt.Sprintf("%.0f%%", 100*float64(count-len(results))/float64(count))},
				{"min/avg/max", fmt.Sprintf("%.2f / %.2f / %.2f ms", min, avg, max)},
			})
			ui.Println("\n" + summary)
			return nil
		},
	}
	c.Flags().IntVarP(&count, "count", "c", 4, "number of probes to send")
	c.Flags().IntVarP(&port, "port", "p", 80, "TCP port to probe")
	c.Flags().DurationVarP(&timeout, "timeout", "t", 2*time.Second, "per-probe timeout")
	return c
}
