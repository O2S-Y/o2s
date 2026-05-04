// Package prod hosts the personal-productivity command group:
// todos, notes, timers and a Bubble Tea focus mode.
package prod

import (
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/cli/prod/clip"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "prod",
		Aliases: []string{"productivity", "p"},
		Short:   "Personal productivity (todos, notes, timer, focus, snippets, clipboard)",
	}
	c.AddCommand(todoCmd(), noteCmd(), timerCmd(), focusCmd(), clip.NewCmd(), snippetCmd())
	return c
}
