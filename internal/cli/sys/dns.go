package sys

import (
	"net"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

type dnsResult struct {
	Domain string   `json:"domain"`
	A      []string `json:"a"`
	AAAA   []string `json:"aaaa"`
	CNAME  string   `json:"cname,omitempty"`
	MX     []string `json:"mx,omitempty"`
	NS     []string `json:"ns,omitempty"`
	TXT    []string `json:"txt,omitempty"`
}

func dnsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dns <domain>",
		Short: "Resolve A / AAAA / CNAME / MX / NS / TXT records",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := args[0]
			out := dnsResult{Domain: d}

			if ips, err := net.LookupIP(d); err == nil {
				for _, ip := range ips {
					if v4 := ip.To4(); v4 != nil {
						out.A = append(out.A, v4.String())
					} else {
						out.AAAA = append(out.AAAA, ip.String())
					}
				}
			}
			if cname, err := net.LookupCNAME(d); err == nil {
				out.CNAME = cname
			}
			if mxs, err := net.LookupMX(d); err == nil {
				for _, m := range mxs {
					out.MX = append(out.MX, m.Host)
				}
			}
			if nss, err := net.LookupNS(d); err == nil {
				for _, n := range nss {
					out.NS = append(out.NS, n.Host)
				}
			}
			if txts, err := net.LookupTXT(d); err == nil {
				out.TXT = txts
			}

			pretty := ui.Headline("DNS — "+d) + "\n" + ui.KV([][2]string{
				{"A", joinDash(out.A)},
				{"AAAA", joinDash(out.AAAA)},
				{"CNAME", orDashStr(out.CNAME)},
				{"MX", joinDash(out.MX)},
				{"NS", joinDash(out.NS)},
				{"TXT", joinDash(out.TXT)},
			})
			return ui.Render(out, pretty)
		},
	}
}

func joinDash(xs []string) string {
	if len(xs) == 0 {
		return "-"
	}
	out := xs[0]
	for i := 1; i < len(xs); i++ {
		out += ", " + xs[i]
	}
	return out
}
func orDashStr(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
