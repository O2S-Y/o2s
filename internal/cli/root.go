// Package cli wires the root cobra command and registers every subcommand group.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/O2S-Y/o2s/internal/cli/dev"
	"github.com/O2S-Y/o2s/internal/cli/files"
	"github.com/O2S-Y/o2s/internal/cli/plugin"
	"github.com/O2S-Y/o2s/internal/cli/prod"
	"github.com/O2S-Y/o2s/internal/cli/sys"
	"github.com/O2S-Y/o2s/internal/config"
	olog "github.com/O2S-Y/o2s/internal/log"
	"github.com/O2S-Y/o2s/internal/tui/dashboard"
	"github.com/O2S-Y/o2s/internal/ui"
	"github.com/O2S-Y/o2s/internal/version"
)

// Global flags exposed at the root level.
type globalFlags struct {
	theme   string
	noColor bool
	json    bool
	verbose bool
}

// NewRootCmd builds the root cobra command tree.
func NewRootCmd() *cobra.Command {
	g := &globalFlags{}

	root := &cobra.Command{
		Use:           "o2s",
		Short:         "O2S - a beautiful, powerful Swiss-army-knife CLI",
		Long:          "O2S CLI bundles dev, sysadmin, file, productivity tools and drop-in plugins into one polished command-line tool.",
		Version:       version.Full(),
		SilenceUsage:  true,
		SilenceErrors: false,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			if g.theme == "" {
				g.theme = cfg.Theme
			}
			if !g.noColor {
				g.noColor = cfg.NoColor
			}
			ui.SetTheme(g.theme)
			ui.SetNoColor(g.noColor)
			ui.SetJSON(g.json)
			olog.SetLevel(g.verbose)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if !cfg.Onboard {
				if err := runOnboarding(&cfg); err != nil {
					return err
				}
				if err := config.Save(cfg); err != nil {
					return err
				}
				ui.SetTheme(cfg.Theme)
			}
			return dashboard.Run(cfg)
		},
	}

	root.PersistentFlags().StringVar(&g.theme, "theme", "", "colour theme: "+joinThemes())
	root.PersistentFlags().BoolVar(&g.noColor, "no-color", false, "disable ANSI colours and styling")
	root.PersistentFlags().BoolVar(&g.json, "json", false, "emit machine-readable JSON where supported")
	root.PersistentFlags().BoolVarP(&g.verbose, "verbose", "v", false, "verbose / debug logging")

	root.SetUsageTemplate(usageTemplate)
	root.SetVersionTemplate("o2s {{.Version}}\n")

	root.AddCommand(versionCmd())
	root.AddCommand(configCmd())
	root.AddCommand(dev.NewCmd())
	root.AddCommand(sys.NewCmd())
	root.AddCommand(files.NewCmd())
	root.AddCommand(prod.NewCmd())
	root.AddCommand(plugin.NewCmd())

	return root
}

func joinThemes() string {
	out := ""
	for i, t := range ui.Themes() {
		if i > 0 {
			out += " | "
		}
		out += t
	}
	return out
}
