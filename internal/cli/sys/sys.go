// Package sys exposes the `o2s sys` command group:
// info, top, ports, ping, dns — system & network utilities.
package sys

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "sys",
		Aliases: []string{"system"},
		Short:   "System & network utilities (info, top, ports, ping, dns)",
	}
	c.AddCommand(infoCmd(), topCmd(), portsCmd(), pingCmd(), dnsCmd())
	return c
}
