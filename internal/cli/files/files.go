// Package files exposes the `o2s files` command group:
// tree, find, hash, convert and dedup utilities.
package files

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "files",
		Aliases: []string{"fs", "f"},
		Short:   "File & data utilities (tree, find, hash, convert, dedup)",
	}
	c.AddCommand(treeCmd(), findCmd(), hashCmd(), convertCmd(), dedupCmd())
	return c
}
