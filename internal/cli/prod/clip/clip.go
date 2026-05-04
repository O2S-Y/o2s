// Package clip hosts clipboard helpers (get / set).
package clip

import (
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/ui"
)

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "clip",
		Aliases: []string{"clipboard"},
		Short:   "Clipboard utilities (get / set)",
	}
	c.AddCommand(clipGetCmd(), clipSetCmd())
	return c
}

func clipGetCmd() *cobra.Command {
	var trim bool
	c := &cobra.Command{
		Use:     "get",
		Aliases: []string{"read"},
		Short:   "Print clipboard contents to stdout",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := clipboard.ReadAll()
			if err != nil {
				return err
			}
			if trim {
				s = strings.TrimSpace(s)
			}
			return ui.Render(map[string]any{"text": s}, s)
		},
	}
	c.Flags().BoolVar(&trim, "trim", false, "trim surrounding whitespace")
	return c
}

func clipSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set [text...]",
		Aliases: []string{"write"},
		Short:   "Set clipboard from args (joined with spaces)",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if err := clipboard.WriteAll(text); err != nil {
				return err
			}
			type out struct {
				OK   bool   `json:"ok"`
				Len  int    `json:"len"`
				Text string `json:"text,omitempty"`
			}
			pretty := ui.T().Success.Render("✔ ") + "clipboard updated (" + ui.T().Muted.Render(strconv.Itoa(len(text))) + " chars)"
			return ui.Render(out{OK: true, Len: len(text)}, pretty)
		},
	}
}
